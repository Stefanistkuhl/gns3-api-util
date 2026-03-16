-- name: GetFileByUUID :one
SELECT
    file_uuid,
    file_path,
    filename,
    size_bytes,
    checksum_sha256,
    content_type,
    scope_label,
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
    file_path,
    filename,
    size_bytes,
    checksum_sha256,
    content_type,
    scope_label,
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
    file_path,
    filename,
    size_bytes,
    checksum_sha256,
    content_type,
    scope_label,
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

-- name: ListFilesByScope :many
SELECT
    file_uuid,
    file_path,
    filename,
    size_bytes,
    checksum_sha256,
    content_type,
    scope_label,
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
    scope_label = ?
ORDER BY
    created_at DESC;

-- name: ListFilesByBucket :many
SELECT
    file_uuid,
    file_path,
    filename,
    size_bytes,
    checksum_sha256,
    content_type,
    scope_label,
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
