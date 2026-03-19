package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type FileDownloadResult struct {
	FilestoreID string `json:"filestore_id"`
	FileUUID    string `json:"file_uuid"`
	OutputPath  string `json:"output_path"`
	Status      string `json:"status"`
}

func (f FileDownloadResult) GetHeaders() []string {
	return []string{"FILESTORE ID", "FILE UUID", "OUTPUT PATH", "STATUS"}
}

func (f FileDownloadResult) GetRow() []string {
	return []string{
		f.FilestoreID,
		f.FileUUID,
		f.OutputPath,
		f.Status,
	}
}

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
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
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
				return fmt.Errorf("cluster %q not found in keys file", clusterName)
			}

			cfg.ClusterEntry.CaCert = targetCluster.CaCert

			filestore, err := selectFilestore(targetCluster, fileStoreName)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Targeting Filestore: %s\n", filestore.URL)
			fmt.Fprintf(cmd.ErrOrStderr(), "Downloading file: %s\n", fileUUID)

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

			result := FileDownloadResult{
				FilestoreID: filestore.ID,
				FileUUID:    fileUUID,
				OutputPath:  outputPath,
				Status:      "Complete",
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
