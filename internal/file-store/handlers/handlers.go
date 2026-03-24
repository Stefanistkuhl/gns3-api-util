package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	backgroundjobs "github.com/0xveya/gns3util/internal/file-store/backroundjobs"
	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
)

type JobRunnerFunc func(ctx context.Context, invokedBy backgroundjobs.Invocator) (any, error)

type FilestoreHandlers struct {
	Store      *db.Store
	Logger     *slog.Logger
	Dirs       *fs.Dirs
	JobRunners map[string]JobRunnerFunc
}

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

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID

	var req models.InitUploadRequest

	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		f.Logger.Warn("Failed to decode request body", "err", decodeErr, "user_id", userID)
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidJSON, decodeErr.Error(), http.StatusBadRequest)
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

	bucketID := req.BucketID
	if bucketID == "" {
		bucketID = models.GlobalBucketID
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
//	@Param			file_uuid	path		string						true	"File UUID from initialization"
//	@Param			Content-Range	header		string						false	"byte range for resumable upload"
//	@Success		200			{object}	models.FinalizeUploadResponse
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		409			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/content [put]
//	@Security		BearerAuth
func (f *FilestoreHandlers) HandleStreamUpload(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}
	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID
	dbID, getDDIDErr := f.Store.GetOwnerOfFileByUUID(r.Context(), fileUUID)
	if getDDIDErr != nil {
		if errors.Is(getDDIDErr, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "file not found", helpers.ErrCodeFileNotFound, "no file found with the provided uuid", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get owner of file by UUID", "err", getDDIDErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to query db for file info", helpers.ErrCodeDBErr, getDDIDErr.Error(), http.StatusInternalServerError)
		return
	}
	if dbID != userID {
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
//	@Param			file_uuid	path		string							true	"File UUID"
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
	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID
	dbID, getDBIDErr := f.Store.GetOwnerOfFileByUUID(r.Context(), fileUUID)
	if getDBIDErr != nil {
		if errors.Is(getDBIDErr, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "file not found", helpers.ErrCodeFileNotFound, "no file found with the provided uuid", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get owner of file by UUID", "err", getDBIDErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to query db for file info", helpers.ErrCodeDBErr, getDBIDErr.Error(), http.StatusInternalServerError)
		return
	}
	if dbID != userID {
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

// DownloadFileHandler downloads a file from the filestore
//
//	@Summary		Download file
//	@Description	Downloads a file with support for range requests (byte-range downloads)
//	@Tags			files
//	@Produce		application/octet-stream
//	@Param			file_uuid	path		string						true	"File UUID"
//	@Param			Range		header		string						false	"byte range request"
//	@Success		200			{file}		binary
//	@Success		206			{file}		binary
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid} [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt", http.StatusInternalServerError)
		return
	}

	file, fileErr := f.Store.GetFileWithBlobByUUID(r.Context(), fileUUID)
	if fileErr != nil {
		if errors.Is(fileErr, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "file not found", helpers.ErrCodeFileNotFound, "no file found with the provided uuid", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get file", "err", fileErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to query db for file", helpers.ErrCodeDBErr, fileErr.Error(), http.StatusInternalServerError)
		return
	}

	if file.OwnerID != claims.UserID {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to download this file", http.StatusForbidden)
		return
	}

	fileSize := file.SizeBytes

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Filename))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fmt.Sprintf("%q", file.BlobSha256.String))
	w.Header().Set("Cache-Control", "public, max-age=3600")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		ranges, err := parseRange(rangeHeader, fileSize)
		if err != nil {
			helpers.WriteAPIError(w, "invalid range header", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
			return
		}

		if len(ranges) == 1 {
			ra := ranges[0]
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.start+ra.length-1, fileSize))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", ra.length))
			w.WriteHeader(http.StatusPartialContent)

			srcFile, openErr := os.Open(file.FilePath) // #nosec G304
			if openErr != nil {
				f.Logger.Error("Failed to open file", "err", openErr, "file_uuid", fileUUID)
				return

			}
			defer srcFile.Close()

			if _, seekErr := srcFile.Seek(ra.start, io.SeekStart); seekErr != nil {
				f.Logger.Error("Failed to seek file", "err", seekErr, "file_uuid", fileUUID)
				helpers.WriteAPIError(w, "Failed to seek file", helpers.ErrCodeInternal, seekErr.Error(), http.StatusBadRequest)
				return
			}

			if _, copyErr := io.CopyN(w, srcFile, ra.length); copyErr != nil && !errors.Is(copyErr, io.EOF) {
				f.Logger.Error("Failed to copy file", "err", copyErr, "file_uuid", fileUUID)
				return
			}

			f.Logger.Info("Partial file downloaded",
				"file_uuid", fileUUID,
				"user_id", claims.UserID,
				"range", fmt.Sprintf("%d-%d", ra.start, ra.start+ra.length-1),
			)
			return
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		helpers.WriteAPIError(w, "Multiple ranges not supported", helpers.ErrCodeInternal, "Multiple range not supported", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))

	f.Logger.Info("File downloaded",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"size_bytes", fileSize,
	)

	http.ServeFile(w, r, file.FilePath)
}

// CreateBucket creates a new storage bucket
//
//	@Summary		Create bucket
//	@Description	Creates a new bucket for organizing files
//	@Tags			buckets
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.CreateBucketRequest	true	"Bucket creation request"
//	@Success		200		{object}	models.CreateBucketResponse
//	@Failure		400		{object}	helpers.APIErrorResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID

	var req models.CreateBucketRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		f.Logger.Warn("Failed to decode request body", "err", decodeErr, "user_id", userID)
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidJSON, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		helpers.WriteAPIError(w, "name is required", helpers.ErrCodeInvalidInput, "missing required field: name", http.StatusBadRequest)
		return
	}

	bucketUUID, uuidErr := uuid.NewV7()
	if uuidErr != nil {
		f.Logger.Error("Failed to generate UUID", "err", uuidErr, "user_id", userID)
		helpers.WriteAPIError(w, "failed to generate a uuid", helpers.ErrCodeFailedToGenerateUUID, "failed to generate a uuid for the bucket", http.StatusInternalServerError)
		return
	}

	params := sqlc_file_store.CreateBucketParams{
		BucketID:       bucketUUID.String(),
		Name:           req.Name,
		OwnerID:        userID,
		IsPublic:       dbutils.NullBool(&req.IsPublic),
		RequiredScopes: dbutils.NullString(req.RequiredScopes),
	}

	bucket, insertErr := f.Store.CreateBucket(r.Context(), params)
	if insertErr != nil {
		f.Logger.Error("Failed to create bucket", "err", insertErr, "user_id", userID, "name", req.Name)
		helpers.WriteAPIError(w, "failed to insert bucket into db", helpers.ErrCodeDBErr, insertErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("Bucket created",
		"bucket_id", bucket.BucketID,
		"name", bucket.Name,
		"user_id", userID,
	)

	writeErr := helpers.WriteJSON(w, models.CreateBucketResponse{
		BucketID:       bucket.BucketID,
		Name:           bucket.Name,
		IsPublic:       bucket.IsPublic.Bool,
		RequiredScopes: bucket.RequiredScopes.String,
		CreatedAt:      dbutils.ParseDBTime(bucket.CreatedAt),
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "bucket_id", bucket.BucketID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// ListBucketFiles lists all files in a bucket
//
//	@Summary		List bucket files
//	@Description	Returns all files in a specific bucket
//	@Tags			buckets
//	@Produce		json
//	@Param			bucket_id	path		string						true	"Bucket ID"
//	@Success		200			{object}	models.ListBucketFilesResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id}/files [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListBucketFiles(w http.ResponseWriter, r *http.Request) {
	bucketID := chi.URLParam(r, "bucket_id")
	if bucketID == "" {
		helpers.WriteAPIError(w, "bucket_id is required", helpers.ErrCodeInvalidInput, "missing bucket_id in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt", http.StatusInternalServerError)
		return
	}

	files, queryErr := f.Store.ListFilesByBucketWithBlob(r.Context(), bucketID)
	if queryErr != nil {
		f.Logger.Error("Failed to list bucket files", "err", queryErr, "bucket_id", bucketID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to query bucket files", helpers.ErrCodeDBErr, queryErr.Error(), http.StatusInternalServerError)
		return
	}

	fileResponses := make([]models.FileInfo, 0, len(files))
	for i := range files {
		file := files[i]
		fileResponses = append(fileResponses, models.FileInfo{
			FileUUID:    file.FileUuid,
			Filename:    file.Filename,
			SizeBytes:   file.SizeBytes,
			ContentType: file.ContentType,
			BlobSHA256:  file.BlobSha256.String,
			Status:      models.FileStatus(file.Status),
			CreatedAt:   dbutils.ParseDBTime(file.CreatedAt),
		})
	}

	writeErr := helpers.WriteJSON(w, models.ListBucketFilesResponse{BucketUUID: bucketID, Files: fileResponses, Count: len(fileResponses)})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "bucket_id", bucketID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// PublicFileHandler downloads a file using a public access token
//
//	@Summary		Download file with public token
//	@Description	Downloads a file using a public access token (no authentication required)
//	@Tags			public
//	@Produce		application/octet-stream
//	@Param			bucket_id		path		string						true	"Public access token"
//	@Param			token	path		string						true	"File UUID"
//	@Success		200			{file}		binary
//	@Success		206			{file}		binary
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/public/{token}/{file_uuid} [get]
func (f *FilestoreHandlers) PublicFileHandler(w http.ResponseWriter, r *http.Request) {
	bucketID := chi.URLParam(r, "bucket_id")
	token := chi.URLParam(r, "token")

	if bucketID == "" {
		helpers.WriteAPIError(
			w,
			"bucket_id is required",
			helpers.ErrCodeInvalidInput,
			"missing bucket_id in URL path",
			http.StatusBadRequest,
		)
		return
	}

	if token == "" {
		helpers.WriteAPIError(
			w,
			"token is required",
			helpers.ErrCodeInvalidInput,
			"missing token in URL path",
			http.StatusBadRequest,
		)
		return
	}

	publicToken, queryErr := f.Store.GetPublicFileToken(r.Context(), token)
	if queryErr != nil {
		if errors.Is(queryErr, sql.ErrNoRows) {
			helpers.WriteAPIError(
				w,
				"token not found",
				helpers.ErrCodeFileNotFound,
				"invalid or expired token",
				http.StatusNotFound,
			)
			return
		}
		f.Logger.Error("Failed to get public token", "err", queryErr, "token", token)
		helpers.WriteAPIError(
			w,
			"failed to query db for token",
			helpers.ErrCodeDBErr,
			queryErr.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	if publicToken.BucketID != bucketID {
		helpers.WriteAPIError(
			w,
			"token not found",
			helpers.ErrCodeFileNotFound,
			"invalid or expired token",
			http.StatusNotFound,
		)
		return
	}

	if publicToken.ExpiresAt != "" {
		expiresAt := dbutils.ParseDBTime(publicToken.ExpiresAt)
		if !expiresAt.IsZero() && expiresAt.Before(time.Now()) {
			helpers.WriteAPIError(
				w,
				"token expired",
				helpers.ErrCodeFileNotFound,
				"token has expired",
				http.StatusNotFound,
			)
			return
		}
	}

	file, fileErr := f.Store.GetFileWithBlobByUUID(r.Context(), publicToken.FileUuid)
	if fileErr != nil {
		if errors.Is(fileErr, sql.ErrNoRows) {
			helpers.WriteAPIError(
				w,
				"file not found",
				helpers.ErrCodeFileNotFound,
				"file associated with token not found",
				http.StatusNotFound,
			)
			return
		}
		f.Logger.Error("Failed to get file", "err", fileErr, "file_uuid", publicToken.FileUuid)
		helpers.WriteAPIError(
			w,
			"failed to query db for file",
			helpers.ErrCodeDBErr,
			fileErr.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	fileSize := file.SizeBytes

	incrementErr := f.Store.IncrementTokenAccessCount(r.Context(), token)
	if incrementErr != nil {
		f.Logger.Warn("Failed to increment token access count", "err", incrementErr, "token", token)
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Filename))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fmt.Sprintf("%q", file.BlobSha256.String))
	w.Header().Set("Cache-Control", "public, max-age=3600")

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		ranges, err := parseRange(rangeHeader, fileSize)
		if err != nil {
			helpers.WriteAPIError(
				w,
				"invalid range header",
				helpers.ErrCodeInvalidInput,
				err.Error(),
				http.StatusBadRequest,
			)
			return
		}

		if len(ranges) == 1 {
			ra := ranges[0]
			w.Header().Set(
				"Content-Range",
				fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.start+ra.length-1, fileSize),
			)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", ra.length))
			w.WriteHeader(http.StatusPartialContent)

			srcFile, openErr := os.Open(file.FilePath) // #nosec G304
			if openErr != nil {
				f.Logger.Error("Failed to open file", "err", openErr, "file_uuid", publicToken.FileUuid)
				helpers.WriteAPIError(
					w,
					"Failed to open file",
					helpers.ErrCodeInternal,
					openErr.Error(),
					http.StatusBadRequest,
				)
				return
			}
			defer srcFile.Close()

			if _, seekErr := srcFile.Seek(ra.start, io.SeekStart); seekErr != nil {
				f.Logger.Error("Failed to seek file", "err", seekErr, "file_uuid", publicToken.FileUuid)
				helpers.WriteAPIError(
					w,
					"Failed to seek file",
					helpers.ErrCodeInternal,
					seekErr.Error(),
					http.StatusBadRequest,
				)
				return
			}

			if _, copyErr := io.CopyN(w, srcFile, ra.length); copyErr != nil && !errors.Is(copyErr, io.EOF) {
				f.Logger.Error("Failed to copy file", "err", copyErr, "file_uuid", publicToken.FileUuid)
				return
			}

			f.Logger.Info(
				"Partial public file downloaded",
				"file_uuid",
				publicToken.FileUuid,
				"token",
				token,
				"range",
				fmt.Sprintf("%d-%d", ra.start, ra.start+ra.length-1),
			)
			return
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		helpers.WriteAPIError(
			w,
			"Multiple ranges not supported",
			helpers.ErrCodeInternal,
			"Multiple range not supported",
			http.StatusRequestedRangeNotSatisfiable,
		)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))

	f.Logger.Info(
		"Public file accessed",
		"file_uuid",
		file.FileUuid,
		"filename",
		file.Filename,
		"token",
		token,
	)

	http.ServeFile(w, r, file.FilePath)
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

// GeneratePublicToken creates a public access token for a file
//
//	@Summary		Generate public token
//	@Description	Creates a temporary public access token for sharing a file without authentication
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Param			file_uuid	path		string						true	"File UUID"
//	@Param			request		body		models.PublicTokenRequest	true	"Token generation request"
//	@Success		200			{object}	models.PublicTokenResponse
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid}/public-token [post]
//	@Security		BearerAuth
func (f *FilestoreHandlers) GeneratePublicToken(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt", http.StatusInternalServerError)
		return
	}
	var req models.PublicTokenRequest

	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		f.Logger.Warn("Failed to decode request body", "err", decodeErr, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "Invalid request body", helpers.ErrCodeInvalidJSON, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	file, getFileErr := f.Store.GetFileByUUID(r.Context(), fileUUID)
	if getFileErr != nil {
		if errors.Is(getFileErr, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "file not found", helpers.ErrCodeFileNotFound, "no file found with the provided uuid", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get file by UUID", "err", getFileErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to query db for file info", helpers.ErrCodeDBErr, getFileErr.Error(), http.StatusInternalServerError)
		return
	}
	bucket, getBucketErr := f.Store.GetBucketByID(r.Context(), req.BucketUUID)
	if getBucketErr != nil {
		if errors.Is(getBucketErr, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "bucket not found", helpers.ErrCodeFileNotFound, "no bucket found with the provided uuid", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get bucket by UUID", "err", getBucketErr, "bucket_uuid", req.BucketUUID)
		helpers.WriteAPIError(w, "failed to query db for bucket info", helpers.ErrCodeDBErr, getBucketErr.Error(), http.StatusInternalServerError)
		return
	}

	isFileOwner := file.OwnerID == claims.UserID
	isBucketOwner := bucket.OwnerID == claims.UserID
	isGlobalBucket := bucket.BucketID == models.GlobalBucketID

	if !isFileOwner && !isBucketOwner && !isGlobalBucket {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to generate token for this file", http.StatusForbidden)
		return
	}

	if file.Status != string(models.FileStatusAvailable) {
		helpers.WriteAPIError(w, "file not available", helpers.ErrCodeInvalidInput, "cannot generate public token for file that is not available", http.StatusBadRequest)
		return
	}

	tokenStr, uuidErr := uuid.NewV7()

	if uuidErr != nil {
		f.Logger.Error("Failed to generate token UUID", "err", uuidErr, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to generate token", helpers.ErrCodeFailedToGenerateUUID, "failed to generate token", http.StatusInternalServerError)
		return
	}

	expiresAt := ""

	if req.ExpiresAt != nil {
		expiresAt = dbutils.FormatDBTime(*req.ExpiresAt)
	}

	params := sqlc_file_store.CreatePublicFileTokenParams{
		Token:     tokenStr.String(),
		FileUuid:  fileUUID,
		BucketID:  file.BucketID,
		ExpiresAt: expiresAt,
	}

	token, insertErr := f.Store.CreatePublicFileToken(r.Context(), params)
	if insertErr != nil {
		f.Logger.Error("Failed to create public token", "err", insertErr, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to create public token", helpers.ErrCodeDBErr, insertErr.Error(), http.StatusInternalServerError)
		return
	}

	f.Logger.Info("Public token generated",
		"file_uuid", fileUUID,
		"token", token.Token,
		"user_id", claims.UserID,
		"expires_at", expiresAt,
	)

	writeErr := helpers.WriteJSON(w, models.PublicTokenResponse{
		Token:      token.Token,
		BucketUUID: req.BucketUUID,
		FileUUID:   fileUUID,
		ExpiresAt:  expiresAt,
		URL:        fmt.Sprintf("https://%s/api/v1/public/files/%s/%s", nwutils.GetFirstNonLoopbackIP(), req.BucketUUID, token.Token),
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteFile deletes a file from the filestore
//
//	@Summary		Delete file
//	@Description	Deletes a file and marks blob as deleted
//	@Tags			files
//	@Produce		json
//	@Param			file_uuid	path		string						true	"File UUID"
//	@Success		200			{object}	models.DeleteFileResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/files/{file_uuid} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt", http.StatusInternalServerError)
		return
	}
	userID := claims.UserID

	dbID, getOwnerErr := f.Store.GetOwnerOfFileByUUID(r.Context(), fileUUID)
	if getOwnerErr != nil {
		if errors.Is(getOwnerErr, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "file not found", helpers.ErrCodeFileNotFound, "no file found with the provided uuid", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get owner of file by UUID", "err", getOwnerErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to query db for file info", helpers.ErrCodeDBErr, getOwnerErr.Error(), http.StatusInternalServerError)
		return
	}

	if dbID != userID {
		helpers.WriteAPIError(w, "forbidden", helpers.ErrCodeForbidden, "you do not have permission to delete this file", http.StatusForbidden)
		return
	}

	var blobDeleted bool
	var blobPath string
	txErr := f.Store.WithTx(r.Context(), func(q *sqlc_file_store.Queries) error {
		blobInfo, getBlobErr := q.GetBlobByFileUUID(r.Context(), fileUUID)
		if getBlobErr != nil {
			// Pending file (no blob yet) or already deleted by a concurrent request.
			if errors.Is(getBlobErr, sql.ErrNoRows) {
				return q.DeleteFile(r.Context(), fileUUID)
			}
			return fmt.Errorf("failed to get blob info for file deletion: %w", getBlobErr)
		}

		blobPath = blobInfo.FilePath

		// Decrement blob refcount first; only delete blob row when it reaches 0.
		if drefErr := q.DecrementBlobRefCount(r.Context(), blobInfo.BlobSha256.String); drefErr != nil {
			return fmt.Errorf("failed to decrement blob refcount: %w", drefErr)
		}

		updatedBlob, gErr := q.GetBlobBySHA256(r.Context(), blobInfo.BlobSha256.String)
		if gErr != nil {
			return fmt.Errorf("failed to get updated blob info: %w", gErr)
		}

		if updatedBlob.RefCount <= 0 {
			if bDelErr := q.DeleteBlob(r.Context(), blobInfo.BlobSha256.String); bDelErr != nil {
				return fmt.Errorf("failed to delete unreferenced blob: %w", bDelErr)
			}
			blobDeleted = true
		}

		// Finally delete the file row.
		if dErr := q.DeleteFile(r.Context(), fileUUID); dErr != nil {
			return fmt.Errorf("failed to delete file row: %w", dErr)
		}
		return nil
	})
	if txErr != nil {
		f.Logger.Error("Failed to delete file in database", "err", txErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to delete file in database", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return
	}

	if blobDeleted && blobPath != "" {
		if rmErr := os.Remove(blobPath); rmErr != nil {
			f.Logger.Warn("Failed to remove blob file from disk", "err", rmErr, "path", blobPath)
		}
	}

	f.Logger.Info("File deleted",
		"file_uuid", fileUUID,
		"user_id", userID,
		"blob_deleted", blobDeleted,
	)

	writeErr := helpers.WriteJSON(w, models.DeleteFileResponse{
		FileUUID: fileUUID,
		Status:   models.FileStatusDeleted, // Hard delete semantics
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteBucket deletes a bucket and all its files
//
//	@Summary		Delete bucket
//	@Description	Deletes a bucket and all files contained within it
//	@Tags			buckets
//	@Produce		json
//	@Param			bucket_id	path		string						true	"Bucket ID"
//	@Success		200			{object}	models.DeleteBucketResponse
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		403			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets/{bucket_id} [delete]
//	@Security		BearerAuth
func (f *FilestoreHandlers) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucketID := chi.URLParam(r, "bucket_id")
	if bucketID == "" {
		helpers.WriteAPIError(w, "bucket_id is required", helpers.ErrCodeInvalidInput, "", http.StatusBadRequest)
		return
	}

	if bucketID == models.GlobalBucketID {
		helpers.WriteAPIError(w, "Forbidden: Cannot delete system bucket", helpers.ErrCodeForbidden, "The global-default bucket is protected", http.StatusForbidden)
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "Unauthorized", helpers.ErrCodeGetClaims, "", http.StatusUnauthorized)
		return
	}

	bucket, err := f.Store.GetBucketByID(r.Context(), bucketID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.WriteAPIError(w, "Bucket not found", helpers.ErrCodeFileNotFound, "no bucket found with provided id", http.StatusNotFound)
			return
		}
		f.Logger.Error("Failed to get bucket", "bucket_id", bucketID, "err", err)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	if bucket.OwnerID != claims.UserID {
		helpers.WriteAPIError(w, "Forbidden", helpers.ErrCodeForbidden, "you do not have permission to delete this bucket", http.StatusForbidden)
		return
	}

	files, err := f.Store.ListFilesByBucketWithBlob(r.Context(), bucketID)
	if err != nil {
		f.Logger.Error("Failed to list bucket files for deletion", "bucket_id", bucketID, "err", err)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	var blobsToRemove []string
	txErr := f.Store.WithTx(r.Context(), func(q *sqlc_file_store.Queries) error {
		for i := range files {
			if drefErr := q.DecrementBlobRefCount(r.Context(), files[i].BlobSha256.String); drefErr != nil {
				return fmt.Errorf("failed to decrement blob refcount for file %s: %w", files[i].FileUuid, drefErr)
			}

			updatedBlob, gErr := q.GetBlobBySHA256(r.Context(), files[i].BlobSha256.String)
			if gErr != nil {
				return fmt.Errorf("failed to get updated blob info for %s: %w", files[i].BlobSha256.String, gErr)
			}

			if updatedBlob.RefCount <= 0 {
				if bDelErr := q.DeleteBlob(r.Context(), files[i].BlobSha256.String); bDelErr != nil {
					return fmt.Errorf("failed to delete unreferenced blob %s: %w", files[i].BlobSha256.String, bDelErr)
				}
				blobsToRemove = append(blobsToRemove, updatedBlob.FilePath)
			}
		}

		if dErr := q.DeleteBucket(r.Context(), bucketID); dErr != nil {
			return fmt.Errorf("failed to delete bucket: %w", dErr)
		}
		return nil
	})

	if txErr != nil {
		f.Logger.Error("Failed to delete bucket in database", "bucket_id", bucketID, "err", txErr)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, txErr.Error(), http.StatusInternalServerError)
		return
	}

	for _, path := range blobsToRemove {
		if rmErr := os.Remove(path); rmErr != nil {
			f.Logger.Warn("Failed to remove blob file from disk", "err", rmErr, "path", path)
		}
	}

	f.Logger.Info("Bucket deleted", "bucket_id", bucketID, "user_id", claims.UserID)

	writeErr := helpers.WriteJSON(w, models.DeleteBucketResponse{BucketUUID: bucketID, Status: "deleted"})

	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "bucket_uuid", bucketID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// ListBuckets lists all buckets owned by the user
//
//	@Summary		List buckets
//	@Description	Returns all buckets accessible to the authenticated user
//	@Tags			buckets
//	@Produce		json
//	@Success		200		{object}	models.ListBucketResponse
//	@Failure		401		{object}	helpers.APIErrorResponse
//	@Failure		500		{object}	helpers.APIErrorResponse
//	@Router			/api/v1/buckets [get]
//	@Security		BearerAuth
func (f *FilestoreHandlers) ListBuckets(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "Unauthorized", helpers.ErrCodeGetClaims, "", http.StatusUnauthorized)
		return
	}

	buckets, err := f.Store.ListBucketsByOwner(r.Context(), claims.UserID)
	if err != nil {
		f.Logger.Error("Failed to list buckets", "err", err, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "Internal error", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
		return
	}

	globalBucket, err := f.Store.GetBucketByID(r.Context(), models.GlobalBucketID)

	resp := make([]models.CreateBucketResponse, 0)

	if err == nil {
		resp = append(resp, models.CreateBucketResponse{
			BucketID:       globalBucket.BucketID,
			Name:           globalBucket.Name,
			IsPublic:       globalBucket.IsPublic.Bool,
			RequiredScopes: globalBucket.RequiredScopes.String,
			CreatedAt:      dbutils.ParseDBTime(globalBucket.CreatedAt),
		})
	}

	for i := range buckets {
		b := buckets[i]
		if b.BucketID == models.GlobalBucketID {
			continue
		}
		resp = append(resp, models.CreateBucketResponse{
			BucketID:       b.BucketID,
			Name:           b.Name,
			IsPublic:       b.IsPublic.Bool,
			RequiredScopes: b.RequiredScopes.String,
			CreatedAt:      dbutils.ParseDBTime(b.CreatedAt),
		})
	}

	writeErr := helpers.WriteJSON(w, models.ListBucketResponse{Buckets: resp, Count: len(resp)})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}
