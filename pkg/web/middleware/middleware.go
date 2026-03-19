package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	"github.com/0xveya/gns3util/internal/file-store/db"
	filerpc "github.com/0xveya/gns3util/internal/file-store/rpc"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/scopes"
)

type (
	ctxKey    string
	bucketKey string
	fileKey   string
)

const (
	ClaimsKey ctxKey    = "user_claims"
	BucketKey bucketKey = "bucket"
	FileKey   fileKey   = "file"
)

type PermissionChecker interface {
	HasEffectivePermission(
		ctx context.Context,
		userID string,
		requiredScope string,
	) (bool, error)
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

func AuthMiddleware(
	mgr *auth.IdentityManager,
) func(http.Handler) http.Handler {
	publicPaths := map[string]struct{}{
		"/api/v1/public/files": {},
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			for prefix := range publicPaths {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}
			rawHeader := r.Header.Get("Authorization")
			if rawHeader == "" ||
				!strings.HasPrefix(rawHeader, "Bearer ") {
				helpers.WriteAPIError(
					w,
					"Unauthorized: Missing or malformed token",
					helpers.ErrCodeUnauthorized,
					"",
					http.StatusUnauthorized,
				)
				return
			}

			tokenStr := strings.TrimSpace(
				strings.TrimPrefix(rawHeader, "Bearer "),
			)
			claims, err := mgr.Validate(tokenStr)
			if err != nil {
				helpers.WriteAPIError(
					w,
					"Unauthorized: Invalid token",
					helpers.ErrCodeUnauthorized,
					err.Error(),
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
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			claims, ok := GetClaims(r)
			if !ok {
				helpers.WriteAPIError(
					w,
					"Unauthorized: No claims found",
					helpers.ErrCodeGetClaims,
					"",
					http.StatusUnauthorized,
				)
				return
			}

			revoked, err := checker.IsTokenRevoked(r.Context(), claims.ID)
			if err != nil {
				helpers.WriteAPIError(
					w,
					"Internal: Failed to verify token revocation",
					helpers.ErrCodeInternal,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}
			if revoked {
				helpers.WriteAPIError(
					w,
					"Unauthorized: Token revoked",
					helpers.ErrCodeUnauthorized,
					"",
					http.StatusUnauthorized,
				)
				return
			}

			isSuperInToken := slices.Contains(
				claims.Scopes,
				scopes.Superuser,
			)
			hasScopeInToken := isSuperInToken ||
				slices.Contains(claims.Scopes, requiredScope)

			if !hasScopeInToken {
				helpers.WriteAPIError(
					w,
					"Forbidden: Scope not in token",
					helpers.ErrCodeForbidden,
					"",
					http.StatusForbidden,
				)
				return
			}

			hasPermission, err := checker.HasEffectivePermission(
				r.Context(),
				claims.UserID,
				requiredScope,
			)
			if err != nil {
				helpers.WriteAPIError(
					w,
					"Internal: Failed to verify permissions",
					helpers.ErrCodeInternal,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			if !hasPermission {
				helpers.WriteAPIError(
					w,
					"Forbidden: Insufficient permissions",
					helpers.ErrCodeForbidden,
					"",
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireScopeRemote(
	masterClient *filerpc.MasterSyncClient,
	requiredScope string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			claims, ok := GetClaims(r)
			if !ok {
				helpers.WriteAPIError(
					w,
					"Unauthorized: No claims found",
					helpers.ErrCodeGetClaims,
					"",
					http.StatusUnauthorized,
				)
				return
			}

			if slices.Contains(claims.Scopes, scopes.Superuser) ||
				slices.Contains(claims.Scopes, requiredScope) {
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := masterClient.CheckPermission(
				r.Context(),
				claims.UserID,
				claims.ID,
				claims.Role,
				requiredScope,
			)
			if err != nil {
				helpers.WriteAPIError(
					w,
					"Auth service unavailable",
					helpers.ErrCodeInternal,
					fmt.Sprintf("DRPC Permission Check Failed: %v", err),
					http.StatusServiceUnavailable,
				)
				return
			}

			if !allowed {
				helpers.WriteAPIError(
					w,
					"Forbidden: Insufficient permissions",
					helpers.ErrCodeForbidden,
					"",
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SetupCommonMiddleware(
	r chi.Router,
	otelEnabled bool,
	appName string,
	logger *slog.Logger,
) {
	if !otelEnabled {
		r.Use(chimiddleware.Logger)
	} else {
		r.Use(otelchi.Middleware(fmt.Sprintf("%s-api", appName)))
		r.Use(LoggingMiddleware(logger))
	}
	r.Use(chimiddleware.Recoverer)
}

func LoggingMiddleware(
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			req *http.Request,
		) {
			ww := chimiddleware.NewWrapResponseWriter(
				w,
				req.ProtoMajor,
			)
			next.ServeHTTP(ww, req)

			logger.InfoContext(
				req.Context(),
				"HTTP Request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", ww.Status(),
			)
		})
	}
}

func BucketAccessMiddleware(
	store *db.Store,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bucketID := chi.URLParam(r, "bucket_id")
			claims, ok := GetClaims(r)
			if !ok {
				helpers.WriteAPIError(w, "Unauthorized: No claims found", helpers.ErrCodeGetClaims, "", http.StatusUnauthorized)
				return
			}

			bucket, err := store.GetBucketByID(r.Context(), bucketID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					helpers.WriteAPIError(w, "Bucket not found", helpers.ErrCodeFileNotFound, err.Error(), http.StatusNotFound)
					return
				}
				logger.Error("Failed to get bucket", "bucket_id", bucketID, "err", err)
				helpers.WriteAPIError(w, "Internal: failed to lookup bucket", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
				return
			}

			isSuperuser := slices.Contains(claims.Scopes, scopes.Superuser)

			isGlobalBucket := bucket.BucketID == models.GlobalBucketID

			if !isSuperuser {
				if !isGlobalBucket {
					isPublic := bucket.IsPublic.Valid && bucket.IsPublic.Bool
					if !isPublic && bucket.OwnerID != claims.UserID {
						helpers.WriteAPIError(w, "Forbidden: Insufficient permissions", helpers.ErrCodeForbidden, "You do not own this bucket", http.StatusForbidden)
						return
					}
				}

				if bucket.RequiredScopes.Valid && bucket.RequiredScopes.String != "" {
					requiredScopes := strings.Split(bucket.RequiredScopes.String, ",")
					hasRequiredScope := false
					for _, scope := range requiredScopes {
						if slices.Contains(claims.Scopes, strings.TrimSpace(scope)) {
							hasRequiredScope = true
							break
						}
					}
					if !hasRequiredScope {
						helpers.WriteAPIError(w, "Forbidden: Missing required scopes", helpers.ErrCodeForbidden, "Token missing required scope for this bucket", http.StatusForbidden)
						return
					}
				}
			}

			ctx := context.WithValue(r.Context(), BucketKey, bucket)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FileAccessMiddleware(
	store *db.Store,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fileUUID := chi.URLParam(r, "file_uuid")
			claims, ok := GetClaims(r)
			if !ok {
				helpers.WriteAPIError(w, "Unauthorized: No claims found", helpers.ErrCodeGetClaims, "", http.StatusUnauthorized)
				return
			}

			file, err := store.GetFileByUUID(r.Context(), fileUUID)
			if err != nil {
				logger.Error("Failed to get file", "file_uuid", fileUUID, "err", err)
				helpers.WriteAPIError(w, "Internal: failed to lookup file", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
				return
			}

			isSuperuser := slices.Contains(claims.Scopes, scopes.Superuser)

			if !isSuperuser {
				if file.OwnerID != claims.UserID {
					helpers.WriteAPIError(w, "Forbidden: Insufficient permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
					return
				}
			}

			bucket, err := store.GetBucketByID(r.Context(), file.BucketID)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), BucketKey, bucket))
			} else {
				logger.Warn("Bucket associated with file not found", "bucket_id", file.BucketID, "file_uuid", fileUUID)
			}

			r = r.WithContext(context.WithValue(r.Context(), FileKey, file))
			next.ServeHTTP(w, r)
		})
	}
}
