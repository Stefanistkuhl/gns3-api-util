package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func (f *FilestoreHandlers) getDownloadObjectByUUID(ctx context.Context, fileUUID string) (*DownloadObject, error) {
	file, err := f.Store.GetFileWithBlobByUUID(ctx, fileUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("db error: %w", err)
	}

	if file.Status != string(models.FileStatusAvailable) {
		return nil, ErrFileUnavailable
	}

	return &DownloadObject{
		Path:        file.FilePath,
		Size:        file.SizeBytes,
		ContentType: file.ContentType,
		Filename:    file.Filename,
		ETag:        file.BlobSha256.String,
		FileStatus:  models.FileStatus(file.Status),
	}, nil
}

func (f *FilestoreHandlers) getPublicFiletoken(ctx context.Context, token string) (*sqlc_file_store.PublicFileToken, error) {
	pubTok, queryErr := f.Store.GetPublicFileToken(ctx, token)

	if queryErr != nil {
		if errors.Is(queryErr, sql.ErrNoRows) {
			return nil, ErrorTokenNotFound
		}
		f.Logger.Error("Failed to get public token", "err", queryErr, "token", token)
		return nil, ErrorFailedToQueryDB
	}
	return &pubTok, nil
}

func (f *FilestoreHandlers) writeDownloadObjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrFileNotFound):
		helpers.WriteAPIError(w, "file not found", helpers.ErrCodeFileNotFound, err.Error(), http.StatusNotFound)

	case errors.Is(err, ErrFileUnavailable):
		helpers.WriteAPIError(w, "file not available", helpers.ErrCodeInvalidInput, err.Error(), http.StatusConflict)
	case errors.Is(ErrorNoBucketID, err):
		helpers.WriteAPIError(w, "bucket_id is required", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
	case errors.Is(ErrorNoPubToken, err):
		helpers.WriteAPIError(w, "token is required", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
	case errors.Is(ErrorTokenNotFound, err):
		helpers.WriteAPIError(w, "token not found", helpers.ErrCodeFileNotFound, err.Error(), http.StatusNotFound)
	case errors.Is(ErrorFailedToQueryDB, err):
		helpers.WriteAPIError(w, "failed to query db", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)
	case errors.Is(ErrorBucketIDMissmatch, err):
		helpers.WriteAPIError(w, "bucket_id mismatch", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
	case errors.Is(ErrorExpiredToken, err):
		helpers.WriteAPIError(w, "token expired", helpers.ErrCodeFileNotFound, err.Error(), http.StatusNotFound)

	case errors.Is(ErrorFileNotAvailable, err):
		helpers.WriteAPIError(w, "file not available", helpers.ErrCodeInvalidInput, err.Error(), http.StatusNotFound)

	default:
		f.Logger.Error("unexpected error", "err", err)
		helpers.WriteAPIError(w, "internal error", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
	}
}
