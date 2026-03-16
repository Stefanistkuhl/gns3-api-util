package objstorecmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func NewListBucketsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-buckets",
		Short: "List all buckets in the filestore",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
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

			filestore, err := selectFilestore(targetCluster, "")
			if err != nil {
				return err
			}

			buckets, err := helpers.RunListBuckets(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				cfg,
			)
			if err != nil {
				return err
			}

			var body []byte

			if cfg.OutputFormat == globals.OutputTable || cfg.OutputFormat == globals.OutputKV {
				type bucketRow struct {
					FilestoreID string    `json:"filestore_id"`
					BucketID    string    `json:"bucket_id"`
					Name        string    `json:"name"`
					IsPublic    bool      `json:"is_public"`
					CreatedAt   time.Time `json:"created_at"`
				}

				rows := make([]bucketRow, 0, len(buckets.Buckets))
				for _, b := range buckets.Buckets {
					rows = append(rows, bucketRow{
						FilestoreID: filestore.ID,
						BucketID:    b.BucketID,
						Name:        b.Name,
						IsPublic:    b.IsPublic,
						CreatedAt:   b.CreatedAt,
					})
				}
				body, err = json.Marshal(rows)
			} else {
				wrapped := struct {
					FilestoreID  string                     `json:"filestore_id"`
					FilestoreURL string                     `json:"filestore_url"`
					Response     *models.ListBucketResponse `json:"response"`
				}{
					FilestoreID:  filestore.ID,
					FilestoreURL: filestore.URL,
					Response:     buckets,
				}
				body, err = json.Marshal(wrapped)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}

			utils.PrintOutput(body, cfg)

			return nil
		},
	}

	return cmd
}
