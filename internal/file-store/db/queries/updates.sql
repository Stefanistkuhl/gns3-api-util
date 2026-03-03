-- name: UpdateFileStatus :exec
UPDATE
    files
SET
    STATUS = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: MarkLastAccessed :exec
UPDATE
    files
SET
    last_accessed_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;

-- name: TombstoneFile :exec
UPDATE
    files
SET
    STATUS = 'tombstoned',
    updated_at = CURRENT_TIMESTAMP
WHERE
    file_uuid = ?;
