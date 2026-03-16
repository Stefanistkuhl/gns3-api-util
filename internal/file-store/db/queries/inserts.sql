-- name: InitFile :one
INSERT INTO
    files (
        file_uuid,
        filename,
        size_bytes,
        content_type,
        scope_label,
        owner_id,
        bucket_id,
        last_accessed_at,
        STATUS,
        retention_period,
        file_path,
        checksum_sha256
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, 'uploading', ?, '', '')
RETURNING
    *;

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
        bucket_id,
        last_accessed_at,
        STATUS,
        retention_period
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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

-- name: CreatePublicFileToken :one
INSERT INTO
    public_file_tokens (
        token,
        file_uuid,
        bucket_id,
        expires_at,
        access_count
    )
VALUES
    (?, ?, ?, ?, 0)
RETURNING
    *;

-- name: IncrementTokenAccessCount :exec
UPDATE
    public_file_tokens
SET
    access_count = access_count + 1,
    last_accessed_at = CURRENT_TIMESTAMP
WHERE
    token = ?;

-- name: CreateBucket :one
INSERT INTO
    buckets (
        bucket_id,
        name,
        owner_id,
        is_public,
        required_scopes,
        bucket_type
    )
VALUES
    (?, ?, ?, ?, ?, ?)
RETURNING
    *;
