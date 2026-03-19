package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type BucketDeleteResult struct {
	FilestoreID string `json:"filestore_id"`
	BucketID    string `json:"bucket_id"`
	Status      string `json:"status"`
}

func (b BucketDeleteResult) GetHeaders() []string {
	return []string{"FILESTORE ID", "BUCKET ID", "STATUS"}
}

func (b BucketDeleteResult) GetRow() []string {
	return []string{
		b.FilestoreID,
		b.BucketID,
		b.Status,
	}
}

func NewDeleteBucketCmd() *cobra.Command {
	var fileStoreName string

	cmd := &cobra.Command{
		Use:   "delete-bucket [bucket_id]",
		Short: "Delete a bucket and all its files (cannot delete global bucket)",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			bucketID := args[0]
			if bucketID == models.GlobalBucketID {
				return fmt.Errorf("operation forbidden: cannot delete the global system bucket")
			}

			clusterName := cfg.Cluster
			keyPath, err := pathutils.ResolveKeyFilePath("")
			if err != nil {
				return err
			}
			kf, err := pathutils.LoadGNS3KeysFile(keyPath)
			if err != nil {
				return err
			}

			var targetCluster *pathutils.ClusterEntry
			for i := range kf.Clusters {
				if kf.Clusters[i].Name == clusterName {
					targetCluster = &kf.Clusters[i]
					break
				}
			}

			if targetCluster == nil {
				return fmt.Errorf("cluster %q not found", clusterName)
			}
			cfg.ClusterEntry.CaCert = targetCluster.CaCert

			filestore, err := selectFilestore(targetCluster, fileStoreName)
			if err != nil {
				return err
			}

			_, err = helpers.RunDeleteBucket(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				bucketID,
				cfg,
			)
			if err != nil {
				return err
			}

			result := BucketDeleteResult{
				FilestoreID: filestore.ID,
				BucketID:    bucketID,
				Status:      "Deleted",
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed: %w", err)
			}

			return printer.PrintObj(result, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	return cmd
}
