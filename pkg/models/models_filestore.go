package models

import "time"

type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusUploading  FileStatus = "uploading"
	FileStatusAvailable  FileStatus = "available"
	FileStatusTombstoned FileStatus = "tombstoned"
)

type InitUploadRequest struct {
	Filename        string `json:"filename"`
	SizeBytes       int64  `json:"size_bytes"`
	ContentType     string `json:"content_type"`
	ScopeLabel      string `json:"scope_label"`
	RetentionPeriod int64  `json:"retention_period"`
}

type InitUploadResponse struct {
	FileUUID  string     `json:"file_uuid"`
	Status    FileStatus `json:"status"`
	UploadURL string     `json:"upload_url"`
	ExpiresAt time.Time  `json:"expires_at"`
}

type FinalizeUploadResponse struct {
	FileUUID       string     `json:"file_uuid"`
	Status         FileStatus `json:"status"`
	ChecksumSHA256 string     `json:"checksum_sha256"`
	SizeBytes      int64      `json:"size_bytes"`
}

type GetUploadStatusResponse struct {
	Offset int64 `json:"offset"`
}
