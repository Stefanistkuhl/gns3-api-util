package schemas

type Label struct {
	Text     *string `json:"text,omitempty"`
	Style    *string `json:"style,omitempty"`
	X        *int    `json:"x,omitempty"`
	Y        *int    `json:"y,omitempty"`
	Rotation *int    `json:"rotation,omitempty"`
}

type NodeCreate struct {
	ComputeID        *string        `json:"compute_id,omitempty"`
	Name             *string        `json:"name,omitempty"`
	NodeType         *string        `json:"node_type,omitempty"`
	Console          *int           `json:"console,omitempty"`
	ConsoleType      *string        `json:"console_type,omitempty"`
	ConsoleAutoStart *bool          `json:"console_auto_start,omitempty"`
	Aux              *int           `json:"aux,omitempty"`
	AuxType          *string        `json:"aux_type,omitempty"`
	Properties       map[string]any `json:"properties,omitempty"`
	Label            *Label         `json:"label,omitempty"`
	Symbol           *string        `json:"symbol,omitempty"`
	X                *int           `json:"x,omitempty"`
	Y                *int           `json:"y,omitempty"`
	Z                *int           `json:"z,omitempty"`
	Locked           *bool          `json:"locked,omitempty"`
	PortNameFormat   *string        `json:"port_name_format,omitempty"`
	PortSegmentSize  *int           `json:"port_segment_size,omitempty"`
	FirstPortName    *string        `json:"first_port_name,omitempty"`
}

type NodeUpdate struct {
	ComputeID        *string        `json:"compute_id,omitempty"`
	Name             *string        `json:"name,omitempty"`
	NodeType         *string        `json:"node_type,omitempty"`
	Console          *int           `json:"console,omitempty"`
	ConsoleType      *string        `json:"console_type,omitempty"`
	ConsoleAutoStart *bool          `json:"console_auto_start,omitempty"`
	Aux              *int           `json:"aux,omitempty"`
	AuxType          *string        `json:"aux_type,omitempty"`
	Properties       map[string]any `json:"properties,omitempty"`
	Label            *Label         `json:"label,omitempty"`
	Symbol           *string        `json:"symbol,omitempty"`
	X                *int           `json:"x,omitempty"`
	Y                *int           `json:"y,omitempty"`
	Z                *int           `json:"z,omitempty"`
	Locked           *bool          `json:"locked,omitempty"`
	PortNameFormat   *string        `json:"port_name_format,omitempty"`
	PortSegmentSize  *int           `json:"port_segment_size,omitempty"`
	FirstPortName    *string        `json:"first_port_name,omitempty"`
}
