package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateProjectCmd() *cobra.Command {
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
		useJSON             string
	)

	cmd := &cobra.Command{
		Use:     utils.UpdateSingleElementCmdName + " [project-name/id]",
		Short:   "Update a Project",
		Long:    "Update a Project with new settings and properties.",
		Example: "gns3util -s https://controller:3080 project update my-project --name new-name",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			projectIDArg := args[0]

			if !utils.IsValidUUIDv4(projectIDArg) {
				resolvedID, err := utils.ResolveID(cfg, "project", projectIDArg, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve project ID: %w", err)
				}
				projectIDArg = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				if name == "" && projectID == "" && path == "" && !autoClose && !autoOpen && !autoStart && sceneHeight == 0 && sceneWidth == 0 && zoom == 0 && !showLayers && !snapToGrid && !showGrid && gridSize == 0 && drawingGridSize == 0 && !showInterfaceLabels && supplierLogo == "" && supplierURL == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				data := schemas.ProjectUpdate{}
				if name != "" {
					data.Name = &name
				}
				if projectID != "" {
					data.ProjectID = &projectID
				}
				if path != "" {
					data.Path = &path
				}
				if autoClose {
					data.AutoClose = &autoClose
				}
				if autoOpen {
					data.AutoOpen = &autoOpen
				}
				if autoStart {
					data.AutoStart = &autoStart
				}
				if sceneHeight != 0 {
					data.SceneHeight = &sceneHeight
				}
				if sceneWidth != 0 {
					data.SceneWidth = &sceneWidth
				}
				if zoom != 0 {
					data.Zoom = &zoom
				}
				if showLayers {
					data.ShowLayers = &showLayers
				}
				if snapToGrid {
					data.SnapToGrid = &snapToGrid
				}
				if showGrid {
					data.ShowGrid = &showGrid
				}
				if gridSize != 0 {
					data.GridSize = &gridSize
				}
				if drawingGridSize != 0 {
					data.DrawingGridSize = &drawingGridSize
				}
				if showInterfaceLabels {
					data.ShowInterfaceLabels = &showInterfaceLabels
				}
				if supplierLogo != "" || supplierURL != "" {
					supplier := schemas.Supplier{}
					if supplierLogo != "" {
						supplier.Logo = &supplierLogo
					}
					if supplierURL != "" {
						supplier.URL = &supplierURL
					}
					data.Supplier = &supplier
				}
				b, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to encode request: %w", err)
				}
				if err := json.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("failed to prepare payload: %w", err)
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "updateProject", []string{projectIDArg}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the project")
	cmd.Flags().StringVarP(&projectID, "project-id", "d", "", "Desired id for the project")
	cmd.Flags().StringVarP(&path, "path", "p", "", "Filepath for the project")
	cmd.Flags().BoolVarP(&autoClose, "auto-close", "c", true, "Close project when last client leaves")
	cmd.Flags().BoolVarP(&autoOpen, "auto-open", "o", false, "Project opens when GNS3 starts")
	cmd.Flags().BoolVar(&autoStart, "auto-start", false, "Project starts when opened")
	cmd.Flags().IntVar(&sceneHeight, "scene-height", 0, "Height of the drawing area")
	cmd.Flags().IntVarP(&sceneWidth, "scene-width", "w", 0, "Width of the drawing area")
	cmd.Flags().IntVarP(&zoom, "zoom", "z", 0, "Zoom of the drawing area")
	cmd.Flags().BoolVarP(&showLayers, "show-layers", "l", false, "Show layers on the drawing area")
	cmd.Flags().BoolVarP(&snapToGrid, "snap-to-grid", "g", false, "Snap to grid on the drawing area")
	cmd.Flags().BoolVar(&showGrid, "show-grid", false, "Show the grid on the drawing area")
	cmd.Flags().IntVar(&gridSize, "grid-size", 0, "Grid size for the drawing area for nodes")
	cmd.Flags().IntVar(&drawingGridSize, "drawing-grid-size", 0, "Grid size for the drawing area for drawings")
	cmd.Flags().BoolVar(&showInterfaceLabels, "show-interface-labels", false, "Show interface labels on the drawing area")
	cmd.Flags().StringVarP(&supplierLogo, "supplier-logo", "", "", "Path to the project supplier logo")
	cmd.Flags().StringVarP(&supplierURL, "supplier-url", "", "", "URL to the project supplier site")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
