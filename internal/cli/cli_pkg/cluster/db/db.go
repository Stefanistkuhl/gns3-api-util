package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

//go:embed schema.sql
var Schema string

type Store struct {
	*sqlc.Queries
	DB *sql.DB
}

func openDB(ctx context.Context, dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	setupQuery := `
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA foreign_keys = ON;
		PRAGMA busy_timeout = 5000;
	`
	if _, err := db.ExecContext(ctx, setupQuery); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to apply sqlite pragmas: %w", err)
	}

	return db, nil
}

func Init() (*Store, error) {
	dir, err := pathutils.GetGNS3Dir()
	if err != nil {
		return nil, fmt.Errorf("get dir: %w", err)
	}
	dbPath := filepath.Join(dir, "clusterData.db")

	_, statErr := os.Stat(dbPath)
	dbExists := !os.IsNotExist(statErr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := openDB(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if !dbExists {
		if _, err := db.ExecContext(ctx, Schema); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("apply initial schema: %w", err)
		}
	}

	return &Store{
		Queries: sqlc.New(db),
		DB:      db,
	}, nil
}

func (s *Store) ReadOnly(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}

	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Log the rollback error but don't override the original error
			fmt.Printf("Warning: failed to rollback transaction: %v\n", rollbackErr)
		}
	}()

	qtx := s.WithTx(tx)

	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}
