package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewNodeCmdGroup() *cobra.Command {
	nodeCmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"n"},
		Short:   "Node operations",
		Long:    `Create, manage, and manipulate GNS3 nodes.`,
	}

	nodeCmd.AddCommand(create.NewCreateNodeCmd())
	nodeCmd.AddCommand(create.NewCreateNodeFromTemplateCmd())
	nodeCmd.AddCommand(create.NewCreateQemuDiskImageCmd())

	nodeCmd.AddCommand(get.NewGetNodeCmd())
	nodeCmd.AddCommand(get.NewGetNodesCmd())
	nodeCmd.AddCommand(get.NewGetNodesAutoIdlePCCmd())
	nodeCmd.AddCommand(get.NewGetNodesAutoIdlePCProposalsCmd())
	nodeCmd.AddCommand(get.NewGetNodeFileCmd())
	nodeCmd.AddCommand(get.NewGetNodeLinksCmd())

	nodeCmd.AddCommand(post.NewNodeDuplicateCmd())
	nodeCmd.AddCommand(post.NewNodeConsoleResetCmd())
	nodeCmd.AddCommand(post.NewNodeIsolateCmd())
	nodeCmd.AddCommand(post.NewNodeUnisolateCmd())
	nodeCmd.AddCommand(post.NewReloadNodesCmd())
	nodeCmd.AddCommand(post.NewStartNodesCmd())
	nodeCmd.AddCommand(post.NewStopNodesCmd())
	nodeCmd.AddCommand(post.NewSuspendNodesCmd())

	nodeCmd.AddCommand(update.NewUpdateNodeCmd())

	nodeCmd.AddCommand(delete.NewDeleteNodeCmd())

	return nodeCmd
}
