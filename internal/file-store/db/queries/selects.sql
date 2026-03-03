-- name: GetFileWithVMData :one
SELECT
    f.*,
    v.virt_type,
    v.format,
    v.vcpus,
    v.ram_mb,
    v.extra_attributes_json
FROM
    files f
    LEFT JOIN vm_images v ON f.file_uuid = v.file_uuid
WHERE
    f.file_uuid = ?
LIMIT
    1;

-- name: ListFilesByScope :many
SELECT
    *
FROM
    files
WHERE
    scope_label = ?
    AND STATUS = 'available'
ORDER BY
    created_at DESC;
