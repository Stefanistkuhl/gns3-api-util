package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/web/auth"
)

type ctxKey string

const ClaimsKey ctxKey = "user_claims"

func AuthMiddleware(mgr *auth.IdentityManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawHeader := r.Header.Get("Authorization")

			if rawHeader == "" || !strings.HasPrefix(rawHeader, "Bearer ") {
				http.Error(w, "Unauthorized: Missing or malformed token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(rawHeader, "Bearer ")

			claims, err := mgr.Validate(tokenStr)
			if err != nil {
				http.Error(w, "Unauthorized: Invalid token signature or expiry", http.StatusUnauthorized)
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

func RequireScope(store *state.StateManager, requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			if !ok {
				http.Error(w, "Unauthorized: No claims found", http.StatusUnauthorized)
				return
			}

			hasInToken := slices.Contains(claims.Scopes, requiredScope)

			if !hasInToken {
				http.Error(w, "Forbidden: Scope not in token", http.StatusForbidden)
				return
			}

			hasPermission, err := store.CheckPermission(r.Context(), claims.UserID, requiredScope)
			if err != nil {
				http.Error(w, "Internal: Failed to verify permissions", http.StatusInternalServerError)
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
