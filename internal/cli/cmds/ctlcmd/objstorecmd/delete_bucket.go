package objstorecmd

import (
	"encoding/json"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
	"github.com/spf13/cobra"
)

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

			var resp *models.DeleteBucketResponse
			resp, err = helpers.RunDeleteBucket(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				bucketID,
				cfg,
			)
			if err != nil {
				return err
			}
			body, err := json.Marshal(resp)
			if err != nil {
				return fmt.Errorf("failed to marshal upload result: %w", err)
			}

			utils.PrintOutput(body, cfg)

			return nil
		},
	}

	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	return cmd
}
