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

// ---- VM Image models ----

// ValidVirtTypes lists the accepted virtualisation back-ends.
var ValidVirtTypes = []string{"qemu", "iou", "docker", "dynamips", "vmware", "virtualbox"}

// ValidBackupTypes lists accepted backup classifications.
var ValidBackupTypes = []string{"full", "incremental"}

// InitVMUploadRequest is the body for POST /api/v1/vms.
// It initialises both the file record and the vm_images metadata row
// in a single request so callers never have to deal with a bare file.
type InitVMUploadRequest struct {
	Filename            string `json:"filename"`
	SizeBytes           int64  `json:"size_bytes"`
	ContentType         string `json:"content_type"`
	RetentionPeriod     *int64 `json:"retention_period,omitempty"`
	BucketID            string `json:"bucket_id,omitempty"`
	VirtType            string `json:"virt_type"`
	Format              string `json:"format"`
	VCPUs               *int64 `json:"vcpus,omitempty"`
	RAMMB               *int64 `json:"ram_mb,omitempty"`
	ExtraAttributesJSON string `json:"extra_attributes_json,omitempty"`
}

// UpdateVMImageRequest is the body for PATCH /api/v1/vms/{file_uuid}.
// All fields are optional; only non-zero values are applied.
type UpdateVMImageRequest struct {
	VirtType            string `json:"virt_type,omitempty"`
	Format              string `json:"format,omitempty"`
	VCPUs               *int64 `json:"vcpus,omitempty"`
	RAMMB               *int64 `json:"ram_mb,omitempty"`
	ExtraAttributesJSON string `json:"extra_attributes_json,omitempty"`
}

// VMImageInfo is the combined file + vm_images view returned by the API.
type VMImageInfo struct {
	FileUUID            string     `json:"file_uuid"`
	Filename            string     `json:"filename"`
	ContentType         string     `json:"content_type"`
	OwnerID             string     `json:"owner_id"`
	BucketID            string     `json:"bucket_id"`
	SizeBytes           int64      `json:"size_bytes"`
	BlobSHA256          string     `json:"blob_sha256,omitempty"`
	Status              FileStatus `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	VirtType            string     `json:"virt_type"`
	Format              string     `json:"format"`
	VCPUs               int64      `json:"vcpus"`
	RAMMB               int64      `json:"ram_mb"`
	ExtraAttributesJSON string     `json:"extra_attributes_json,omitempty"`
}

// ListVMImagesResponse wraps a slice of VMImageInfo.
type ListVMImagesResponse struct {
	VMs   []VMImageInfo `json:"vms"`
	Count int           `json:"count"`
}

// ---- Backup models ----

// InitBackupUploadRequest is the body for POST /api/v1/backups.
type InitBackupUploadRequest struct {
	Filename         string `json:"filename"`
	SizeBytes        int64  `json:"size_bytes"`
	ContentType      string `json:"content_type"`
	RetentionPeriod  *int64 `json:"retention_period,omitempty"`
	BucketID         string `json:"bucket_id,omitempty"`
	SourceNodeID     string `json:"source_node_id,omitempty"`
	BackupType       string `json:"backup_type"` // "full" | "incremental"
	IsCompressed     bool   `json:"is_compressed"`
	IsEncrypted      bool   `json:"is_encrypted"`
	ParentBackupUUID string `json:"parent_backup_uuid,omitempty"`
}

// UpdateBackupRequest is the body for PATCH /api/v1/backups/{file_uuid}.
type UpdateBackupRequest struct {
	SourceNodeID     string `json:"source_node_id,omitempty"`
	BackupType       string `json:"backup_type,omitempty"`
	IsCompressed     *bool  `json:"is_compressed,omitempty"`
	IsEncrypted      *bool  `json:"is_encrypted,omitempty"`
	ParentBackupUUID string `json:"parent_backup_uuid,omitempty"`
}

// BackupInfo is the combined file + backups view returned by the API.
type BackupInfo struct {
	FileUUID         string     `json:"file_uuid"`
	Filename         string     `json:"filename"`
	ContentType      string     `json:"content_type"`
	OwnerID          string     `json:"owner_id"`
	BucketID         string     `json:"bucket_id"`
	SizeBytes        int64      `json:"size_bytes"`
	BlobSHA256       string     `json:"blob_sha256,omitempty"`
	Status           FileStatus `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	SourceNodeID     string     `json:"source_node_id,omitempty"`
	BackupType       string     `json:"backup_type"`
	IsCompressed     bool       `json:"is_compressed"`
	IsEncrypted      bool       `json:"is_encrypted"`
	ParentBackupUUID string     `json:"parent_backup_uuid,omitempty"`
}

// ListBackupsResponse wraps a slice of BackupInfo.
type ListBackupsResponse struct {
	Backups []BackupInfo `json:"backups"`
	Count   int          `json:"count"`
}

// ---- Project file models ----

// ValidCompressionTypes lists the compression formats supported by GNS3 project export.
var ValidCompressionTypes = []string{"deflate", "bz2", "xz", "zstd", "none"}

// InitProjectFileUploadRequest is the body for POST /api/v1/project-files.
// It reflects the options available on the GNS3 project export command.
type InitProjectFileUploadRequest struct {
	Filename          string `json:"filename"`
	SizeBytes         int64  `json:"size_bytes"`
	ContentType       string `json:"content_type"`
	RetentionPeriod   *int64 `json:"retention_period,omitempty"`
	BucketID          string `json:"bucket_id,omitempty"`
	ProjectID         string `json:"project_id"`
	ProjectName       string `json:"project_name,omitempty"`
	VersionTag        string `json:"version_tag,omitempty"`
	IsReadOnly        bool   `json:"is_read_only"`
	IncludeSnapshots  bool   `json:"include_snapshots"`
	IncludeImages     bool   `json:"include_images"`
	ResetMacAddresses bool   `json:"reset_mac_addresses"`
	KeepComputeIds    bool   `json:"keep_compute_ids"`
	Compression       string `json:"compression"` // deflate|bz2|xz|zstd|none
}

// UpdateProjectFileRequest is the body for PATCH /api/v1/project-files/{file_uuid}.
// All fields are optional; only explicitly provided values are applied.
type UpdateProjectFileRequest struct {
	ProjectName       string `json:"project_name,omitempty"`
	VersionTag        string `json:"version_tag,omitempty"`
	IsReadOnly        *bool  `json:"is_read_only,omitempty"`
	IncludeSnapshots  *bool  `json:"include_snapshots,omitempty"`
	IncludeImages     *bool  `json:"include_images,omitempty"`
	ResetMacAddresses *bool  `json:"reset_mac_addresses,omitempty"`
	KeepComputeIds    *bool  `json:"keep_compute_ids,omitempty"`
	Compression       string `json:"compression,omitempty"`
}

// ProjectFileInfo is the combined file + project_files view returned by the API.
type ProjectFileInfo struct {
	FileUUID          string     `json:"file_uuid"`
	Filename          string     `json:"filename"`
	ContentType       string     `json:"content_type"`
	OwnerID           string     `json:"owner_id"`
	BucketID          string     `json:"bucket_id"`
	SizeBytes         int64      `json:"size_bytes"`
	BlobSHA256        string     `json:"blob_sha256,omitempty"`
	Status            FileStatus `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ProjectID         string     `json:"project_id"`
	ProjectName       string     `json:"project_name,omitempty"`
	VersionTag        string     `json:"version_tag,omitempty"`
	IsReadOnly        bool       `json:"is_read_only"`
	IncludeSnapshots  bool       `json:"include_snapshots"`
	IncludeImages     bool       `json:"include_images"`
	ResetMacAddresses bool       `json:"reset_mac_addresses"`
	KeepComputeIds    bool       `json:"keep_compute_ids"`
	Compression       string     `json:"compression"`
}

// ListProjectFilesResponse wraps a slice of ProjectFileInfo.
type ListProjectFilesResponse struct {
	ProjectFiles []ProjectFileInfo `json:"project_files"`
	Count        int               `json:"count"`
}
