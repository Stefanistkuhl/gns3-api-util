package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/add"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewGroupCmdGroup() *cobra.Command {
	groupCmd := &cobra.Command{
		Use:   "group",
		Short: "Group operations",
		Long:  `Create, manage, and manipulate GNS3 groups.`,
	}

	// Create subcommands
	groupCmd.AddCommand(create.NewCreateGroupCmd())

	// Get subcommands
	groupCmd.AddCommand(get.NewGetGroupCmd())
	groupCmd.AddCommand(get.NewGetGroupsCmd())
	groupCmd.AddCommand(get.NewGetGroupMembersCmd())

	// Update subcommands
	groupCmd.AddCommand(update.NewUpdateGroupCmd())

	// Delete subcommands
	groupCmd.AddCommand(delete.NewDeleteGroupCmd())

	// Add subcommands
	groupCmd.AddCommand(add.NewAddGroupMemberCmd())

	return groupCmd
}
