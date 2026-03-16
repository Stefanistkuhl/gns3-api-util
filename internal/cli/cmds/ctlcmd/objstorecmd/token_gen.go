package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func NewGeneratePublicTokenCmd() *cobra.Command {
	var (
		fileStoreName  string
		expiresInHours int
	)

	cmd := &cobra.Command{
		Use:   "token-gen [file_uuid]",
		Short: "Generate a public token for file access",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
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

			fileUUID := args[0]
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

			fmt.Printf("Generating public token for: %s\n",
				fileUUID)

			resp, err := helpers.RunGeneratePublicToken(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				fileUUID,
				expiresInHours,
				cfg,
			)
			if err != nil {
				return err
			}

			fmt.Printf("\nPublic Token Generated!\n")
			fmt.Printf("Token: %s\n", resp.Token)
			fmt.Printf("Expires: %s\n", resp.ExpiresAt)
			fmt.Printf("URL: %s\n", resp.URL)

			return nil
		},
	}

	cmd.Flags().IntVar(
		&expiresInHours,
		"expires",
		24,
		"Token expiration in hours",
	)
	cmd.Flags().StringVar(
		&fileStoreName,
		"filestore-id",
		"",
		"Specific filestore node ID",
	)

	return cmd
}
