package handlers

import "github.com/0xveya/gns3util/pkg/models"

type DownloadObject struct {
	Path        string
	Size        int64
	ContentType string
	Filename    string
	ETag        string
	FileStatus  models.FileStatus
}

type UploadObject struct {
	FileUUID    string
	TempPath    string
	FinalPath   string
	FinalHash   string
	Offset      int64
	Written     int64
	TotalSize   int64
	ContentType string
}

type UploadResult struct {
	FileUUID  string
	BlobHash  string
	SizeBytes int64
}
