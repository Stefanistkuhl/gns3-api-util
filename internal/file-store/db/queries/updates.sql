-- name: UpdateFileStatus :exec
UPDATE
    files
SET
    STATUS = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: UpdateFileLastAccessedAt :exec
UPDATE
    files
SET
    last_accessed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: UpdateFileMetadata :exec
UPDATE
    files
SET
    filename = ?,
    content_type = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: FinalizeFile :one
UPDATE
    files
SET
    blob_sha256 = ?,
    STATUS = 'available',
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?
RETURNING
    *;

-- name: MarkFileTombstoned :exec
UPDATE
    files
SET
    STATUS = 'tombstoned',
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: UpdateFileBucket :exec
UPDATE
    files
SET
    bucket_id = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: UpdateBucket :exec
UPDATE
    buckets
SET
    name = ?,
    is_public = ?,
    required_scopes = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    bucket_id = ?;

-- name: IncrementBlobRefCount :exec
UPDATE
    blobs
SET
    ref_count = ref_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE
    sha256 = ?;

-- name: DecrementBlobRefCount :exec
UPDATE
    blobs
SET
    ref_count = ref_count - 1,
    updated_at = CURRENT_TIMESTAMP
WHERE
    sha256 = ?;

-- name: UpdateBlobIndexMetadata :exec
UPDATE
    blobs
SET
    file_path = ?,
    size_bytes = ?,
    ref_count = ?,
    last_verified_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE
    sha256 = ?;

-- name: TombstoneFilesByBlobSHA :exec
UPDATE
    files
SET
    STATUS = 'tombstoned',
    updated_at = CURRENT_TIMESTAMP
WHERE
    blob_sha256 = ?
    AND STATUS != 'tombstoned';

-- name: IncrementTokenAccessCount :exec
UPDATE
    public_file_tokens
SET
    access_count = access_count + 1
WHERE
    token = ?;

-- name: UpdateProjectFile :exec
UPDATE
    project_files
SET
    project_name = ?,
    version_tag = ?,
    is_read_only = ?,
    include_snapshots = ?,
    include_images = ?,
    reset_mac_addresses = ?,
    keep_compute_ids = ?,
    compression = ?
WHERE
    file_uuid = ?;

-- name: UpdateVMImage :exec
UPDATE
    vm_images
SET
    virt_type = ?,
    format = ?,
    vcpus = ?,
    ram_mb = ?,
    extra_attributes_json = ?
WHERE
    file_uuid = ?;

-- name: UpdateBackup :exec
UPDATE
    backups
SET
    source_node_id = ?,
    backup_type = ?,
    is_compressed = ?,
    is_encrypted = ?,
    parent_backup_uuid = ?
WHERE
    file_uuid = ?;
