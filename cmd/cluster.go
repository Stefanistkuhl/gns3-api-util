package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/clustercmd"
)

func NewClusterCmdGroup() *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "cluster operations",
		Long:  `Create and organize your GNS3 servers inside of a cluster`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			if err := validateGlobalFlags(); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	clusterCmd.AddCommand(clustercmd.NewCreateClusterCmd())
	clusterCmd.AddCommand(clustercmd.NewAddNodeCmd())
	clusterCmd.AddCommand(clustercmd.NewAddNodesCmd())
	clusterCmd.AddCommand(clustercmd.NewLsClusterCmd())
	clusterCmd.AddCommand(clustercmd.NewClusterConfigmdGroup())
	return clusterCmd
}
