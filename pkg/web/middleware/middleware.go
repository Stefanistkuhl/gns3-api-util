package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/scopes"
)

type ctxKey string

const ClaimsKey ctxKey = "user_claims"

type PermissionChecker interface {
	HasEffectivePermission(
		ctx context.Context,
		userID string,
		requiredScope string,
	) (bool, error)
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

func AuthMiddleware(mgr *auth.IdentityManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawHeader := r.Header.Get("Authorization")
			if rawHeader == "" || !strings.HasPrefix(rawHeader, "Bearer ") {
				http.Error(
					w,
					"Unauthorized: Missing or malformed token",
					http.StatusUnauthorized,
				)
				return
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(rawHeader, "Bearer "))
			claims, err := mgr.Validate(tokenStr)
			if err != nil {
				http.Error(
					w,
					"Unauthorized: Invalid token signature or expiry",
					http.StatusUnauthorized,
				)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(r *http.Request) (*auth.Claims, bool) {
	claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
	return claims, ok
}

func RequireScope(
	checker PermissionChecker,
	requiredScope string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			if !ok {
				http.Error(w, "Unauthorized: No claims found", http.StatusUnauthorized)
				return
			}

			revoked, err := checker.IsTokenRevoked(r.Context(), claims.ID)
			if err != nil {
				http.Error(
					w,
					"Internal: Failed to verify token revocation",
					http.StatusInternalServerError,
				)
				return
			}
			if revoked {
				http.Error(w, "Unauthorized: Token revoked", http.StatusUnauthorized)
				return
			}

			isSuperInToken := slices.Contains(claims.Scopes, scopes.Superuser)
			hasScopeInToken := isSuperInToken ||
				slices.Contains(claims.Scopes, requiredScope)

			if !hasScopeInToken {
				http.Error(w, "Forbidden: Scope not in token", http.StatusForbidden)
				return
			}

			hasPermission, err := checker.HasEffectivePermission(
				r.Context(),
				claims.UserID,
				requiredScope,
			)
			if err != nil {
				http.Error(
					w,
					"Internal: Failed to verify permissions",
					http.StatusInternalServerError,
				)
				return
			}

			if !hasPermission {
				http.Error(w, "Forbidden: Permission revoked", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
