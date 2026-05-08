package handlers

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func uploadTempPath(tmpDir, fileUUID string) string {
	return filepath.Join(tmpDir, fmt.Sprintf("%s.tmp", fileUUID))
}

func getExistingUploadOffset(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return info.Size()
}

func validateUploadOffset(contentRange string, expectedOffset int64) error {
	if contentRange == "" {
		return nil
	}

	parts := strings.Split(contentRange, " ")
	if len(parts) != 2 || parts[0] != "bytes" {
		return ErrInvalidContentRange
	}

	rangeParts := strings.Split(parts[1], "-")
	if len(rangeParts) < 1 {
		return ErrInvalidContentRange
	}

	clientStart, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return ErrInvalidContentRange
	}

	if clientStart != expectedOffset {
		return fmt.Errorf("%w: expected start at %d, client sent %d",
			ErrUploadOffsetMismatch,
			expectedOffset,
			clientStart,
		)
	}

	return nil
}

func hashExistingPartialFile(path string, hasher hash.Hash) error {
	existingFile, openErr := os.Open(path) // #nosec G304
	if openErr != nil {
		return fmt.Errorf("%w: %w", ErrOpenUploadTempFile, openErr)
	}

	if _, resumeCopyErr := io.Copy(hasher, existingFile); resumeCopyErr != nil {
		if closeErr := existingFile.Close(); closeErr != nil {
			return ErrorCloseUploadTempFile
		}
		return ErrHashExistingPartialFile
	}
	if closeErr := existingFile.Close(); closeErr != nil {
		return ErrorCloseUploadTempFile
	}
	return nil
}

func openUploadTempFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenUploadTempFile, err)
	}
	return file, nil
}

func seekUploadEnd(file *os.File) error {
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("%w: %w", ErrSeekUploadTempFile, err)
	}
	return nil
}

func streamUploadChunk(dst io.Writer, body io.Reader, hasher hash.Hash) (int64, error) {
	teeReader := io.TeeReader(body, hasher)
	written, copyErr := io.Copy(dst, teeReader)
	if copyErr != nil {
		return 0, fmt.Errorf("%w: %w", ErrStreamUploadChunk, copyErr)
	}
	return written, nil
}

func syncUploadFile(file *os.File) error {
	syncErr := file.Sync()
	if syncErr != nil {
		return fmt.Errorf("%w: %w", ErrSyncUploadTempFile, syncErr)
	}
	return nil
}

func (f *FilestoreHandlers) markFileUploading(ctx context.Context, fileUUID string) error {
	statusErr := f.Store.UpdateFileStatus(
		ctx,
		sqlc_file_store.UpdateFileStatusParams{
			Status:   string(models.FileStatusUploading),
			FileUuid: fileUUID,
		},
	)
	if statusErr != nil {
		return fmt.Errorf("%w: %w", ErrMarkFileUploading, statusErr)
	}
	return nil
}

func (f *FilestoreHandlers) moveUploadToBlob(tmpPath, finalHash string) (string, error) {
	dstDir, checkDirErr := f.Dirs.CreateShardDirsIfNeed(finalHash)
	if checkDirErr != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateShardDirs, checkDirErr)
	}
	finalPath := filepath.Join(dstDir, finalHash)
	renameErr := os.Rename(tmpPath, finalPath)
	if renameErr != nil {
		return "", fmt.Errorf("%w: %w", ErrMoveUploadToBlob, renameErr)
	}
	return finalPath, nil
}

func (f *FilestoreHandlers) finalizeUploadDB(
	ctx context.Context,
	fileUUID string,
	finalHash string,
	finalPath string,
	totalSize int64,
) error {
	finalizeErr := f.Store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
		upsertErr := q.UpsertBlob(ctx, sqlc_file_store.UpsertBlobParams{
			Sha256:    finalHash,
			FilePath:  finalPath,
			SizeBytes: totalSize,
		})
		if upsertErr != nil {
			return fmt.Errorf("%w failed to upsert blob: %w", ErrFinalizeUploadDB, upsertErr)
		}

		_, fErr := q.FinalizeFile(ctx, sqlc_file_store.FinalizeFileParams{
			BlobSha256: dbutils.NullString(&finalHash),
			FileUuid:   fileUUID,
		})
		if fErr != nil {
			return fmt.Errorf("%w failed to finalize file: %w", ErrFinalizeUploadDB, fErr)
		}
		return nil
	})
	return finalizeErr
}

func (f *FilestoreHandlers) writeUploadObjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUploadOffsetMismatch):
		// 409 Conflict: client upload offset does not match the existing upload offset.
		helpers.WriteAPIError(w, "upload offset mismatch", helpers.ErrCodeOffsetMismatch, err.Error(), http.StatusConflict)

	case errors.Is(err, ErrInvalidContentRange):
		// 400 Bad Request: Content-Range header is missing, malformed, or inconsistent.
		helpers.WriteAPIError(w, "invalid content range", helpers.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)

	case errors.Is(err, ErrFinalizeUploadDB),
		errors.Is(err, ErrMarkFileUploading):
		// 500 Internal Server Error: database/state update failed.
		helpers.WriteAPIError(w, "upload database error", helpers.ErrCodeDBErr, err.Error(), http.StatusInternalServerError)

	case errors.Is(err, ErrHashExistingPartialFile),
		errors.Is(err, ErrOpenUploadTempFile),
		errors.Is(err, ErrSeekUploadTempFile),
		errors.Is(err, ErrStreamUploadChunk),
		errors.Is(err, ErrSyncUploadTempFile),
		errors.Is(err, ErrCreateShardDirs),
		errors.Is(err, ErrMoveUploadToBlob),
		errors.Is(err, ErrRemoveUploadTempFile),
		errors.Is(err, ErrorCloseUploadTempFile),
		errors.Is(err, ErrorUpsertBlob):
		// 500 Internal Server Error: filesystem, hashing, streaming, or blob storage operation failed.
		helpers.WriteAPIError(w, "upload failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)

	default:
		// 500 Internal Server Error: unexpected upload failure.
		helpers.WriteAPIError(w, "upload failed", helpers.ErrCodeInternal, err.Error(), http.StatusInternalServerError)
	}
}
