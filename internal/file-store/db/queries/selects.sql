-- name: GetFileByUUID :one
SELECT
    file_uuid,
    blob_sha256,
    filename,
    content_type,
    owner_id,
    bucket_id,
    created_at,
    updated_at,
    last_accessed_at,
    STATUS,
    retention_period
FROM
    files
WHERE
    file_uuid = ?;

-- name: ListFiles :many
SELECT
    file_uuid,
    blob_sha256,
    filename,
    content_type,
    owner_id,
    bucket_id,
    created_at,
    updated_at,
    last_accessed_at,
    STATUS,
    retention_period
FROM
    files
ORDER BY
    created_at DESC;

-- name: ListFilesByOwner :many
SELECT
    file_uuid,
    blob_sha256,
    filename,
    content_type,
    owner_id,
    bucket_id,
    created_at,
    updated_at,
    last_accessed_at,
    STATUS,
    retention_period
FROM
    files
WHERE
    owner_id = ?
ORDER BY
    created_at DESC;

-- name: ListFilesByBucket :many
SELECT
    file_uuid,
    blob_sha256,
    filename,
    content_type,
    owner_id,
    bucket_id,
    created_at,
    updated_at,
    last_accessed_at,
    STATUS,
    retention_period
FROM
    files
WHERE
    bucket_id = ?
ORDER BY
    created_at DESC;

-- name: GetVMImageByFileUUID :one
SELECT
    file_uuid,
    virt_type,
    format,
    vcpus,
    ram_mb,
    extra_attributes_json
FROM
    vm_images
WHERE
    file_uuid = ?;

-- name: GetBackupByFileUUID :one
SELECT
    file_uuid,
    source_node_id,
    backup_type,
    is_compressed,
    is_encrypted,
    parent_backup_uuid
FROM
    backups
WHERE
    file_uuid = ?;

-- name: GetProjectFileByFileUUID :one
SELECT
    file_uuid,
    project_id,
    version_tag,
    is_read_only
FROM
    project_files
WHERE
    file_uuid = ?;

-- name: GetOwnerOfFileByUUID :one
SELECT
    owner_id
FROM
    files
WHERE
    file_uuid = ?;

-- name: GetBucketByID :one
SELECT
    bucket_id,
    name,
    owner_id,
    is_public,
    required_scopes,
    created_at,
    updated_at
FROM
    buckets
WHERE
    bucket_id = ?;

-- name: ListBucketsByOwner :many
SELECT
    bucket_id,
    name,
    owner_id,
    is_public,
    required_scopes,
    created_at,
    updated_at
FROM
    buckets
WHERE
    owner_id = ?
ORDER BY
    created_at DESC;

-- name: GetPublicFileToken :one
SELECT
    token,
    file_uuid,
    bucket_id,
    created_at,
    expires_at,
    access_count
FROM
    public_file_tokens
WHERE
    token = ?;

-- name: GetPublicFileTokens :many
SELECT
    token,
    file_uuid,
    bucket_id,
    created_at,
    expires_at,
    access_count
FROM
    public_file_tokens
WHERE
    file_uuid = ?
ORDER BY
    created_at DESC;

-- name: GetDefaultBucketForOwner :one
SELECT
    bucket_id,
    name,
    owner_id,
    is_public,
    required_scopes,
    created_at,
    updated_at
FROM
    buckets
WHERE
    owner_id = ?
    AND name = 'default'
LIMIT
    1;

-- name: GetFilesWithStatus :many
SELECT
    file_uuid,
    blob_sha256,
    filename,
    content_type,
    owner_id,
    bucket_id,
    created_at,
    updated_at,
    last_accessed_at,
    STATUS,
    retention_period
FROM
    files
WHERE
    STATUS = ?;

-- name: GetFilesWithPassedRetention :many
SELECT
    file_uuid
FROM
    files
WHERE
    retention_period > 0
    AND datetime(created_at, '+' || retention_period || ' hours') < datetime('now');

-- name: GetBlobBySHA256 :one
SELECT
    sha256,
    file_path,
    size_bytes,
    ref_count,
    created_at,
    updated_at,
    last_verified_at
FROM
    blobs
WHERE
    sha256 = ?;

-- name: GetUnreferencedBlobs :many
SELECT
    sha256,
    file_path
FROM
    blobs
WHERE
    ref_count <= 0;

-- name: ListBlobsForIndex :many
SELECT
    sha256,
    file_path,
    size_bytes,
    ref_count
FROM
    blobs;

-- name: ListFileBlobRefs :many
SELECT
    file_uuid,
    blob_sha256,
    STATUS
FROM
    files
WHERE
    blob_sha256 IS NOT NULL;

-- name: ListTmpUploadStatuses :many
SELECT
    file_uuid,
    STATUS
FROM
    files
WHERE
    STATUS IN ('pending', 'uploading');

-- name: ListFilesByBucketWithBlob :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.last_accessed_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes
FROM
    files f
    JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    f.bucket_id = ?
ORDER BY
    f.created_at DESC;

-- name: GetFileWithBlobByUUID :one
SELECT
    f.file_uuid,
    f.blob_sha256,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.last_accessed_at,
    f.status,
    f.retention_period,
    b.file_path,
    b.size_bytes,
    b.ref_count
FROM
    files f
    JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    f.file_uuid = ?;

-- name: GetBlobByFileUUID :one
SELECT
    blob_sha256,
    blobs.file_path
FROM
    files
    JOIN blobs ON blobs.sha256 = files.blob_sha256
WHERE
    file_uuid = ?;

-- name: GetProjectFileWithFile :one
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    pf.project_id,
    pf.project_name,
    pf.version_tag,
    pf.is_read_only,
    pf.include_snapshots,
    pf.include_images,
    pf.reset_mac_addresses,
    pf.keep_compute_ids,
    pf.compression
FROM
    project_files pf
    JOIN files f ON f.file_uuid = pf.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    pf.file_uuid = ?;

-- name: ListProjectFiles :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    pf.project_id,
    pf.project_name,
    pf.version_tag,
    pf.is_read_only,
    pf.include_snapshots,
    pf.include_images,
    pf.reset_mac_addresses,
    pf.keep_compute_ids,
    pf.compression
FROM
    project_files pf
    JOIN files f ON f.file_uuid = pf.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
ORDER BY
    f.created_at DESC;

-- name: ListProjectFilesByOwner :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    pf.project_id,
    pf.project_name,
    pf.version_tag,
    pf.is_read_only,
    pf.include_snapshots,
    pf.include_images,
    pf.reset_mac_addresses,
    pf.keep_compute_ids,
    pf.compression
FROM
    project_files pf
    JOIN files f ON f.file_uuid = pf.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    f.owner_id = ?
ORDER BY
    f.created_at DESC;

-- name: ListProjectFilesByProject :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    pf.project_id,
    pf.project_name,
    pf.version_tag,
    pf.is_read_only,
    pf.include_snapshots,
    pf.include_images,
    pf.reset_mac_addresses,
    pf.keep_compute_ids,
    pf.compression
FROM
    project_files pf
    JOIN files f ON f.file_uuid = pf.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    pf.project_id = ?
ORDER BY
    f.created_at DESC;

-- name: GetVMImageWithFile :one
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    v.virt_type,
    v.format,
    v.vcpus,
    v.ram_mb,
    v.extra_attributes_json
FROM
    vm_images v
    JOIN files f ON f.file_uuid = v.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    v.file_uuid = ?;

-- name: ListVMImages :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    v.virt_type,
    v.format,
    v.vcpus,
    v.ram_mb,
    v.extra_attributes_json
FROM
    vm_images v
    JOIN files f ON f.file_uuid = v.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
ORDER BY
    f.created_at DESC;

-- name: ListVMImagesByOwner :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    v.virt_type,
    v.format,
    v.vcpus,
    v.ram_mb,
    v.extra_attributes_json
FROM
    vm_images v
    JOIN files f ON f.file_uuid = v.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    f.owner_id = ?
ORDER BY
    f.created_at DESC;

-- name: GetBackupWithFile :one
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    bk.source_node_id,
    bk.backup_type,
    bk.is_compressed,
    bk.is_encrypted,
    bk.parent_backup_uuid
FROM
    backups bk
    JOIN files f ON f.file_uuid = bk.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    bk.file_uuid = ?;

-- name: ListBackups :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    bk.source_node_id,
    bk.backup_type,
    bk.is_compressed,
    bk.is_encrypted,
    bk.parent_backup_uuid
FROM
    backups bk
    JOIN files f ON f.file_uuid = bk.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
ORDER BY
    f.created_at DESC;

-- name: ListBackupsByOwner :many
SELECT
    f.file_uuid,
    f.filename,
    f.content_type,
    f.owner_id,
    f.bucket_id,
    f.created_at,
    f.updated_at,
    f.status,
    f.retention_period,
    f.blob_sha256,
    b.size_bytes,
    bk.source_node_id,
    bk.backup_type,
    bk.is_compressed,
    bk.is_encrypted,
    bk.parent_backup_uuid
FROM
    backups bk
    JOIN files f ON f.file_uuid = bk.file_uuid
    LEFT JOIN blobs b ON b.sha256 = f.blob_sha256
WHERE
    f.owner_id = ?
ORDER BY
    f.created_at DESC;
