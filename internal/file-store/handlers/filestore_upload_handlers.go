package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// HandleInitUpload initializes a new file upload session
//
//	@Summary		Initialize file upload
//	@Description	Creates a new upload session and returns a file UUID and upload URL
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.InitUploadRequest	true	"Upload initialization request"
//	@Success		200		{object}	models.InitUploadResponse
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) HandleInitUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Alt-Svc", `h3=":443"; ma=2592000`)

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}
	userID := claims.UserID

	var req models.InitUploadRequest
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

	fileUUID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		f.Logger.Error("Failed to generate UUID", "err", uuidErr, "user_id", userID)
		helpers.WriteAPIError(w, "failed to generate a uuid", helpers.ErrCodeFailedToGenerateUUID, "somehow failed to generate a uuid for the uploaded file if somehow someone manages to trigger this i am very impressed", http.StatusInternalServerError)
		return
	}

	pathBucketID := chi.URLParam(r, "bucket_id")
	switch {
	case pathBucketID != "" && req.BucketID != "" && req.BucketID != pathBucketID:
		helpers.WriteAPIError(w, "bucket_id mismatch", helpers.ErrCodeInvalidInput, "bucket_id in request body must match the route bucket_id", http.StatusBadRequest)
		return
	case pathBucketID != "":
		req.BucketID = pathBucketID
	case req.BucketID == "":
		req.BucketID = models.GlobalBucketID
	}
	bucketID := req.BucketID

	bucket, ok := lookupBucketOrErr(r.Context(), w, f.Store, bucketID)
	if !ok {
		return
	}

	allowed, err := hasBucketPermission(r.Context(), f.Store, claims, &bucket, "write")
	if err != nil {
		f.Logger.Error("bucket write permission check failed", "err", err, "bucket_id", bucketID, "user_id", userID)
		helpers.WriteAPIError(w, "failed to verify bucket permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to upload to this bucket", http.StatusForbidden)
		return
	}

	params := sqlc_file_store.InitFileParams{
		FileUuid:        fileUUID.String(),
		Filename:        req.Filename,
		ContentType:     req.ContentType,
		OwnerID:         userID,
		BucketID:        bucketID,
		LastAccessedAt:  dbutils.FormatDBTime(time.Now()),
		RetentionPeriod: dbutils.NullInt64(req.RetentionPeriod),
	}

	file, insertErr := f.Store.InitFile(r.Context(), params)
	if insertErr != nil {
		f.Logger.Error("Failed to init file", "err", insertErr, "user_id", userID, "filename", req.Filename)
		helpers.WriteAPIError(w, "failed to insert values into db", helpers.ErrCodeDBErr, insertErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("File initialized",
		"file_uuid", file.FileUuid,
		"filename", file.Filename,
		"user_id", userID,
	)

	expiresAt := time.Now().Add(24 * time.Hour)
	if file.RetentionPeriod.Valid {
		expiresAt = time.Now().Add(time.Duration(file.RetentionPeriod.Int64) * time.Hour)
	}

	status := models.FileStatusPending
	f.mustWriteResponse(w, models.InitUploadResponse{
		FileUUID:  file.FileUuid,
		Status:    status,
		UploadURL: fmt.Sprintf("/api/v1/files/%s/content", file.FileUuid),
		ExpiresAt: expiresAt,
	}, map[string]any{"file_uuid": file.FileUuid, "user_id": userID})
}

// HandleStreamUpload streams file content to the filestore
//
//	@Summary		Stream file upload
//	@Description	Uploads file content with resumable upload support via Content-Range header
//	@Tags			files
//	@Accept			octet-stream
//	@Produce		json
//	@Param			file_uuid		path		string	true	"File UUID from initialization"
//	@Param			Content-Range	header		string	false	"byte range for resumable upload"
//	@Success		200				{object}	models.FinalizeUploadResponse
//	@Failure		400				{object}	helpers.APIErrorResponse
//	@Failure		401				{object}	helpers.APIErrorResponse
//	@Failure		403				{object}	helpers.APIErrorResponse
//	@Failure		404				{object}	helpers.APIErrorResponse
//	@Failure		409				{object}	helpers.APIErrorResponse
//	@Failure		500				{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/content [put]
//	@Security		BearerAuth
func (f *FilestoreHandlers) HandleStreamUpload(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	file, ok := lookupFileOrErr(r.Context(), w, f.Store, fileUUID)
	if !ok {
		return
	}

	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "write")
	if err != nil {
		f.Logger.Error("file write permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify file permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to upload to this file", http.StatusForbidden)
		return
	}
	path := uploadTempPath(f.Dirs.TmpDir, fileUUID)
	offset := getExistingUploadOffset(path)
	if err := validateUploadOffset(r.Header.Get("Content-Range"), offset); err != nil {
		f.writeUploadObjectError(w, err)
		return
	}

	markErr := f.markFileUploading(r.Context(), fileUUID)
	if markErr != nil {
		f.writeUploadObjectError(w, markErr)
	}

	hasher := sha256.New()
	if offset > 0 {
		if err := hashExistingPartialFile(path, hasher); err != nil {
			f.writeUploadObjectError(w, err)
			return
		}
	}

	dst, openErr := openUploadTempFile(path)
	if openErr != nil {
		f.writeUploadObjectError(w, openErr)
	}

	seekErr := seekUploadEnd(dst)
	if seekErr != nil {
		f.writeUploadObjectError(w, seekErr)
	}

	var moved bool
	defer func() {
		// TODO: make this in resuable cleanup function
		closeErr := dst.Close()
		if closeErr != nil {
			f.Logger.Error("Failed to close file", "err", closeErr, "file_uuid", fileUUID)
		}
		if !moved {
			removeErr := os.Remove(path)
			if removeErr != nil {
				f.Logger.Error("Failed to remove temporary file", "err", removeErr, "file_uuid", fileUUID, "path", path)
			}
		}
	}()

	written, streamErr := streamUploadChunk(dst, r.Body, hasher)
	if streamErr != nil {
		f.writeUploadObjectError(w, streamErr)
	}

	syncErr := syncUploadFile(dst)
	if syncErr != nil {
		f.writeUploadObjectError(w, syncErr)
	}

	finalHash := hex.EncodeToString(hasher.Sum(nil))
	totalSize := offset + written

	finalPath, moveErr := f.moveUploadToBlob(path, finalHash)
	if moveErr != nil {
		f.writeUploadObjectError(w, moveErr)
	}
	moved = true

	finalizeErr := f.finalizeUploadDB(r.Context(), fileUUID, finalHash, finalPath, totalSize)
	if finalizeErr != nil {
		f.writeUploadObjectError(w, finalizeErr)
	}

	ret := models.FinalizeUploadResponse{
		FileUUID:   fileUUID,
		Status:     models.FileStatusAvailable,
		BlobSHA256: finalHash,
		SizeBytes:  totalSize,
	}
	f.mustWriteResponse(w, ret, map[string]any{
		"file_uuid": fileUUID,
	})
}

// GetUploadStatus retrieves the current upload progress
//
//	@Summary		Get upload status
//	@Description	Returns the current byte offset for an in-progress upload (for resumable uploads)
//	@Tags			files
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Success		200			{object}	models.GetUploadStatusResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/status [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GetUploadStatus(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}
	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	file, ok := lookupFileOrErr(r.Context(), w, f.Store, fileUUID)
	if !ok {
		return
	}

	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "write")
	if err != nil {
		f.Logger.Error("file status permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify file permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to upload to this file", http.StatusForbidden)
		return
	}

	path := filepath.Join(f.Dirs.TmpDir, fmt.Sprintf("%s.tmp", fileUUID))
	var offset int64
	if info, statErr := os.Stat(path); statErr == nil {
		offset = info.Size()
	}

	f.mustWriteResponse(w, models.GetUploadStatusResponse{
		Offset: offset,
	}, map[string]any{"file_uuid": fileUUID})
}
