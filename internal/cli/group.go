package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/add"
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewGroupCmdGroup() *cobra.Command {
	groupCmd := &cobra.Command{
		Use:     "group",
		Aliases: []string{"g"},
		Short:   "Group operations",
		Long:    `Create, manage, and manipulate GNS3 groups.`,
	}

	groupCmd.AddCommand(create.NewCreateGroupCmd())

	groupCmd.AddCommand(get.NewGetGroupCmd())
	groupCmd.AddCommand(get.NewGetGroupsCmd())
	groupCmd.AddCommand(get.NewGetGroupMembersCmd())

	groupCmd.AddCommand(update.NewUpdateGroupCmd())

	groupCmd.AddCommand(delete.NewDeleteGroupCmd())

	groupCmd.AddCommand(add.NewAddGroupMemberCmd())

	return groupCmd
}
