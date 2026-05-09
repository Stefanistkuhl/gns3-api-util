package db

import (
	"context"
	"sync"
	"testing"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/mig.db"
	store, err := NewStore(dbPath, true) // run migrations
	require.NoError(t, err)
	defer store.DB.Close()

	// check if the tables actually exist
	_, err = store.DB.ExecContext(ctx, "SELECT sha256 FROM blobs LIMIT 0")
	assert.NoError(t, err)
}

func TestMVCC_Integrity(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/mvcc.db"
	store, err := NewStore(dbPath, true)
	require.NoError(t, err)
	defer store.DB.Close()

	sha := "test-sha-256"
	// setup base blob
	_, err = store.DB.ExecContext(ctx, "INSERT INTO blobs (sha256, file_path, size_bytes, ref_count) VALUES (?, ?, ?, ?)",
		sha, "/tmp/test", 1024, 0)
	require.NoError(t, err)

	const numWorkers = 5
	const incrementsPerWorker = 5
	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal
			for j := 0; j < incrementsPerWorker; j++ {
				// use real increment func inside the concurrent tx
				err := store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
					return q.IncrementBlobRefCount(ctx, sha)
				})
				if err != nil {
					t.Logf("worker failed to increment: %v", err)
				}
			}
		}()
	}

	// unleash everyone
	close(startSignal)
	wg.Wait()

	var finalCount int64
	err = store.DB.QueryRowContext(ctx, "SELECT ref_count FROM blobs WHERE sha256 = ?", sha).Scan(&finalCount)
	require.NoError(t, err)

	// should be 25, no lost updates allowed
	assert.Equal(t, int64(numWorkers*incrementsPerWorker), finalCount, "lost data in atomic increments")
}

func TestConstraints(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/constraints.db"
	store, err := NewStore(dbPath, true)
	require.NoError(t, err)
	defer store.DB.Close()

	t.Run("Status_Check", func(t *testing.T) {
		// need a real file for the update to actually hit something
		fileID := "test-uuid"
		_, err := store.DB.ExecContext(ctx, `
			INSERT INTO files (file_uuid, filename, content_type, owner_id, bucket_id, status)
			VALUES (?, 'test.txt', 'text/plain', 'user', '00000000-0000-0000-0000-000000000000', 'pending')`,
			fileID)
		require.NoError(t, err)

		// try updating to garbage status
		err = store.UpdateFileStatus(ctx, sqlc_file_store.UpdateFileStatusParams{
			Status:   "broken_enum_value",
			FileUuid: fileID,
		})

		// sqlite should scream about the check constraint
		assert.Error(t, err, "status check failed to stop garbage value")
	})

	t.Run("Ref_Count_Check", func(t *testing.T) {
		sha := "integrity-sha"
		// start at 0
		_, err := store.DB.ExecContext(ctx,
			"INSERT INTO blobs (sha256, file_path, size_bytes, ref_count) VALUES (?, ?, ?, 0)",
			sha, "/p", 10)
		require.NoError(t, err)

		// this should fail immediately (0 - 1 = -1)
		err = store.DecrementBlobRefCount(ctx, sha)
		assert.Error(t, err, "ref_count dropped below zero")

		// make sure it stayed at 0
		var count int
		_ = store.DB.QueryRowContext(ctx, "SELECT ref_count FROM blobs WHERE sha256 = ?", sha).Scan(&count)
		assert.Equal(t, 0, count)
	})
}

func TestWithTx_PanicSafety(t *testing.T) {
	ctx := context.Background()
	dbPath := t.TempDir() + "/panic.db"
	store, err := NewStore(dbPath, true)
	require.NoError(t, err)

	sha := "panic-sha"
	_, _ = store.DB.ExecContext(ctx, "INSERT INTO blobs (sha256, file_path, size_bytes, ref_count) VALUES (?, ?, ?, 0)", sha, "/p", 10)

	// crash the tx
	assert.Panics(t, func() {
		_ = store.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
			_ = q.IncrementBlobRefCount(ctx, sha)
			panic("boom")
		})
	})

	// check if rollback cleaned up the mess
	var count int
	_ = store.DB.QueryRow("SELECT ref_count FROM blobs WHERE sha256 = ?", sha).Scan(&count)
	assert.Equal(t, 0, count, "tx didn't rollback after panic")
}
