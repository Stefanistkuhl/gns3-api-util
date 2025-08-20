package schemas

type UserGroupCreate struct {
	Name *string `json:"name"`
}

type UserGroupUpdate struct {
	Name *string `json:"name,omitempty"`
}
