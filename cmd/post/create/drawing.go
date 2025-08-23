package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateDrawingCmd() *cobra.Command {
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
		Use:     utils.CreateSingleElementCmdName + " [project-name/id]",
		Short:   "Create a drawing",
		Long:    "Create a drawing in a project with specified properties",
		Example: "gns3util -s https://controller:3080 drawing create my-project --x 100 --y 200 --svg '<svg>...</svg>'",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			projectID := args[0]
			var payload map[string]any
			if useJSON == "" {
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
			utils.ExecuteAndPrintWithBody(cfg, "createDrawing", []string{projectID}, payload)
			return nil
		},
	}
	cmd.Flags().IntVar(&x, "x", 0, "X position")
	cmd.Flags().IntVar(&y, "y", 0, "Y position")
	cmd.Flags().IntVar(&z, "z", 1, "Z position (layer)")
	cmd.Flags().BoolVar(&locked, "locked", false, "Lock the drawing")
	cmd.Flags().IntVarP(&rotation, "rotation", "r", 0, "Rotation of the drawing")
	cmd.Flags().StringVar(&svg, "svg", "", "Raw SVG data")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
