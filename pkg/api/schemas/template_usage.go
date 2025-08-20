package schemas

type TemplateUsage struct {
	X         *int    `json:"x,omitempty"`
	Y         *int    `json:"y,omitempty"`
	Name      *string `json:"name,omitempty"`
	ComputeID *string `json:"compute_id,omitempty"`
}
