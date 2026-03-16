package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/add"
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewRoleCmdGroup() *cobra.Command {
	roleCmd := &cobra.Command{
		Use:     "role",
		Aliases: []string{"r"},
		Short:   "Role operations",
		Long:    `Create, manage, and manipulate GNS3 roles.`,
	}

	roleCmd.AddCommand(create.NewCreateRoleCmd())

	roleCmd.AddCommand(get.NewGetRoleCmd())
	roleCmd.AddCommand(get.NewGetRolesCmd())
	roleCmd.AddCommand(get.NewGetRolePrivsCmd())
	roleCmd.AddCommand(get.NewGetPrivilegesCmd())

	roleCmd.AddCommand(update.NewUpdateRoleCmd())

	roleCmd.AddCommand(delete.NewDeleteRoleCmd())
	roleCmd.AddCommand(delete.NewDeleteRolePrivilegeCmd())

	roleCmd.AddCommand(add.NewAddPrivilegeCmd())

	return roleCmd
}
