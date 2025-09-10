package schemas

import (
	"github.com/google/uuid"
)

type UserGroupResponse struct {
	UserGroupID uuid.UUID `json:"user_group_id"`
	Name        string    `json:"name"`
}

type UserResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	IsActive     bool      `json:"is_active"`
	Email        *string   `json:"email,omitempty"`
	FullName     *string   `json:"full_name,omitempty"`
	CreatedAt    *string   `json:"created_at,omitempty"`
	UpdatedAt    *string   `json:"updated_at,omitempty"`
	LastLogin    *string   `json:"last_login,omitempty"`
	IsSuperadmin bool      `json:"is_superadmin"`
}

type ProjectResponse struct {
	ProjectID           string    `json:"project_id"`
	Name                string    `json:"name"`
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

type ResourcePoolResponse struct {
	ResourcePoolID string `json:"resource_pool_id"`
	Name           string `json:"name"`
}

type RoleResponse struct {
	RoleID string `json:"role_id"`
	Name   string `json:"name"`
}

type ACLResponse struct {
	ACLID     string  `json:"acl_id"`
	ACEType   string  `json:"ace_type"`
	Path      string  `json:"path"`
	Propagate bool    `json:"propagate"`
	Allowed   bool    `json:"allowed"`
	GroupID   *string `json:"group_id,omitempty"`
	RoleID    *string `json:"role_id,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
}
