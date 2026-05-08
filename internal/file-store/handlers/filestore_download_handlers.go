package handlers

import (
	"fmt"
	"net/http"

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

	obj, err := f.getDownloadObjectByUUID(r.Context(), fileUUID)
	if err != nil {
		f.writeDownloadObjectError(w, err)
		return
	}

	writeDownloadHeaders(w, obj)

	if r.Header.Get("Range") != "" {
		if err := writeRangeObject(w, r, obj); err != nil {
			f.writeRangeObjectError(w, err)
			return
		}

		f.Logger.Info("Partial file downloaded",
			"file_uuid", fileUUID,
			"user_id", claims.UserID,
		)

		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", obj.Size))

	f.Logger.Info("File downloaded",
		"file_uuid", fileUUID,
		"user_id", claims.UserID,
		"size_bytes", obj.Size,
	)

	http.ServeFile(w, r, obj.Path)
}
