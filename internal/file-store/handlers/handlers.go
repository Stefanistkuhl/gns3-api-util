package handlers

import (
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

	"github.com/go-chi/chi"
	"github.com/google/uuid"

	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
)

type FilestoreHandlers struct {
	Store  *db.Store
	Logger *slog.Logger
	Dirs   *fs.Dirs
}

func (f *FilestoreHandlers) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	writeErr := helpers.WriteJSON(w, map[string]any{
		"status": "ok",
	})
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

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

	params := sqlc_file_store.InitFileParams{
		FileUuid:        fileUUID.String(),
		Filename:        req.Filename,
		SizeBytes:       req.SizeBytes,
		ContentType:     req.ContentType,
		ScopeLabel:      req.ScopeLabel,
		OwnerID:         userID,
		LastAccessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		RetentionPeriod: dbutils.NullInt64(&req.RetentionPeriod),
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
		"size_bytes", file.SizeBytes,
	)

	expiresAt := time.Now().Add(24 * time.Hour)
	if file.RetentionPeriod.Valid {
		expiresAt = time.Now().Add(time.Duration(file.RetentionPeriod.Int64) * time.Hour)
	}

	writeErr := helpers.WriteJSON(w, models.InitUploadResponse{
		FileUUID:  file.FileUuid,
		Status:    file.Status,
		UploadURL: fmt.Sprintf("/api/v1/files/%s/content", file.FileUuid),
		ExpiresAt: expiresAt,
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", file.FileUuid)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

func (f *FilestoreHandlers) HandleStreamUpload(w http.ResponseWriter, r *http.Request) {
	fileUUID := chi.URLParam(r, "file_uuid")
	if fileUUID == "" {
		helpers.WriteAPIError(w, "file_uuid is required", helpers.ErrCodeInvalidInput, "missing file_uuid in URL path", http.StatusBadRequest)
		return
	}
	claims, ok := middleware.GetClaims(r)
	if !ok {
		helpers.WriteAPIError(w, "failed to get claims from jwt", helpers.ErrCodeGetClaims, "failed to get claims from jwt even though this is past middleware and shouldn't happen", http.StatusInternalServerError)
		http.Error(w, "Failed to get claims from JWT", http.StatusInternalServerError)
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

	params := sqlc_file_store.FinalizeFileParams{
		ChecksumSha256: finalHash,
		FileUuid:       fileUUID,
		SizeBytes:      totalSize,
		FilePath:       finalPath,
	}
	_, finalizeErr := f.Store.FinalizeFile(r.Context(), params)
	if finalizeErr != nil {
		f.Logger.Error("Failed to finalize file in database", "err", finalizeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to finalize file in database", helpers.ErrCodeDBErr, finalizeErr.Error(), http.StatusInternalServerError)
		return
	}
	ret := models.FinalizeUploadResponse{
		FileUUID:       fileUUID,
		Status:         models.FileStatusAvailable,
		ChecksumSHA256: finalHash,
		SizeBytes:      totalSize,
	}
	writeErr := helpers.WriteJSON(w, ret)
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}

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

func (f *FilestoreHandlers) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	writeErr := helpers.WriteJSON(w, map[string]any{
		"status": "ok",
	})
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}
