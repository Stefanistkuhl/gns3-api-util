package schemas

type ComputeCreate struct {
	Protocol  *string `json:"protocol,omitempty"`
	Host      *string `json:"host,omitempty"`
	Port      *int    `json:"port,omitempty"`
	User      *string `json:"user,omitempty"`
	Password  *string `json:"password,omitempty"`
	Name      *string `json:"name,omitempty"`
	ComputeID *string `json:"compute_id,omitempty"`
}

type ComputeUpdate struct {
	Protocol  *string `json:"protocol,omitempty"`
	Host      *string `json:"host,omitempty"`
	Port      *int    `json:"port,omitempty"`
	User      *string `json:"user,omitempty"`
	Password  *string `json:"password,omitempty"`
	Name      *string `json:"name,omitempty"`
	ComputeID *string `json:"compute_id,omitempty"`
}
