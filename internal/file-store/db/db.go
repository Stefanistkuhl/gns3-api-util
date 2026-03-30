package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	_ "turso.tech/database/tursogo"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	*sqlc_file_store.Queries
	DB     *sql.DB
	tracer trace.Tracer
}

type TracedDB struct {
	inner  sqlc_file_store.DBTX
	tracer trace.Tracer
}

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func NewStore(dbPath string, isPrimary bool) (*Store, error) {
	if dbPath == "" {
		dbPath = filepath.Join("data", "file-store.db")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	if isPrimary {
		migrationDB, err := sql.Open("turso", dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open migration db: %w", err)
		}
		defer migrationDB.Close()

		if err := runMigrations(migrationDB); err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
	}

	db, err := sql.Open("turso", dbPath)
	if err != nil {
		return nil, err
	}

	setupSQL := `
        PRAGMA foreign_keys = ON;
        PRAGMA journal_mode = 'experimental_mvcc';
        PRAGMA busy_timeout = 5000;
	`
	if _, err := db.ExecContext(context.Background(), setupSQL); err != nil {
		return nil, fmt.Errorf("failed to initialize turso engine: %w", err)
	}

	tracer := otel.Tracer("file-store-db")
	return &Store{
		Queries: sqlc_file_store.New(&TracedDB{inner: db, tracer: tracer}),
		DB:      db,
		tracer:  tracer,
	}, nil
}

func runMigrations(db *sql.DB) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (s *Store) HasEffectiveBucketPermission(
	ctx context.Context,
	bucketID, userID, permLevel string,
	roleNames []string,
) (bool, error) {
	ok, err := s.HasBucketPermission(ctx, sqlc_file_store.HasBucketPermissionParams{
		BucketID:    bucketID,
		PrincipalID: userID,
		Permission:  permLevel,
	})
	if err != nil {
		return false, fmt.Errorf("bucket user permission check: %w", err)
	}
	if ok {
		return true, nil
	}

	if len(roleNames) == 0 {
		return false, nil
	}
	placeholders := strings.Repeat("?,", len(roleNames))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf( //nolint:gosec // G201: placeholders only contains "?" chars, no user data
		`SELECT COUNT(*) > 0
		FROM bucket_permissions
		WHERE bucket_id = ?
		  AND principal_type = 'role'
		  AND principal_id IN (%s)
		  AND permission = ?
		  AND (expires_at IS NULL OR expires_at = '' OR expires_at > CURRENT_TIMESTAMP)`,
		placeholders,
	)
	args := make([]any, 0, 2+len(roleNames))
	args = append(args, bucketID)
	for _, r := range roleNames {
		args = append(args, r)
	}
	args = append(args, permLevel)
	var has bool
	if err := s.DB.QueryRowContext(ctx, query, args...).Scan(&has); err != nil {
		return false, fmt.Errorf("bucket role permission check: %w", err)
	}
	return has, nil
}

func (s *Store) HasEffectiveFilePermission(
	ctx context.Context,
	fileUUID, userID, permLevel string,
	roleNames []string,
) (bool, error) {
	ok, err := s.HasFilePermission(ctx, sqlc_file_store.HasFilePermissionParams{
		FileUuid:    fileUUID,
		PrincipalID: userID,
		Permission:  permLevel,
	})
	if err != nil {
		return false, fmt.Errorf("file user permission check: %w", err)
	}
	if ok {
		return true, nil
	}

	if len(roleNames) == 0 {
		return false, nil
	}
	placeholders := strings.Repeat("?,", len(roleNames))
	placeholders = placeholders[:len(placeholders)-1]
	query := fmt.Sprintf( //nolint:gosec // G201: placeholders only contains "?" chars, no user data
		`SELECT COUNT(*) > 0
		FROM file_permissions
		WHERE file_uuid = ?
		  AND principal_type = 'role'
		  AND principal_id IN (%s)
		  AND permission = ?
		  AND (expires_at IS NULL OR expires_at = '' OR expires_at > CURRENT_TIMESTAMP)`,
		placeholders,
	)
	args := make([]any, 0, 2+len(roleNames))
	args = append(args, fileUUID)
	for _, r := range roleNames {
		args = append(args, r)
	}
	args = append(args, permLevel)
	var has bool
	if err := s.DB.QueryRowContext(ctx, query, args...).Scan(&has); err != nil {
		return false, fmt.Errorf("file role permission check: %w", err)
	}
	return has, nil
}

// ListBucketsAccessibleToUser returns all buckets that are visible to a user:
// buckets they own, plus buckets where they (or one of their roles) has any
// non-expired permission entry.
func (s *Store) ListBucketsAccessibleToUser(
	ctx context.Context,
	userID string,
	roleNames []string,
) ([]sqlc_file_store.ListBucketsByOwnerRow, error) {
	// Base query: owned or directly permitted to the user.
	query := `
SELECT bucket_id, name, owner_id, is_public, required_scopes, created_at, updated_at
FROM buckets
WHERE owner_id = ?
  OR bucket_id IN (
      SELECT bucket_id FROM bucket_permissions
      WHERE principal_type = 'user'
        AND principal_id = ?
        AND (expires_at IS NULL OR expires_at = '' OR expires_at > CURRENT_TIMESTAMP)
  )`

	args := []any{userID, userID}

	if len(roleNames) > 0 {
		placeholders := strings.Repeat("?,", len(roleNames))
		placeholders = placeholders[:len(placeholders)-1]
		query += fmt.Sprintf( //nolint:gosec // G201: only "?" placeholders, no user data interpolated
			`
  OR bucket_id IN (
      SELECT bucket_id FROM bucket_permissions
      WHERE principal_type = 'role'
        AND principal_id IN (%s)
        AND (expires_at IS NULL OR expires_at = '' OR expires_at > CURRENT_TIMESTAMP)
  )`, placeholders)
		for _, r := range roleNames {
			args = append(args, r)
		}
	}

	query += "\nORDER BY created_at DESC"

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list accessible buckets: %w", err)
	}
	defer rows.Close()

	var items []sqlc_file_store.ListBucketsByOwnerRow
	for rows.Next() {
		var i sqlc_file_store.ListBucketsByOwnerRow
		if err := rows.Scan(
			&i.BucketID,
			&i.Name,
			&i.OwnerID,
			&i.IsPublic,
			&i.RequiredScopes,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return items, rows.Err()
}

func (s *Store) WithTx(
	ctx context.Context,
	fn func(*sqlc_file_store.Queries) error,
) (err error) {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	q := sqlc_file_store.New(&TracedDB{
		inner:  tx,
		tracer: s.tracer,
	})

	err = fn(q)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (t *TracedDB) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"db.Exec",
		trace.WithAttributes(attribute.String("db.statement", query)),
	)
	defer span.End()

	res, err := t.inner.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
	}
	return res, err
}

func (t *TracedDB) PrepareContext(
	ctx context.Context,
	query string,
) (*sql.Stmt, error) {
	return t.inner.PrepareContext(ctx, query)
}

func (t *TracedDB) QueryContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"db.Query",
		trace.WithAttributes(attribute.String("db.statement", query)),
	)
	defer span.End()

	rows, err := t.inner.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
	}
	return rows, err
}

func (t *TracedDB) QueryRowContext(
	ctx context.Context,
	query string,
	args ...any,
) *sql.Row {
	_, span := t.tracer.Start(
		ctx,
		"db.QueryRow",
		trace.WithAttributes(attribute.String("db.statement", query)),
	)
	defer span.End()

	return t.inner.QueryRowContext(ctx, query, args...)
}
