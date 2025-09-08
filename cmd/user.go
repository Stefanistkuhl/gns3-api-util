package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewUserCmdGroup() *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User operations",
		Long:  `Create, manage, and manipulate GNS3 users.`,
	}

	// Create subcommands
	userCmd.AddCommand(create.NewCreateUserCmd())

	// Get subcommands
	userCmd.AddCommand(get.NewGetUserCmd())
	userCmd.AddCommand(get.NewGetUsersCmd())
	userCmd.AddCommand(get.NewGetMeCmd())
	userCmd.AddCommand(get.NewGetGroupMembershipsCmd())

	// Post subcommands
	userCmd.AddCommand(post.NewUserAuthenticateCmd())

	// Update subcommands
	userCmd.AddCommand(update.NewUpdateMeCmd())
	userCmd.AddCommand(update.NewUpdateUserCmd())
	userCmd.AddCommand(update.NewChangePasswordCmd())

	// Delete subcommands
	userCmd.AddCommand(delete.NewDeleteUserCmd())

	return userCmd
}
