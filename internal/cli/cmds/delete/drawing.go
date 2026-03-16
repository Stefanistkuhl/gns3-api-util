package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteDrawingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [project-name/id] [drawing-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a drawing from a project",
		Long:    `Delete a drawing from a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 drawing delete my-project my-drawing",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			drawingID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					return err
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(drawingID) {
				return fmt.Errorf("drawing ID must be a valid UUID")
			}

			utils.ExecuteAndPrint(cfg, "deleteDrawing", []string{projectID, drawingID})
			return nil
		},
	}

	return cmd
}
