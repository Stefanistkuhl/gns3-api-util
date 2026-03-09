package scopes

import (
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		resource Resource
		expected string
	}{
		{"read vms", Read, VMs, "read:vms"},
		{"write backups", Write, Backups, "write:backups"},
		{"delete configs", Delete, Configs, "delete:configs"},
		{"admin system", Admin, System, "admin:system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.action, tt.resource)
			if got != tt.expected {
				t.Errorf("New(%v, %v) = %v, want %v", tt.action, tt.resource, got, tt.expected)
			}
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		expected string
	}{
		{"all vms", VMs, "*:vms"},
		{"all backups", Backups, "*:backups"},
		{"all configs", Configs, "*:configs"},
		{"all system", System, "*:system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := All(tt.resource)
			if got != tt.expected {
				t.Errorf("All(%v) = %v, want %v", tt.resource, got, tt.expected)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{"superuser", Superuser, true},
		{"read vms", "read:vms", true},
		{"write backups", "write:backups", true},
		{"delete configs", "delete:configs", true},
		{"admin system", "admin:system", true},
		{"wildcard action vms", "*:vms", true},
		{"all vms", "*:vms", true},
		{"all backups", "*:backups", true},
		{"all configs", "*:configs", true},
		{"all system", "*:system", true},

		{"empty", "", false},
		{"no colon", "read", false},
		{"too many colons", "read:vms:extra", false},
		{"invalid action", "invalid:vms", false},
		{"invalid resource", "read:invalid", false},
		{"both invalid", "invalid:invalid", false},
		{"only colon", ":", false},
		{"empty action", ":vms", false},
		{"empty resource", "read:", false},
		{"spaces", "read : vms", false},
		{"wrong wildcard", "read:*", false},
		{"double wildcard", "*:*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValid(tt.scope)
			if got != tt.expected {
				t.Errorf("IsValid(%q) = %v, want %v", tt.scope, got, tt.expected)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if Superuser != "superuser" {
		t.Errorf("Superuser = %v, want %v", Superuser, "superuser")
	}

	resources := []Resource{VMs, Backups, Configs, System}
	expectedResources := []Resource{"vms", "backups", "configs", "system"}
	for i, resource := range resources {
		if resource != expectedResources[i] {
			t.Errorf("Resource constant %d = %v, want %v", i, resource, expectedResources[i])
		}
	}

	actions := []Action{Read, Write, Delete, Admin}
	expectedActions := []Action{"read", "write", "delete", "admin"}
	for i, action := range actions {
		if action != expectedActions[i] {
			t.Errorf("Action constant %d = %v, want %v", i, action, expectedActions[i])
		}
	}
}

func TestValidActionsAndResources(t *testing.T) {
	expectedActions := []Action{Read, Write, Delete, Admin}
	for _, expected := range expectedActions {
		found := slices.Contains(validActions, expected)
		if !found {
			t.Errorf("Expected action %v not found in validActions", expected)
		}
	}

	expectedResources := []Resource{VMs, Backups, Configs, System}
	for _, expected := range expectedResources {
		found := slices.Contains(validResources, expected)
		if !found {
			t.Errorf("Expected resource %v not found in validResources", expected)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("scope with extra whitespace", func(t *testing.T) {
		got := IsValid(" read:vms ")
		if got {
			t.Error("IsValid() should return false for scope with whitespace")
		}
	})

	t.Run("scope with special characters", func(t *testing.T) {
		got := IsValid("read:vms@domain")
		if got {
			t.Error("IsValid() should return false for scope with special characters")
		}
	})

	t.Run("numeric values", func(t *testing.T) {
		got := IsValid("123:456")
		if got {
			t.Error("IsValid() should return false for numeric values")
		}
	})

	t.Run("mixed case", func(t *testing.T) {
		got := IsValid("Read:VMs")
		if got {
			t.Error("IsValid() should return false for mixed case")
		}
	})
}
