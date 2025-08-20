package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateProjectCmd() *cobra.Command {
	var (
		name                string
		projectID           string
		path                string
		autoClose           bool
		autoOpen            bool
		autoStart           bool
		sceneHeight         int
		sceneWidth          int
		zoom                int
		showLayers          bool
		snapToGrid          bool
		showGrid            bool
		gridSize            int
		drawingGridSize     int
		showInterfaceLabels bool
		supplierLogo        string
		supplierURL         string
		closeAfterCreation  bool
		useJSON             string
	)

	var cmd = &cobra.Command{
		Use:     "new",
		Short:   "Create a Project",
		Long:    "Create a new project on the controller.",
		Example: "gns3util -s https://controller:3080 project new -n some_name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			var payload map[string]any
			if useJSON == "" {
				if name == "" {
					return fmt.Errorf("for this command -n/--name is required or provide --use-json")
				}
				data := schemas.ProjectCreate{}
				data.Name = &name
				if projectID != "" {
					pid := projectID
					data.ProjectID = &pid
				}
				if path != "" {
					p := path
					data.Path = &p
				}
				if autoClose {
					v := true
					data.AutoClose = &v
				}
				if autoOpen {
					v := true
					data.AutoOpen = &v
				}
				if autoStart {
					v := true
					data.AutoStart = &v
				}
				if sceneHeight != 0 {
					v := sceneHeight
					data.SceneHeight = &v
				}
				if sceneWidth != 0 {
					v := sceneWidth
					data.SceneWidth = &v
				}
				if zoom != 0 {
					v := zoom
					data.Zoom = &v
				}
				if showLayers {
					v := true
					data.ShowLayers = &v
				}
				if snapToGrid {
					v := true
					data.SnapToGrid = &v
				}
				if showGrid {
					v := true
					data.ShowGrid = &v
				}
				if gridSize != 0 {
					v := gridSize
					data.GridSize = &v
				}
				if drawingGridSize != 0 {
					v := drawingGridSize
					data.DrawingGridSize = &v
				}
				if showInterfaceLabels {
					v := true
					data.ShowInterfaceLabels = &v
				}
				if supplierLogo != "" || supplierURL != "" {
					supplier := schemas.Supplier{}
					if supplierLogo != "" {
						v := supplierLogo
						supplier.Logo = &v
					}
					if supplierURL != "" {
						v := supplierURL
						supplier.URL = &v
					}
					data.Supplier = &supplier
				}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createProject", nil, payload)
			if closeAfterCreation && projectID != "" {
				utils.ExecuteAndPrint(cfg, "closeProject", []string{projectID})
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the project")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Desired id for the project, leave empty for a generated one")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Filesystem path for the project")
	cmd.Flags().BoolVar(&autoClose, "auto-close", true, "Close project when last client leaves")
	cmd.Flags().BoolVar(&autoOpen, "auto-open", false, "Open project when GNS3 starts")
	cmd.Flags().BoolVar(&autoStart, "auto-start", false, "Start project when opened")
	cmd.Flags().IntVar(&sceneHeight, "scene-height", 0, "Height of the drawing area")
	cmd.Flags().IntVar(&sceneWidth, "scene-width", 0, "Width of the drawing area")
	cmd.Flags().IntVar(&zoom, "zoom", 0, "Zoom of the drawing area")
	cmd.Flags().BoolVar(&showLayers, "show-layers", false, "Show layers on the drawing area")
	cmd.Flags().BoolVar(&snapToGrid, "snap-to-grid", false, "Snap to grid on the drawing area")
	cmd.Flags().BoolVar(&showGrid, "show-grid", false, "Show the grid on the drawing area")
	cmd.Flags().IntVar(&gridSize, "grid-size", 0, "Grid size for nodes")
	cmd.Flags().IntVar(&drawingGridSize, "drawing-grid-size", 0, "Grid size for drawings")
	cmd.Flags().BoolVar(&showInterfaceLabels, "show-interface-labels", false, "Show interface labels on the drawing area")
	cmd.Flags().StringVar(&supplierLogo, "supplier-logo", "", "Path to the project supplier logo")
	cmd.Flags().StringVar(&supplierURL, "supplier-url", "", "URL to the project supplier site")
	cmd.Flags().BoolVar(&closeAfterCreation, "close-after-creation", true, "Close the project after creation")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
