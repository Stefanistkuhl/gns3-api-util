-- name: InitFile :one
INSERT INTO
    files (
        file_uuid,
        filename,
        content_type,
        owner_id,
        bucket_id,
        last_accessed_at,
        STATUS,
        retention_period
    )
VALUES
    (?, ?, ?, ?, ?, ?, 'pending', ?)
RETURNING
    *;

-- name: InsertFile :exec
INSERT INTO
    files (
        file_uuid,
        blob_sha256,
        filename,
        content_type,
        owner_id,
        bucket_id,
        last_accessed_at,
        STATUS,
        retention_period
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?);

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
        project_name,
        version_tag,
        is_read_only,
        include_snapshots,
        include_images,
        reset_mac_addresses,
        keep_compute_ids,
        compression
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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

-- name: UpsertBlob :exec
INSERT INTO
    blobs (
        sha256,
        file_path,
        size_bytes,
        ref_count
    )
VALUES
    (?, ?, ?, 1) ON CONFLICT(sha256) DO
UPDATE
SET
    ref_count = ref_count + 1,
    updated_at = CURRENT_TIMESTAMP;
