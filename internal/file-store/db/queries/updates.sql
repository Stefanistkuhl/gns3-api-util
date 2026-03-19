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

-- name: IncrementTokenAccessCount :exec
UPDATE
    public_file_tokens
SET
    access_count = access_count + 1
WHERE
    token = ?;
