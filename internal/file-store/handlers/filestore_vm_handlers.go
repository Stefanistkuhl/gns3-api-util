package handlers

import (
	"database/sql"
	"errors"
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

// vmInfoFromRow converts a joined VM+file row into the API response type.
func vmInfoFromRow(r *sqlc_file_store.GetVMImageWithFileRow) models.VMImageInfo {
	return models.VMImageInfo{
		FileUUID:            r.FileUuid,
		Filename:            r.Filename,
		ContentType:         r.ContentType,
		OwnerID:             r.OwnerID,
		BucketID:            r.BucketID,
		SizeBytes:           r.SizeBytes.Int64,
		BlobSHA256:          r.BlobSha256.String,
		Status:              models.FileStatus(r.Status),
		CreatedAt:           dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:           dbutils.ParseDBTime(r.UpdatedAt),
		VirtType:            r.VirtType,
		Format:              r.Format,
		VCPUs:               r.Vcpus.Int64,
		RAMMB:               r.RamMb.Int64,
		ExtraAttributesJSON: r.ExtraAttributesJson.String,
	}
}

// vmInfoFromListRow converts a list row (same shape, different type alias) into
// the API response type.
func vmInfoFromListRow(r *sqlc_file_store.ListVMImagesRow) models.VMImageInfo {
	return models.VMImageInfo{
		FileUUID:            r.FileUuid,
		Filename:            r.Filename,
		ContentType:         r.ContentType,
		OwnerID:             r.OwnerID,
		BucketID:            r.BucketID,
		SizeBytes:           r.SizeBytes.Int64,
		BlobSHA256:          r.BlobSha256.String,
		Status:              models.FileStatus(r.Status),
		CreatedAt:           dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:           dbutils.ParseDBTime(r.UpdatedAt),
		VirtType:            r.VirtType,
		Format:              r.Format,
		VCPUs:               r.Vcpus.Int64,
		RAMMB:               r.RamMb.Int64,
		ExtraAttributesJSON: r.ExtraAttributesJson.String,
	}
}

// vmInfoFromOwnerRow converts an owner-filtered list row.
func vmInfoFromOwnerRow(r *sqlc_file_store.ListVMImagesByOwnerRow) models.VMImageInfo {
	return models.VMImageInfo{
		FileUUID:            r.FileUuid,
		Filename:            r.Filename,
		ContentType:         r.ContentType,
		OwnerID:             r.OwnerID,
		BucketID:            r.BucketID,
		SizeBytes:           r.SizeBytes.Int64,
		BlobSHA256:          r.BlobSha256.String,
		Status:              models.FileStatus(r.Status),
		CreatedAt:           dbutils.ParseDBTime(r.CreatedAt),
		UpdatedAt:           dbutils.ParseDBTime(r.UpdatedAt),
		VirtType:            r.VirtType,
		Format:              r.Format,
		VCPUs:               r.Vcpus.Int64,
		RAMMB:               r.RamMb.Int64,
		ExtraAttributesJSON: r.ExtraAttributesJson.String,
	}
}

// lookupVMOrErr fetches the combined VM+file row and writes an appropriate HTTP
// error on failure. The second return value reports success.
func lookupVMOrErr(
	w http.ResponseWriter,
	r *http.Request,
	f *FilestoreHandlers,
	fileUUID string,
) (sqlc_file_store.GetVMImageWithFileRow, bool) {
	row, err := f.Store.GetVMImageWithFile(r.Context(), fileUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "VM image not found", helpers.ErrCodeFileNotFound, "", http.StatusNotFound)
		} else {
			helpers.WriteAPIError(w, "failed to look up VM image", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		}
		return sqlc_file_store.GetVMImageWithFileRow{}, false
	}
	return row, true
}

// fileFromVMRow constructs a minimal sqlc_file_store.File for permission helpers
// so we can reuse hasFilePermission without an extra DB round-trip.
func fileFromVMRow(row *sqlc_file_store.GetVMImageWithFileRow) sqlc_file_store.File {
	return sqlc_file_store.File{
		FileUuid:   row.FileUuid,
		OwnerID:    row.OwnerID,
		BucketID:   row.BucketID,
		BlobSha256: row.BlobSha256,
	}
}

// ---- handlers ----

// InitVMUpload initialises a VM image upload.
//
//	@Summary		Initialise VM image upload
//	@Description	Creates a file record and a vm_images metadata row atomically, returning a file UUID and upload URL.
//	@Tags			vms
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.InitVMUploadRequest	true	"VM upload request"
//	@Success		200		{object}	models.InitUploadResponse
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		403		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/vms [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) InitVMUpload(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	var req models.InitVMUploadRequest
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
	if !slices.Contains(models.ValidVirtTypes, req.VirtType) {
		helpers.WriteAPIError(w, fmt.Sprintf("virt_type must be one of: %v", models.ValidVirtTypes), helpers.ErrCodeInvalidInput, "invalid virt_type", http.StatusBadRequest)
		return
	}
	if req.Format == "" {
		helpers.WriteAPIError(w, "format is required", helpers.ErrCodeInvalidInput, "missing required field: format", http.StatusBadRequest)
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

		return q.InsertVMImage(r.Context(), sqlc_file_store.InsertVMImageParams{
			FileUuid:            fileUUID.String(),
			VirtType:            req.VirtType,
			Format:              req.Format,
			Vcpus:               dbutils.NullInt64(req.VCPUs),
			RamMb:               dbutils.NullInt64(req.RAMMB),
			ExtraAttributesJson: dbutils.NullString(nullableString(req.ExtraAttributesJSON)),
		})
	})
	if txErr != nil {
		f.Logger.Error("failed to initialise VM image", "err", txErr, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to initialise VM image", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("VM image initialised",
		"file_uuid", file.FileUuid,
		"filename", file.Filename,
		"virt_type", req.VirtType,
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
	}, map[string]any{"file_uuid": file.FileUuid, "user_id": claims.UserID})
}

// GetVMImage returns combined file and VM metadata for a single VM image.
//
//	@Summary		Get VM image
//	@Description	Returns the combined file and vm_images metadata for the requested UUID.
//	@Tags			vms
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.VMImageInfo
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/vms/{file_uuid} [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GetVMImage(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupVMOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromVMRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "read")
	if err != nil {
		f.Logger.Error("VM image read permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to read this VM image", http.StatusForbidden)
		return
	}

	f.Logger.Info("VM image read", "file_uuid", fileUUID, "user_id", claims.UserID)
	f.mustWriteResponse(w, vmInfoFromRow(&row), map[string]any{"file_uuid": fileUUID})
}

// ListVMImages lists VM images accessible to the authenticated user.
//
//	@Summary		List VM images
//	@Description	Returns all VM images owned by the caller. Admins see all VM images.
//	@Tags			vms
//	@Produce		json
//	@Success		200	{object}	models.ListVMImagesResponse
//	@Failure		401	{object}	helpers.APIErrorResponse
//	@Failure		500	{object}	helpers.APIErrorResponse
//	@Router			/api/v1/vms [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListVMImages(w http.ResponseWriter, r *http.Request) {
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	var items []models.VMImageInfo

	if isAdminClaims(claims) {
		rows, err := f.Store.ListVMImages(r.Context())
		if err != nil {
			f.Logger.Error("failed to list all VM images", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list VM images", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.VMImageInfo, 0, len(rows))
		for i := range rows {
			items = append(items, vmInfoFromListRow(&rows[i]))
		}
	} else {
		rows, err := f.Store.ListVMImagesByOwner(r.Context(), claims.UserID)
		if err != nil {
			f.Logger.Error("failed to list VM images by owner", "err", err, "user_id", claims.UserID)
			helpers.WriteAPIError(w, "failed to list VM images", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
			return
		}
		items = make([]models.VMImageInfo, 0, len(rows))
		for i := range rows {
			items = append(items, vmInfoFromOwnerRow(&rows[i]))
		}
	}

	f.Logger.Info("VM image list", "user_id", claims.UserID, "count", len(items))
	f.mustWriteResponse(w, models.ListVMImagesResponse{VMs: items, Count: len(items)}, map[string]any{"user_id": claims.UserID})
}

// UpdateVMImage updates the vm_images metadata for an existing VM image.
//
//	@Summary		Update VM image metadata
//	@Description	Patches vm_images fields (virt_type, format, vcpus, ram_mb, extra_attributes_json). Requires write permission on the file.
//	@Tags			vms
//	@Accept			json
//	@Produce		json
//	@Param			file_uuid	path		string						true	"File UUID"
//	@Param			request		body		models.UpdateVMImageRequest	true	"Update request"
//	@Success		200			{object}	models.VMImageInfo
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/vms/{file_uuid} [patch]
//	@Security		BearerAuth
func (f *FilestoreHandlers) UpdateVMImage(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupVMOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromVMRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "write")
	if err != nil {
		f.Logger.Error("VM image write permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to update this VM image", http.StatusForbidden)
		return
	}

	var req models.UpdateVMImageRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// Merge: keep existing values for any field the caller left blank.
	newVirtType := row.VirtType
	if req.VirtType != "" {
		if !slices.Contains(models.ValidVirtTypes, req.VirtType) {
			helpers.WriteAPIError(w, fmt.Sprintf("virt_type must be one of: %v", models.ValidVirtTypes), helpers.ErrCodeInvalidInput, "invalid virt_type", http.StatusBadRequest)
			return
		}
		newVirtType = req.VirtType
	}

	newFormat := row.Format
	if req.Format != "" {
		newFormat = req.Format
	}

	newVCPUs := row.Vcpus
	if req.VCPUs != nil {
		newVCPUs = dbutils.NullInt64(req.VCPUs)
	}

	newRAMMB := row.RamMb
	if req.RAMMB != nil {
		newRAMMB = dbutils.NullInt64(req.RAMMB)
	}

	newExtra := row.ExtraAttributesJson
	if req.ExtraAttributesJSON != "" {
		newExtra = dbutils.NullString(nullableString(req.ExtraAttributesJSON))
	}

	if err := f.Store.UpdateVMImage(r.Context(), sqlc_file_store.UpdateVMImageParams{
		VirtType:            newVirtType,
		Format:              newFormat,
		Vcpus:               newVCPUs,
		RamMb:               newRAMMB,
		ExtraAttributesJson: newExtra,
		FileUuid:            fileUUID,
	}); err != nil {
		f.Logger.Error("failed to update VM image", "err", err, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to update VM image", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("VM image updated",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"virt_type", newVirtType,
	)

	// Re-fetch so the response reflects the authoritative DB state.
	updated, ok := lookupVMOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}
	f.mustWriteResponse(w, vmInfoFromRow(&updated), map[string]any{"file_uuid": fileUUID})
}

// DeleteVMImage deletes a VM image and its underlying file/blob.
//
//	@Summary		Delete VM image
//	@Description	Deletes the VM image record, its parent file, and the blob when no other references remain.
//	@Tags			vms
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.DeleteFileResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/vms/{file_uuid} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DeleteVMImage(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	row, ok := lookupVMOrErr(w, r, f, fileUUID)
	if !ok {
		return
	}

	file := fileFromVMRow(&row)
	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "admin")
	if err != nil {
		f.Logger.Error("VM image delete permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to delete this VM image", http.StatusForbidden)
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

	f.Logger.Info("VM image deleted",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"blob_deleted", blobDeleted,
	)

	f.mustWriteResponse(w, models.DeleteFileResponse{
		FileUUID: fileUUID,
		Status:   models.FileStatusDeleted,
	}, map[string]any{"file_uuid": fileUUID, "user_id": claims.UserID, "blob_deleted": blobDeleted})
}

// ---- internal shared deletion helper ----

// deleteFileAndBlob performs the transactional file+blob cleanup used by both
// DeleteVMImage and DeleteBackup. It returns (blobPath, blobDeleted, ok); on
// any DB error it writes the HTTP error itself and returns ("", false, false).
// A pending file with no blob yet is handled correctly: ok=true, blobDeleted=false.
func deleteFileAndBlob(
	r *http.Request,
	w http.ResponseWriter,
	f *FilestoreHandlers,
	fileUUID string,
) (blobPath string, blobDeleted, ok bool) {
	txErr := f.Store.WithTx(r.Context(), func(q *sqlc_file_store.Queries) error {
		blobInfo, getBlobErr := q.GetBlobByFileUUID(r.Context(), fileUUID)
		if getBlobErr != nil {
			if errors.Is(getBlobErr, sql.ErrNoRows) {
				// Pending file with no blob yet — just delete the file row.
				return q.DeleteFile(r.Context(), fileUUID)
			}
			return fmt.Errorf("get blob info: %w", getBlobErr)
		}

		blobPath = blobInfo.FilePath

		if drefErr := q.DecrementBlobRefCount(r.Context(), blobInfo.BlobSha256.String); drefErr != nil {
			return fmt.Errorf("decrement blob refcount: %w", drefErr)
		}

		updated, gErr := q.GetBlobBySHA256(r.Context(), blobInfo.BlobSha256.String)
		if gErr != nil {
			return fmt.Errorf("re-fetch blob: %w", gErr)
		}

		if updated.RefCount <= 0 {
			if bDelErr := q.DeleteBlob(r.Context(), blobInfo.BlobSha256.String); bDelErr != nil {
				return fmt.Errorf("delete unreferenced blob: %w", bDelErr)
			}
			blobDeleted = true
		}

		if dErr := q.DeleteFile(r.Context(), fileUUID); dErr != nil {
			return fmt.Errorf("delete file row: %w", dErr)
		}
		return nil
	})
	if txErr != nil {
		f.Logger.Error("failed to delete file in DB", "err", txErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to delete file", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return "", false, false
	}
	return blobPath, blobDeleted, true
}

// nullableString returns nil when s is empty, otherwise a pointer to s.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
