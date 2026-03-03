package scopes

import "fmt"

type (
	Resource string
	Action   string
)

const (
	VMs     Resource = "vms"
	Backups Resource = "backups"
	Configs Resource = "configs"
	System  Resource = "system"

	Read   Action = "read"
	Write  Action = "write"
	Delete Action = "delete"
	Admin  Action = "admin"
)

func New(a Action, r Resource) string {
	return fmt.Sprintf("%s:%s", a, r)
}

func All(r Resource) string {
	return fmt.Sprintf("*:%s", r)
}
