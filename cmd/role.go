package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/add"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewRoleCmdGroup() *cobra.Command {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Role operations",
		Long:  `Create, manage, and manipulate GNS3 roles.`,
	}

	// Create subcommands
	roleCmd.AddCommand(create.NewCreateRoleCmd())

	// Get subcommands
	roleCmd.AddCommand(get.NewGetRoleCmd())
	roleCmd.AddCommand(get.NewGetRolesCmd())
	roleCmd.AddCommand(get.NewGetRolePrivsCmd())
	roleCmd.AddCommand(get.NewGetPrivilegesCmd())

	// Update subcommands
	roleCmd.AddCommand(update.NewUpdateRoleCmd())

	// Delete subcommands
	roleCmd.AddCommand(delete.NewDeleteRoleCmd())
	roleCmd.AddCommand(delete.NewDeleteRolePrivilegeCmd())

	// Add subcommands
	roleCmd.AddCommand(add.NewAddPrivilegeCmd())

	return roleCmd
}
