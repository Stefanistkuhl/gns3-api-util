package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
