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
