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
