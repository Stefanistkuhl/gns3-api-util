package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewDrawingCmdGroup() *cobra.Command {
	drawingCmd := &cobra.Command{
		Use:   "drawing",
		Short: "Drawing operations",
		Long:  `Create, manage, and manipulate GNS3 drawings.`,
	}

	// Create subcommands
	drawingCmd.AddCommand(create.NewCreateDrawingCmd())

	// Get subcommands
	drawingCmd.AddCommand(get.NewGetDrawingsCmd())
	drawingCmd.AddCommand(get.NewGetDrawingCmd())

	// Update subcommands
	drawingCmd.AddCommand(update.NewUpdateDrawingCmd())

	// Delete subcommands
	drawingCmd.AddCommand(delete.NewDeleteDrawingCmd())

	return drawingCmd
}
