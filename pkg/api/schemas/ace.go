package schemas

type ACECreate struct {
	ACEType   *string `json:"ace_type,omitempty"`
	Path      *string `json:"path,omitempty"`
	Propagate *bool   `json:"propagate,omitempty"`
	Allowed   *bool   `json:"allowed,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	GroupID   *string `json:"group_id,omitempty"`
	RoleID    *string `json:"role_id,omitempty"`
}

type ACEUpdate struct {
	ACEType   *string `json:"ace_type,omitempty"`
	Path      *string `json:"path,omitempty"`
	Propagate *bool   `json:"propagate,omitempty"`
	Allowed   *bool   `json:"allowed,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	GroupID   *string `json:"group_id,omitempty"`
	RoleID    *string `json:"role_id,omitempty"`
}
