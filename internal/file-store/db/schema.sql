CREATE TABLE IF NOT EXISTS blobs (
    sha256 TEXT PRIMARY KEY NOT NULL,
    file_path TEXT UNIQUE NOT NULL,
    size_bytes INTEGER NOT NULL,
    ref_count INTEGER NOT NULL DEFAULT 0 CHECK(ref_count >= 0),
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    last_verified_at TEXT
);

CREATE TABLE IF NOT EXISTS files (
    file_uuid TEXT PRIMARY KEY NOT NULL,
    blob_sha256 TEXT,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    bucket_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    last_accessed_at TEXT,
    STATUS TEXT NOT NULL CHECK(
        STATUS IN (
            'pending',
            'uploading',
            'available',
            'tombstoned'
        )
    ) DEFAULT 'pending',
    retention_period INTEGER DEFAULT NULL CHECK (
        retention_period IS NULL
        OR retention_period >= 0
    ),
    FOREIGN KEY(blob_sha256) REFERENCES blobs(sha256) ON DELETE RESTRICT,
    FOREIGN KEY(bucket_id) REFERENCES buckets(bucket_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS vm_images (
    file_uuid TEXT PRIMARY KEY NOT NULL,
    virt_type TEXT CHECK(
        virt_type IN (
            'qemu',
            'iou',
            'docker',
            'dynamips',
            'vmware',
            'virtualbox'
        )
    ) NOT NULL,
    format TEXT NOT NULL,
    vcpus INTEGER DEFAULT 1,
    ram_mb INTEGER DEFAULT 512,
    extra_attributes_json TEXT,
    FOREIGN KEY(file_uuid) REFERENCES files(file_uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS backups (
    file_uuid TEXT PRIMARY KEY NOT NULL,
    source_node_id TEXT,
    backup_type TEXT CHECK(
        backup_type IN ('full', 'incremental')
    ) NOT NULL,
    is_compressed BOOLEAN DEFAULT TRUE,
    is_encrypted BOOLEAN DEFAULT FALSE,
    parent_backup_uuid TEXT,
    FOREIGN KEY(file_uuid) REFERENCES files(file_uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_files (
    file_uuid TEXT PRIMARY KEY NOT NULL,
    project_id TEXT NOT NULL,
    version_tag TEXT,
    is_read_only BOOLEAN DEFAULT FALSE,
    FOREIGN KEY(file_uuid) REFERENCES files(file_uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS buckets (
    bucket_id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    bucket_type TEXT NOT NULL DEFAULT 'standard',
    is_public BOOLEAN DEFAULT FALSE,
    required_scopes TEXT DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS public_file_tokens (
    token TEXT PRIMARY KEY NOT NULL,
    file_uuid TEXT NOT NULL,
    bucket_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    expires_at TEXT,
    access_count INTEGER NOT NULL DEFAULT 0 CHECK (access_count >= 0),
    FOREIGN KEY(file_uuid) REFERENCES files(file_uuid) ON DELETE CASCADE,
    FOREIGN KEY(bucket_id) REFERENCES buckets(bucket_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bucket_permissions (
    id TEXT PRIMARY KEY NOT NULL,
    bucket_id TEXT NOT NULL,
    principal_type TEXT NOT NULL CHECK(principal_type IN ('user', 'group', 'role')),
    principal_id TEXT NOT NULL,
    permission TEXT NOT NULL CHECK(permission IN ('read', 'write', 'admin')),
    granted_by TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    expires_at TEXT,
    FOREIGN KEY(bucket_id) REFERENCES buckets(bucket_id) ON DELETE CASCADE,
    UNIQUE(
        bucket_id,
        principal_type,
        principal_id,
        permission
    )
);

CREATE TABLE IF NOT EXISTS file_permissions (
    id TEXT PRIMARY KEY NOT NULL,
    file_uuid TEXT NOT NULL,
    principal_type TEXT NOT NULL CHECK(principal_type IN ('user', 'group', 'role')),
    principal_id TEXT NOT NULL,
    permission TEXT NOT NULL CHECK(permission IN ('read', 'write', 'admin')),
    granted_by TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    expires_at TEXT,
    FOREIGN KEY(file_uuid) REFERENCES files(file_uuid) ON DELETE CASCADE,
    UNIQUE(
        file_uuid,
        principal_type,
        principal_id,
        permission
    )
);

CREATE INDEX IF NOT EXISTS idx_blob_sha256 ON blobs(sha256);

CREATE INDEX IF NOT EXISTS idx_files_blob_sha256 ON files(blob_sha256);

CREATE INDEX IF NOT EXISTS idx_files_status ON files(STATUS);

CREATE INDEX IF NOT EXISTS idx_files_bucket ON files(bucket_id);

CREATE INDEX IF NOT EXISTS idx_vm_format ON vm_images(format);

CREATE INDEX IF NOT EXISTS idx_public_tokens ON public_file_tokens(token);

CREATE INDEX IF NOT EXISTS idx_public_uuid ON public_file_tokens(file_uuid);

CREATE INDEX IF NOT EXISTS idx_blobs_ref_count ON blobs(ref_count);

CREATE INDEX IF NOT EXISTS idx_bucket_perms_bucket ON bucket_permissions(bucket_id);

CREATE INDEX IF NOT EXISTS idx_bucket_perms_principal ON bucket_permissions(principal_id);

CREATE INDEX IF NOT EXISTS idx_file_perms_file ON file_permissions(file_uuid);

CREATE INDEX IF NOT EXISTS idx_file_perms_principal ON file_permissions(principal_id);

INSERT INTO
    buckets (bucket_id, name, owner_id, is_public)
VALUES
    (
        '00000000-0000-0000-0000-000000000000',
        'global-default',
        'system',
        TRUE
    ) ON CONFLICT (bucket_id) DO NOTHING;
