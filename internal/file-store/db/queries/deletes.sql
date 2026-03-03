-- name: DeleteFilePermanent :exec
DELETE FROM
    files
WHERE
    file_uuid = ?;
