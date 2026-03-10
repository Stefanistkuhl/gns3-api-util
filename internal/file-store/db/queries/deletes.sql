-- name: DeleteFile :exec
DELETE FROM
    files
WHERE
    file_uuid = ?;
