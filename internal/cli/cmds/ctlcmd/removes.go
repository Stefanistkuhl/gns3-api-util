package ctlcmd

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/spf13/cobra"
)

func NewRemoveCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove related cluster operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(NewRemoveClusterCMD())
	return cmd
}

func NewRemoveClusterCMD() *cobra.Command {
	var removeClusterName string
	var force bool

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Remove a cluster from the keyfile",
		Long: `Removes a cluster configuration and its associated nodes from the keyfile. 
If no name is provided via flags, an interactive list will be shown.`,
		Example: `  gns3util remove-cluster --name production-west
  gns3util remove-cluster (launches interactive TUI)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return err
			}

			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return err
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return err
			}

			targetCluster := removeClusterName
			if targetCluster == "" {
				var names []string
				for i := range kf.Clusters {
					c := kf.Clusters[i]
					names = append(names, c.Name)
				}

				if len(names) == 0 {
					return fmt.Errorf("no clusters found in keyfile")
				}

				choices := fuzzy.NewFuzzyFinderWithTitle(names, false, "Select cluster to remove")

				if len(choices) == 0 {
					return fmt.Errorf("selection cancelled")
				}

				targetCluster = choices[0]
			}

			if !kf.CheckIfClusterExists(targetCluster) {
				return fmt.Errorf("cluster %q not found in keyfile", targetCluster)
			}

			if !force && !cfg.Insecure {
				msg := fmt.Sprintf("Are you sure you want to remove cluster %q?", targetCluster)
				if !utils.ConfirmPrompt(msg, false) {
					return fmt.Errorf("operation cancelled")
				}
			}

			kf.RemoveClusterByName(targetCluster)

			if err := pathutils.SaveKeysFile(keyFilePath, kf); err != nil {
				return fmt.Errorf("failed to update keyfile: %w", err)
			}

			fmt.Printf("Successfully removed cluster %q\n", targetCluster)
			return nil
		},
	}

	cmd.Flags().StringVarP(&removeClusterName, "name", "n", "", "Name of the cluster to remove")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
