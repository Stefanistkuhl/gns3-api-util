package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"path/filepath"

	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

//go:embed schema.sql
var Schema string

var ErrNoDb = errors.New("no db")
var ErrQeuryDb = errors.New("failed to query the db")
var ErrClusterExists = errors.New("Cluster already exists")

func openDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func InitIfNeeded() (*sql.DB, error) {
	dir, err := utils.GetGNS3Dir()
	if err != nil {
		return nil, fmt.Errorf("get dir: %w", err)
	}
	dbPath := filepath.Join(dir, "clusterData.db")

	_, statErr := os.Stat(dbPath)
	if os.IsNotExist(statErr) {
		db, err := openDB(dbPath)
		if err != nil {
			return nil, err
		}
		tx, err := db.Begin()
		if err != nil {
			db.Close()
			return nil, err
		}
		if _, err := tx.Exec(Schema); err != nil {
			tx.Rollback()
			db.Close()
			return nil, fmt.Errorf("apply schema: %w", err)
		}
		if err := tx.Commit(); err != nil {
			db.Close()
			return nil, err
		}
		return db, nil
	}

	// Exists: open without applying schema
	return openDB(dbPath)
}

func CheckIfCluterExists(name string) error {
	dir, err := utils.GetGNS3Dir()
	if err != nil {
		return ErrNoDb
	}

	dbPath := filepath.Join(dir, "clusterData.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return ErrNoDb
	}

	db, err := InitIfNeeded()
	if err != nil {
		return err
	}
	defer db.Close()

	var exists bool
	if err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM clusters WHERE name = ? LIMIT 1)", name,
	).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return ErrClusterExists
	}
	return nil
}

func QueryRows[T any](
	db *sql.DB,
	query string,
	scanFn func(*sql.Rows) (T, error),
	args ...any,
) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}

func UpdateRows(
	db *sql.DB,
	query string,
	args ...any,
) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func CreateClusters(conn *sql.DB, clusters []ClusterName) ([]ClusterName, error) {
	var ids []ClusterName

	tx, err := conn.Begin()
	if err != nil {
		return ids, fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO clusters (name, description)
        VALUES (?, ?)
    `)
	if err != nil {
		tx.Rollback()
		return ids, fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, c := range clusters {
		res, err := stmt.Exec(c.Name, c.Desc)
		if err != nil {
			tx.Rollback()
			return ids, fmt.Errorf("insert cluster %s failed: %w", c.Name, err)
		}
		id, idErr := res.LastInsertId()
		if idErr != nil {
			return ids, fmt.Errorf("insert cluster %s failed: %w", c.Name, err)

		}
		a := ClusterName{
			Id:   int(id),
			Name: c.Name,
			Desc: c.Desc,
		}
		ids = append(ids, a)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return ids, nil
}

func InsertNodesIntoClusters(conn *sql.DB, nodes []NodeDataAll) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO nodes (cluster_id, protocol, host, port, weight, max_groups, auth_user)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, n := range nodes {
		_, err := stmt.Exec(n.ClusterID, n.Protocol, n.Host, n.Port, n.Weight, n.MaxGroups, n.User)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert node %s:%d failed: %w", n.Host, n.Port, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func InsertNodes(clusterID int, nodes []NodeData) error {
	dbConn, err := InitIfNeeded()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer dbConn.Close()

	tx, err := dbConn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO nodes (cluster_id, protocol, host, port, weight, max_groups, auth_user)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, n := range nodes {
		_, err := stmt.Exec(clusterID, n.Protocol, n.Host, n.Port, n.Weight, n.MaxGroups, n.User)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert node %s:%d failed: %w", n.Host, n.Port, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func GetNodes(conn *sql.DB) ([]NodeDataAll, error) {
	return QueryRows(conn,
		"SELECT node_id, cluster_id, protocol, auth_user, host, port, weight, max_groups FROM nodes ORDER BY cluster_id",
		func(rows *sql.Rows) (NodeDataAll, error) {
			var n NodeDataAll
			err := rows.Scan(&n.ID, &n.ClusterID, &n.Protocol, &n.User, &n.Host, &n.Port, &n.Weight, &n.MaxGroups)
			return n, err
		},
	)

}

func GetClusters(conn *sql.DB) ([]ClusterName, error) {
	return QueryRows(conn,
		"SELECT cluster_id, name, description FROM clusters ORDER BY cluster_id",
		func(rows *sql.Rows) (ClusterName, error) {
			var c ClusterName
			err := rows.Scan(&c.Id, &c.Name, &c.Desc)
			return c, err
		},
	)

}
