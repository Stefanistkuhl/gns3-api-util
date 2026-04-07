package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/go-chi/chi/v5"
)

// DownloadFileHandler downloads a file from the filestore
//
//	@Summary		Download file
//	@Description	Downloads a file with support for range requests (byte-range downloads)
//	@Tags			files
//	@Produce		application/octet-stream
//	@Param			file_uuid	path		string	true	"File UUID"
//	@Param			Range		header		string	false	"byte range request"
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

	claims, ok := mustClaims(w, r)
	if !ok {
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

	fileSize := file.SizeBytes
	if file.Status != string(models.FileStatusAvailable) {
		helpers.WriteAPIError(w, "file not available", helpers.ErrCodeInvalidInput, "file is not available for download", http.StatusConflict)
		return
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
