package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type FilestoreHandlers struct {
	Store      *db.Store
	Logger     *slog.Logger
	Dirs       *fs.Dirs
	JobRunners map[string]JobRunnerFunc
}

// DeleteFile deletes a file from the filestore
//
//	@Summary		Delete file
//	@Description	Deletes a file and marks blob as deleted
//	@Tags			files
//	@Produce		json
//	@Param			file_uuid	path		string	true	"File UUID"
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

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	file, ok := lookupFileOrErr(r.Context(), w, f.Store, fileUUID)
	if !ok {
		return
	}

	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "admin")
	if err != nil {
		f.Logger.Error("file delete permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify file permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
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
		"user_id", claims.UserID,
		"blob_deleted", blobDeleted,
	)

	writeErr := helpers.WriteJSON(w, models.DeleteFileResponse{
		FileUUID: fileUUID,
		Status:   models.FileStatusDeleted,
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}
