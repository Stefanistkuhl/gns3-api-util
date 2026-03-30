package scopes

import (
	"slices"
	"strings"
)

type Action = string

type Resource = string

const (
	// Superuser is a special scope that grants all permissions.
	Superuser = "superuser"

	// Actions
	Read    Action = "read"
	Write   Action = "write"
	Execute Action = "execute"
	Delete  Action = "delete"
	Admin   Action = "admin"

	// Resources
	VMs     Resource = "vms"
	Backups Resource = "backups"
	Configs Resource = "configs"
	System  Resource = "system"
	Metrics Resource = "metrics"
	Files   Resource = "files"
	Jobs    Resource = "jobs"
)

var (
	validActions   = []Action{Read, Write, Execute, Delete, Admin}
	validResources = []Resource{VMs, Backups, Configs, System, Metrics, Files, Jobs}
)

func New(action Action, resource Resource) string {
	return action + ":" + resource
}

func All(resource Resource) string {
	return "*:" + resource
}

func IsValid(scope string) bool {
	if scope == Superuser {
		return true
	}

	before, after, ok := strings.Cut(scope, ":")
	if !ok {
		return false
	}

	action := before
	resource := after

	if action == "" || resource == "" {
		return false
	}

	// No additional colons allowed in the resource segment.
	if strings.Contains(resource, ":") {
		return false
	}

	validAction := action == "*" || slices.Contains(validActions, action)
	validResource := slices.Contains(validResources, resource)

	return validAction && validResource
}
