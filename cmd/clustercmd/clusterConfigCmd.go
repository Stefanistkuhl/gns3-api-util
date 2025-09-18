package clustercmd

import (
	"github.com/spf13/cobra"
)

func NewClusterConfigmdGroup() *cobra.Command {
	clusterConfigCmd := &cobra.Command{
		Use:   "config",
		Short: "cluster config operations",
		Long:  `commands to manage your cluster config`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	//sync
	//check
	//apply
	//edit
	//export
	//import
	clusterConfigCmd.AddCommand(NewSyncClusterConfigCmdGroup())
	return clusterConfigCmd
}
