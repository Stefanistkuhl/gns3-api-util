package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewLinkCmdGroup() *cobra.Command {
	linkCmd := &cobra.Command{
		Use:     "link",
		Aliases: []string{"l"},
		Short:   "Link operations",
		Long:    `Create, manage, and manipulate GNS3 links.`,
	}

	linkCmd.AddCommand(create.NewCreateLinkCmd())

	linkCmd.AddCommand(get.NewGetLinksCmd())
	linkCmd.AddCommand(get.NewGetLinkCmd())
	linkCmd.AddCommand(get.NewGetLinkIfaceCmd())
	linkCmd.AddCommand(get.NewGetLinkFiltersCmd())

	linkCmd.AddCommand(post.NewResetLinkCmd())
	linkCmd.AddCommand(post.NewStartCaptureCmd())
	linkCmd.AddCommand(post.NewStopCaptureCmd())

	linkCmd.AddCommand(update.NewUpdateLinkCmd())

	linkCmd.AddCommand(delete.NewDeleteLinkCmd())

	return linkCmd
}
