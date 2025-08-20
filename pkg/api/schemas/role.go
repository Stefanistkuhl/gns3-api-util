package schemas

type RoleCreate struct {
	Name        *string `json:"name"`
	Description *string `json:"description,omitempty"`
}

type RoleUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
