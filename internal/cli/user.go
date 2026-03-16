package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewUserCmdGroup() *cobra.Command {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "User operations",
		Long:    `Create, manage, and manipulate GNS3 users.`,
	}

	userCmd.AddCommand(create.NewCreateUserCmd())

	userCmd.AddCommand(get.NewGetUserCmd())
	userCmd.AddCommand(get.NewGetUsersCmd())
	userCmd.AddCommand(get.NewGetMeCmd())
	userCmd.AddCommand(get.NewGetGroupMembershipsCmd())

	userCmd.AddCommand(post.NewUserAuthenticateCmd())

	userCmd.AddCommand(update.NewUpdateMeCmd())
	userCmd.AddCommand(update.NewUpdateUserCmd())
	userCmd.AddCommand(update.NewChangePasswordCmd())

	userCmd.AddCommand(delete.NewDeleteUserCmd())

	return userCmd
}
