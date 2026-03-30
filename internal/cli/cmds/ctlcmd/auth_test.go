package ctlcmd

import (
	"testing"
)

func TestNewAuthCmd_Structure(t *testing.T) {
	cmd := NewAuthCmd()

	if cmd == nil {
		t.Fatal("NewAuthCmd() returned nil")
		return
	}
	if cmd.Use != "auth" {
		t.Errorf("Use = %v, want auth", cmd.Use)
	}

	expected := map[string]bool{
		"status":           false,
		"perms":            false,
		"set-default-user": false,
		"add-user":         false,
	}
	for _, sub := range cmd.Commands() {
		expected[sub.Name()] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q not found under 'ctl auth'", name)
		}
	}
}

func TestNewAuthPermsCmd(t *testing.T) {
	cmd := NewAuthPermsCmd()

	if cmd == nil {
		t.Fatal("NewAuthPermsCmd() returned nil")
		return
	}
	if cmd.Use != "perms" {
		t.Errorf("Use = %v, want perms", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
	if cmd.Annotations["auth-mode"] != "flexible" {
		t.Errorf("auth-mode = %v, want flexible", cmd.Annotations["auth-mode"])
	}
}

func TestNewSetDefaultClusterUserCmd(t *testing.T) {
	cmd := NewSetDefaultClusterUserCmd()

	if cmd == nil {
		t.Fatal("NewSetDefaultClusterUserCmd() returned nil")
		return
	}
	if cmd.Use != "set-default-user [username]" {
		t.Errorf("Use = %v, want set-default-user [username]", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("RunE should not be nil")
	}

	if cmd.Annotations["auth-mode"] != "cluster-only" {
		t.Errorf("auth-mode = %v, want cluster-only", cmd.Annotations["auth-mode"])
	}

	if cmd.Flags().Lookup("server") != nil {
		t.Error("set-default-user (ctl) should not have its own --server flag")
	}
	if cmd.Flags().Lookup("cluster") != nil {
		t.Error("set-default-user (ctl) should not have its own --cluster flag")
	}
}

func TestNewAddClusterUserCmd(t *testing.T) {
	cmd := NewAddClusterUserCmd()

	if cmd == nil {
		t.Fatal("NewAddClusterUserCmd() returned nil")
		return
	}
	if cmd.Use != "add-user [token]" {
		t.Errorf("Use = %v, want add-user [token]", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
	if cmd.Annotations["auth-mode"] != "cluster-only" {
		t.Errorf("auth-mode = %v, want cluster-only", cmd.Annotations["auth-mode"])
	}
}
