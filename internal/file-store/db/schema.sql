PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS files (
    file_uuid TEXT PRIMARY KEY NOT NULL,
    file_path TEXT NOT NULL DEFAULT '',
    filename TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    checksum_sha256 TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL,
    scope_label TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    bucket_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at DATETIME,
    STATUS TEXT CHECK(
        STATUS IN (
            'pending',
            'uploading',
            'available',
            'tombstoned'
        )
    ) DEFAULT 'pending',
    retention_period INTEGER DEFAULT NULL,
    FOREIGN KEY(bucket_id) REFERENCES buckets(bucket_id) ON DELETE
    SET
        NULL
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
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public_file_tokens (
    token TEXT PRIMARY KEY NOT NULL,
    file_uuid TEXT NOT NULL,
    bucket_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    access_count INTEGER DEFAULT 0,
    FOREIGN KEY(file_uuid) REFERENCES files(file_uuid) ON DELETE CASCADE,
    FOREIGN KEY(bucket_id) REFERENCES buckets(bucket_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_files_uuid ON files(file_uuid);

CREATE INDEX IF NOT EXISTS idx_files_scope ON files(scope_label);

CREATE INDEX IF NOT EXISTS idx_files_status ON files(STATUS);

CREATE INDEX IF NOT EXISTS idx_files_bucket ON files(bucket_id);

CREATE INDEX IF NOT EXISTS idx_vm_format ON vm_images(format);

CREATE INDEX IF NOT EXISTS idx_public_tokens ON public_file_tokens(token);

INSERT INTO
    buckets (bucket_id, name, owner_id, is_public)
VALUES
    (
        '00000000-0000-0000-0000-000000000000',
        'global-default',
        'system',
        TRUE
    ) ON CONFLICT (bucket_id) DO NOTHING;
