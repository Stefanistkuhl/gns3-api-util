package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	writeErr := helpers.WriteJSON(w, models.InitUploadResponse{
		FileUUID:  file.FileUuid,
		Status:    status,
		UploadURL: fmt.Sprintf("/api/v1/files/%s/content", file.FileUuid),
		ExpiresAt: expiresAt,
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", file.FileUuid)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
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
	path := filepath.Join(f.Dirs.TmpDir, fmt.Sprintf("%s.tmp", fileUUID))
	var offset int64
	if info, statErr := os.Stat(path); statErr == nil {
		offset = info.Size()
	}
	if cr := r.Header.Get("Content-Range"); cr != "" {
		parts := strings.Split(cr, " ")
		if len(parts) == 2 && parts[0] == "bytes" {
			rangeParts := strings.Split(parts[1], "-")
			if len(rangeParts) >= 1 {
				clientStart, _ := strconv.ParseInt(rangeParts[0], 10, 64)
				if clientStart != offset {
					helpers.WriteAPIError(w, "upload offset mismatch", helpers.ErrCodeOffsetMismatch,
						fmt.Sprintf("expected start at %d, client sent %d", offset, clientStart),
						http.StatusConflict)
					return
				}
			}
		}
	}
	statusErr := f.Store.UpdateFileStatus(
		r.Context(),
		sqlc_file_store.UpdateFileStatusParams{
			Status:   string(models.FileStatusUploading),
			FileUuid: fileUUID,
		},
	)
	if statusErr != nil {
		f.Logger.Error("Failed to update file status", "err", statusErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to update file status", helpers.ErrCodeInternal, "failed to update file status", http.StatusInternalServerError)
		return
	}
	hasher := sha256.New()
	if offset > 0 {
		existingFile, openErr := os.Open(path) // #nosec G304
		if openErr != nil {
			f.Logger.Error("Failed to open existing partial file", "err", openErr)
			helpers.WriteAPIError(w, "failed to resume upload", helpers.ErrCodeInternal, openErr.Error(), http.StatusInternalServerError)
			return
		}

		if _, resumeCopyErr := io.Copy(hasher, existingFile); resumeCopyErr != nil {
			if closeErr := existingFile.Close(); closeErr != nil {
				f.Logger.Error("Failed to close existing file", "err", closeErr)
			}
			f.Logger.Error("Failed to hash existing partial file", "err", resumeCopyErr)
			helpers.WriteAPIError(w, "failed to hash existing data", helpers.ErrCodeInternal, resumeCopyErr.Error(), http.StatusInternalServerError)
			return
		}
		if closeErr := existingFile.Close(); closeErr != nil {
			f.Logger.Error("Failed to close existing file", "err", closeErr)
		}
	}

	dst, createFileErr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304
	if createFileErr != nil {
		f.Logger.Error("Failed to create file on disk", "err", createFileErr, "file_uuid", fileUUID, "path", path)
		helpers.WriteAPIError(w, "failed to create file on disk", helpers.ErrCodeInternal, createFileErr.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := dst.Seek(0, io.SeekEnd); err != nil {
		if closeErr := dst.Close(); closeErr != nil {
			f.Logger.Error("Failed to close file", "err", closeErr)
		}
		f.Logger.Error("Failed to seek to end", "err", err)
		helpers.WriteAPIError(w, "failed to seek file", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	var moved bool
	defer func() {
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
	teeReader := io.TeeReader(r.Body, hasher)
	written, copyErr := io.Copy(dst, teeReader)
	if copyErr != nil {
		f.Logger.Error("Streaming failed", "err", copyErr)
		helpers.WriteAPIError(w, "upload interrupted", helpers.ErrCodeInternal, copyErr.Error(), http.StatusInternalServerError)
		return
	}
	if syncErr := dst.Sync(); syncErr != nil {
		f.Logger.Error("Failed to sync file", "err", syncErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to sync file", helpers.ErrCodeInternal, syncErr.Error(), http.StatusInternalServerError)
		return
	}
	finalHash := hex.EncodeToString(hasher.Sum(nil))
	totalSize := offset + written

	dstDir, checkDirErr := f.Dirs.CreateShardDirsIfNeed(finalHash)
	if checkDirErr != nil {
		f.Logger.Error("Failed to create shard directories", "err", checkDirErr, "file_uuid", fileUUID, "hash", finalHash)
		helpers.WriteAPIError(w, "failed to create shard directories", helpers.ErrCodeInternal, checkDirErr.Error(), http.StatusInternalServerError)
		return
	}
	finalPath := filepath.Join(dstDir, finalHash)
	moved = true
	renameErr := os.Rename(path, finalPath)
	if renameErr != nil {
		f.Logger.Error("Failed to move file to object store", "err", renameErr, "file_uuid", fileUUID, "hash", finalHash)
		helpers.WriteAPIError(w, "failed to move file to object store", helpers.ErrCodeInternal, renameErr.Error(), http.StatusInternalServerError)
		return
	}

	finalizeErr := f.Store.WithTx(r.Context(), func(q *sqlc_file_store.Queries) error {
		upsertErr := q.UpsertBlob(r.Context(), sqlc_file_store.UpsertBlobParams{
			Sha256:    finalHash,
			FilePath:  finalPath,
			SizeBytes: totalSize,
		})
		if upsertErr != nil {
			return fmt.Errorf("failed to upsert blob: %w", upsertErr)
		}

		_, fErr := q.FinalizeFile(r.Context(), sqlc_file_store.FinalizeFileParams{
			BlobSha256: dbutils.NullString(&finalHash),
			FileUuid:   fileUUID,
		})
		if fErr != nil {
			return fmt.Errorf("failed to finalize file: %w", fErr)
		}
		return nil
	})

	if finalizeErr != nil {
		f.Logger.Error("Failed to finalize upload in database", "err", finalizeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to finalize upload in database", helpers.ErrCodeDBErr, finalizeErr.Error(), http.StatusInternalServerError)
		return
	}

	ret := models.FinalizeUploadResponse{
		FileUUID:   fileUUID,
		Status:     models.FileStatusAvailable,
		BlobSHA256: finalHash,
		SizeBytes:  totalSize,
	}
	writeErr := helpers.WriteJSON(w, ret)
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
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

	writeErr := helpers.WriteJSON(w, models.GetUploadStatusResponse{
		Offset: offset,
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}
