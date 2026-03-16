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
    scope_label = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: FinalizeFile :one
UPDATE
    files
SET
    checksum_sha256 = ?,
    size_bytes = ?,
    file_path = ?,
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
