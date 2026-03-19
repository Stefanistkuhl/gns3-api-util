package objstorecmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type PublicTokenResult struct {
	FileUUID   string `json:"file_uuid"`
	BUcketUUID string `json:"bucket_uuid"`
	Token      string `json:"token"`
	ExpiresAt  string `json:"expires_at"`
	URL        string `json:"url"`
}

func (p *PublicTokenResult) GetHeaders() []string {
	return []string{"FILE UUID", "TOKEN", "EXPIRES AT", "URL"}
}

func (p *PublicTokenResult) GetRow() []string {
	return []string{
		p.FileUUID,
		p.Token,
		p.ExpiresAt,
		p.URL,
	}
}

func NewGeneratePublicTokenCmd() *cobra.Command {
	var (
		fileStoreName  string
		expiresInHours int
	)

	cmd := &cobra.Command{
		Use:   "token-gen [file_uuid] [bucket_id]",
		Short: "Generate a public token for file access",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fileUUID := args[0]
			bucketID := models.GlobalBucketID
			if len(args) > 1 {
				bucketID = args[1]
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
			var expiry *time.Time
			if expiresInHours == 0 {
				expiry = nil
			} else {
				t := time.Now().Add(time.Hour * time.Duration(expiresInHours))
				expiry = &t
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Generating public token for: %s\n", fileUUID)

			resp, err := helpers.RunGeneratePublicToken(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				fileUUID,
				bucketID,
				expiry,
				cfg,
			)
			if err != nil {
				return err
			}

			result := PublicTokenResult{
				FileUUID:   fileUUID,
				BUcketUUID: resp.BucketUUID,
				Token:      resp.Token,
				ExpiresAt:  resp.ExpiresAt,
				URL:        resp.URL,
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed: %w", err)
			}

			return printer.PrintObj(&result, cmd.OutOrStdout())
		},
	}

	cmd.Flags().IntVar(&expiresInHours, "expires", 24, "Token expiration in hours (0 for no expiry)")
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}
