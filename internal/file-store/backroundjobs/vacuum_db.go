package backgroundjobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (j *DBVacuumJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()

	tracer := otel.Tracer("backgroundjobs")

	for {
		select {
		case <-ctx.Done():
			j.Logger.Info("vacuum job stopped")
			return
		case <-ticker.C:
			runCtx, span := tracer.Start(ctx, "DBVacuumJob.Iteration")

			_, err := j.ExecuteIteration(runCtx, InvocatorInterval)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
		}
	}
}

func (j *DBVacuumJob) ExecuteIteration(ctx context.Context, invokedBy Invocator) (*VacuumIterationResult, error) {
	result := &VacuumIterationResult{}

	result.TombstoneExpired = j.tombstoneExpiredFiles(ctx, invokedBy)

	result.CleanupTombstoned = j.cleanupFilesWithStatus(
		ctx,
		models.FileStatusTombstoned,
		j.TombstoneRemoveInterval,
		invokedBy,
	)

	result.CleanupPending = j.cleanupFilesWithStatus(
		ctx,
		models.FileStatusPending,
		j.PendingFileRemoveInterval,
		invokedBy,
	)

	result.CleanupStalledUploads = j.cleanupStalledUploads(ctx, invokedBy)

	result.CleanupOrphanedBlobs = j.cleanupOrphanedBlobs(ctx, invokedBy)

	result.DeleteExpiredTokens = j.deleteExpiredTokens(ctx, invokedBy)

	result.DeleteOldTmpFiles = j.deleteOldTmpFiles(ctx, invokedBy)

	return result, nil
}

func errStr(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}

func (j *DBVacuumJob) cleanupFilesWithStatus(
	ctx context.Context,
	status models.FileStatus,
	olderThan time.Duration,
	invokedBy Invocator,
) CleanupFilesWithStatusResult {
	result := CleanupFilesWithStatusResult{
		Success:   true,
		InvokedBy: invokedBy,
	}

	ctx, span := otel.Tracer("backgroundjobs").Start(ctx, "cleanupFilesWithStatus")
	defer span.End()
	span.SetAttributes(attribute.String("file.status", string(status)))

	files, err := j.Store.GetFilesWithStatus(ctx, string(status))
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result
	}

	for i := range files {
		updatedAt := dbutils.ParseDBTime(files[i].UpdatedAt)
		if time.Since(updatedAt) < olderThan {
			continue
		}

		var blobPath string
		var blobDeleted bool

		err := j.Store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
			if files[i].BlobSha256.Valid {
				blobInfo, getBlobErr := q.GetBlobBySHA256(ctx, files[i].BlobSha256.String)
				if getBlobErr != nil && !errors.Is(getBlobErr, sql.ErrNoRows) {
					return fmt.Errorf(
						"failed to get blob info for file %s: %w",
						files[i].FileUuid,
						getBlobErr,
					)
				}

				if getBlobErr == nil {
					blobPath = blobInfo.FilePath

					if drefErr := q.DecrementBlobRefCount(
						ctx,
						files[i].BlobSha256.String,
					); drefErr != nil {
						return fmt.Errorf(
							"failed to decrement blob refcount: %w",
							drefErr,
						)
					}

					updatedBlob, updErr := q.GetBlobBySHA256(ctx, files[i].BlobSha256.String)
					if updErr != nil {
						return fmt.Errorf(
							"failed to re-read blob after decrement: %w",
							updErr,
						)
					}

					if updatedBlob.RefCount <= 0 {
						if delErr := q.DeleteBlob(ctx, files[i].BlobSha256.String); delErr != nil {
							return fmt.Errorf(
								"failed to delete unreferenced blob: %w",
								delErr,
							)
						}
						blobDeleted = true
					}
				}
			}

			return q.DeleteFile(ctx, files[i].FileUuid)
		})
		if err != nil {
			j.Logger.ErrorContext(
				ctx,
				"failed to cleanup file",
				"file_uuid",
				files[i].FileUuid,
				"status",
				status,
				"err",
				err,
			)
			continue
		}

		result.DeletedFiles = append(result.DeletedFiles, files[i].FileUuid)

		if files[i].BlobSha256.Valid {
			blob := Blob{
				UUID:       files[i].BlobSha256.String,
				FilePath:   blobPath,
				WasDeleted: blobDeleted,
			}
			result.Blobs = append(result.Blobs, blob)
		}

		if blobDeleted && blobPath != "" {
			if rmErr := os.Remove(blobPath); rmErr != nil && !os.IsNotExist(rmErr) {
				j.Logger.ErrorContext(
					ctx,
					"failed to remove blob file",
					"file_uuid",
					files[i].FileUuid,
					"path",
					blobPath,
					"err",
					rmErr,
				)
			}
		}
	}

	return result
}

func (j *DBVacuumJob) cleanupStalledUploads(ctx context.Context, invokedBy Invocator) CleanupStalledUploadsResults {
	result := CleanupStalledUploadsResults{
		Success:   true,
		InvokedBy: invokedBy,
	}

	files, err := j.Store.GetFilesWithStatus(ctx, string(models.FileStatusUploading))
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result
	}

	for i := range files {
		updatedAt := dbutils.ParseDBTime(files[i].UpdatedAt)
		if time.Since(updatedAt) < j.NoWriteSinceUploadInterval {
			continue
		}

		tmpPath := filepath.Join(j.Dirs.TmpDir, fmt.Sprintf("%s.tmp", files[i].FileUuid))
		stat, statErr := os.Stat(tmpPath)
		if statErr == nil && time.Since(stat.ModTime()) < j.NoWriteSinceUploadInterval {
			continue
		}

		var blobPath string
		var blobDeleted bool

		err := j.Store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
			if files[i].BlobSha256.Valid {
				blobInfo, getBlobErr := q.GetBlobBySHA256(ctx, files[i].BlobSha256.String)
				if getBlobErr != nil && !errors.Is(getBlobErr, sql.ErrNoRows) {
					return fmt.Errorf(
						"failed to get blob for stalled upload %s: %w",
						files[i].FileUuid,
						getBlobErr,
					)
				}

				if getBlobErr == nil {
					blobPath = blobInfo.FilePath

					if drefErr := q.DecrementBlobRefCount(
						ctx,
						files[i].BlobSha256.String,
					); drefErr != nil {
						return fmt.Errorf(
							"failed to decrement blob refcount: %w",
							drefErr,
						)
					}

					updatedBlob, updErr := q.GetBlobBySHA256(ctx, files[i].BlobSha256.String)
					if updErr != nil {
						return fmt.Errorf(
							"failed to re-read blob after decrement: %w",
							updErr,
						)
					}

					if updatedBlob.RefCount <= 0 {
						if delErr := q.DeleteBlob(ctx, files[i].BlobSha256.String); delErr != nil {
							return fmt.Errorf(
								"failed to delete unreferenced blob: %w",
								delErr,
							)
						}
						blobDeleted = true
					}
				}
			}

			return q.DeleteFile(ctx, files[i].FileUuid)
		})
		if err != nil {
			j.Logger.ErrorContext(
				ctx,
				"failed to delete stalled upload from db",
				"file_uuid",
				files[i].FileUuid,
				"err",
				err,
			)
			continue
		}

		result.DeletedFiles = append(result.DeletedFiles, files[i].FileUuid)

		if files[i].BlobSha256.Valid {
			blob := Blob{
				UUID:       files[i].BlobSha256.String,
				FilePath:   blobPath,
				WasDeleted: blobDeleted,
			}
			result.Blobs = append(result.Blobs, blob)
		}

		if statErr == nil {
			if rmErr := os.Remove(tmpPath); rmErr != nil && !os.IsNotExist(rmErr) {
				j.Logger.ErrorContext(
					ctx,
					"failed to remove stalled tmp file",
					"path",
					tmpPath,
					"err",
					rmErr,
				)
			}
		}

		if blobDeleted && blobPath != "" {
			if rmErr := os.Remove(blobPath); rmErr != nil && !os.IsNotExist(rmErr) {
				j.Logger.ErrorContext(
					ctx,
					"failed to remove stalled blob file",
					"path",
					blobPath,
					"err",
					rmErr,
				)
			}
		}
	}

	return result
}

func (j *DBVacuumJob) cleanupOrphanedBlobs(ctx context.Context, invokedBy Invocator) CleanupOrphanedBlobResults {
	result := CleanupOrphanedBlobResults{
		Success:   true,
		InvokedBy: invokedBy,
	}

	blobs, err := j.Store.GetUnreferencedBlobs(ctx)
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result
	}

	for _, blob := range blobs {
		var shouldRemove bool

		err := j.Store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
			b, gErr := q.GetBlobBySHA256(ctx, blob.Sha256)
			if gErr != nil {
				if errors.Is(gErr, sql.ErrNoRows) {
					return nil
				}
				return gErr
			}

			if b.RefCount > 0 {
				return nil
			}

			if delErr := q.DeleteBlob(ctx, blob.Sha256); delErr != nil {
				return delErr
			}

			shouldRemove = true
			return nil
		})
		if err != nil {
			j.Logger.ErrorContext(
				ctx,
				"failed to delete orphaned blob from db",
				"sha256",
				blob.Sha256,
				"err",
				err,
			)
			continue
		}

		if shouldRemove {
			result.Blobs = append(result.Blobs, Blob{
				UUID:       blob.Sha256,
				FilePath:   blob.FilePath,
				WasDeleted: true,
			})

			if rmErr := os.Remove(blob.FilePath); rmErr != nil && !os.IsNotExist(rmErr) {
				j.Logger.ErrorContext(
					ctx,
					"failed to remove orphaned blob file",
					"path",
					blob.FilePath,
					"err",
					rmErr,
				)
			} else {
				j.Logger.InfoContext(
					ctx,
					"removed orphaned blob",
					"sha256",
					blob.Sha256,
					"path",
					blob.FilePath,
				)
			}
		}
	}

	return result
}

func (j *DBVacuumJob) tombstoneExpiredFiles(ctx context.Context, invokedBy Invocator) MarkExpiredFilesResults {
	result := MarkExpiredFilesResults{
		Success:   true,
		InvokedBy: invokedBy,
	}

	ids, err := j.Store.GetFilesWithPassedRetention(ctx)
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result
	}

	for _, id := range ids {
		if err := j.Store.MarkFileTombstoned(ctx, id); err != nil {
			j.Logger.ErrorContext(
				ctx,
				"failed to mark file tombstoned",
				"file_uuid",
				id,
				"err",
				err,
			)
			continue
		}
		result.FileUUIDS = append(result.FileUUIDS, id)
	}

	return result
}

func (j *DBVacuumJob) deleteExpiredTokens(ctx context.Context, invokedBy Invocator) DeleteExpiredTokensResult {
	result := DeleteExpiredTokensResult{
		Success:   true,
		InvokedBy: invokedBy,
	}

	if err := j.Store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
		return q.DeleteExpiredPublicTokens(ctx)
	}); err != nil {
		j.Logger.ErrorContext(ctx, "failed to delete expired public tokens", "err", err)
		result.Success = false
		result.Error = errStr(err)
	}

	return result
}

func (j *DBVacuumJob) deleteOldTmpFiles(ctx context.Context, invokedBy Invocator) DeleteOldTmpFilesResults {
	result := DeleteOldTmpFilesResults{
		Success:   true,
		InvokedBy: invokedBy,
	}

	entries, err := os.ReadDir(j.Dirs.TmpDir)
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result
	}

	for _, entry := range entries {
		path := filepath.Join(j.Dirs.TmpDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()) > j.PendingFileRemoveInterval {
			if err := os.RemoveAll(path); err != nil {
				j.Logger.ErrorContext(ctx, "failed to remove old tmp entry", "path", path, "err", err)
			} else {
				result.FilePaths = append(result.FilePaths, path)
			}
		}
	}

	return result
}
