package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/utils/globals"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
)

// isNotFound reports whether err is a sql.ErrNoRows sentinel.
func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func mustClaims(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w,
			"failed to get claims from jwt",
			helpers.ErrCodeGetClaims,
			"claims missing past AuthMiddleware – this is a bug",
			http.StatusInternalServerError,
		)
	}
	return claims, ok
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request, v *T) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidJSON, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func lookupBucketOrErr(
	ctx context.Context,
	w http.ResponseWriter,
	store *db.Store,
	bucketID string,
) (sqlc_file_store.GetBucketByIDRow, bool) {
	bucket, err := store.GetBucketByID(ctx, bucketID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "Bucket not found", helpers.ErrCodeFileNotFound, "", http.StatusNotFound)
		} else {
			helpers.WriteAPIError(w, "Internal: failed to look up bucket", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		}
		return sqlc_file_store.GetBucketByIDRow{}, false
	}
	return bucket, true
}

func lookupFileOrErr(
	ctx context.Context,
	w http.ResponseWriter,
	store *db.Store,
	fileUUID string,
) (sqlc_file_store.File, bool) {
	file, err := store.GetFileByUUID(ctx, fileUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "File not found", helpers.ErrCodeFileNotFound, "", http.StatusNotFound)
		} else {
			helpers.WriteAPIError(w, "Internal: failed to look up file", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		}
		return sqlc_file_store.File{}, false
	}
	return file, true
}

func canManageBucketPermissions(
	ctx context.Context,
	store *db.Store,
	claims *auth.Claims,
	bucket *sqlc_file_store.GetBucketByIDRow,
) (bool, error) {
	return hasBucketPermission(ctx, store, claims, bucket, "admin")
}

func canManageFilePermissions(
	ctx context.Context,
	store *db.Store,
	claims *auth.Claims,
	file *sqlc_file_store.File,
) (bool, error) {
	return hasFilePermission(ctx, store, claims, file, "admin")
}

func roleNamesFromClaims(claims *auth.Claims) []string {
	if claims.Role == "" {
		return nil
	}
	return []string{claims.Role}
}

func isAdminClaims(claims *auth.Claims) bool {
	return slices.Contains(roleNamesFromClaims(claims), globals.RoleAdmin)
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

func bucketRoleRequirementSatisfied(bucket *sqlc_file_store.GetBucketByIDRow, claims *auth.Claims) bool {
	if bucket == nil || isAdminClaims(claims) {
		return true
	}
	if !bucket.RequiredScopes.Valid || bucket.RequiredScopes.String == "" {
		return true
	}

	roleNames := roleNamesFromClaims(claims)
	for _, required := range strings.Split(bucket.RequiredScopes.String, ",") {
		if slices.Contains(roleNames, strings.TrimSpace(required)) {
			return true
		}
	}
	return false
}

func hasBucketPermission(
	ctx context.Context,
	store *db.Store,
	claims *auth.Claims,
	bucket *sqlc_file_store.GetBucketByIDRow,
	required string,
) (bool, error) {
	if isAdminClaims(claims) {
		return true, nil
	}
	if bucket.OwnerID == claims.UserID {
		return true, nil
	}
	if !bucketRoleRequirementSatisfied(bucket, claims) {
		return false, nil
	}

	roleNames := roleNamesFromClaims(claims)
	for _, perm := range permissionImplies(required) {
		ok, err := store.HasEffectiveBucketPermission(ctx, bucket.BucketID, claims.UserID, perm, roleNames)
		if err != nil {
			return false, fmt.Errorf("bucket permission check: %w", err)
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

func hasFilePermission(
	ctx context.Context,
	store *db.Store,
	claims *auth.Claims,
	file *sqlc_file_store.File,
	required string,
) (bool, error) {
	if isAdminClaims(claims) {
		return true, nil
	}
	if file.OwnerID == claims.UserID {
		return true, nil
	}

	var bucket *sqlc_file_store.GetBucketByIDRow
	if file.BucketID != "" {
		loadedBucket, err := store.GetBucketByID(ctx, file.BucketID)
		if err != nil {
			return false, fmt.Errorf("load file bucket: %w", err)
		}
		bucket = &loadedBucket
		if !bucketRoleRequirementSatisfied(bucket, claims) {
			return false, nil
		}
	}

	roleNames := roleNamesFromClaims(claims)
	for _, perm := range permissionImplies(required) {
		ok, err := store.HasEffectiveFilePermission(ctx, file.FileUuid, claims.UserID, perm, roleNames)
		if err != nil {
			return false, fmt.Errorf("file permission check: %w", err)
		}
		if ok {
			return true, nil
		}
	}

	if bucket != nil {
		for _, perm := range permissionImplies(required) {
			ok, err := store.HasEffectiveBucketPermission(ctx, file.BucketID, claims.UserID, perm, roleNames)
			if err != nil {
				return false, fmt.Errorf("bucket permission check: %w", err)
			}
			if ok {
				return true, nil
			}
		}
	}

	return false, nil
}

type httpRange struct {
	start  int64
	length int64
}

func parseRange(s string, size int64) ([]httpRange, error) {
	if !strings.HasPrefix(s, "bytes=") {
		return nil, fmt.Errorf("invalid range format")
	}

	ranges := []httpRange{}
	for ra := range strings.SplitSeq(s[6:], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}

		parts := strings.Split(ra, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format")
		}

		start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range start")
		}

		end, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range end")
		}

		if start > end || start >= size {
			return nil, fmt.Errorf("invalid range")
		}

		if end >= size {
			end = size - 1
		}

		ranges = append(ranges, httpRange{start: start, length: end - start + 1})
	}

	return ranges, nil
}

func (f *FilestoreHandlers) mustWriteResponse(
	w http.ResponseWriter,
	data any,
	logFields map[string]any,
) {
	if err := helpers.WriteJSON(w, data); err != nil {
		fields := make([]any, 0, 2+2*len(logFields))
		fields = append(fields, "err", err)

		for k, v := range logFields {
			fields = append(fields, k, v)
		}

		f.Logger.Error("Failed to write response", fields...)

		helpers.WriteAPIError(
			w,
			"failed to write response",
			helpers.ErrCodeInternal,
			err.Error(),
			http.StatusInternalServerError,
		)
	}
}
