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
