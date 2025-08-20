package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteDrawingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [project-name/id] [drawing-name/id]",
		Short: "Delete a drawing from a project",
		Long:  `Delete a drawing from a project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete drawing my-project my-drawing
gns3util -s https://controller:3080 delete drawing 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			drawingID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(drawingID) {
				fmt.Println("Drawing ID must be a valid UUID")
				return
			}

			utils.ExecuteAndPrint(cfg, "deleteDrawing", []string{projectID, drawingID})
		},
	}

	return cmd
}
