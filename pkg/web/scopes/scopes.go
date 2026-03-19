package scopes

import (
	"fmt"
	"slices"
	"strings"
)

type (
	Resource string
	Action   string
)

const (
	Superuser          = "superuser"
	VMs       Resource = "vms"
	Backups   Resource = "backups"
	Configs   Resource = "configs"
	System    Resource = "system"
	Metrics   Resource = "metrics"
	Files     Resource = "files"
	Jobs      Resource = "jobs"

	Admin   Action = "admin"
	Read    Action = "read"
	Write   Action = "write"
	Execute Action = "exceute"
	Delete  Action = "delete"
)

func New(a Action, r Resource) string {
	return fmt.Sprintf("%s:%s", a, r)
}

func All(r Resource) string {
	return fmt.Sprintf("*:%s", r)
}

var (
	validActions   = []Action{Read, Write, Delete, Admin, Execute}
	validResources = []Resource{VMs, Backups, Configs, System, Metrics, Files, Jobs}
)

func IsValid(scope string) bool {
	if scope == Superuser {
		return true
	}

	parts := strings.Split(scope, ":")
	if len(parts) != 2 {
		return false
	}

	act := Action(parts[0])
	res := Resource(parts[1])

	isValidAction := string(act) == "*" || slices.Contains(validActions, act)
	isValidResource := slices.Contains(validResources, res)

	return isValidAction && isValidResource
}
