package cli

import (
	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cmds/ctlcmd"
	"github.com/0xveya/gns3util/internal/cli/cmds/ctlcmd/jobscmd"
	"github.com/0xveya/gns3util/internal/cli/cmds/ctlcmd/objstorecmd"
)

func NewCtlCmdGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster-control",
		Aliases: []string{"ctl"},
		Short:   "gns3util control plane operations",
		Long:    `gns3util control plane operations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	clusterGroup := &cobra.Group{ID: "cluster", Title: "Cluster Management Commands:"}
	storeGroup := &cobra.Group{ID: "storage", Title: "Storage Commands:"}
	opsGroup := &cobra.Group{ID: "operations", Title: "Operations Commands:"}

	cmd.AddGroup(clusterGroup, storeGroup, opsGroup)

	createCmd := ctlcmd.NewCreateCmd()
	createCmd.GroupID = "cluster"

	authCmd := ctlcmd.NewAuthCmd()
	authCmd.GroupID = "cluster"

	addClusterCmd := ctlcmd.NewAddClusterCMD()
	addClusterCmd.GroupID = "cluster"

	addDiscoverCmd := ctlcmd.NewAddDiscoverCMD()
	addDiscoverCmd.GroupID = "cluster"

	removeCmd := ctlcmd.NewRemoveCMD()
	removeCmd.GroupID = "cluster"

	objCmd := objstorecmd.NewObjCmd()
	objCmd.GroupID = "storage"

	jobsCmd := jobscmd.NewJobsCmd()
	jobsCmd.GroupID = "operations"

	cmd.AddCommand(createCmd, authCmd, addClusterCmd, addDiscoverCmd, removeCmd, objCmd, jobsCmd)

	return cmd
}
