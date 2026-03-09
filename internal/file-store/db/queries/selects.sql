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

-- name: ListUserPermissions :many
SELECT
    user_id,
    scope,
    granted_at,
    updated_at
FROM
    user_permissions
WHERE
    user_id = ?
ORDER BY
    scope ASC;

-- name: UserHasPermission :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            user_permissions
        WHERE
            user_id = ?
            AND scope = ?
    ) AS has_permission;

-- name: UserIsSuperuser :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            user_permissions
        WHERE
            user_id = ?
            AND scope = 'superuser'
    ) AS is_superuser;

-- name: ListAllUserPermissions :many
SELECT
    user_id,
    scope,
    granted_at,
    updated_at
FROM
    user_permissions
ORDER BY
    user_id ASC,
    scope ASC;

-- name: GetClusterNode :one
SELECT
    node_id,
    node_name,
    node_kind,
    api_url,
    drpc_addr,
    advertise_addr,
    STATUS,
    last_heartbeat_at,
    created_at,
    updated_at
FROM
    cluster_nodes
WHERE
    node_id = ?;

-- name: ListClusterNodes :many
SELECT
    node_id,
    node_name,
    node_kind,
    api_url,
    drpc_addr,
    advertise_addr,
    STATUS,
    last_heartbeat_at,
    created_at,
    updated_at
FROM
    cluster_nodes
ORDER BY
    node_name ASC;

-- name: ListClusterNodesByKind :many
SELECT
    node_id,
    node_name,
    node_kind,
    api_url,
    drpc_addr,
    advertise_addr,
    STATUS,
    last_heartbeat_at,
    created_at,
    updated_at
FROM
    cluster_nodes
WHERE
    node_kind = ?
ORDER BY
    node_name ASC;

-- name: GetClusterKV :one
SELECT
    KEY,
    value,
    version,
    updated_at
FROM
    cluster_kv
WHERE
    KEY = ?;

-- name: ListClusterKV :many
SELECT
    KEY,
    value,
    version,
    updated_at
FROM
    cluster_kv
ORDER BY
    KEY ASC;

-- name: ListClusterKVByPrefix :many
SELECT
    KEY,
    value,
    version,
    updated_at
FROM
    cluster_kv
WHERE
    KEY LIKE (? || '%')
ORDER BY
    KEY ASC;

-- name: IsTokenRevoked :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            revoked_tokens
        WHERE
            jti = ?
    ) AS is_revoked;

-- name: GetSyncState :one
SELECT
    replica_name,
    last_full_sync_at,
    last_incremental_sync_at,
    last_source_revision,
    last_status,
    last_error
FROM
    sync_state
WHERE
    replica_name = ?;

-- name: GetOwnerOfFileByUUID :one
SELECT
    owner_id
FROM
    files
WHERE
    file_uuid = ?;
