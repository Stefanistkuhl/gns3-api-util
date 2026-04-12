package objstorecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/pkg/models"
)

// Scaffolded by tools/scaffoldctl from ".codegen/scaffoldctl/packages/backups.yaml"; edit command bodies as needed.
func NewBackupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backups",
		Aliases: []string{"backup"},
		Short:   "Manage backup objects in the filestore",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	queryGroup := &cobra.Group{ID: "query", Title: "Query commands:"}
	actionGroup := &cobra.Group{ID: "action", Title: "Action commands:"}
	cmd.AddGroup(
		queryGroup,
		actionGroup,
	)
	deleteBackupCmd := newDeleteBackupCmd()
	deleteBackupCmd.GroupID = "action"
	getBackupCmd := newGetBackupCmd()
	getBackupCmd.GroupID = "query"
	listBackupsCmd := newListBackupsCmd()
	listBackupsCmd.GroupID = "query"
	updateBackupCmd := newUpdateBackupCmd()
	updateBackupCmd.GroupID = "action"
	uploadBackupCmd := newUploadBackupCmd()
	uploadBackupCmd.GroupID = "action"

	cmd.AddCommand(
		deleteBackupCmd,
		getBackupCmd,
		listBackupsCmd,
		updateBackupCmd,
		uploadBackupCmd,
	)

	return cmd
}

type backupRow struct {
	FileUUID         string    `json:"file_uuid"`
	Filename         string    `json:"filename"`
	Status           string    `json:"status"`
	BackupType       string    `json:"backup_type"`
	SourceNodeID     string    `json:"source_node_id"`
	IsCompressed     bool      `json:"is_compressed"`
	IsEncrypted      bool      `json:"is_encrypted"`
	ParentBackupUUID string    `json:"parent_backup_uuid"`
	CreatedAt        time.Time `json:"created_at"`
}

func (b *backupRow) GetHeaders() []string {
	return []string{
		"FILE UUID",
		"FILENAME",
		"STATUS",
		"BACKUP TYPE",
		"SOURCE NODE",
		"COMPRESSED",
		"ENCRYPTED",
		"PARENT BACKUP",
		"CREATED AT",
	}
}

func (b *backupRow) GetRow() []string {
	return []string{
		b.FileUUID,
		b.Filename,
		b.Status,
		b.BackupType,
		b.SourceNodeID,
		fmt.Sprint(b.IsCompressed),
		fmt.Sprint(b.IsEncrypted),
		b.ParentBackupUUID,
		b.CreatedAt.Format(time.RFC3339),
	}
}

type backupUploadResult struct {
	FileUUID string `json:"file_uuid"`
	Status   string `json:"status"`
}

func (b *backupUploadResult) GetHeaders() []string {
	return []string{
		"FILE UUID",
		"STATUS",
	}
}

func (b *backupUploadResult) GetRow() []string {
	return []string{
		b.FileUUID,
		b.Status,
	}
}

type backupDeleteResult struct {
	FileUUID string `json:"file_uuid"`
	Status   string `json:"status"`
}

func (b *backupDeleteResult) GetHeaders() []string {
	return []string{
		"FILE UUID",
		"STATUS",
	}
}

func (b *backupDeleteResult) GetRow() []string {
	return []string{
		b.FileUUID,
		b.Status,
	}
}

func newDeleteBackupCmd() *cobra.Command {
	var fileStoreName string
	cmd := &cobra.Command{
		Use:     "delete <file-uuid>",
		Aliases: []string{"rm"},
		Short:   "Delete a backup object",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			fileUUID := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			resp, err := client.DeleteBackup(cmd.Context(), fileUUID)
			if err != nil {
				return err
			}
			return printRowOrObject[backupDeleteResult](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newGetBackupCmd() *cobra.Command {
	var fileStoreName string
	cmd := &cobra.Command{
		Use:   "get <file-uuid>",
		Short: "Get backup details",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			fileUUID := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			resp, err := client.GetBackup(cmd.Context(), fileUUID)
			if err != nil {
				return err
			}
			return printRowOrObject[backupRow](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newListBackupsCmd() *cobra.Command {
	var fileStoreName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backup objects",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			resp, err := client.ListBackups(cmd.Context())
			if err != nil {
				return err
			}
			return printRowsOrObject[backupRow](cmd, cfg, resp.Backups, "No backups found.")
		},
	}
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newUpdateBackupCmd() *cobra.Command {
	var (
		fileStoreName    string
		sourceNodeID     string
		backupType       string
		isCompressed     bool
		isEncrypted      bool
		parentBackupUUID string
	)
	cmd := &cobra.Command{
		Use:   "update <file-uuid>",
		Short: "Update backup metadata",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			fileUUID := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			req := &models.UpdateBackupRequest{}
			if !cmd.Flags().Changed("source-node-id") && !cmd.Flags().Changed("backup-type") && !cmd.Flags().Changed("compressed") && !cmd.Flags().Changed("encrypted") && !cmd.Flags().Changed("parent-backup-uuid") {
				return fmt.Errorf("no update flags were provided")
			}
			if cmd.Flags().Changed("compressed") {
				req.IsCompressed = &isCompressed
			}
			if cmd.Flags().Changed("encrypted") {
				req.IsEncrypted = &isEncrypted
			}
			if cmd.Flags().Changed("source-node-id") {
				req.SourceNodeID = sourceNodeID
			}
			if cmd.Flags().Changed("backup-type") {
				req.BackupType = backupType
			}
			if cmd.Flags().Changed("parent-backup-uuid") {
				req.ParentBackupUUID = parentBackupUUID
			}
			resp, err := client.UpdateBackup(cmd.Context(), fileUUID, req)
			if err != nil {
				return err
			}
			return printRowOrObject[backupRow](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVar(&sourceNodeID, "source-node-id", "", "Source node ID for this backup")
	cmd.Flags().StringVar(&backupType, "backup-type", "", "Backup type: full or incremental")
	cmd.Flags().BoolVar(&isCompressed, "compressed", false, "Set whether the backup is compressed")
	cmd.Flags().BoolVar(&isEncrypted, "encrypted", false, "Set whether the backup is encrypted")
	cmd.Flags().StringVar(&parentBackupUUID, "parent-backup-uuid", "", "Parent backup UUID for incremental backups")
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newUploadBackupCmd() *cobra.Command {
	var (
		fileStoreName    string
		contentType      string
		retentionPeriod  int64
		bucketID         string
		sourceNodeID     string
		backupType       string
		isCompressed     bool
		isEncrypted      bool
		parentBackupUUID string
	)
	cmd := &cobra.Command{
		Use:     "upload <file-path>",
		Aliases: []string{"add"},
		Short:   "Upload a backup object",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			filePath := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			file, err := os.Open(filePath) // #nosec G304
			if err != nil {
				return err
			}
			defer file.Close()
			stat, err := file.Stat()
			if err != nil {
				return err
			}
			req := &models.InitBackupUploadRequest{
				Filename:         filepath.Base(filePath),
				SizeBytes:        stat.Size(),
				ContentType:      contentType,
				BucketID:         bucketID,
				SourceNodeID:     sourceNodeID,
				BackupType:       backupType,
				IsCompressed:     isCompressed,
				IsEncrypted:      isEncrypted,
				ParentBackupUUID: parentBackupUUID,
			}
			if req.BucketID == "" {
				req.BucketID = models.GlobalBucketID
			}
			if cmd.Flags().Changed("retention") {
				req.RetentionPeriod = &retentionPeriod
			}
			initResp, err := client.InitBackupUpload(cmd.Context(), req)
			if err != nil {
				return err
			}
			status, err := client.GetUploadStatus(cmd.Context(), initResp.FileUUID)
			if err != nil {
				return err
			}
			if status.Offset >= stat.Size() {
				return printRowOrObject[backupUploadResult](cmd, cfg, &models.FinalizeUploadResponse{
					FileUUID: initResp.FileUUID,
					Status:   models.FileStatusAvailable,
				})
			}
			if status.Offset > 0 {
				if _, seekErr := file.Seek(status.Offset, 0); seekErr != nil {
					return fmt.Errorf("failed to seek local file: %w", seekErr)
				}
			}
			resp, err := client.StreamUpload(cmd.Context(), initResp.FileUUID, file, status.Offset)
			if err != nil {
				return err
			}
			return printRowOrObject[backupUploadResult](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVarP(&contentType, "type", "t", "application/octet-stream", "Explicit Content-Type for the backup file")
	cmd.Flags().Int64VarP(&retentionPeriod, "retention", "r", 0, "Retention period in hours")
	cmd.Flags().StringVar(&bucketID, "bucket-id", "", "Target bucket ID")
	cmd.Flags().StringVar(&sourceNodeID, "source-node-id", "", "Source node ID for this backup")
	cmd.Flags().StringVar(&backupType, "backup-type", "", "Backup type: full or incremental")
	cmd.Flags().BoolVar(&isCompressed, "compressed", false, "Mark the backup as compressed")
	cmd.Flags().BoolVar(&isEncrypted, "encrypted", false, "Mark the backup as encrypted")
	cmd.Flags().StringVar(&parentBackupUUID, "parent-backup-uuid", "", "Parent backup UUID for incremental backups")
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	if err := cmd.MarkFlagRequired("backup-type"); err != nil {
		panic(err)
	}

	return cmd
}
