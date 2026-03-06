PRAGMA foreign_keys = ON;

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
    retention_period INTEGER DEFAULT NULL
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

CREATE TABLE IF NOT EXISTS user_permissions (
    user_id TEXT NOT NULL,
    scope TEXT NOT NULL,
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, scope)
);

CREATE TABLE IF NOT EXISTS cluster_nodes (
    node_id TEXT PRIMARY KEY NOT NULL,
    node_name TEXT NOT NULL,
    node_kind TEXT NOT NULL,
    api_url TEXT,
    drpc_addr TEXT,
    advertise_addr TEXT,
    STATUS TEXT NOT NULL DEFAULT 'unknown',
    last_heartbeat_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cluster_kv (
    KEY TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS revoked_tokens (
    jti TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    revoked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME
);

CREATE TABLE IF NOT EXISTS sync_state (
    replica_name TEXT PRIMARY KEY NOT NULL,
    last_full_sync_at DATETIME,
    last_incremental_sync_at DATETIME,
    last_source_revision INTEGER NOT NULL DEFAULT 0,
    last_status TEXT NOT NULL DEFAULT 'never',
    last_error TEXT
);

CREATE INDEX IF NOT EXISTS idx_files_scope ON files(scope_label);

CREATE INDEX IF NOT EXISTS idx_files_status ON files(STATUS);

CREATE INDEX IF NOT EXISTS idx_vm_format ON vm_images(format);

CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions(user_id);

CREATE INDEX IF NOT EXISTS idx_user_permissions_scope ON user_permissions(scope);

CREATE INDEX IF NOT EXISTS idx_cluster_nodes_kind ON cluster_nodes(node_kind);

CREATE INDEX IF NOT EXISTS idx_cluster_nodes_status ON cluster_nodes(STATUS);

CREATE INDEX IF NOT EXISTS idx_revoked_tokens_user_id ON revoked_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_revoked_tokens_expires_at ON revoked_tokens(expires_at);
