package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewLinkCmdGroup() *cobra.Command {
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Link operations",
		Long:  `Create, manage, and manipulate GNS3 links.`,
	}

	// Create subcommands
	linkCmd.AddCommand(create.NewCreateLinkCmd())

	// Get subcommands
	linkCmd.AddCommand(get.NewGetLinksCmd())
	linkCmd.AddCommand(get.NewGetLinkCmd())
	linkCmd.AddCommand(get.NewGetLinkIfaceCmd())
	linkCmd.AddCommand(get.NewGetLinkFiltersCmd())

	// Post subcommands
	linkCmd.AddCommand(post.NewResetLinkCmd())
	linkCmd.AddCommand(post.NewStartCaptureCmd())
	linkCmd.AddCommand(post.NewStopCaptureCmd())

	// Update subcommands
	linkCmd.AddCommand(update.NewUpdateLinkCmd())

	// Delete subcommands
	linkCmd.AddCommand(delete.NewDeleteLinkCmd())

	return linkCmd
}
