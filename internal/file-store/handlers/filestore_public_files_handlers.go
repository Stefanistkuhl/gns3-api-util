package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
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
		f.writeDownloadObjectError(w, ErrorNoBucketID)
		return
	}

	if token == "" {
		f.writeDownloadObjectError(w, ErrorNoPubToken)
		return
	}

	publicToken, queryErr := f.getPublicFiletoken(r.Context(), token)
	if queryErr != nil {
		f.writeDownloadObjectError(w, queryErr)
		return
	}

	if publicToken.BucketID != bucketID {
		f.writeDownloadObjectError(w, ErrorBucketIDMissmatch)
		return
	}

	if publicToken.ExpiresAt != "" {
		expiresAt := dbutils.ParseDBTime(publicToken.ExpiresAt)
		if !expiresAt.IsZero() && expiresAt.Before(time.Now()) {
			f.writeDownloadObjectError(w, ErrorExpiredToken)
			return
		}
	}

	file, fileErr := f.getDownloadObjectByUUID(r.Context(), publicToken.FileUuid)
	if fileErr != nil {
		f.writeDownloadObjectError(w, fileErr)
		return
	}

	fileSize := file.Size
	if file.FileStatus != models.FileStatusAvailable {
		f.writeDownloadObjectError(w, ErrorFileNotAvailable)
		return
	}

	incrementErr := f.Store.IncrementTokenAccessCount(r.Context(), token)
	if incrementErr != nil {
		f.Logger.Warn("Failed to increment token access count", "err", incrementErr, "token", token)
	}

	writeDownloadHeaders(w, file)

	if r.Header.Get("Range") != "" {
		if err := writeRangeObject(w, r, file); err != nil {
			f.writeRangeObjectError(w, err)
			return
		}

		f.Logger.Info("Partial file downloaded",
			"file_uuid", file.Path,
		)

		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))

	f.Logger.Info(
		"Public file accessed",
		"filename",
		file.Filename,
		"token",
		token,
	)

	http.ServeFile(w, r, file.Filename)
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

	f.mustWriteResponse(w, models.PublicTokenResponse{
		Token:      token.Token,
		BucketUUID: file.BucketID,
		FileUUID:   fileUUID,
		ExpiresAt:  expiresAt,
		URL:        fmt.Sprintf("https://%s/api/v1/public/files/%s/%s", nwutils.GetFirstNonLoopbackIP(), file.BucketID, token.Token),
	}, map[string]any{"file_uuid": fileUUID, "token": token.Token, "user_id": claims.UserID, "expires_at": expiresAt})
}
