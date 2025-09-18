package db

import "database/sql"

type ClusterName struct {
	Id   int
	Name string
	Desc sql.NullString
}

type NodeData struct {
	User      string
	Protocol  string
	Host      string
	Port      int
	Weight    int
	MaxGroups int
}

type NodeDataAll struct {
	ID        int
	ClusterID int
	User      string
	Protocol  string
	Host      string
	Port      int
	Weight    int
	MaxGroups int
}
