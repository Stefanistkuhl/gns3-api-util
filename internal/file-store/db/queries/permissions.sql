-- name: InsertBucketPermission :one
INSERT INTO bucket_permissions (
    id,
    bucket_id,
    principal_type,
    principal_id,
    permission,
    granted_by,
    expires_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: DeleteBucketPermission :exec
DELETE FROM bucket_permissions
WHERE id = ? AND bucket_id = ?;

-- name: ListBucketPermissions :many
SELECT
    id,
    bucket_id,
    principal_type,
    principal_id,
    permission,
    granted_by,
    created_at,
    expires_at
FROM bucket_permissions
WHERE bucket_id = ?
ORDER BY created_at DESC;

-- name: GetBucketPermission :one
SELECT
    id,
    bucket_id,
    principal_type,
    principal_id,
    permission,
    granted_by,
    created_at,
    expires_at
FROM bucket_permissions
WHERE id = ?;

-- name: HasBucketPermission :one
SELECT COUNT(*) > 0 AS has_permission
FROM bucket_permissions
WHERE bucket_id = ?
  AND principal_type = 'user'
  AND principal_id = ?
  AND permission = ?
  AND (expires_at IS NULL OR expires_at = '' OR expires_at > CURRENT_TIMESTAMP);


-- name: InsertFilePermission :one
INSERT INTO file_permissions (
    id,
    file_uuid,
    principal_type,
    principal_id,
    permission,
    granted_by,
    expires_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: DeleteFilePermission :exec
DELETE FROM file_permissions
WHERE id = ? AND file_uuid = ?;

-- name: ListFilePermissions :many
SELECT
    id,
    file_uuid,
    principal_type,
    principal_id,
    permission,
    granted_by,
    created_at,
    expires_at
FROM file_permissions
WHERE file_uuid = ?
ORDER BY created_at DESC;

-- name: GetFilePermission :one
SELECT
    id,
    file_uuid,
    principal_type,
    principal_id,
    permission,
    granted_by,
    created_at,
    expires_at
FROM file_permissions
WHERE id = ?;

-- name: HasFilePermission :one
SELECT COUNT(*) > 0 AS has_permission
FROM file_permissions
WHERE file_uuid = ?
  AND principal_type = 'user'
  AND principal_id = ?
  AND permission = ?
  AND (expires_at IS NULL OR expires_at = '' OR expires_at > CURRENT_TIMESTAMP);

