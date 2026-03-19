package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type FileRow struct {
	FilestoreID string `json:"filestore_id"`
	BucketUUID  string `json:"bucket_uuid"`
	models.FileInfo
}

func (f *FileRow) GetHeaders() []string {
	return []string{"FILESTORE ID", "BUCKET UUID", "FILENAME", "FILE UUID", "SIZE (BYTES)"}
}

func (f *FileRow) GetRow() []string {
	return []string{
		f.FilestoreID,
		f.BucketUUID,
		f.Filename,
		f.FileUUID,
		fmt.Sprintf("%d", f.SizeBytes),
	}
}

func NewListBucketFilesCmd() *cobra.Command {
	var fileStoreName string

	cmd := &cobra.Command{
		Use:   "list-files [bucket_id]",
		Short: "List files in a specific bucket",
		Long:  "List files in a specific bucket if no bucket is specified the default bucket is used",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			bucketID := models.GlobalBucketID
			if len(args) > 0 {
				bucketID = args[0]
			}
			clusterName := cfg.Cluster

			keyPath, err := pathutils.ResolveKeyFilePath("")
			if err != nil {
				return err
			}
			kf, err := pathutils.LoadGNS3KeysFile(keyPath)
			if err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			var targetCluster *pathutils.ClusterEntry
			for i := range kf.Clusters {
				if kf.Clusters[i].Name == clusterName {
					targetCluster = &kf.Clusters[i]
					break
				}
			}

			if targetCluster == nil {
				return fmt.Errorf("cluster %q not found in keys file", clusterName)
			}

			cfg.ClusterEntry.CaCert = targetCluster.CaCert

			filestore, err := selectFilestore(targetCluster, fileStoreName)
			if err != nil {
				return err
			}

			files, err := helpers.RunListBucketFiles(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				bucketID,
				cfg,
			)
			if err != nil {
				return err
			}

			var records []utils.TableRecord
			for _, f := range files.Files {
				records = append(records, &FileRow{
					FilestoreID: filestore.ID,
					BucketUUID:  files.BucketUUID,
					FileInfo:    f,
				})
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed (use --output=table or json): %w", err)
			}

			if cfg.OutputFormat == globals.OutputJSON {
				wrapped := struct {
					FilestoreID  string                          `json:"filestore_id"`
					FilestoreURL string                          `json:"filestore_url"`
					Response     *models.ListBucketFilesResponse `json:"response"`
				}{
					FilestoreID:  filestore.ID,
					FilestoreURL: filestore.URL,
					Response:     files,
				}
				return printer.PrintObj(wrapped, cmd.OutOrStdout())
			}

			return printer.PrintObj(records, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	return cmd
}
