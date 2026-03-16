package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/spf13/cobra"
)

func NewImageCmdGroup() *cobra.Command {
	imageCmd := &cobra.Command{
		Use:     "image",
		Aliases: []string{"img", "i"},
		Short:   "Image operations",
		Long:    `Create, manage, and manipulate GNS3 images.`,
	}

	imageCmd.AddCommand(create.NewCreateQemuImageCmd())

	imageCmd.AddCommand(get.NewGetImagesCmd())
	imageCmd.AddCommand(get.NewGetImageCmd())

	imageCmd.AddCommand(post.NewImageCmdGroup())

	imageCmd.AddCommand(delete.NewDeleteImageCmd())
	imageCmd.AddCommand(delete.NewDeletePruneImagesCmd())

	return imageCmd
}
