package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/clustercmd"
	"github.com/spf13/cobra"
)

func NewClusterCmdGroup() *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clu"},
		Short:   "cluster operations",
		Long:    `Create and organize your GNS3 servers inside of a cluster`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := validateGlobalFlags(); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	clusterCmd.AddCommand(clustercmd.NewCreateClusterCmd())
	clusterCmd.AddCommand(clustercmd.NewAddNodeCmd())
	clusterCmd.AddCommand(clustercmd.NewAddNodesCmd())
	clusterCmd.AddCommand(clustercmd.NewLsClusterCmd())
	clusterCmd.AddCommand(clustercmd.NewClusterConfigmdGroup())
	return clusterCmd
}
