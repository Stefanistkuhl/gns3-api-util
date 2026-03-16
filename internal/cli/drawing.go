package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewDrawingCmdGroup() *cobra.Command {
	drawingCmd := &cobra.Command{
		Use:     "drawing",
		Aliases: []string{"draw", "d"},
		Short:   "Drawing operations",
		Long:    `Create, manage, and manipulate GNS3 drawings.`,
	}

	drawingCmd.AddCommand(create.NewCreateDrawingCmd())

	drawingCmd.AddCommand(get.NewGetDrawingsCmd())
	drawingCmd.AddCommand(get.NewGetDrawingCmd())

	drawingCmd.AddCommand(update.NewUpdateDrawingCmd())

	drawingCmd.AddCommand(delete.NewDeleteDrawingCmd())

	return drawingCmd
}
