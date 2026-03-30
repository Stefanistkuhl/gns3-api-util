package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PublicFileHandler downloads a file using a public access token
//
//	@Summary		Download file with public token
//	@Description	Downloads a file using a public access token (no authentication required)
//	@Tags			public
//	@Produce		application/octet-stream
//	@Param			bucket_id	path		string	true	"Public access token"
//	@Param			token		path		string	true	"File UUID"
//	@Success		200			{file}		binary
//	@Success		206			{file}		binary
//	@Failure		400			{object}	helpers.APIErrorResponse
//	@Failure		401			{object}	helpers.APIErrorResponse
//	@Failure		404			{object}	helpers.APIErrorResponse
//	@Failure		500			{object}	helpers.APIErrorResponse
//	@Router			/public/{bucket_id}/{token} [get]
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

	claims, ok := mustClaims(w, r)
	if !ok {
		return
	}

	var req models.PublicTokenRequest
	if !decodeJSON(w, r, &req) {
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
	if req.BucketUUID != "" && req.BucketUUID != file.BucketID {
		helpers.WriteAPIError(w, "bucket_uuid mismatch", helpers.ErrCodeInvalidInput, "bucket_uuid must match the file's parent bucket", http.StatusBadRequest)
		return
	}

	allowed, err := hasFilePermission(r.Context(), f.Store, claims, &file, "write")
	if err != nil {
		f.Logger.Error("public token permission check failed", "err", err, "file_uuid", fileUUID, "user_id", claims.UserID)
		helpers.WriteAPIError(w, "failed to verify file permissions", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
		return
	}
	if !allowed {
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
		BucketUUID: file.BucketID,
		FileUUID:   fileUUID,
		ExpiresAt:  expiresAt,
		URL:        fmt.Sprintf("https://%s/api/v1/public/files/%s/%s", nwutils.GetFirstNonLoopbackIP(), file.BucketID, token.Token),
	})
	if writeErr != nil {
		f.Logger.Error("Failed to write response", "err", writeErr, "file_uuid", fileUUID)
		helpers.WriteAPIError(w, "failed to write response", helpers.ErrCodeInternal, writeErr.Error(), http.StatusInternalServerError)
		return
	}
}
