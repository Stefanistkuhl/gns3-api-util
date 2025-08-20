package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
)

func NewImageCmdGroup() *cobra.Command {
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Image operations",
		Long:  `Create, manage, and manipulate GNS3 images.`,
	}

	// Create subcommands
	imageCmd.AddCommand(create.NewCreateQemuImageCmd())

	// Get subcommands
	imageCmd.AddCommand(get.NewGetImagesCmd())
	imageCmd.AddCommand(get.NewGetImageCmd())

	// Post subcommands
	imageCmd.AddCommand(post.NewImageCmdGroup())

	// Delete subcommands
	imageCmd.AddCommand(delete.NewDeleteImageCmd())
	imageCmd.AddCommand(delete.NewDeletePruneImagesCmd())

	return imageCmd
}
