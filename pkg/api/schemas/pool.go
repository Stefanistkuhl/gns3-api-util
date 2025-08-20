package schemas

type ResourcePoolCreate struct {
	Name *string `json:"name"`
}

type ResourcePoolUpdate struct {
	Name *string `json:"name,omitempty"`
}
