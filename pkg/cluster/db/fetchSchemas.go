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

type CreateClustersAndNodes struct {
	Clusters []ClusterAndNodes
}

type ClusterAndNodes struct {
	Cluster ClusterName
	Nodes   []NodeDataAll
}

type NodeAndGroupIds struct {
	Assignments []AssignmentIds
}

type AssignmentIds struct {
	NodeID   int
	GroupIDs []int
}

type InsertedClassData struct {
	ClassID  int
	GroupIDs map[string]int
	UserIDs  []int
}
type UserData struct {
	Username  string
	FullName  string
	Email     string
	Password  string
	GroupName string
}

type GroupData struct {
	Name     string
	Students []UserData
}

type NodeGroupsForClass struct {
	NodeURL string
	Groups  []GroupData
}

type ExerciseItem struct {
	Name        string
	ProjectUUID string
	GroupName   string
	State       string
}

type NodeExercisesForClass struct {
	NodeURL   string
	Exercises []ExerciseItem
}
