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
	FileStatusDeleted    FileStatus = "deleted"
)

func (f FileStatus) String() string {
	return string(f)
}

type InitUploadRequest struct {
	Filename        string `json:"filename"`
	SizeBytes       int64  `json:"size_bytes"`
	ContentType     string `json:"content_type"`
	RetentionPeriod *int64 `json:"retention_period"`
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
	FileUUID   string     `json:"file_uuid"`
	Status     FileStatus `json:"status"`
	BlobSHA256 string     `json:"blob_sha256"`
	SizeBytes  int64      `json:"size_bytes"`
}

type GetUploadStatusResponse struct {
	Offset int64 `json:"offset"`
}

type CreateBucketRequest struct {
	Name           string  `json:"name"`
	IsPublic       bool    `json:"is_public,omitempty"`
	RequiredScopes *string `json:"required_scopes,omitempty"`
}

type CreateBucketResponse struct {
	BucketID       string    `json:"bucket_id"`
	Name           string    `json:"name"`
	IsPublic       bool      `json:"is_public"`
	RequiredScopes string    `json:"required_scopes"`
	CreatedAt      time.Time `json:"created_at"`
}

type FileInfo struct {
	FileUUID    string     `json:"file_uuid"`
	Filename    string     `json:"filename"`
	SizeBytes   int64      `json:"size_bytes"`
	ContentType string     `json:"content_type"`
	BlobSHA256  string     `json:"blob_sha256"`
	Status      FileStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PublicTokenRequest struct {
	FileUUID   string     `json:"file_uuid"`
	BucketUUID string     `json:"bucket_uuid"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

type PublicTokenResponse struct {
	Token      string `json:"token"`
	FileUUID   string `json:"file_uuid"`
	BucketUUID string `json:"bucket_uuid"`
	ExpiresAt  string `json:"expires_at"`
	URL        string `json:"url"`
}

type DeleteFileResponse struct {
	FileUUID string     `json:"file_uuid"`
	Status   FileStatus `json:"status"`
}

type UpdateBucketRequest struct {
	// Name is the new display name for the bucket. Required.
	Name string `json:"name"`
	// IsPublic controls whether the bucket is publicly readable.
	IsPublic bool `json:"is_public"`
	// RequiredScopes is an optional comma-separated list of role names a
	// bearer must hold to access this bucket.
	RequiredScopes *string `json:"required_scopes,omitempty"`
}

type UpdateBucketResponse struct {
	BucketID       string    `json:"bucket_id"`
	Name           string    `json:"name"`
	IsPublic       bool      `json:"is_public"`
	RequiredScopes string    `json:"required_scopes"`
	UpdatedAt      time.Time `json:"updated_at"`
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

type BlobInfo struct {
	SHA256         string     `json:"sha256"`
	FilePath       string     `json:"file_path"`
	SizeBytes      int64      `json:"size_bytes"`
	RefCount       int64      `json:"ref_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`
}

type FileDetailResponse struct {
	FileUUID    string     `json:"file_uuid"`
	Filename    string     `json:"filename"`
	SizeBytes   int64      `json:"size_bytes"`
	ContentType string     `json:"content_type"`
	BlobSHA256  string     `json:"blob_sha256"`
	FilePath    string     `json:"file_path"`
	Status      FileStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type GrantPermissionRequest struct {
	PrincipalType string `json:"principal_type"` // "user", "group", or "role"
	PrincipalID   string `json:"principal_id"`
	Permission    string `json:"permission"` // "read", "write", or "admin"
	ExpiresAt     string `json:"expires_at"` // optional RFC3339 or empty
}

type PermissionEntry struct {
	ID            string `json:"id"`
	PrincipalType string `json:"principal_type"`
	PrincipalID   string `json:"principal_id"`
	Permission    string `json:"permission"`
	GrantedBy     string `json:"granted_by"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

type ListPermissionsResponse struct {
	Permissions []PermissionEntry `json:"permissions"`
	Count       int               `json:"count"`
}

type RevokePermissionResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}
