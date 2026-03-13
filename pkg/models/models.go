package models

type NodeType int

const (
	NodeTypeMaster NodeType = iota
	NodeTypeWorker
	NodeTypeFilestore
)
