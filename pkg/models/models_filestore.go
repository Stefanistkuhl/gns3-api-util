package models

import "time"

type FileStatus string

const (
	GlobalBucketID = "00000000-0000-0000-0000-000000000000"
	SystemOwnerID  = "system"
)

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusUploading  FileStatus = "uploading"
	FileStatusAvailable  FileStatus = "available"
	FileStatusTombstoned FileStatus = "tombstoned"
)

func (f FileStatus) String() string {
	return string(f)
}

type InitUploadRequest struct {
	Filename        string `json:"filename"`
	SizeBytes       int64  `json:"size_bytes"`
	ContentType     string `json:"content_type"`
	ScopeLabel      string `json:"scope_label"`
	RetentionPeriod int64  `json:"retention_period"`
	BucketID        string `json:"bucket_id,omitempty"`
}

// TODO: add back vms and backups

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

type CreateBucketRequest struct {
	Name           string `json:"name"`
	IsPublic       bool   `json:"is_public,omitempty"`
	RequiredScopes string `json:"required_scopes,omitempty"`
}

type CreateBucketResponse struct {
	BucketID       string    `json:"bucket_id"`
	Name           string    `json:"name"`
	IsPublic       bool      `json:"is_public"`
	RequiredScopes string    `json:"required_scopes"`
	CreatedAt      time.Time `json:"created_at"`
}

type FileInfo struct {
	FileUUID    string    `json:"file_uuid"`
	Filename    string    `json:"filename"`
	SizeBytes   int64     `json:"size_bytes"`
	ContentType string    `json:"content_type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type PublicTokenResponse struct {
	Token     string    `json:"token"`
	FileUUID  string    `json:"file_uuid"`
	ExpiresAt time.Time `json:"expires_at"`
	URL       string    `json:"url"`
}

type DeleteFileResponse struct {
	FileUUID string `json:"file_uuid"`
	Status   string `json:"status"`
}

type DeleteBucketResponse struct {
	BucketUUID string `json:"bucket_uuid"`
	Status     string `json:"status"`
}

type ListBucketResponse struct {
	Buckets []CreateBucketResponse `json:"buckets"`
	Count   int                    `json:"count"`
}

type ListBucketFilesResponse struct {
	BucketUUID string     `json:"bucket_uuid"`
	Files      []FileInfo `json:"files"`
	Count      int        `json:"count"`
}
