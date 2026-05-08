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

// projectFileInfoFromRow converts a single-fetch joined row into the API type.
func projectFileInfoFromRow(r *sqlc_file_store.GetProjectFileWithFileRow) models.ProjectFileInfo {
	return models.ProjectFileInfo{
		FileUUID:          r.FileUuid,
		Filename:          r.Filename,
		ContentType:       r.ContentType,
		OwnerID:           r.OwnerID,
		BucketID:          r.BucketID,
		SizeBytes:         r.SizeBytes.Int64,
		BlobSHA256:        r.BlobSha256.String,
		Status:            models.FileStatus(r.Status),
		CreatedAt:         dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:         dbutils.ParseDBTime(r.UpdatedAt),
		ProjectID:         r.ProjectID,
		ProjectName:       r.ProjectName.String,
		VersionTag:        r.VersionTag.String,
		IsReadOnly:        r.IsReadOnly.Bool,
		IncludeSnapshots:  r.IncludeSnapshots,
		IncludeImages:     r.IncludeImages,
		ResetMacAddresses: r.ResetMacAddresses,
		KeepComputeIds:    r.KeepComputeIds,
		Compression:       r.Compression,
	}
}

// projectFileInfoFromListRow converts a list row (all project files).
func projectFileInfoFromListRow(r *sqlc_file_store.ListProjectFilesRow) models.ProjectFileInfo {
	return models.ProjectFileInfo{
		FileUUID:          r.FileUuid,
		Filename:          r.Filename,
		ContentType:       r.ContentType,
		OwnerID:           r.OwnerID,
		BucketID:          r.BucketID,
		SizeBytes:         r.SizeBytes.Int64,
		BlobSHA256:        r.BlobSha256.String,
		Status:            models.FileStatus(r.Status),
		CreatedAt:         dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:         dbutils.ParseDBTime(r.UpdatedAt),
		ProjectID:         r.ProjectID,
		ProjectName:       r.ProjectName.String,
		VersionTag:        r.VersionTag.String,
		IsReadOnly:        r.IsReadOnly.Bool,
		IncludeSnapshots:  r.IncludeSnapshots,
		IncludeImages:     r.IncludeImages,
		ResetMacAddresses: r.ResetMacAddresses,
		KeepComputeIds:    r.KeepComputeIds,
		Compression:       r.Compression,
	}
}

// projectFileInfoFromOwnerRow converts an owner-filtered list row.
func projectFileInfoFromOwnerRow(r *sqlc_file_store.ListProjectFilesByOwnerRow) models.ProjectFileInfo {
	return models.ProjectFileInfo{
		FileUUID:          r.FileUuid,
		Filename:          r.Filename,
		ContentType:       r.ContentType,
		OwnerID:           r.OwnerID,
		BucketID:          r.BucketID,
		SizeBytes:         r.SizeBytes.Int64,
		BlobSHA256:        r.BlobSha256.String,
		Status:            models.FileStatus(r.Status),
		CreatedAt:         dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:         dbutils.ParseDBTime(r.UpdatedAt),
		ProjectID:         r.ProjectID,
		ProjectName:       r.ProjectName.String,
		VersionTag:        r.VersionTag.String,
		IsReadOnly:        r.IsReadOnly.Bool,
		IncludeSnapshots:  r.IncludeSnapshots,
		IncludeImages:     r.IncludeImages,
		ResetMacAddresses: r.ResetMacAddresses,
		KeepComputeIds:    r.KeepComputeIds,
		Compression:       r.Compression,
	}
}

// projectFileInfoFromProjectRow converts a project-id-filtered list row.
func projectFileInfoFromProjectRow(r *sqlc_file_store.ListProjectFilesByProjectRow) models.ProjectFileInfo {
	return models.ProjectFileInfo{
		FileUUID:          r.FileUuid,
		Filename:          r.Filename,
		ContentType:       r.ContentType,
		OwnerID:           r.OwnerID,
		BucketID:          r.BucketID,
		SizeBytes:         r.SizeBytes.Int64,
		BlobSHA256:        r.BlobSha256.String,
		Status:            models.FileStatus(r.Status),
		CreatedAt:         dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:         dbutils.ParseDBTime(r.UpdatedAt),
		ProjectID:         r.ProjectID,
		ProjectName:       r.ProjectName.String,
		VersionTag:        r.VersionTag.String,
		IsReadOnly:        r.IsReadOnly.Bool,
		IncludeSnapshots:  r.IncludeSnapshots,
		IncludeImages:     r.IncludeImages,
		ResetMacAddresses: r.ResetMacAddresses,
		KeepComputeIds:    r.KeepComputeIds,
		Compression:       r.Compression,
	}
}

// lookupProjectFileOrErr fetches the combined project_file+file row and writes
// an appropriate HTTP error on failure.
func lookupProjectFileOrErr(
	w http.ResponseWriter,
	r *http.Request,
	f *FilestoreHandlers,
	fileUUID string,
) (sqlc_file_store.GetProjectFileWithFileRow, bool) {
	row, err := f.Store.GetProjectFileWithFile(r.Context(), fileUUID)
	if err != nil {
		if isNotFound(err) {
			helpers.WriteAPIError(w, "project file not found", helpers.ErrCodeFileNotFound, "", http.StatusNotFound)
		} else {
			helpers.WriteAPIError(w, "failed to look up project file", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		}
		return sqlc_file_store.GetProjectFileWithFileRow{}, false
	}
	return row, true
}

// fileFromProjectFileRow constructs a minimal File for permission helpers.
func fileFromProjectFileRow(row *sqlc_file_store.GetProjectFileWithFileRow) sqlc_file_store.File {
	return sqlc_file_store.File{
		FileUuid:   row.FileUuid,
		OwnerID:    row.OwnerID,
		BucketID:   row.BucketID,
		BlobSha256: row.BlobSha256,
	}
}

// defaultCompression returns "zstd" when the given string is empty.
func defaultCompression(s string) string {
	if s == "" {
		return "zstd"
	}
	return s
}

// ---- handlers ----

// InitProjectFileUpload initialises a GNS3 project-export upload.
//
//	@Summary		Initialise project file upload
//	@Description	Creates a file record and a project_files metadata row atomically.
//	@Description	The fields mirror the options available on the GNS3 project export command.
//	@Tags			project-files
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.InitProjectFileUploadRequest	true	"Project file upload request"
//	@Success		200		{object}	models.InitUploadResponse
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		403		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/project-files [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) InitProjectFileUpload(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	var req models.InitProjectFileUploadRequest
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
	if req.ProjectID == "" {
		helpers.WriteAPIError(w, "project_id is required", helpers.ErrCodeInvalidInput, "missing required field: project_id", http.StatusBadRequest)
		return
	}

	compression := defaultCompression(req.Compression)
	if !slices.Contains(models.ValidCompressionTypes, compression) {
		helpers.WriteAPIError(w,
			fmt.Sprintf("compression must be one of: %v", models.ValidCompressionTypes),
			helpers.ErrCodeInvalidInput, "invalid compression", http.StatusBadRequest)
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

	isReadOnly := req.IsReadOnly
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

		return q.InsertProjectFile(r.Context(), sqlc_file_store.InsertProjectFileParams{
			FileUuid:          fileUUID.String(),
			ProjectID:         req.ProjectID,
			ProjectName:       dbutils.NullString(nullableString(req.ProjectName)),
			VersionTag:        dbutils.NullString(nullableString(req.VersionTag)),
			IsReadOnly:        dbutils.NullBool(&isReadOnly),
			IncludeSnapshots:  req.IncludeSnapshots,
			IncludeImages:     req.IncludeImages,
			ResetMacAddresses: req.ResetMacAddresses,
			KeepComputeIds:    req.KeepComputeIds,
			Compression:       compression,
		})
	})
	if txErr != nil {
		f.Logger.Error("failed to initialise project file", "err", txErr, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to initialise project file", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("project file initialised",
		"file_uuid", file.FileUuid,
		"filename", file.Filename,
		"project_id", req.ProjectID,
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
	}, map[string]any{"file_uuid": file.FileUuid, "project_id": req.ProjectID, "user_id": claims.UserID})
}

// GetProjectFile returns combined file and project_files metadata for a single record.
//
//	@Summary		Get project file
//	@Description	Returns the combined file and project_files metadata for the requested UUID.
//	@Tags			project-files
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.ProjectFileInfo
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/project-files/{file_uuid} [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GetProjectFile(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupProjectFileOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromProjectFileRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "read")
	if err != nil {
		f.Logger.Error("project file read permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to read this project file", http.StatusForbidden)
		return
	}

	f.Logger.Info("project file read", "file_uuid", fileUUID, "user_id", claims.UserID)
	f.mustWriteResponse(w, projectFileInfoFromRow(&row), map[string]any{"file_uuid": fileUUID})
}

// ListProjectFiles lists project files accessible to the authenticated user.
//
//	@Summary		List project files
//	@Description	Returns project files owned by the caller, optionally filtered by project_id query param.
//	@Description	Admins see all project files. The optional ?project_id= filter is applied after access control.
//	@Tags			project-files
//	@Produce		json
//	@Param			project_id	query		string	false	"Filter by GNS3 project UUID"
//	@Success		200			{object}	models.ListProjectFilesResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/project-files [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	// Optional project_id filter from query string.
	projectIDFilter := r.URL.Query().Get("project_id")

	var items []models.ProjectFileInfo

	// If a project_id filter is provided we always use the per-project query
	// (visible to anyone who can read individual files for that project, so we
	// still gate on the caller's files access below).
	switch {
	case projectIDFilter != "":
		rows, err := f.Store.ListProjectFilesByProject(r.Context(), projectIDFilter)
		if err != nil {
			f.Logger.Error("failed to list project files by project", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list project files", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.ProjectFileInfo, 0, len(rows))
		for i := range rows {
			pf := projectFileInfoFromProjectRow(&rows[i])
			// Non-admins only see their own within the project result.
			if isAdminClaims(claims) || pf.OwnerID == claims.UserID {
				items = append(items, pf)
			}
		}
	case isAdminClaims(claims):
		rows, err := f.Store.ListProjectFiles(r.Context())
		if err != nil {
			f.Logger.Error("failed to list all project files", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list project files", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.ProjectFileInfo, 0, len(rows))
		for i := range rows {
			items = append(items, projectFileInfoFromListRow(&rows[i]))
		}
	default:
		rows, err := f.Store.ListProjectFilesByOwner(r.Context(), claims.UserID)
		if err != nil {
			f.Logger.Error("failed to list project files by owner", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list project files", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.ProjectFileInfo, 0, len(rows))
		for i := range rows {
			items = append(items, projectFileInfoFromOwnerRow(&rows[i]))
		}
	}

	f.Logger.Info("project file list", "user_id", claims.UserID, "count", len(items))
	f.mustWriteResponse(w, models.ListProjectFilesResponse{ProjectFiles: items, Count: len(items)}, map[string]any{"user_id": claims.UserID})
}

// UpdateProjectFile patches the project_files metadata for an existing record.
//
//	@Summary		Update project file metadata
//	@Description	Patches project_files fields (project_name, version_tag, is_read_only, export flags, compression).
//	@Description	Requires write permission on the file.
//	@Tags			project-files
//	@Accept			json
//	@Produce		json
//	@Param			file_uuid	path		string							true	"File UUID"
//	@Param			request		body		models.UpdateProjectFileRequest	true	"Update request"
//	@Success		200			{object}	models.ProjectFileInfo
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/project-files/{file_uuid} [patch]
//	@Security		BearerAuth
func (f *FilestoreHandlers) UpdateProjectFile(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupProjectFileOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromProjectFileRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "write")
	if err != nil {
		f.Logger.Error("project file write permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to update this project file", http.StatusForbidden)
		return
	}

	var req models.UpdateProjectFileRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// merge stuff and keep existing values for unset fields
	newProjectName := row.ProjectName
	if req.ProjectName != "" {
		newProjectName = dbutils.NullString(nullableString(req.ProjectName))
	}

	newVersionTag := row.VersionTag
	if req.VersionTag != "" {
		newVersionTag = dbutils.NullString(nullableString(req.VersionTag))
	}

	newIsReadOnly := row.IsReadOnly
	if req.IsReadOnly != nil {
		newIsReadOnly = dbutils.NullBool(req.IsReadOnly)
	}

	newIncludeSnapshots := row.IncludeSnapshots
	if req.IncludeSnapshots != nil {
		newIncludeSnapshots = *req.IncludeSnapshots
	}

	newIncludeImages := row.IncludeImages
	if req.IncludeImages != nil {
		newIncludeImages = *req.IncludeImages
	}

	newResetMAC := row.ResetMacAddresses
	if req.ResetMacAddresses != nil {
		newResetMAC = *req.ResetMacAddresses
	}

	newKeepComputeIds := row.KeepComputeIds
	if req.KeepComputeIds != nil {
		newKeepComputeIds = *req.KeepComputeIds
	}

	newCompression := row.Compression
	if req.Compression != "" {
		if !slices.Contains(models.ValidCompressionTypes, req.Compression) {
			helpers.WriteAPIError(w,
				fmt.Sprintf("compression must be one of: %v", models.ValidCompressionTypes),
				helpers.ErrCodeInvalidInput, "invalid compression", http.StatusBadRequest)
			return
		}
		newCompression = req.Compression
	}

	if err := f.Store.UpdateProjectFile(r.Context(), sqlc_file_store.UpdateProjectFileParams{
		ProjectName:       newProjectName,
		VersionTag:        newVersionTag,
		IsReadOnly:        newIsReadOnly,
		IncludeSnapshots:  newIncludeSnapshots,
		IncludeImages:     newIncludeImages,
		ResetMacAddresses: newResetMAC,
		KeepComputeIds:    newKeepComputeIds,
		Compression:       newCompression,
		FileUuid:          fileUUID,
	}); err != nil {
		f.Logger.Error("failed to update project file", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to update project file", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("project file updated",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"project_id", row.ProjectID,
	)

	updated, ok := lookupProjectFileOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}
	f.mustWriteResponse(w, projectFileInfoFromRow(&updated), map[string]any{"file_uuid": fileUUID})
}

// DeleteProjectFile deletes a project file and its underlying file/blob.
//
//	@Summary		Delete project file
//	@Description	Deletes the project_files record, its parent file, and the blob when no other references remain.
//	@Tags			project-files
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.DeleteFileResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/project-files/{file_uuid} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DeleteProjectFile(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupProjectFileOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromProjectFileRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "admin")
	if err != nil {
		f.Logger.Error("project file delete permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to delete this project file", http.StatusForbidden)
		return
	}

	blobPath, blobDeleted, ok := deleteFileAndBlob(r, w, f, fileUUID)
	if !ok {
		return
	}

	if blobDeleted && blobPath != "" {
		if rmErr := os.Remove(blobPath); rmErr != nil {
			f.Logger.Warn("failed to remove blob from disk", "err", rmErr, "path", blobPath)
		}
	}

	f.Logger.Info("project file deleted",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"project_id", row.ProjectID,
		"blob_deleted", blobDeleted,
	)

	f.mustWriteResponse(w, models.DeleteFileResponse{
		FileUUID: fileUUID,
		Status:   models.FileStatusDeleted,
	}, map[string]any{"file_uuid": fileUUID, "user_id": claims.UserID, "project_id": row.ProjectID, "blob_deleted": blobDeleted})
}
