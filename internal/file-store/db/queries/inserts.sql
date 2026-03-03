-- name: CreateFile :one
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
        STATUS,
        retention_period
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING
    *;

-- name: LinkVMImage :exec
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
