-- name: DeleteFile :exec
DELETE FROM
    files
WHERE
    file_uuid = ?;

-- name: DeleteBucket :exec
DELETE FROM
    buckets
WHERE
    bucket_id = ?;

-- name: DeletePublicFileToken :exec
DELETE FROM
    public_file_tokens
WHERE
    token = ?;

-- name: DeleteExpiredPublicTokens :exec
DELETE FROM
    public_file_tokens
WHERE
    expires_at IS NOT NULL
    AND expires_at < CURRENT_TIMESTAMP;

-- name: DeleteVMImage :exec
DELETE FROM
    vm_images
WHERE
    file_uuid = ?;

-- name: DeleteBackup :exec
DELETE FROM
    backups
WHERE
    file_uuid = ?;

-- name: DeleteProjectFile :exec
DELETE FROM
    project_files
WHERE
    file_uuid = ?;

-- name: DeleteBlob :exec
DELETE FROM
    blobs
WHERE
    sha256 = ?;

-- name: DeleteBlobIfUnreferenced :exec
DELETE FROM
    blobs
WHERE
    sha256 = ?
    AND ref_count <= 0;
