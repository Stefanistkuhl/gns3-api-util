-- name: DeleteFile :exec
DELETE FROM
    files
WHERE
    file_uuid = ?;

-- name: DeleteUserPermission :exec
DELETE FROM
    user_permissions
WHERE
    user_id = ?
    AND scope = ?;

-- name: DeleteAllUserPermissionsForUser :exec
DELETE FROM
    user_permissions
WHERE
    user_id = ?;

-- name: DeleteAllUserPermissions :exec
DELETE FROM
    user_permissions;

-- name: DeleteClusterNode :exec
DELETE FROM
    cluster_nodes
WHERE
    node_id = ?;

-- name: DeleteAllClusterNodes :exec
DELETE FROM
    cluster_nodes;

-- name: DeleteClusterKV :exec
DELETE FROM
    cluster_kv
WHERE
    KEY = ?;

-- name: DeleteAllClusterKV :exec
DELETE FROM
    cluster_kv;

-- name: DeleteExpiredRevokedTokens :exec
DELETE FROM
    revoked_tokens
WHERE
    expires_at IS NOT NULL
    AND expires_at < CURRENT_TIMESTAMP;
