package backgroundjobs

import (
	"log/slog"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/fs"
)

type Invocator string

const (
	InvocatorMaster   = "master"
	InvocatorInterval = "interval"
)

type DBVacuumJob struct {
	Store                      *db.Store
	Interval                   time.Duration
	Logger                     *slog.Logger
	TombstoneRemoveInterval    time.Duration
	PendingFileRemoveInterval  time.Duration
	NoWriteSinceUploadInterval time.Duration
	Dirs                       fs.Dirs
}

type Blob struct {
	UUID       string `json:"blob_uuid"`
	RefCount   int    `json:"blob_refcount"`
	WasDeleted bool   `json:"was_deleted"`
	FilePath   string `json:"blob_file_path"`
}

type CleanupFilesWithStatusResult struct {
	Success      bool      `json:"success"`
	Error        *string   `json:"error"`
	InvokedBy    Invocator `json:"invoked_by"`
	DeletedFiles []string  `json:"deleted_files"`
	Blobs        []Blob    `json:"changed_blobs"`
}

type CleanupStalledUploadsResults struct {
	Success      bool      `json:"success"`
	Error        *string   `json:"error"`
	InvokedBy    Invocator `json:"invoked_by"`
	DeletedFiles []string  `json:"deleted_files"`
	Blobs        []Blob    `json:"changed_blobs"`
}

type CleanupOrphanedBlobResults struct {
	Success   bool      `json:"success"`
	Error     *string   `json:"error"`
	InvokedBy Invocator `json:"invoked_by"`
	Blobs     []Blob    `json:"changed_blobs"`
}

type MarkExpiredFilesResults struct {
	Success   bool      `json:"success"`
	Error     *string   `json:"error"`
	InvokedBy Invocator `json:"invoked_by"`
	FileUUIDS []string  `json:"file_uuids"`
}

type DeleteOldTmpFilesResults struct {
	Success   bool      `json:"success"`
	Error     *string   `json:"error"`
	InvokedBy Invocator `json:"invoked_by"`
	FilePaths []string  `json:"file_paths"`
}

type DeleteExpiredTokensResult struct {
	Success   bool      `json:"success"`
	Error     *string   `json:"error"`
	InvokedBy Invocator `json:"invoked_by"`
}

type VacuumIterationResult struct {
	TombstoneExpired      MarkExpiredFilesResults      `json:"tombstone_expired"`
	CleanupTombstoned     CleanupFilesWithStatusResult `json:"cleanup_tombstoned"`
	CleanupPending        CleanupFilesWithStatusResult `json:"cleanup_pending"`
	CleanupStalledUploads CleanupStalledUploadsResults `json:"cleanup_stalled_uploads"`
	CleanupOrphanedBlobs  CleanupOrphanedBlobResults   `json:"cleanup_orphaned_blobs"`
	DeleteExpiredTokens   DeleteExpiredTokensResult    `json:"delete_expired_tokens"`
	DeleteOldTmpFiles     DeleteOldTmpFilesResults     `json:"delete_old_tmp_files"`
}
