-- Rollback: SQLite does not support DROP COLUMN before 3.35.0; recreate table.
CREATE TABLE project_files_old AS SELECT file_uuid, project_id, version_tag, is_read_only FROM project_files;
DROP TABLE project_files;
ALTER TABLE project_files_old RENAME TO project_files;
DROP INDEX IF EXISTS idx_project_files_project;
