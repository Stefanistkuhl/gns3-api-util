package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	filerpc "github.com/0xveya/gns3util/internal/file-store/rpc"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/scopes"
)

type ctxKey string

const ClaimsKey ctxKey = "user_claims"

type PermissionChecker interface {
	HasEffectivePermission(ctx context.Context, userID string, requiredScope string) (bool, error)
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

func AuthMiddleware(mgr *auth.IdentityManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawHeader := r.Header.Get("Authorization")
			if rawHeader == "" || !strings.HasPrefix(rawHeader, "Bearer ") {
				helpers.WriteAPIError(w, "Unauthorized: Missing or malformed token", helpers.ErrCodeUnauthorized, "", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(rawHeader, "Bearer "))
			claims, err := mgr.Validate(tokenStr)
			if err != nil {
				helpers.WriteAPIError(w, "Unauthorized: Invalid token", helpers.ErrCodeUnauthorized, err.Error(), http.StatusUnauthorized)
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

func RequireScope(checker PermissionChecker, requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			if !ok {
				helpers.WriteAPIError(w, "Unauthorized: No claims found", helpers.ErrCodeUnauthorized, "", http.StatusUnauthorized)
				return
			}

			revoked, err := checker.IsTokenRevoked(r.Context(), claims.ID)
			if err != nil {
				helpers.WriteAPIError(w, "Internal: Failed to verify token revocation", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
				return
			}
			if revoked {
				helpers.WriteAPIError(w, "Unauthorized: Token revoked", helpers.ErrCodeUnauthorized, "", http.StatusUnauthorized)
				return
			}

			isSuperInToken := slices.Contains(claims.Scopes, scopes.Superuser)
			hasScopeInToken := isSuperInToken || slices.Contains(claims.Scopes, requiredScope)

			if !hasScopeInToken {
				helpers.WriteAPIError(w, "Forbidden: Scope not in token", helpers.ErrCodeForbidden, "", http.StatusForbidden)
				return
			}

			hasPermission, err := checker.HasEffectivePermission(r.Context(), claims.UserID, requiredScope)
			if err != nil {
				helpers.WriteAPIError(w, "Internal: Failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				helpers.WriteAPIError(w, "Forbidden: Insufficient permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireScopeRemote(masterClient *filerpc.MasterSyncClient, requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			if !ok {
				http.Error(w, "Unauthorized: No claims found", http.StatusUnauthorized)
				return
			}

			// Ask master: is this user valid AND has this scope?
			allowed, err := masterClient.CheckPermission(r.Context(), claims.UserID, claims.ID, requiredScope)
			if err != nil {
				http.Error(w, "Auth service unavailable", http.StatusServiceUnavailable)
				return
			}

			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SetupCommonMiddleware(r chi.Router, otelEnabled bool, appName string, logger *slog.Logger) {
	if !otelEnabled {
		r.Use(chimiddleware.Logger)
	} else {
		r.Use(otelchi.Middleware(fmt.Sprintf("%s-api", appName)))
		r.Use(LoggingMiddleware(logger))
	}
	r.Use(chimiddleware.Recoverer)
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ww := chimiddleware.NewWrapResponseWriter(w, req.ProtoMajor)
			next.ServeHTTP(ww, req)

			logger.InfoContext(req.Context(), "HTTP Request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", ww.Status(),
			)
		})
	}
}
