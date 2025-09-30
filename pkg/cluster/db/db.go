package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

//go:embed schema.sql
var Schema string

var ErrNoDb = errors.New("no db")
var ErrQeuryDb = errors.New("failed to query the db")
var ErrClusterExists = errors.New("cluster already exists")
var ErrNodeExists = errors.New("node already exists")
var ErrClassExists = errors.New("class already exists")

func openDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
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
			_ = db.Close()
			return nil, err
		}
		if _, err := tx.Exec(Schema); err != nil {
			_ = tx.Rollback()
			_ = db.Close()
			return nil, fmt.Errorf("apply schema: %w", err)
		}
		if err := tx.Commit(); err != nil {
			_ = db.Close()
			return nil, err
		}
		return db, nil
	}

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
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}()

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

func CheckIfNodeExists(conn *sql.DB, clusterID int, host string, port int) error {
	var exists bool
	if err := conn.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM nodes WHERE cluster_id = ? AND host = ? AND port = ? LIMIT 1)", clusterID, host, port,
	).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrNodeExists
	}
	return nil
}

func CheckIfClassExists(conn *sql.DB, clusterID int, className string) error {
	var exists bool
	if err := conn.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM classes WHERE cluster_id = ? AND name = ? LIMIT 1)",
		clusterID, className,
	).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return ErrClassExists
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
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()

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
		_ = tx.Rollback()
		return ids, fmt.Errorf("prepare stmt: %w", err)
	}
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
	}()

	for _, c := range clusters {
		res, err := stmt.Exec(c.Name, c.Desc)
		if err != nil {
			_ = tx.Rollback()
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
		_ = tx.Rollback()
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
	}()

	for _, n := range nodes {
		_, err := stmt.Exec(n.ClusterID, n.Protocol, n.Host, n.Port, n.Weight, n.MaxGroups, n.User)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert node %s:%d failed: %w", n.Host, n.Port, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func InsertNodes(clusterID int, nodes []NodeData) ([]NodeDataAll, error) {
	dbConn, err := InitIfNeeded()
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer func() {
		if dbConn != nil {
			_ = dbConn.Close()
		}
	}()

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO nodes (cluster_id, protocol, host, port, weight, max_groups, auth_user)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("prepare stmt: %w", err)
	}
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
	}()

	var insertedNodes []NodeDataAll
	for _, n := range nodes {
		res, err := stmt.Exec(clusterID, n.Protocol, n.Host, n.Port, n.Weight, n.MaxGroups, n.User)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("insert node %s:%d failed: %w", n.Host, n.Port, err)
		}

		nodeID, idErr := res.LastInsertId()
		if idErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("get node insert id: %w", idErr)
		}

		insertedNode := NodeDataAll{
			ID:        int(nodeID),
			ClusterID: clusterID,
			User:      n.User,
			Protocol:  n.Protocol,
			Host:      n.Host,
			Port:      n.Port,
			Weight:    n.Weight,
			MaxGroups: n.MaxGroups,
		}
		insertedNodes = append(insertedNodes, insertedNode)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return insertedNodes, nil
}

func AssignGroupsToNode(conn *sql.DB, ids NodeAndGroupIds) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Prepare(`
        INSERT INTO group_assignments (node_id, group_id)
        VALUES (?, ?)
    `)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
	}()

	for _, a := range ids.Assignments {
		for _, g := range a.GroupIDs {
			_, err := stmt.Exec(a.NodeID, g)
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("insert assingment failed: %w", err)
			}
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

func InsertClassIntoDB(conn *sql.DB, clusterID int, data schemas.Class) (InsertedClassData, error) {
	tx, conErr := conn.Begin()
	if conErr != nil {
		return InsertedClassData{}, fmt.Errorf("begin tx: %w", conErr)
	}

	res, createClassErr := tx.Exec(`
  INSERT INTO classes (cluster_id, name, description)
  VALUES (?, ?, ?)
`, clusterID, data.Name, data.Desc)
	if createClassErr != nil {
		return InsertedClassData{}, fmt.Errorf("insert class: %w", createClassErr)
	}

	classID, classErr := res.LastInsertId()
	if classErr != nil {
		return InsertedClassData{}, fmt.Errorf("get last insert id: %w", classErr)
	}

	result := InsertedClassData{
		ClassID:  int(classID),
		GroupIDs: make(map[string]int),
		UserIDs:  make([]int, 0),
	}

	stmtGrp, grpErr := tx.Prepare(`
	       INSERT INTO groups (class_id, name)
	       VALUES (?, ?)
	   `)
	if grpErr != nil {
		_ = tx.Rollback()
		return InsertedClassData{}, fmt.Errorf("prepare group stmt: %w", grpErr)
	}
	defer func() {
		if stmtGrp != nil {
			_ = stmtGrp.Close()
		}
	}()

	stmtUsr, usrErr := tx.Prepare(`
	       INSERT INTO users (group_id, username, full_name, default_password)
	       VALUES (?, ?, ?, ?)
	   `)
	if usrErr != nil {
		_ = tx.Rollback()
		return InsertedClassData{}, fmt.Errorf("prepare user stmt: %w", usrErr)
	}
	defer func() {
		if stmtUsr != nil {
			_ = stmtUsr.Close()
		}
	}()

	for _, g := range data.Groups {
		grpRes, err := stmtGrp.Exec(classID, g.Name)
		if err != nil {
			_ = tx.Rollback()
			return InsertedClassData{}, fmt.Errorf("insert group %s failed: %w", g.Name, err)
		}

		groupID, groupIDErr := grpRes.LastInsertId()
		if groupIDErr != nil {
			_ = tx.Rollback()
			return InsertedClassData{}, fmt.Errorf("get group insert id: %w", groupIDErr)
		}

		result.GroupIDs[g.Name] = int(groupID)

		for _, u := range g.Students {
			usrRes, err := stmtUsr.Exec(groupID, u.UserName, u.FullName, u.Password)
			if err != nil {
				_ = tx.Rollback()
				return InsertedClassData{}, fmt.Errorf("insert user %v failed: %w", u.FullName, err)
			}

			userID, userIDErr := usrRes.LastInsertId()
			if userIDErr != nil {
				_ = tx.Rollback()
				return InsertedClassData{}, fmt.Errorf("get user insert id: %w", userIDErr)
			}

			result.UserIDs = append(result.UserIDs, int(userID))
		}
	}

	if err := tx.Commit(); err != nil {
		return InsertedClassData{}, fmt.Errorf("commit tx: %w", err)
	}

	return result, nil
}

func DeleteClassFromDB(conn *sql.DB, clusterID int, className string) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var classID int
	err = tx.QueryRow(
		"SELECT class_id FROM classes WHERE cluster_id = ? AND name = ?",
		clusterID, className,
	).Scan(&classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("class %s not found in database", className)
		}
		return fmt.Errorf("failed to find class: %w", err)
	}

	groupRows, err := tx.Query(
		"SELECT group_id FROM groups WHERE class_id = ?",
		classID,
	)
	if err != nil {
		return fmt.Errorf("failed to get groups for class: %w", err)
	}
	defer func() {
		if groupRows != nil {
			_ = groupRows.Close()
		}
	}()

	var groupIDs []int
	for groupRows.Next() {
		var groupID int
		err := groupRows.Scan(&groupID)
		if err != nil {
			return fmt.Errorf("failed to scan group id: %w", err)
		}
		groupIDs = append(groupIDs, groupID)
	}
	if err := groupRows.Err(); err != nil {
		return fmt.Errorf("failed to iterate over group rows: %w", err)
	}

	nodeRows, err := tx.Query(
		"SELECT node_id FROM nodes WHERE cluster_id = ?",
		clusterID,
	)
	if err != nil {
		return fmt.Errorf("failed to get nodes for cluster: %w", err)
	}
	defer func() {
		if err := nodeRows.Close(); err != nil {
			fmt.Printf("failed to close node rows: %v", err)
		}
	}()

	var nodeIDs []int
	for nodeRows.Next() {
		var nodeID int
		err := nodeRows.Scan(&nodeID)
		if err != nil {
			return fmt.Errorf("failed to scan node id: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	if err := nodeRows.Err(); err != nil {
		return fmt.Errorf("failed to iterate over node rows: %w", err)
	}

	// Unassign all groups from all nodes in the cluster
	for _, nodeID := range nodeIDs {
		unassignErr := unassignGroupsFromNodeWithTx(conn, tx, NodeAndGroupIds{
			Assignments: []AssignmentIds{{NodeID: nodeID, GroupIDs: groupIDs}},
		})
		if unassignErr != nil {
			return fmt.Errorf("failed to unassign groups: %w", unassignErr)
		}
	}

	result, err := tx.Exec("DELETE FROM classes WHERE class_id = ?", classID)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("class %s not found in database", className)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

type ClassData struct {
	ID          int    `json:"id"`
	ClusterID   int    `json:"cluster_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetClassesFromDB(conn *sql.DB, clusterID int) ([]ClassData, error) {
	return QueryRows(conn,
		"SELECT class_id, cluster_id, name, description FROM classes WHERE cluster_id = ? ORDER BY name",
		func(rows *sql.Rows) (ClassData, error) {
			var c ClassData
			err := rows.Scan(&c.ID, &c.ClusterID, &c.Name, &c.Description)
			return c, err
		},
		clusterID,
	)
}

func UnassignGroupsFromNode(conn *sql.DB, ids NodeAndGroupIds) error {
	return unassignGroupsFromNodeWithTx(conn, nil, ids)
}

func unassignGroupsFromNodeWithTx(conn *sql.DB, tx *sql.Tx, ids NodeAndGroupIds) error {
	var err error
	var stmt *sql.Stmt
	var commitTx bool

	if tx == nil {
		tx, err = conn.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		commitTx = true
	}

	stmt, err = tx.Prepare(`
        DELETE FROM group_assignments
        WHERE node_id = ? AND group_id = ?
    `)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
	}()

	for _, g := range ids.Assignments {
		for _, groupID := range g.GroupIDs {
			_, err := stmt.Exec(g.NodeID, groupID)
			if err != nil {
				return fmt.Errorf("delete assignment failed: %w", err)
			}
		}
	}

	if commitTx {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
	}

	return nil
}

func GetNodeGroupNamesForClass(conn *sql.DB, clusterID int, className string) ([]NodeGroupsForClass, error) {
	rows, err := conn.Query(
		`
SELECT
  n.protocol || '://' || n.host || ':' || CAST(n.port AS TEXT) AS node_url,
  g.name AS group_name,
  u.username,
  u.full_name,
  u.default_password
FROM classes c
JOIN groups g
  ON g.class_id = c.class_id
JOIN users u
  ON u.group_id = g.group_id
JOIN group_assignments ga
  ON ga.group_id = g.group_id
JOIN nodes n
  ON n.node_id = ga.node_id
WHERE c.cluster_id = ?
  AND c.name = ?
ORDER BY n.node_id, g.group_id, u.user_id;
`, clusterID, className)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()

	results := make([]NodeGroupsForClass, 0)
	var currentNodeURL string
	var currentNodeGroups *NodeGroupsForClass

	for rows.Next() {
		var nodeURL string
		var groupName, username, fullName, password string
		err := rows.Scan(&nodeURL, &groupName, &username, &fullName, &password)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if currentNodeURL != nodeURL {
			if currentNodeGroups != nil {
				results = append(results, *currentNodeGroups)
			}
			currentNodeGroups = &NodeGroupsForClass{
				NodeURL: nodeURL,
				Groups:  make([]GroupData, 0),
			}
			currentNodeURL = nodeURL
		}

		var groupIndex = -1
		for i, group := range currentNodeGroups.Groups {
			if group.Name == groupName {
				groupIndex = i
				break
			}
		}

		var group GroupData
		if groupIndex == -1 {
			group = GroupData{
				Name:     groupName,
				Students: make([]UserData, 0),
			}
			currentNodeGroups.Groups = append(currentNodeGroups.Groups, group)
			groupIndex = len(currentNodeGroups.Groups) - 1
		} else {
			group = currentNodeGroups.Groups[groupIndex]
		}

		userData := UserData{
			Username:  username,
			FullName:  fullName,
			Email:     "",
			Password:  password,
			GroupName: groupName,
		}
		group.Students = append(group.Students, userData)
		currentNodeGroups.Groups[groupIndex] = group
	}

	if currentNodeGroups != nil {
		results = append(results, *currentNodeGroups)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}
	return results, nil
}

func DeleteExerciseRecord(conn *sql.DB, projectUUID string) error {
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	if projectUUID == "" {
		return fmt.Errorf("project UUID cannot be empty")
	}

	result, err := conn.Exec(`
        DELETE FROM exercises 
        WHERE project_uuid = ?`,
		projectUUID)
	if err != nil {
		return fmt.Errorf("failed to delete exercise record: %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	return nil
}

func InsertExerciseRecord(conn *sql.DB, projectUUID, groupName, exerciseName, state string) error {
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var groupID int
	err = tx.QueryRow(`
        SELECT g.group_id 
        FROM groups g
        JOIN classes c ON g.class_id = c.class_id
        WHERE g.name = ?`, groupName).Scan(&groupID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("group not found: %s", groupName)
		}
		return fmt.Errorf("failed to get group ID: %w", err)
	}

	if len(projectUUID) > 8 {
		projectUUID = projectUUID[:8]
	} else if len(projectUUID) < 8 {
		return fmt.Errorf("project UUID must be at least 8 characters")
	}

	_, err = tx.Exec(`
        INSERT INTO exercises (project_uuid, group_id, name, state)
        VALUES (?, ?, ?, ?)`,
		projectUUID, groupID, exerciseName, state)

	if err != nil {
		return fmt.Errorf("failed to insert exercise record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetNodeExercisesForClass(conn *sql.DB, clusterID int, className string) ([]NodeExercisesForClass, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if strings.TrimSpace(className) != "" {
		rows, err = conn.Query(`
SELECT
  n.protocol || '://' || n.host || ':' || CAST(n.port AS TEXT) AS node_url,
  e.name AS exercise_name,
  e.project_uuid,
  g.name AS group_name,
  e.state
FROM classes c
JOIN groups g ON g.class_id = c.class_id
JOIN exercises e ON e.group_id = g.group_id
JOIN group_assignments ga ON ga.group_id = g.group_id
JOIN nodes n ON n.node_id = ga.node_id
WHERE c.cluster_id = ?
  AND c.name = ?
  AND e.state <> 'deleted'
ORDER BY n.node_id, g.group_id, e.exercise_id;`, clusterID, className)
	} else {
		rows, err = conn.Query(`
SELECT
  n.protocol || '://' || n.host || ':' || CAST(n.port AS TEXT) AS node_url,
  e.name AS exercise_name,
  e.project_uuid,
  g.name AS group_name,
  e.state
FROM classes c
JOIN groups g ON g.class_id = c.class_id
JOIN exercises e ON e.group_id = g.group_id
JOIN group_assignments ga ON ga.group_id = g.group_id
JOIN nodes n ON n.node_id = ga.node_id
WHERE c.cluster_id = ?
  AND e.state <> 'deleted'
ORDER BY n.node_id, g.group_id, e.exercise_id;`, clusterID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query exercises: %w", err)
	}
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()

	results := make([]NodeExercisesForClass, 0)
	var current *NodeExercisesForClass
	var currentURL string

	for rows.Next() {
		var nodeURL, exName, projUUID, groupName, state string
		if err := rows.Scan(&nodeURL, &exName, &projUUID, &groupName, &state); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if nodeURL != currentURL {
			if current != nil {
				results = append(results, *current)
			}
			current = &NodeExercisesForClass{NodeURL: nodeURL, Exercises: make([]ExerciseItem, 0)}
			currentURL = nodeURL
		}
		current.Exercises = append(current.Exercises, ExerciseItem{
			Name:        exName,
			ProjectUUID: projUUID,
			GroupName:   groupName,
			State:       state,
		})
	}
	if current != nil {
		results = append(results, *current)
	}
	return results, rows.Err()
}
