package backgroundjobs

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	dbpkg "github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/pkg/models"
)

type ObjectIndexJob struct {
	Store  *dbpkg.Store
	Logger *slog.Logger
	Dirs   fs.Dirs
}

type IndexedObjectFile struct {
	SHA256    string `json:"sha256"`
	FilePath  string `json:"file_path"`
	SizeBytes int64  `json:"size_bytes"`
}

type IndexedTmpFile struct {
	FileUUID   string `json:"file_uuid"`
	FilePath   string `json:"file_path"`
	SizeBytes  int64  `json:"size_bytes"`
	FileStatus string `json:"file_status,omitempty"`
}

type MissingBlobReference struct {
	SHA256    string   `json:"sha256"`
	FileUUIDs []string `json:"file_uuids"`
}

type ObjectIndexResult struct {
	Success               bool                   `json:"success"`
	Error                 *string                `json:"error"`
	InvokedBy             Invocator              `json:"invoked_by"`
	ScannedObjects        int                    `json:"scanned_objects"`
	ScannedTmpFiles       int                    `json:"scanned_tmp_files"`
	VerifiedBlobs         int                    `json:"verified_blobs"`
	RepairedBlobs         []Blob                 `json:"repaired_blobs"`
	RemovedBlobRows       []string               `json:"removed_blob_rows"`
	TombstonedFiles       []string               `json:"tombstoned_files"`
	MissingBlobRefs       []MissingBlobReference `json:"missing_blob_refs"`
	OrphanedObjectFiles   []IndexedObjectFile    `json:"orphaned_object_files"`
	OrphanedTmpFiles      []IndexedTmpFile       `json:"orphaned_tmp_files"`
	UnexpectedDiskEntries []string               `json:"unexpected_disk_entries"`
}

type blobIndexRow = sqlc_file_store.ListBlobsForIndexRow
type fileBlobRef = sqlc_file_store.ListFileBlobRefsRow

func (j *ObjectIndexJob) ExecuteIteration(ctx context.Context, invokedBy Invocator) (*ObjectIndexResult, error) {
	result := &ObjectIndexResult{
		Success:   true,
		InvokedBy: invokedBy,
	}

	diskObjects, unexpected, err := j.scanObjectFiles()
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result, err
	}
	result.UnexpectedDiskEntries = append(result.UnexpectedDiskEntries, unexpected...)
	result.ScannedObjects = len(diskObjects)

	blobRows, err := j.loadBlobRows(ctx)
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result, err
	}

	fileRefs, err := j.loadFileBlobRefs(ctx)
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result, err
	}

	for sha, blobRow := range blobRows {
		refs := fileRefs[sha]
		actualRefCount := int64(len(refs))
		diskFile, existsOnDisk := diskObjects[sha]

		if !existsOnDisk {
			if actualRefCount == 0 {
				if delErr := j.deleteBlobRow(ctx, sha); delErr != nil {
					result.Success = false
					result.Error = errStr(delErr)
					return result, delErr
				}
				result.RemovedBlobRows = append(result.RemovedBlobRows, sha)
				continue
			}

			fileUUIDs := make([]string, 0, len(refs))
			for _, ref := range refs {
				fileUUIDs = append(fileUUIDs, ref.FileUuid)
			}
			result.MissingBlobRefs = append(result.MissingBlobRefs, MissingBlobReference{
				SHA256:    sha,
				FileUUIDs: fileUUIDs,
			})

			tombstoned, tombErr := j.tombstoneReferencedFiles(ctx, sha, refs)
			if tombErr != nil {
				result.Success = false
				result.Error = errStr(tombErr)
				return result, tombErr
			}
			result.TombstonedFiles = append(result.TombstonedFiles, tombstoned...)

			if blobRow.RefCount != actualRefCount {
				if updErr := j.updateBlobMetadata(ctx, sha, blobRow.FilePath, blobRow.SizeBytes, actualRefCount); updErr != nil {
					result.Success = false
					result.Error = errStr(updErr)
					return result, updErr
				}
			}

			continue
		}

		result.VerifiedBlobs++
		delete(diskObjects, sha)

		if blobRow.FilePath != diskFile.FilePath || blobRow.SizeBytes != diskFile.SizeBytes || blobRow.RefCount != actualRefCount {
			if updErr := j.updateBlobMetadata(ctx, sha, diskFile.FilePath, diskFile.SizeBytes, actualRefCount); updErr != nil {
				result.Success = false
				result.Error = errStr(updErr)
				return result, updErr
			}
			result.RepairedBlobs = append(result.RepairedBlobs, Blob{
				UUID:     sha,
				RefCount: int(actualRefCount),
				FilePath: diskFile.FilePath,
			})
		}
	}

	for _, diskFile := range diskObjects {
		result.OrphanedObjectFiles = append(result.OrphanedObjectFiles, diskFile)
	}

	tmpFiles, tmpUnexpected, scannedTmpFiles, err := j.scanTmpFiles(ctx)
	if err != nil {
		result.Success = false
		result.Error = errStr(err)
		return result, err
	}
	result.ScannedTmpFiles = scannedTmpFiles
	result.OrphanedTmpFiles = tmpFiles
	result.UnexpectedDiskEntries = append(result.UnexpectedDiskEntries, tmpUnexpected...)

	return result, nil
}

func (j *ObjectIndexJob) scanObjectFiles() (objectFiles map[string]IndexedObjectFile, unexpected []string, err error) {
	objectFiles = make(map[string]IndexedObjectFile)
	unexpected = make([]string, 0)

	err = filepath.WalkDir(j.Dirs.ObjectDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if !looksLikeSHA256(name) {
			unexpected = append(unexpected, path)
			return nil
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}

		if _, exists := objectFiles[name]; exists {
			unexpected = append(unexpected, path)
			return nil
		}

		expectedPath := expectedObjectPath(j.Dirs.ObjectDir, name)
		if filepath.Clean(path) != expectedPath {
			unexpected = append(unexpected, path)
		}

		objectFiles[name] = IndexedObjectFile{
			SHA256:    name,
			FilePath:  path,
			SizeBytes: info.Size(),
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("scan object dir: %w", err)
		return nil, nil, err
	}

	return objectFiles, unexpected, nil
}

func (j *ObjectIndexJob) scanTmpFiles(ctx context.Context) (orphaned []IndexedTmpFile, unexpected []string, scannedCount int, err error) {
	tmpStatuses, err := j.loadTmpUploadStatuses(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	entries, err := os.ReadDir(j.Dirs.TmpDir)
	if err != nil {
		err = fmt.Errorf("scan tmp dir: %w", err)
		return nil, nil, 0, err
	}

	orphaned = make([]IndexedTmpFile, 0)
	unexpected = make([]string, 0)
	scannedCount = 0

	for _, entry := range entries {
		if entry.IsDir() {
			unexpected = append(unexpected, filepath.Join(j.Dirs.TmpDir, entry.Name()))
			continue
		}

		path := filepath.Join(j.Dirs.TmpDir, entry.Name())
		if !strings.HasSuffix(entry.Name(), ".tmp") {
			unexpected = append(unexpected, path)
			continue
		}
		scannedCount++

		info, infoErr := entry.Info()
		if infoErr != nil {
			err = fmt.Errorf("stat tmp entry %s: %w", path, infoErr)
			return nil, nil, 0, err
		}

		fileUUID := strings.TrimSuffix(entry.Name(), ".tmp")
		status, known := tmpStatuses[fileUUID]
		if known {
			continue
		}

		orphaned = append(orphaned, IndexedTmpFile{
			FileUUID:   fileUUID,
			FilePath:   path,
			SizeBytes:  info.Size(),
			FileStatus: status,
		})
	}

	return orphaned, unexpected, scannedCount, nil
}

func (j *ObjectIndexJob) loadBlobRows(ctx context.Context) (map[string]blobIndexRow, error) {
	rows, err := j.Store.ListBlobsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("query blobs: %w", err)
	}

	blobs := make(map[string]blobIndexRow)
	for _, row := range rows {
		blobs[row.Sha256] = row
	}

	return blobs, nil
}

func (j *ObjectIndexJob) loadFileBlobRefs(ctx context.Context) (map[string][]fileBlobRef, error) {
	rows, err := j.Store.ListFileBlobRefs(ctx)
	if err != nil {
		return nil, fmt.Errorf("query file blob refs: %w", err)
	}

	refs := make(map[string][]fileBlobRef)
	for _, row := range rows {
		if !row.BlobSha256.Valid || row.BlobSha256.String == "" {
			continue
		}
		refs[row.BlobSha256.String] = append(refs[row.BlobSha256.String], row)
	}

	return refs, nil
}

func (j *ObjectIndexJob) loadTmpUploadStatuses(ctx context.Context) (map[string]string, error) {
	rows, err := j.Store.ListTmpUploadStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("query tmp upload statuses: %w", err)
	}

	statuses := make(map[string]string)
	for _, row := range rows {
		statuses[row.FileUuid] = row.Status
	}

	return statuses, nil
}

func (j *ObjectIndexJob) updateBlobMetadata(ctx context.Context, sha256, filePath string, sizeBytes, refCount int64) error {
	err := j.Store.UpdateBlobIndexMetadata(ctx, sqlc_file_store.UpdateBlobIndexMetadataParams{
		FilePath:  filePath,
		SizeBytes: sizeBytes,
		RefCount:  refCount,
		Sha256:    sha256,
	})
	if err != nil {
		return fmt.Errorf("update blob %s metadata: %w", sha256, err)
	}
	return nil
}

func (j *ObjectIndexJob) deleteBlobRow(ctx context.Context, sha256 string) error {
	err := j.Store.DeleteBlob(ctx, sha256)
	if err != nil {
		return fmt.Errorf("delete blob %s: %w", sha256, err)
	}
	return nil
}

func (j *ObjectIndexJob) tombstoneReferencedFiles(ctx context.Context, sha256 string, refs []fileBlobRef) ([]string, error) {
	tombstoned := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.Status == string(models.FileStatusTombstoned) {
			continue
		}
		tombstoned = append(tombstoned, ref.FileUuid)
	}

	if len(tombstoned) == 0 {
		return tombstoned, nil
	}

	err := j.Store.TombstoneFilesByBlobSHA(ctx, sql.NullString{String: sha256, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("tombstone files for missing blob %s: %w", sha256, err)
	}

	return tombstoned, nil
}

func expectedObjectPath(rootDir, sha256 string) string {
	return filepath.Clean(filepath.Join(rootDir, sha256[:2], sha256[2:4], sha256))
}

func looksLikeSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
