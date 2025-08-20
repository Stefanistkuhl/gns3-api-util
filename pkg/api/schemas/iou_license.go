package schemas

type IOULicense struct {
	IOURCContent *string `json:"iourc_content,omitempty"`
	LicenseCheck *bool   `json:"license_check,omitempty"`
}
