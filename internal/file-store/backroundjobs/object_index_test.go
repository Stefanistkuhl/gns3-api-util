package backgroundjobs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
)

func newTestObjectIndexJob(t *testing.T) (*ObjectIndexJob, *db.Store, fs.Dirs) {
	t.Helper()

	root := t.TempDir()
	store, err := db.NewStore(filepath.Join(root, "filestore.db"), true)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.DB.Close()
	})

	dirs, err := fs.CreateDirStructure(filepath.Join(root, "data"))
	if err != nil {
		t.Fatalf("CreateDirStructure() error = %v", err)
	}

	return &ObjectIndexJob{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Dirs:   dirs,
	}, store, dirs
}

func initAvailableFileWithBlob(t *testing.T, store *db.Store, dirs fs.Dirs, fileUUID, sha string, content []byte) string {
	t.Helper()

	ctx := context.Background()
	_, err := store.InitFile(ctx, sqlc_file_store.InitFileParams{
		FileUuid:       fileUUID,
		Filename:       "blob.bin",
		ContentType:    "application/octet-stream",
		OwnerID:        "owner",
		BucketID:       models.GlobalBucketID,
		LastAccessedAt: dbutils.FormatDBTime(time.Now()),
	})
	if err != nil {
		t.Fatalf("InitFile() error = %v", err)
	}

	dstDir, err := dirs.CreateShardDirsIfNeed(sha)
	if err != nil {
		t.Fatalf("CreateShardDirsIfNeed() error = %v", err)
	}
	path := filepath.Join(dstDir, sha)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
		if upsertErr := q.UpsertBlob(ctx, sqlc_file_store.UpsertBlobParams{
			Sha256:    sha,
			FilePath:  path,
			SizeBytes: int64(len(content)),
		}); upsertErr != nil {
			return upsertErr
		}

		_, finalizeErr := q.FinalizeFile(ctx, sqlc_file_store.FinalizeFileParams{
			BlobSha256: dbutils.NullString(&sha),
			FileUuid:   fileUUID,
		})
		return finalizeErr
	}); err != nil {
		t.Fatalf("WithTx() error = %v", err)
	}

	return path
}

func TestObjectIndexJobRepairsBlobMetadataFromDisk(t *testing.T) {
	job, store, dirs := newTestObjectIndexJob(t)
	sha := strings.Repeat("a", 64)
	path := initAvailableFileWithBlob(t, store, dirs, "file-repair", sha, []byte("hello world"))

	if _, err := store.DB.ExecContext(
		context.Background(),
		`UPDATE blobs SET file_path = ?, size_bytes = ?, ref_count = ? WHERE sha256 = ?`,
		"/wrong/path",
		int64(1),
		int64(99),
		sha,
	); err != nil {
		t.Fatalf("ExecContext() error = %v", err)
	}

	result, err := job.ExecuteIteration(context.Background(), InvocatorStartup)
	if err != nil {
		t.Fatalf("ExecuteIteration() error = %v", err)
	}

	blob, err := store.GetBlobBySHA256(context.Background(), sha)
	if err != nil {
		t.Fatalf("GetBlobBySHA256() error = %v", err)
	}
	if blob.FilePath != path {
		t.Fatalf("blob.FilePath = %q, want %q", blob.FilePath, path)
	}
	if blob.SizeBytes != int64(len("hello world")) {
		t.Fatalf("blob.SizeBytes = %d, want %d", blob.SizeBytes, len("hello world"))
	}
	if blob.RefCount != 1 {
		t.Fatalf("blob.RefCount = %d, want 1", blob.RefCount)
	}
	if len(result.RepairedBlobs) != 1 {
		t.Fatalf("len(result.RepairedBlobs) = %d, want 1", len(result.RepairedBlobs))
	}
}

func TestObjectIndexJobTombstonesFilesWithMissingBlob(t *testing.T) {
	job, store, dirs := newTestObjectIndexJob(t)
	sha := strings.Repeat("b", 64)
	path := initAvailableFileWithBlob(t, store, dirs, "file-missing", sha, []byte("gone"))

	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	result, err := job.ExecuteIteration(context.Background(), InvocatorStartup)
	if err != nil {
		t.Fatalf("ExecuteIteration() error = %v", err)
	}

	file, err := store.GetFileByUUID(context.Background(), "file-missing")
	if err != nil {
		t.Fatalf("GetFileByUUID() error = %v", err)
	}
	if file.Status != string(models.FileStatusTombstoned) {
		t.Fatalf("file.Status = %q, want %q", file.Status, models.FileStatusTombstoned)
	}
	if len(result.TombstonedFiles) != 1 || result.TombstonedFiles[0] != "file-missing" {
		t.Fatalf("result.TombstonedFiles = %#v, want [file-missing]", result.TombstonedFiles)
	}
	if len(result.MissingBlobRefs) != 1 || result.MissingBlobRefs[0].SHA256 != sha {
		t.Fatalf("result.MissingBlobRefs = %#v, want sha %q", result.MissingBlobRefs, sha)
	}
}
