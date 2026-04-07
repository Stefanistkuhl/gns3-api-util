-- Expand project_files with GNS3 export metadata so stored project archives
-- are queryable by project, name, and export settings.
ALTER TABLE project_files ADD COLUMN project_name TEXT;
ALTER TABLE project_files ADD COLUMN include_snapshots BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_files ADD COLUMN include_images BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_files ADD COLUMN reset_mac_addresses BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_files ADD COLUMN keep_compute_ids BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_files ADD COLUMN compression TEXT NOT NULL DEFAULT 'zstd';

-- Index for the most common lookup: "give me all exports for project X"
CREATE INDEX IF NOT EXISTS idx_project_files_project ON project_files(project_id);
