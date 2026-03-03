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
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	_ "modernc.org/sqlite"
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

func NewStore() (*Store, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "file-store.db")
	}

	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close DB: %w", closeErr)
		}
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	tracer := otel.Tracer("db")
	tracedDB := &TracedDB{inner: db, tracer: tracer}

	return &Store{
		Queries: sqlc_file_store.New(tracedDB),
		DB:      db,
		tracer:  tracer,
	}, nil
}

func (s *Store) WithTx(ctx context.Context, fn func(*sqlc_file_store.Queries) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	tracedTx := &TracedDB{inner: tx, tracer: s.tracer}
	q := sqlc_file_store.New(tracedTx)

	if err := fn(q); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
		}
		return err
	}
	return tx.Commit()
}

func runMigrations(db *sql.DB) error {
	source, _ := iofs.New(migrationsFS, "migrations")
	driver, _ := sqlite3.WithInstance(db, &sqlite3.Config{})
	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (t *TracedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, span := t.tracer.Start(ctx, "db.Exec", trace.WithAttributes(attribute.String("db.statement", query)))
	defer span.End()
	res, err := t.inner.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
	}
	return res, err
}

func (t *TracedDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return t.inner.PrepareContext(ctx, query)
}

func (t *TracedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	ctx, span := t.tracer.Start(ctx, "db.Query", trace.WithAttributes(attribute.String("db.statement", query)))
	defer span.End()
	rows, err := t.inner.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
	}
	return rows, err
}

func (t *TracedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	_, span := t.tracer.Start(ctx, "db.QueryRow", trace.WithAttributes(attribute.String("db.statement", query)))
	defer span.End()
	return t.inner.QueryRowContext(ctx, query, args...)
}
