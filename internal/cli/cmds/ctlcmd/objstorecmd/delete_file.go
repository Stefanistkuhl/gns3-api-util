package objstorecmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func NewDeleteFileCmd() *cobra.Command {
	var fileStoreName string

	cmd := &cobra.Command{
		Use:   "delete-file [file_uuid]",
		Short: "Delete a file from the filestore",
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

			var resp *models.DeleteFileResponse
			resp, err = helpers.RunDeleteFile(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				fileUUID,
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

	cmd.Flags().StringVar(
		&fileStoreName,
		"filestore-id",
		"",
		"Specific filestore node ID",
	)

	return cmd
}
