package schemas

type Version struct {
	ControllerHost string `json:"controller_host"`
	Version        string `json:"version"`
	Local          bool   `json:"local"`
}
