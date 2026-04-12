package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/dbutils"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/middleware"
)

func newTestFilestoreHandlers(t *testing.T) (*FilestoreHandlers, *db.Store, fs.Dirs) {
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

	return &FilestoreHandlers{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Dirs:   &dirs,
	}, store, dirs
}

func withClaims(req *http.Request, claims *auth.Claims) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), middleware.ClaimsKey, claims))
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func createTestBucket(t *testing.T, store *db.Store, bucketID, ownerID string, requiredScopes *string) {
	t.Helper()

	_, err := store.CreateBucket(context.Background(), sqlc_file_store.CreateBucketParams{
		BucketID:       bucketID,
		Name:           "bucket-" + bucketID,
		OwnerID:        ownerID,
		BucketType:     "standard",
		IsPublic:       sql.NullBool{Bool: false, Valid: true},
		RequiredScopes: dbutils.NullString(requiredScopes),
	})
	if err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}
}

func initTestFile(t *testing.T, store *db.Store, fileUUID, ownerID, bucketID, status string) {
	t.Helper()

	_, err := store.InitFile(context.Background(), sqlc_file_store.InitFileParams{
		FileUuid:       fileUUID,
		Filename:       "test.bin",
		ContentType:    "application/octet-stream",
		OwnerID:        ownerID,
		BucketID:       bucketID,
		LastAccessedAt: dbutils.FormatDBTime(time.Now()),
	})
	if err != nil {
		t.Fatalf("InitFile() error = %v", err)
	}
	if status != "" && status != string(models.FileStatusPending) {
		if err := store.UpdateFileStatus(context.Background(), sqlc_file_store.UpdateFileStatusParams{
			Status:   status,
			FileUuid: fileUUID,
		}); err != nil {
			t.Fatalf("UpdateFileStatus() error = %v", err)
		}
	}
}

func grantBucketPermission(t *testing.T, store *db.Store, bucketID, principalID, permission string) {
	t.Helper()

	if _, err := store.InsertBucketPermission(context.Background(), sqlc_file_store.InsertBucketPermissionParams{
		ID:            bucketID + "-" + principalID + "-" + permission,
		BucketID:      bucketID,
		PrincipalType: "user",
		PrincipalID:   principalID,
		Permission:    permission,
		GrantedBy:     "owner",
		ExpiresAt:     "",
	}); err != nil {
		t.Fatalf("InsertBucketPermission() error = %v", err)
	}
}

func grantFilePermission(t *testing.T, store *db.Store, fileUUID, principalID, permission string) {
	t.Helper()

	if _, err := store.InsertFilePermission(context.Background(), sqlc_file_store.InsertFilePermissionParams{
		ID:            fileUUID + "-" + principalID + "-" + permission,
		FileUuid:      fileUUID,
		PrincipalType: "user",
		PrincipalID:   principalID,
		Permission:    permission,
		GrantedBy:     "owner",
		ExpiresAt:     "",
	}); err != nil {
		t.Fatalf("InsertFilePermission() error = %v", err)
	}
}

func TestHandleInitUploadRejectsBucketIDMismatch(t *testing.T) {
	ctx := context.Background()
	handlers, store, _ := newTestFilestoreHandlers(t)
	createTestBucket(t, store, "bucket-a", "owner", nil)
	createTestBucket(t, store, "bucket-b", "owner", nil)

	body := bytes.NewBufferString(`{"filename":"test.bin","size_bytes":10,"content_type":"application/octet-stream","bucket_id":"bucket-b"}`)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/buckets/bucket-a/files", body)
	req = withClaims(req, &auth.Claims{UserID: "owner"})
	req = withURLParam(req, "bucket_id", "bucket-a")

	rr := httptest.NewRecorder()
	handlers.HandleInitUpload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestHandleInitUploadRejectsUnauthorizedBucketWrite(t *testing.T) {
	ctx := context.Background()
	handlers, store, _ := newTestFilestoreHandlers(t)
	createTestBucket(t, store, "bucket-a", "owner", nil)

	body := bytes.NewBufferString(`{"filename":"test.bin","size_bytes":10,"content_type":"application/octet-stream","bucket_id":"bucket-a"}`)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/files", body)
	req = withClaims(req, &auth.Claims{UserID: "alice"})

	rr := httptest.NewRecorder()
	handlers.HandleInitUpload(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestHandleInitUploadAllowsBucketWriteACL(t *testing.T) {
	ctx := context.Background()
	handlers, store, _ := newTestFilestoreHandlers(t)
	createTestBucket(t, store, "bucket-a", "owner", nil)
	grantBucketPermission(t, store, "bucket-a", "alice", "write")

	body := bytes.NewBufferString(`{"filename":"test.bin","size_bytes":10,"content_type":"application/octet-stream","bucket_id":"bucket-a"}`)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/files", body)
	req = withClaims(req, &auth.Claims{UserID: "alice"})

	rr := httptest.NewRecorder()
	handlers.HandleInitUpload(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp models.InitUploadResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	file, err := store.GetFileByUUID(context.Background(), resp.FileUUID)
	if err != nil {
		t.Fatalf("GetFileByUUID() error = %v", err)
	}
	if file.BucketID != "bucket-a" {
		t.Fatalf("file.BucketID = %q, want %q", file.BucketID, "bucket-a")
	}
}

func TestGetUploadStatusAllowsBucketAdmin(t *testing.T) {
	ctx := context.Background()
	handlers, store, dirs := newTestFilestoreHandlers(t)
	createTestBucket(t, store, "bucket-a", "owner", nil)
	initTestFile(t, store, "file-1", "owner", "bucket-a", string(models.FileStatusPending))
	grantBucketPermission(t, store, "bucket-a", "alice", "admin")

	tmpPath := filepath.Join(dirs.TmpDir, "file-1.tmp")
	if err := os.WriteFile(tmpPath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/files/file-1/status", http.NoBody)
	req = withClaims(req, &auth.Claims{UserID: "alice"})
	req = withURLParam(req, "file_uuid", "file-1")

	rr := httptest.NewRecorder()
	handlers.GetUploadStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp models.GetUploadStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp.Offset != 5 {
		t.Fatalf("Offset = %d, want %d", resp.Offset, 5)
	}
}

func TestGeneratePublicTokenAllowsFileWriteACL(t *testing.T) {
	ctx := context.Background()
	handlers, store, _ := newTestFilestoreHandlers(t)
	createTestBucket(t, store, "bucket-a", "owner", nil)
	initTestFile(t, store, "file-1", "owner", "bucket-a", string(models.FileStatusAvailable))
	grantFilePermission(t, store, "file-1", "alice", "write")

	body := bytes.NewBufferString(`{"file_uuid":"file-1","bucket_uuid":"bucket-a"}`)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/files/file-1/public-token", body)
	req = withClaims(req, &auth.Claims{UserID: "alice"})
	req = withURLParam(req, "file_uuid", "file-1")

	rr := httptest.NewRecorder()
	handlers.GeneratePublicToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp models.PublicTokenResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp.BucketUUID != "bucket-a" {
		t.Fatalf("BucketUUID = %q, want %q", resp.BucketUUID, "bucket-a")
	}
}

func TestDownloadFileRejectsTombstonedFile(t *testing.T) {
	ctx := context.Background()
	handlers, store, dirs := newTestFilestoreHandlers(t)
	initTestFile(t, store, "file-1", "owner", models.GlobalBucketID, string(models.FileStatusPending))

	sha := strings.Repeat("c", 64)
	dstDir, err := dirs.CreateShardDirsIfNeed(sha)
	if err != nil {
		t.Fatalf("CreateShardDirsIfNeed() error = %v", err)
	}
	path := filepath.Join(dstDir, sha)
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := store.WithTx(context.Background(), func(q *sqlc_file_store.Queries) error {
		if upsertErr := q.UpsertBlob(context.Background(), sqlc_file_store.UpsertBlobParams{
			Sha256:    sha,
			FilePath:  path,
			SizeBytes: 5,
		}); upsertErr != nil {
			return upsertErr
		}

		_, finalizeErr := q.FinalizeFile(context.Background(), sqlc_file_store.FinalizeFileParams{
			BlobSha256: dbutils.NullString(&sha),
			FileUuid:   "file-1",
		})
		return finalizeErr
	}); err != nil {
		t.Fatalf("WithTx() error = %v", err)
	}

	if err := store.MarkFileTombstoned(context.Background(), "file-1"); err != nil {
		t.Fatalf("MarkFileTombstoned() error = %v", err)
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/files/file-1", http.NoBody)
	req = withClaims(req, &auth.Claims{UserID: "owner"})
	req = withURLParam(req, "file_uuid", "file-1")

	rr := httptest.NewRecorder()
	handlers.DownloadFileHandler(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusConflict, rr.Body.String())
	}
}
