-- name: InitFile :one
INSERT INTO
    files (
        file_uuid,
        filename,
        size_bytes,
        content_type,
        scope_label,
        owner_id,
        last_accessed_at,
        STATUS,
        retention_period
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, 'uploading', ?)
RETURNING
    *;

-- name: InsertVMImage :exec
INSERT INTO
    vm_images (
        file_uuid,
        virt_type,
        format,
        vcpus,
        ram_mb,
        extra_attributes_json
    )
VALUES
    (?, ?, ?, ?, ?, ?);

-- name: InsertBackup :exec
INSERT INTO
    backups (
        file_uuid,
        source_node_id,
        backup_type,
        is_compressed,
        is_encrypted,
        parent_backup_uuid
    )
VALUES
    (?, ?, ?, ?, ?, ?);

-- name: InsertFile :exec
INSERT INTO
    files (
        file_uuid,
        file_path,
        filename,
        size_bytes,
        checksum_sha256,
        content_type,
        scope_label,
        owner_id,
        last_accessed_at,
        STATUS,
        retention_period
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertProjectFile :exec
INSERT INTO
    project_files (
        file_uuid,
        project_id,
        version_tag,
        is_read_only
    )
VALUES
    (?, ?, ?, ?);

-- name: InsertUserPermission :exec
INSERT INTO
    user_permissions (
        user_id,
        scope,
        granted_at,
        updated_at
    )
VALUES
    (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: UpsertClusterNode :exec
INSERT INTO
    cluster_nodes (
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
    )
VALUES
    (
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT(node_id) DO
UPDATE
SET
    node_name = excluded.node_name,
    node_kind = excluded.node_kind,
    api_url = excluded.api_url,
    drpc_addr = excluded.drpc_addr,
    advertise_addr = excluded.advertise_addr,
    STATUS = excluded.status,
    last_heartbeat_at = excluded.last_heartbeat_at,
    updated_at = CURRENT_TIMESTAMP;

-- name: UpsertClusterKV :exec
INSERT INTO
    cluster_kv (
        KEY,
        value,
        version,
        updated_at
    )
VALUES
    (?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(KEY) DO
UPDATE
SET
    value = excluded.value,
    version = excluded.version,
    updated_at = CURRENT_TIMESTAMP;

-- name: RevokeToken :exec
INSERT INTO
    revoked_tokens (
        jti,
        user_id,
        revoked_at,
        expires_at
    )
VALUES
    (?, ?, CURRENT_TIMESTAMP, ?) ON CONFLICT(jti) DO
UPDATE
SET
    user_id = excluded.user_id,
    expires_at = excluded.expires_at;

-- name: UpsertSyncState :exec
INSERT INTO
    sync_state (
        replica_name,
        last_full_sync_at,
        last_incremental_sync_at,
        last_source_revision,
        last_status,
        last_error
    )
VALUES
    (?, ?, ?, ?, ?, ?) ON CONFLICT(replica_name) DO
UPDATE
SET
    last_full_sync_at = excluded.last_full_sync_at,
    last_incremental_sync_at = excluded.last_incremental_sync_at,
    last_source_revision = excluded.last_source_revision,
    last_status = excluded.last_status,
    last_error = excluded.last_error;
