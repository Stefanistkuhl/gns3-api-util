package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewNodeCmdGroup() *cobra.Command {
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Node operations",
		Long:  `Create, manage, and manipulate GNS3 nodes.`,
	}

	// Create subcommands
	nodeCmd.AddCommand(create.NewCreateNodeCmd())
	nodeCmd.AddCommand(create.NewCreateNodeFromTemplateCmd())
	nodeCmd.AddCommand(create.NewCreateQemuDiskImageCmd())

	// Get subcommands
	nodeCmd.AddCommand(get.NewGetNodeCmd())
	nodeCmd.AddCommand(get.NewGetNodesCmd())
	nodeCmd.AddCommand(get.NewGetNodesAutoIdlePCCmd())
	nodeCmd.AddCommand(get.NewGetNodesAutoIdlePCProposalsCmd())
	nodeCmd.AddCommand(get.NewGetNodeFileCmd())
	nodeCmd.AddCommand(get.NewGetNodeLinksCmd())

	// Post subcommands
	nodeCmd.AddCommand(post.NewNodeDuplicateCmd())
	nodeCmd.AddCommand(post.NewNodeConsoleResetCmd())
	nodeCmd.AddCommand(post.NewNodeIsolateCmd())
	nodeCmd.AddCommand(post.NewNodeUnisolateCmd())
	nodeCmd.AddCommand(post.NewReloadNodesCmd())
	nodeCmd.AddCommand(post.NewStartNodesCmd())
	nodeCmd.AddCommand(post.NewStopNodesCmd())
	nodeCmd.AddCommand(post.NewSuspendNodesCmd())

	// Update subcommands
	nodeCmd.AddCommand(update.NewUpdateNodeCmd())

	// Delete subcommands
	nodeCmd.AddCommand(delete.NewDeleteNodeCmd())

	return nodeCmd
}
