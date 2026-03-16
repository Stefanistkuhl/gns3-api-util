package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func NewDownloadCmd() *cobra.Command {
	var fileStoreName string

	cmd := &cobra.Command{
		Use:     "download [file_uuid] [output_path]",
		Aliases: []string{"d", "get"},
		Short:   "Download a file from the cluster filestore",
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

			fileUUID := args[0]
			outputPath := args[1]
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

			fmt.Printf("Targeting Filestore: %s\n", filestore.URL)
			fmt.Printf("Downloading file: %s\n", fileUUID)

			err = helpers.RunDownload(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				fileUUID,
				outputPath,
				cfg,
			)
			if err != nil {
				return err
			}

			fmt.Printf("\nDownload Complete!\n")
			fmt.Printf("File saved to: %s\n", outputPath)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&fileStoreName,
		"filestore-id",
		"",
		"Specific filestore node ID",
	)

	return cmd
}
