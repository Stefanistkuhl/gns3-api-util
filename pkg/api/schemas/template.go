package schemas

type TemplateCreate struct {
	TemplateID        *string `json:"template_id,omitempty"`
	Name              *string `json:"name,omitempty"`
	Version           *string `json:"version,omitempty"`
	Category          *string `json:"category,omitempty"`
	DefaultNameFormat *string `json:"default_name_format,omitempty"`
	Symbol            *string `json:"symbol,omitempty"`
	TemplateType      *string `json:"template_type,omitempty"`
	ComputeID         *string `json:"compute_id,omitempty"`
	Usage             *string `json:"usage,omitempty"`
}

type TemplateUpdate struct {
	TemplateID        *string `json:"template_id,omitempty"`
	Name              *string `json:"name,omitempty"`
	Version           *string `json:"version,omitempty"`
	Category          *string `json:"category,omitempty"`
	DefaultNameFormat *string `json:"default_name_format,omitempty"`
	Symbol            *string `json:"symbol,omitempty"`
	TemplateType      *string `json:"template_type,omitempty"`
	ComputeID         *string `json:"compute_id,omitempty"`
	Usage             *string `json:"usage,omitempty"`
}
