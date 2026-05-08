package handlers

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

// ---- shared helpers ----

// backupInfoFromRow converts a single-fetch joined row into the API type.
func backupInfoFromRow(r *sqlc_file_store.GetBackupWithFileRow) models.BackupInfo {
	return models.BackupInfo{
		FileUUID:         r.FileUuid,
		Filename:         r.Filename,
		ContentType:      r.ContentType,
		OwnerID:          r.OwnerID,
		BucketID:         r.BucketID,
		SizeBytes:        r.SizeBytes.Int64,
		BlobSHA256:       r.BlobSha256.String,
		Status:           models.FileStatus(r.Status),
		CreatedAt:        dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:        dbutils.ParseDBTime(r.UpdatedAt),
		SourceNodeID:     r.SourceNodeID.String,
		BackupType:       r.BackupType,
		IsCompressed:     r.IsCompressed.Bool,
		IsEncrypted:      r.IsEncrypted.Bool,
		ParentBackupUUID: r.ParentBackupUuid.String,
	}
}

// backupInfoFromListRow converts a list row (all backups).
func backupInfoFromListRow(r *sqlc_file_store.ListBackupsRow) models.BackupInfo {
	return models.BackupInfo{
		FileUUID:         r.FileUuid,
		Filename:         r.Filename,
		ContentType:      r.ContentType,
		OwnerID:          r.OwnerID,
		BucketID:         r.BucketID,
		SizeBytes:        r.SizeBytes.Int64,
		BlobSHA256:       r.BlobSha256.String,
		Status:           models.FileStatus(r.Status),
		CreatedAt:        dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:        dbutils.ParseDBTime(r.UpdatedAt),
		SourceNodeID:     r.SourceNodeID.String,
		BackupType:       r.BackupType,
		IsCompressed:     r.IsCompressed.Bool,
		IsEncrypted:      r.IsEncrypted.Bool,
		ParentBackupUUID: r.ParentBackupUuid.String,
	}
}

// backupInfoFromOwnerRow converts an owner-filtered list row.
func backupInfoFromOwnerRow(r *sqlc_file_store.ListBackupsByOwnerRow) models.BackupInfo {
	return models.BackupInfo{
		FileUUID:         r.FileUuid,
		Filename:         r.Filename,
		ContentType:      r.ContentType,
		OwnerID:          r.OwnerID,
		BucketID:         r.BucketID,
		SizeBytes:        r.SizeBytes.Int64,
		BlobSHA256:       r.BlobSha256.String,
		Status:           models.FileStatus(r.Status),
		CreatedAt:        dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:        dbutils.ParseDBTime(r.UpdatedAt),
		SourceNodeID:     r.SourceNodeID.String,
		BackupType:       r.BackupType,
		IsCompressed:     r.IsCompressed.Bool,
		IsEncrypted:      r.IsEncrypted.Bool,
		ParentBackupUUID: r.ParentBackupUuid.String,
	}
}

// lookupBackupOrErr fetches the combined backup+file row and writes an
// appropriate HTTP error on failure. The second return value reports success.
func lookupBackupOrErr(
	w http.ResponseWriter,
	r *http.Request,
	f *FilestoreHandlers,
	fileUUID string,
) (sqlc_file_store.GetBackupWithFileRow, bool) {
	row, err := f.Store.GetBackupWithFile(r.Context(), fileUUID)
	if err != nil {
		if isNotFound(err) {
			helpers.WriteAPIError(w, "backup not found", helpers.ErrCodeFileNotFound, "", http.StatusNotFound)
		} else {
			helpers.WriteAPIError(w, "failed to look up backup", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		}
		return sqlc_file_store.GetBackupWithFileRow{}, false
	}
	return row, true
}

// fileFromBackupRow constructs a minimal File for permission helpers.
func fileFromBackupRow(row *sqlc_file_store.GetBackupWithFileRow) sqlc_file_store.File {
	return sqlc_file_store.File{
		FileUuid:   row.FileUuid,
		OwnerID:    row.OwnerID,
		BucketID:   row.BucketID,
		BlobSha256: row.BlobSha256,
	}
}

// ---- handlers ----

// InitBackupUpload initialises a backup upload.
//
//	@Summary		Initialise backup upload
//	@Description	Creates a file record and a backups metadata row atomically, returning a file UUID and upload URL.
//	@Tags			backups
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.InitBackupUploadRequest	true	"Backup upload request"
//	@Success		200		{object}	models.InitUploadResponse
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		403		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/backups [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) InitBackupUpload(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	var req models.InitBackupUploadRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Filename == "" {
		helpers.WriteAPIError(w, "filename is required", helpers.ErrCodeInvalidInput, "missing required field: filename", http.StatusBadRequest)
		return
	}
	if req.SizeBytes <= 0 {
		helpers.WriteAPIError(w, "size_bytes must be greater than 0", helpers.ErrCodeInvalidInput, "invalid value for size_bytes", http.StatusBadRequest)
		return
	}
	if !slices.Contains(models.ValidBackupTypes, req.BackupType) {
		helpers.WriteAPIError(w, fmt.Sprintf("backup_type must be one of: %v", models.ValidBackupTypes), helpers.ErrCodeInvalidInput, "invalid backup_type", http.StatusBadRequest)
		return
	}
	if req.BackupType == "incremental" && req.ParentBackupUUID == "" {
		helpers.WriteAPIError(w, "parent_backup_uuid is required for incremental backups", helpers.ErrCodeInvalidInput, "missing parent_backup_uuid", http.StatusBadRequest)
		return
	}

	bucketID := req.BucketID
	if bucketID == "" {
		bucketID = models.GlobalBucketID
	}

	bucket, ok := lookupBucketOrErr(r.Context(), w, f.Store, bucketID)
	if !ok {
		return
	}

	allowed, err := hasBucketPermission(r.Context(), f.Store, claims, &bucket, "write")
	if err != nil {
		f.Logger.Error("bucket write permission check failed", "err", err, "bucket_id", bucketID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify bucket permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to upload to this bucket", http.StatusForbidden)
		return
	}

	fileUUID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		f.Logger.Error("failed to generate UUID", "err", uuidErr, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to generate a UUID", helpers.ErrCodeFailedToGenerateUUID, uuidErr.Error(), http.StatusInternalServerError)
		return
	}

	var file sqlc_file_store.File
	txErr := f.Store.WithTx(r.Context(), func(q *sqlc_file_store.Queries) error {
		var initErr error
		file, initErr = q.InitFile(r.Context(), sqlc_file_store.InitFileParams{
			FileUuid:        fileUUID.String(),
			Filename:        req.Filename,
			ContentType:     req.ContentType,
			OwnerID:         claims.UserID,
			BucketID:        bucketID,
			LastAccessedAt:  dbutils.FormatDBTime(time.Now()),
			RetentionPeriod: dbutils.NullInt64(req.RetentionPeriod),
		})
		if initErr != nil {
			return fmt.Errorf("init file: %w", initErr)
		}

		isCompressed := req.IsCompressed
		isEncrypted := req.IsEncrypted
		return q.InsertBackup(r.Context(), sqlc_file_store.InsertBackupParams{
			FileUuid:         fileUUID.String(),
			SourceNodeID:     dbutils.NullString(nullableString(req.SourceNodeID)),
			BackupType:       req.BackupType,
			IsCompressed:     dbutils.NullBool(&isCompressed),
			IsEncrypted:      dbutils.NullBool(&isEncrypted),
			ParentBackupUuid: dbutils.NullString(nullableString(req.ParentBackupUUID)),
		})
	})
	if txErr != nil {
		f.Logger.Error("failed to initialise backup", "err", txErr, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to initialise backup", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("backup initialised",
		"file_uuid", file.FileUuid,
		"filename", file.Filename,
		"backup_type", req.BackupType,
		"user_id", claims.UserID,
	)

	expiresAt := time.Now().Add(24 * time.Hour)
	if file.RetentionPeriod.Valid {
		expiresAt = time.Now().Add(time.Duration(file.RetentionPeriod.Int64) * time.Hour)
	}
	f.mustWriteResponse(w, models.InitUploadResponse{
		FileUUID:  file.FileUuid,
		Status:    models.FileStatusPending,
		UploadURL: fmt.Sprintf("/api/v1/files/%s/content", file.FileUuid),
		ExpiresAt: expiresAt,
	}, map[string]any{
		"file_uuid":   file.FileUuid,
		"filename":    file.Filename,
		"backup_type": req.BackupType,
		"user_id":     claims.UserID,
	})
}

// GetBackup returns combined file and backup metadata for a single backup.
//
//	@Summary		Get backup
//	@Description	Returns the combined file and backups metadata for the requested UUID.
//	@Tags			backups
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.BackupInfo
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/backups/{file_uuid} [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GetBackup(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupBackupOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromBackupRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "read")
	if err != nil {
		f.Logger.Error("backup read permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to read this backup", http.StatusForbidden)
		return
	}
	f.Logger.Info("GetBackup", "user_id", claims.UserID, "file_uuid", fileUUID)
	f.mustWriteResponse(w, backupInfoFromRow(&row), map[string]any{
		"file_uuid": fileUUID,
		"user_id":   claims.UserID,
	})
}

// ListBackups lists backups accessible to the authenticated user.
//
//	@Summary		List backups
//	@Description	Returns all backups owned by the caller. Admins see all backups.
//	@Tags			backups
//	@Produce		json
//	@Success		200	{object}	models.ListBackupsResponse
//	@Failure		401	{object}	helpers.APIErrorResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/backups [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListBackups(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	var items []models.BackupInfo

	if isAdminClaims(claims) {
		rows, err := f.Store.ListBackups(r.Context())
		if err != nil {
			f.Logger.Error("failed to list all backups", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list backups", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.BackupInfo, 0, len(rows))
		for i := range rows {
			items = append(items, backupInfoFromListRow(&rows[i]))
		}
	} else {
		rows, err := f.Store.ListBackupsByOwner(r.Context(), claims.UserID)
		if err != nil {
			f.Logger.Error("failed to list backups by owner", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list backups", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.BackupInfo, 0, len(rows))
		for i := range rows {
			items = append(items, backupInfoFromOwnerRow(&rows[i]))
		}
	}
	f.Logger.Info("ListBackups", "user_id", claims.UserID, "count", len(items))

	f.mustWriteResponse(w, models.ListBackupsResponse{Backups: items, Count: len(items)}, map[string]any{
		"user_id": claims.UserID,
	})
}

// UpdateBackup patches the backups metadata for an existing backup.
//
//	@Summary		Update backup metadata
//	@Description	Patches backups fields (source_node_id, backup_type, is_compressed, is_encrypted, parent_backup_uuid). Requires write permission on the file.
//	@Tags			backups
//	@Accept			json
//	@Produce		json
//	@Param			file_uuid	path		string						true	"File UUID"
//	@Param			request		body		models.UpdateBackupRequest	true	"Update request"
//	@Success		200			{object}	models.BackupInfo
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/backups/{file_uuid} [patch]
//	@Security		BearerAuth
func (f *FilestoreHandlers) UpdateBackup(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupBackupOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromBackupRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "write")
	if err != nil {
		f.Logger.Error("backup write permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to update this backup", http.StatusForbidden)
		return
	}

	var req models.UpdateBackupRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// Merge: keep existing values for unset fields.
	newBackupType := row.BackupType
	if req.BackupType != "" {
		if !slices.Contains(models.ValidBackupTypes, req.BackupType) {
			helpers.WriteAPIError(w, fmt.Sprintf("backup_type must be one of: %v", models.ValidBackupTypes), helpers.ErrCodeInvalidInput, "invalid backup_type", http.StatusBadRequest)
			return
		}
		newBackupType = req.BackupType
	}

	newSourceNodeID := row.SourceNodeID
	if req.SourceNodeID != "" {
		newSourceNodeID = dbutils.NullString(nullableString(req.SourceNodeID))
	}

	newIsCompressed := row.IsCompressed
	if req.IsCompressed != nil {
		newIsCompressed = dbutils.NullBool(req.IsCompressed)
	}

	newIsEncrypted := row.IsEncrypted
	if req.IsEncrypted != nil {
		newIsEncrypted = dbutils.NullBool(req.IsEncrypted)
	}

	newParent := row.ParentBackupUuid
	if req.ParentBackupUUID != "" {
		newParent = dbutils.NullString(nullableString(req.ParentBackupUUID))
	}

	if err := f.Store.UpdateBackup(r.Context(), sqlc_file_store.UpdateBackupParams{
		SourceNodeID:     newSourceNodeID,
		BackupType:       newBackupType,
		IsCompressed:     newIsCompressed,
		IsEncrypted:      newIsEncrypted,
		ParentBackupUuid: newParent,
		FileUuid:         fileUUID,
	}); err != nil {
		f.Logger.Error("failed to update backup", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to update backup", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("backup updated",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"backup_type", newBackupType,
	)

	updated, ok := lookupBackupOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}
	f.mustWriteResponse(w, backupInfoFromRow(&updated), map[string]any{
		"file_uuid":   fileUUID,
		"user_id":     claims.UserID,
		"backup_type": newBackupType,
	})
}

// DeleteBackup deletes a backup and its underlying file/blob.
//
//	@Summary		Delete backup
//	@Description	Deletes the backup record, its parent file, and the blob when no other references remain.
//	@Tags			backups
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.DeleteFileResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/backups/{file_uuid} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupBackupOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromBackupRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "admin")
	if err != nil {
		f.Logger.Error("backup delete permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to delete this backup", http.StatusForbidden)
		return
	}

	blobPath, blobDeleted, ok := deleteFileAndBlob(r, w, f, fileUUID)
	if !ok {
		// deleteFileAndBlob already wrote the error response.
		return
	}

	if blobDeleted && blobPath != "" {
		if rmErr := os.Remove(blobPath); rmErr != nil {
			f.Logger.Warn("failed to remove blob from disk", "err", rmErr, "path", blobPath)
		}
	}

	f.Logger.Info("backup deleted",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"blob_deleted", blobDeleted,
	)

	f.mustWriteResponse(w, models.DeleteFileResponse{FileUUID: fileUUID, Status: models.FileStatusDeleted}, map[string]any{
		"file_uuid":    fileUUID,
		"user_id":      claims.UserID,
		"blob_deleted": blobDeleted,
	})
}
