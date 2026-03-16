package objstorecmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

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
			cfg, err := config.GetGlobalOptionsFromContext(
				cmd.Context(),
			)
			if err != nil {
				return fmt.Errorf(
					"failed to get global options: %w",
					err,
				)
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
				return fmt.Errorf(
					"cluster %q not found in keys file",
					clusterName,
				)
			}

			cfg.ClusterEntry.CaCert = targetCluster.CaCert

			filestore, err := selectFilestore(
				targetCluster,
				fileStoreName,
			)
			if err != nil {
				return err
			}

			uploadReq := models.InitUploadRequest{
				ContentType:     contentType,
				RetentionPeriod: retention,
				Filename:        filepath.Base(filePath),
			}

			fmt.Printf("Targeting Filestore: %s\n",
				filestore.URL)
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

			fmt.Printf("\nUpload to Bucket Complete!\n")
			fmt.Printf("UUID: %s\nHash: %s\nSize: %d bytes\n",
				resp.FileUUID,
				resp.ChecksumSHA256,
				resp.SizeBytes,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&fileStoreName,
		"filestore-id",
		"",
		"Specific filestore node ID",
	)
	cmd.Flags().StringVarP(
		&contentType,
		"type",
		"t",
		"application/octet-stream",
		"Content-Type",
	)
	cmd.Flags().Int64VarP(
		&retention,
		"retention",
		"r",
		0,
		"Retention period in hours",
	)

	return cmd
}
