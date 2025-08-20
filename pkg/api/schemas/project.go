package schemas

type Supplier struct {
	Logo *string `json:"logo,omitempty"`
	URL  *string `json:"url,omitempty"`
}

type ProjectCreate struct {
	Name                *string   `json:"name,omitempty"`
	ProjectID           *string   `json:"project_id,omitempty"`
	Path                *string   `json:"path,omitempty"`
	AutoClose           *bool     `json:"auto_close,omitempty"`
	AutoOpen            *bool     `json:"auto_open,omitempty"`
	AutoStart           *bool     `json:"auto_start,omitempty"`
	SceneHeight         *int      `json:"scene_height,omitempty"`
	SceneWidth          *int      `json:"scene_width,omitempty"`
	Zoom                *int      `json:"zoom,omitempty"`
	ShowLayers          *bool     `json:"show_layers,omitempty"`
	SnapToGrid          *bool     `json:"snap_to_grid,omitempty"`
	ShowGrid            *bool     `json:"show_grid,omitempty"`
	GridSize            *int      `json:"grid_size,omitempty"`
	DrawingGridSize     *int      `json:"drawing_grid_size,omitempty"`
	ShowInterfaceLabels *bool     `json:"show_interface_labels,omitempty"`
	Supplier            *Supplier `json:"supplier,omitempty"`
}

type ProjectUpdate struct {
	Name                *string   `json:"name,omitempty"`
	ProjectID           *string   `json:"project_id,omitempty"`
	Path                *string   `json:"path,omitempty"`
	AutoClose           *bool     `json:"auto_close,omitempty"`
	AutoOpen            *bool     `json:"auto_open,omitempty"`
	AutoStart           *bool     `json:"auto_start,omitempty"`
	SceneHeight         *int      `json:"scene_height,omitempty"`
	SceneWidth          *int      `json:"scene_width,omitempty"`
	Zoom                *int      `json:"zoom,omitempty"`
	ShowLayers          *bool     `json:"show_layers,omitempty"`
	SnapToGrid          *bool     `json:"snap_to_grid,omitempty"`
	ShowGrid            *bool     `json:"show_grid,omitempty"`
	GridSize            *int      `json:"grid_size,omitempty"`
	DrawingGridSize     *int      `json:"drawing_grid_size,omitempty"`
	ShowInterfaceLabels *bool     `json:"show_interface_labels,omitempty"`
	Supplier            *Supplier `json:"supplier,omitempty"`
}

type ProjectDuplicate struct {
	Name        string  `json:"name"`
	ProjectID   *string `json:"project_id,omitempty"`
	Path        *string `json:"path,omitempty"`
	AutoClose   *bool   `json:"auto_close,omitempty"`
	AutoOpen    *bool   `json:"auto_open,omitempty"`
	AutoStart   *bool   `json:"auto_start,omitempty"`
	SceneHeight *int    `json:"scene_height,omitempty"`
	SceneWidth  *int    `json:"scene_width,omitempty"`
	Zoom        *int    `json:"zoom,omitempty"`
	ShowLayers  *bool   `json:"show_layers,omitempty"`
	SnapToGrid  *bool   `json:"snap_to_grid,omitempty"`
	ShowGrid    *bool   `json:"show_grid,omitempty"`
}
