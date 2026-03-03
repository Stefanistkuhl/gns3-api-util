CREATE TABLE IF NOT EXISTS files (
    file_uuid TEXT PRIMARY KEY NOT NULL,
    file_path TEXT NOT NULL,
    filename TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    content_type TEXT NOT NULL,
    scope_label TEXT NOT NULL,
    owner_id TEXT NOT NULL,
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
    retention_period INTEGER DEFAULT NULL -- Seconds until expiry
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
    backup_type TEXT CHECK(FORMAT IN ('full', 'incremental')) NOT NULL,
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

CREATE INDEX IF NOT EXISTS idx_files_scope ON files(scope_label);

CREATE INDEX IF NOT EXISTS idx_files_status ON files(STATUS);

CREATE INDEX IF NOT EXISTS idx_vm_format ON vm_images(format);
