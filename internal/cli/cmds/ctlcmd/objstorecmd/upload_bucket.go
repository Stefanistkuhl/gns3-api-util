package objstorecmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type BucketUploadResult struct {
	BucketID   string `json:"bucket_id"`
	Filename   string `json:"filename"`
	FileUUID   string `json:"file_uuid"`
	BlobSHA256 string `json:"blob_sha256"`
	SizeBytes  int64  `json:"size_bytes"`
}

func (b BucketUploadResult) GetHeaders() []string {
	return []string{"BUCKET ID", "FILENAME", "FILE UUID", "SIZE (BYTES)", "HASH"}
}

func (b BucketUploadResult) GetRow() []string {
	return []string{
		b.BucketID,
		b.Filename,
		b.FileUUID,
		fmt.Sprintf("%d", b.SizeBytes),
		b.BlobSHA256,
	}
}

func NewUploadToBucketCmd() *cobra.Command {
	var (
		fileStoreName string
		contentType   string
		retention     int64
	)

	cmd := &cobra.Command{
		Use:   "upload-bucket [bucket_id] [file_path]",
		Short: "Upload a file to a specific bucket",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			bucketID := args[0]
			filePath := args[1]
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

			uploadReq := models.InitUploadRequest{
				ContentType: contentType,
				Filename:    filepath.Base(filePath),
				BucketID:    bucketID,
			}
			if cmd.Flags().Changed("retention") {
				uploadReq.RetentionPeriod = &retention
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Targeting Filestore: %s\n", filestore.URL)
			fmt.Fprintf(cmd.ErrOrStderr(), "Uploading to bucket %s...\n", bucketID)

			resp, err := helpers.RunUploadToBucket(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				bucketID,
				filePath,
				&uploadReq,
				cfg,
			)
			if err != nil {
				return err
			}

			result := BucketUploadResult{
				BucketID:   bucketID,
				Filename:   filepath.Base(filePath),
				FileUUID:   resp.FileUUID,
				BlobSHA256: resp.BlobSHA256,
				SizeBytes:  resp.SizeBytes,
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed: %w", err)
			}

			return printer.PrintObj(result, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	cmd.Flags().StringVarP(&contentType, "type", "t", "application/octet-stream", "Content-Type")
	cmd.Flags().Int64VarP(&retention, "retention", "r", 0, "Retention period in hours (unset = permanent)")

	return cmd
}
