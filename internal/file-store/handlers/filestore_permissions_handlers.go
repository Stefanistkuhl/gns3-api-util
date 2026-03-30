package handlers

import (
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

var (
	validPrincipalTypes = []string{"user", "group", "role"}
	validPermissions    = []string{"read", "write", "admin"}
)

// GrantBucketPermission grants a permission entry on a bucket.
//
//	@Summary		Grant bucket permission
//	@Description	Grants read/write/admin permission on a bucket to a user, group, or role.
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Param			bucket_id	path		string							true	"Bucket ID"
//	@Param			request		body		models.GrantPermissionRequest	true	"Grant permission request"
//	@Success		200			{object}	models.PermissionEntry
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id}/permissions [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GrantBucketPermission(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	bucketID := chi.URLParam(r, "bucket_id")
	bucket, ok := lookupBucketOrErr(r.Context(), w, f.Store, bucketID)
	if !ok {
		return
	}

	canManage, err := canManageBucketPermissions(r.Context(), f.Store, claims, &bucket)
	if err != nil {
		f.Logger.Error("permission management check failed", "err", err, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canManage {
		helpers.WriteAPIError(w, "Forbidden: only the bucket owner or an admin grantee can manage permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		return
	}

	var req models.GrantPermissionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if !slices.Contains(validPrincipalTypes, req.PrincipalType) {
		helpers.WriteAPIError(w, "invalid principal_type: must be user, group, or role", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}
	if req.PrincipalID == "" {
		helpers.WriteAPIError(w, "principal_id is required", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}
	if !slices.Contains(validPermissions, req.Permission) {
		helpers.WriteAPIError(w, "invalid permission: must be read, write, or admin", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		f.Logger.Error("failed to generate UUID for bucket permission", "err", err)
		helpers.WriteAPIError(w, "failed to generate permission ID", helpers.ErrCodeFailedToGenerateUUID, err.Error(), http.StatusInternalServerError)
		return
	}

	entry, err := f.Store.InsertBucketPermission(r.Context(), sqlc_file_store.InsertBucketPermissionParams{
		ID:            id.String(),
		BucketID:      bucketID,
		PrincipalType: req.PrincipalType,
		PrincipalID:   req.PrincipalID,
		Permission:    req.Permission,
		GrantedBy:     claims.UserID,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		f.Logger.Error("failed to insert bucket permission", "err", err, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "failed to insert permission", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("bucket permission granted",
		"bucket_id", bucketID,
		"principal_type", req.PrincipalType,
		"principal_id", req.PrincipalID,
		"permission", req.Permission,
		"granted_by", claims.UserID,
	)

	writeJSON(w, r, models.PermissionEntry{
		ID:            entry.ID,
		PrincipalType: entry.PrincipalType,
		PrincipalID:   entry.PrincipalID,
		Permission:    entry.Permission,
		GrantedBy:     entry.GrantedBy,
		CreatedAt:     entry.CreatedAt,
		ExpiresAt:     entry.ExpiresAt,
	}, f.Logger)
}

// RevokeBucketPermission deletes a specific permission entry from a bucket.
//
//	@Summary		Revoke bucket permission
//	@Description	Removes a permission entry from a bucket by permission ID.
//	@Tags			buckets
//	@Produce		json
//	@Param			bucket_id	path		string	true	"Bucket ID"
//	@Param			perm_id		path		string	true	"Permission entry ID"
//	@Success		200			{object}	models.RevokePermissionResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id}/permissions/{perm_id} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) RevokeBucketPermission(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	bucketID := chi.URLParam(r, "bucket_id")
	permID := chi.URLParam(r, "perm_id")

	bucket, ok := lookupBucketOrErr(r.Context(), w, f.Store, bucketID)
	if !ok {
		return
	}

	canManage, err := canManageBucketPermissions(r.Context(), f.Store, claims, &bucket)
	if err != nil {
		f.Logger.Error("permission management check failed", "err", err, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canManage {
		helpers.WriteAPIError(w, "Forbidden: only the bucket owner or an admin grantee can manage permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		return
	}

	if err := f.Store.DeleteBucketPermission(r.Context(), sqlc_file_store.DeleteBucketPermissionParams{
		ID:       permID,
		BucketID: bucketID,
	}); err != nil {
		f.Logger.Error("failed to delete bucket permission", "err", err, "perm_id", permID)
		helpers.WriteAPIError(w, "failed to delete permission", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("bucket permission revoked", "perm_id", permID, "bucket_id", bucketID, "revoked_by", claims.UserID)

	writeJSON(w, r, models.RevokePermissionResponse{ID: permID, Deleted: true}, f.Logger)
}

// ListBucketPermissions returns all delegated permission entries for a bucket.
//
//	@Summary		List bucket permissions
//	@Description	Returns all permission entries on a bucket. Requires owner or admin permission.
//	@Tags			buckets
//	@Produce		json
//	@Param			bucket_id	path		string	true	"Bucket ID"
//	@Success		200			{object}	models.ListPermissionsResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id}/permissions [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListBucketPermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	bucketID := chi.URLParam(r, "bucket_id")
	bucket, ok := lookupBucketOrErr(r.Context(), w, f.Store, bucketID)
	if !ok {
		return
	}

	canManage, err := canManageBucketPermissions(r.Context(), f.Store, claims, &bucket)
	if err != nil {
		f.Logger.Error("permission management check failed", "err", err, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canManage {
		helpers.WriteAPIError(w, "Forbidden: only the bucket owner or an admin grantee can list permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		return
	}

	rows, err := f.Store.ListBucketPermissions(r.Context(), bucketID)
	if err != nil {
		f.Logger.Error("failed to list bucket permissions", "err", err, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "failed to list permissions", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	entries := make([]models.PermissionEntry, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		entries = append(entries, models.PermissionEntry{
			ID:            row.ID,
			PrincipalType: row.PrincipalType,
			PrincipalID:   row.PrincipalID,
			Permission:    row.Permission,
			GrantedBy:     row.GrantedBy,
			CreatedAt:     row.CreatedAt,
			ExpiresAt:     row.ExpiresAt,
		})
	}

	writeJSON(w, r, models.ListPermissionsResponse{Permissions: entries, Count: len(entries)}, f.Logger)
}

// GrantFilePermission grants a permission entry on a specific file.
//
//	@Summary		Grant file permission
//	@Description	Grants read/write/admin permission on a specific file to a user, group, or role.
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Param			file_uuid	path		string							true	"File UUID"
//	@Param			request		body		models.GrantPermissionRequest	true	"Grant permission request"
//	@Success		200			{object}	models.PermissionEntry
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/permissions [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GrantFilePermission(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	fileUUID := chi.URLParam(r, "file_uuid")
	file, ok := lookupFileOrErr(r.Context(), w, f.Store, fileUUID)
	if !ok {
		return
	}

	canManage, err := canManageFilePermissions(r.Context(), f.Store, claims, &file)
	if err != nil {
		f.Logger.Error("file permission management check failed", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canManage {
		helpers.WriteAPIError(w, "Forbidden: only the file owner or an admin grantee can manage permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		return
	}

	var req models.GrantPermissionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if !slices.Contains(validPrincipalTypes, req.PrincipalType) {
		helpers.WriteAPIError(w, "invalid principal_type: must be user, group, or role", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}
	if req.PrincipalID == "" {
		helpers.WriteAPIError(w, "principal_id is required", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}
	if !slices.Contains(validPermissions, req.Permission) {
		helpers.WriteAPIError(w, "invalid permission: must be read, write, or admin", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		f.Logger.Error("failed to generate UUID for file permission", "err", err)
		helpers.WriteAPIError(w, "failed to generate permission ID", helpers.ErrCodeFailedToGenerateUUID, err.Error(), http.StatusInternalServerError)
		return
	}

	entry, err := f.Store.InsertFilePermission(r.Context(), sqlc_file_store.InsertFilePermissionParams{
		ID:            id.String(),
		FileUuid:      fileUUID,
		PrincipalType: req.PrincipalType,
		PrincipalID:   req.PrincipalID,
		Permission:    req.Permission,
		GrantedBy:     claims.UserID,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		f.Logger.Error("failed to insert file permission", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to insert permission", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("file permission granted",
		"file_uuid", fileUUID,
		"principal_type", req.PrincipalType,
		"principal_id", req.PrincipalID,
		"permission", req.Permission,
		"granted_by", claims.UserID,
	)

	writeJSON(w, r, models.PermissionEntry{
		ID:            entry.ID,
		PrincipalType: entry.PrincipalType,
		PrincipalID:   entry.PrincipalID,
		Permission:    entry.Permission,
		GrantedBy:     entry.GrantedBy,
		CreatedAt:     entry.CreatedAt,
		ExpiresAt:     entry.ExpiresAt,
	}, f.Logger)
}

// RevokeFilePermission deletes a specific permission entry from a file.
//
//	@Summary		Revoke file permission
//	@Description	Removes a permission entry from a file by permission ID.
//	@Tags			files
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Param			perm_id		path		string	true	"Permission entry ID"
//	@Success		200			{object}	models.RevokePermissionResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/permissions/{perm_id} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) RevokeFilePermission(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	fileUUID := chi.URLParam(r, "file_uuid")
	permID := chi.URLParam(r, "perm_id")

	file, ok := lookupFileOrErr(r.Context(), w, f.Store, fileUUID)
	if !ok {
		return
	}

	canManage, err := canManageFilePermissions(r.Context(), f.Store, claims, &file)
	if err != nil {
		f.Logger.Error("file permission management check failed", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canManage {
		helpers.WriteAPIError(w, "Forbidden: only the file owner or an admin grantee can manage permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		return
	}

	if err := f.Store.DeleteFilePermission(r.Context(), sqlc_file_store.DeleteFilePermissionParams{
		ID:       permID,
		FileUuid: fileUUID,
	}); err != nil {
		f.Logger.Error("failed to delete file permission", "err", err, "perm_id", permID)
		helpers.WriteAPIError(w, "failed to delete permission", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("file permission revoked", "perm_id", permID, "file_uuid", fileUUID, "revoked_by", claims.UserID)

	writeJSON(w, r, models.RevokePermissionResponse{ID: permID, Deleted: true}, f.Logger)
}

// ListFilePermissions returns all delegated permission entries for a file.
//
//	@Summary		List file permissions
//	@Description	Returns all permission entries on a file. Requires owner or admin permission.
//	@Tags			files
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.ListPermissionsResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/permissions [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListFilePermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	fileUUID := chi.URLParam(r, "file_uuid")
	file, ok := lookupFileOrErr(r.Context(), w, f.Store, fileUUID)
	if !ok {
		return
	}

	canManage, err := canManageFilePermissions(r.Context(), f.Store, claims, &file)
	if err != nil {
		f.Logger.Error("file permission management check failed", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "Internal: permission check failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canManage {
		helpers.WriteAPIError(w, "Forbidden: only the file owner or an admin grantee can list permissions", helpers.ErrCodeForbidden, "", http.StatusForbidden)
		return
	}

	rows, err := f.Store.ListFilePermissions(r.Context(), fileUUID)
	if err != nil {
		f.Logger.Error("failed to list file permissions", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to list permissions", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	entries := make([]models.PermissionEntry, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		entries = append(entries, models.PermissionEntry{
			ID:            row.ID,
			PrincipalType: row.PrincipalType,
			PrincipalID:   row.PrincipalID,
			Permission:    row.Permission,
			GrantedBy:     row.GrantedBy,
			CreatedAt:     row.CreatedAt,
			ExpiresAt:     row.ExpiresAt,
		})
	}

	writeJSON(w, r, models.ListPermissionsResponse{Permissions: entries, Count: len(entries)}, f.Logger)
}
