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
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	filerpc "github.com/0xveya/gns3util/internal/file-store/rpc"
	sharedpb "github.com/0xveya/gns3util/internal/shared/pb"
	"github.com/0xveya/gns3util/pkg/utils/globals"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/helpers"
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

type RBACChecker interface {
	CheckPermission(
		ctx context.Context,
		userID, jti string,
		roles []string,
		action sharedpb.Action,
		resource sharedpb.Resource,
	) (bool, error)
}

type LocalPermissionChecker interface {
	HasEffectivePermission(
		ctx context.Context,
		userID string,
		action sharedpb.Action,
		resource sharedpb.Resource,
	) (bool, error)
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

type localRBACAdapter struct {
	checker LocalPermissionChecker
}

func (a *localRBACAdapter) CheckPermission(
	ctx context.Context,
	userID, jti string,
	_ []string,
	action sharedpb.Action,
	resource sharedpb.Resource,
) (bool, error) {
	revoked, err := a.checker.IsTokenRevoked(ctx, jti)
	if err != nil {
		return false, fmt.Errorf("token revocation check: %w", err)
	}
	if revoked {
		return false, nil
	}
	return a.checker.HasEffectivePermission(ctx, userID, action, resource)
}

func GetClaims(r *http.Request) (*auth.Claims, bool) {
	claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
	return claims, ok
}

func mustGetClaims(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	claims, ok := GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "Unauthorized: No claims found", helpers.ErrCodeGetClaims, "", http.StatusUnauthorized)
	}
	return claims, ok
}

func claimsRoleNames(claims *auth.Claims) []string {
	if claims.Role == "" {
		return nil
	}
	return []string{claims.Role}
}

func isAdmin(claims *auth.Claims) bool {
	return slices.Contains(claimsRoleNames(claims), globals.RoleAdmin)
}

func AuthMiddleware(mgr *auth.IdentityManager) func(http.Handler) http.Handler {
	publicPaths := map[string]struct{}{
		"/api/v1/public/files": {},
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for prefix := range publicPaths {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

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

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ClaimsKey, claims)))
		})
	}
}

func RequirePermission(
	checker RBACChecker,
	action sharedpb.Action,
	resource sharedpb.Resource,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := mustGetClaims(w, r)
			if !ok {
				return
			}

			allowed, err := checker.CheckPermission(
				r.Context(),
				claims.UserID,
				claims.ID,
				claimsRoleNames(claims),
				action,
				resource,
			)
			if err != nil {
				helpers.WriteAPIError(w, "Auth service unavailable", helpers.ErrCodeInternal,
					fmt.Sprintf("permission check failed: %v", err), http.StatusServiceUnavailable)
				return
			}
			if !allowed {
				helpers.WriteAPIError(w, "Forbidden: Insufficient permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireScope(
	checker LocalPermissionChecker,
	action sharedpb.Action,
	resource sharedpb.Resource,
) func(http.Handler) http.Handler {
	return RequirePermission(&localRBACAdapter{checker: checker}, action, resource)
}

func RequireScopeRemote(
	masterClient *filerpc.MasterSyncClient,
	action sharedpb.Action,
	resource sharedpb.Resource,
) func(http.Handler) http.Handler {
	return RequirePermission(masterClient, action, resource)
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

func BucketAccessMiddleware(store *db.Store, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := mustGetClaims(w, r)
			if !ok {
				return
			}

			bucketID := chi.URLParam(r, "bucket_id")
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

			// For non-admin users, enforce any bucket-level role requirement.
			// Ownership and ACL-based access are handled downstream by
			// HasBucketPermission, so we intentionally do NOT reject
			// non-owners here - the ACL table may grant them access.
			if !isAdmin(claims) && bucket.RequiredScopes.Valid && bucket.RequiredScopes.String != "" {
				userRoles := claimsRoleNames(claims)
				hasRole := false
				for _, required := range strings.Split(bucket.RequiredScopes.String, ",") {
					if slices.Contains(userRoles, strings.TrimSpace(required)) {
						hasRole = true
						break
					}
				}
				if !hasRole {
					helpers.WriteAPIError(w, "Forbidden: Missing required role", helpers.ErrCodeForbidden, "Token missing required role for this bucket", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), BucketKey, bucket)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FileAccessMiddleware(store *db.Store, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := mustGetClaims(w, r)
			if !ok {
				return
			}

			fileUUID := chi.URLParam(r, "file_uuid")
			file, err := store.GetFileByUUID(r.Context(), fileUUID)
			if err != nil {
				logger.Error("Failed to get file", "file_uuid", fileUUID, "err", err)
				helpers.WriteAPIError(w, "Internal: failed to lookup file", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
				return
			}

			bucket, err := store.GetBucketByID(r.Context(), file.BucketID)
			if err == nil {
				if !isAdmin(claims) && bucket.RequiredScopes.Valid && bucket.RequiredScopes.String != "" {
					userRoles := claimsRoleNames(claims)
					hasRole := false
					for _, required := range strings.Split(bucket.RequiredScopes.String, ",") {
						if slices.Contains(userRoles, strings.TrimSpace(required)) {
							hasRole = true
							break
						}
					}
					if !hasRole {
						helpers.WriteAPIError(w, "Forbidden: Missing required role", helpers.ErrCodeForbidden, "Token missing required role for this bucket", http.StatusForbidden)
						return
					}
				}
				r = r.WithContext(context.WithValue(r.Context(), BucketKey, bucket))
			} else {
				logger.Warn("Bucket associated with file not found", "bucket_id", file.BucketID, "file_uuid", fileUUID)
			}

			r = r.WithContext(context.WithValue(r.Context(), FileKey, file))
			next.ServeHTTP(w, r)
		})
	}
}

func permissionImplies(required string) []string {
	switch required {
	case "read":
		return []string{"read", "write", "admin"}
	case "write":
		return []string{"write", "admin"}
	case "admin":
		return []string{"admin"}
	default:
		return []string{required}
	}
}

// HasBucketPermission enforces a minimum permission level on the bucket identified
// by the {bucket_id} URL parameter. It expects BucketAccessMiddleware to have
// already loaded the bucket into context.
//
// Access is granted when ANY of the following is true:
//  1. The user is a global admin (role "admin").
//  2. The user owns the bucket.
//  3. The user (or one of their roles) has a matching entry in bucket_permissions
//     at or above the required level (admin > write > read).
func HasBucketPermission(store *db.Store, required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := mustGetClaims(w, r)
			if !ok {
				return
			}

			// Global admin always passes
			if isAdmin(claims) {
				next.ServeHTTP(w, r)
				return
			}

			bucket, hasBucket := r.Context().Value(BucketKey).(sqlc_file_store.GetBucketByIDRow)
			if !hasBucket {
				// Fallback: load bucket directly.
				bucketID := chi.URLParam(r, "bucket_id")
				var err error
				bucket, err = store.GetBucketByID(r.Context(), bucketID)
				if err != nil {
					helpers.WriteAPIError(w, "Bucket not found", helpers.ErrCodeFileNotFound, err.Error(), http.StatusNotFound)
					return
				}
			}

			// Owner always passes
			if bucket.OwnerID == claims.UserID {
				next.ServeHTTP(w, r)
				return
			}

			// Check acl for each implied permission level
			roleNames := claimsRoleNames(claims)
			for _, perm := range permissionImplies(required) {
				ok, err := store.HasEffectiveBucketPermission(r.Context(), bucket.BucketID, claims.UserID, perm, roleNames)
				if err != nil {
					helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
					return
				}
				if ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			helpers.WriteAPIError(w, "Forbidden: Insufficient permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		})
	}
}

// thanks claude for explaining this in a comment instead of only living temporarely in my brain
// HasFilePermission enforces a minimum permission level on the file identified
// by the {file_uuid} URL parameter.  It expects FileAccessMiddleware to have
// already loaded the file into context.
//
// Access is granted when any of the following is true:
//  1. The user is a global admin (role "admin").
//  2. The user owns the file.
//  3. The user (or one of their roles) has a matching file-level ACL entry.
//  4. The user (or one of their roles) has a matching bucket-level ACL entry
//     (file permissions inherit from the parent bucket).
func HasFilePermission(store *db.Store, required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := mustGetClaims(w, r)
			if !ok {
				return
			}

			// Global admin always passes
			if isAdmin(claims) {
				next.ServeHTTP(w, r)
				return
			}

			file, hasFile := r.Context().Value(FileKey).(sqlc_file_store.File)
			if !hasFile {
				fileUUID := chi.URLParam(r, "file_uuid")
				var err error
				file, err = store.GetFileByUUID(r.Context(), fileUUID)
				if err != nil {
					helpers.WriteAPIError(w, "File not found", helpers.ErrCodeFileNotFound, err.Error(), http.StatusNotFound)
					return
				}
			}

			// Owner always passes
			if file.OwnerID == claims.UserID {
				next.ServeHTTP(w, r)
				return
			}

			roleNames := claimsRoleNames(claims)
			implied := permissionImplies(required)

			// Check file-level acl
			for _, perm := range implied {
				ok, err := store.HasEffectiveFilePermission(r.Context(), file.FileUuid, claims.UserID, perm, roleNames)
				if err != nil {
					helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
					return
				}
				if ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Inherit from bucket-level acl
			if file.BucketID != "" {
				for _, perm := range implied {
					ok, err := store.HasEffectiveBucketPermission(r.Context(), file.BucketID, claims.UserID, perm, roleNames)
					if err != nil {
						helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
						return
					}
					if ok {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			helpers.WriteAPIError(w, "Forbidden: Insufficient permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		})
	}
}
