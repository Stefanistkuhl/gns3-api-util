package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateDrawingCmd() *cobra.Command {
	var (
		x        int
		y        int
		z        int
		locked   bool
		rotation int
		svg      string
		useJSON  string
	)

	cmd := &cobra.Command{
		Use:     utils.UpdateSingleElementCmdName + " [project-name/id] [drawing-name/id]",
		Short:   "Update a drawing",
		Long:    "Update a drawing in a project.",
		Example: "gns3util -s https://controller:3080 drawing update my-project my-drawing -s 'some svg'",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			projectID := args[0]
			drawingID := args[1]

			if !utils.IsValidUUIDv4(projectID) {
				resolvedID, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve project ID: %w", err)
				}
				projectID = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				// Check if at least one field is provided
				if x == 0 && y == 0 && z == 1 && !locked && rotation == 0 && svg == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				payload = map[string]any{}
				if x != 0 {
					payload["x"] = x
				}
				if y != 0 {
					payload["y"] = y
				}
				if z != 1 {
					payload["z"] = z
				}
				if locked {
					payload["locked"] = true
				}
				if rotation != 0 {
					payload["rotation"] = rotation
				}
				if svg != "" {
					payload["svg"] = svg
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "updateDrawing", []string{projectID, drawingID}, payload)
			return nil
		},
	}

	cmd.Flags().IntVarP(&x, "x", "x", 0, "X-Position of the drawing")
	cmd.Flags().IntVarP(&y, "y", "y", 0, "Y-Position of the drawing")
	cmd.Flags().IntVarP(&z, "z", "z", 1, "Z-Position (layer) of the drawing")
	cmd.Flags().BoolVarP(&locked, "locked", "l", false, "Lock the drawing")
	cmd.Flags().IntVarP(&rotation, "rotation", "r", 0, "Rotation of the drawing")
	cmd.Flags().StringVarP(&svg, "svg", "", "", "Raw SVG data for the drawing")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
