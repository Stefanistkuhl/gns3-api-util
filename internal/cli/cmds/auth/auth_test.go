package auth

import (
	"testing"
)

func TestNewAuthCmdGroup(t *testing.T) {
	cmd := NewAuthCmdGroup()

	if cmd == nil {
		t.Fatal("NewAuthCmdGroup() returned nil")
		return
	}

	if cmd.Use != "auth" {
		t.Errorf("Use = %v, want %v", cmd.Use, "auth")
	}
	if cmd.Short != "Authentication commands" {
		t.Errorf("Short = %v, want %v", cmd.Short, "Authentication commands")
	}
	if cmd.Long != "Authentication commands" {
		t.Errorf("Long = %v, want %v", cmd.Long, "Authentication commands")
	}

	expectedSubcommands := []string{"status", "login", "set-default-user"}
	foundSubcommands := make(map[string]bool)

	for _, subcmd := range cmd.Commands() {
		foundSubcommands[subcmd.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !foundSubcommands[expected] {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

func TestNewAuthStatusCmd(t *testing.T) {
	cmd := NewAuthStatusCmd()

	if cmd == nil {
		t.Fatal("NewAuthStatusCmd() returned nil")
		return
	}

	if cmd.Use != "status" {
		t.Errorf("Use = %v, want %v", cmd.Use, "status")
	}
	if cmd.Short != "Check the current status of your Authentication" {
		t.Errorf("Short = %v, want %v", cmd.Short, "Check the current status of your Authentication")
	}
	if cmd.Long != "Check the current status of your Authentication" {
		t.Errorf("Long = %v, want %v", cmd.Long, "Check the current status of your Authentication")
	}

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}
}

func TestNewAuthLoginCmd(t *testing.T) {
	cmd := NewAuthLoginCmd()

	if cmd == nil {
		t.Fatal("NewAuthLoginCmd() returned nil")
		return
	}

	if cmd.Use != "login" {
		t.Errorf("Use = %v, want %v", cmd.Use, "login")
	}
	if cmd.Short != "Log in as user" {
		t.Errorf("Short = %v, want %v", cmd.Short, "Log in as user")
	}
	if cmd.Long != "Log in as a user" {
		t.Errorf("Long = %v, want %v", cmd.Long, "Log in as a user")
	}

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}

	if cmd.PreRun == nil {
		t.Error("PreRun function should not be nil")
	}

	userFlag := cmd.Flags().Lookup("user")
	if userFlag == nil {
		t.Error("User flag not found")
	} else {
		if userFlag.Name != "user" {
			t.Errorf("User flag name = %v, want %v", userFlag.Name, "user")
		}
		if userFlag.Shorthand != "u" {
			t.Errorf("User flag shorthand = %v, want %v", userFlag.Shorthand, "u")
		}
	}

	passwordFlag := cmd.Flags().Lookup("password")
	if passwordFlag == nil {
		t.Error("Password flag not found")
	} else {
		if passwordFlag.Name != "password" {
			t.Errorf("Password flag name = %v, want %v", passwordFlag.Name, "password")
		}
		if passwordFlag.Shorthand != "p" {
			t.Errorf("Password flag shorthand = %v, want %v", passwordFlag.Shorthand, "p")
		}
	}
}

func TestAuthCommandIntegration(t *testing.T) {
	authCmd := NewAuthCmdGroup()

	loginCmd, _, err := authCmd.Find([]string{"login"})
	if err != nil {
		t.Errorf("Failed to find login subcommand: %v", err)
	}
	if loginCmd.Use != "login" {
		t.Errorf("Found command use = %v, want %v", loginCmd.Use, "login")
	}

	statusCmd, _, err := authCmd.Find([]string{"status"})
	if err != nil {
		t.Errorf("Failed to find status subcommand: %v", err)
	}
	if statusCmd.Use != "status" {
		t.Errorf("Found command use = %v, want %v", statusCmd.Use, "status")
	}
}

func TestAuthCommandHelp(t *testing.T) {
	cmd := NewAuthCmdGroup()

	err := cmd.Help()
	if err != nil {
		t.Errorf("Help() returned error: %v", err)
	}
}

func TestNewSetDefaultUserCmd(t *testing.T) {
	cmd := NewSetDefaultUserCmd()

	if cmd == nil {
		t.Fatal("NewSetDefaultUserCmd() returned nil")
		return
	}
	if cmd.Use != "set-default-user [username]" {
		t.Errorf("Use = %v, want set-default-user [username]", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("RunE should not be nil")
	}

	// The standalone version reads --server from the global flag (cfg.Server),
	// so no local --server flag should exist on this command.
	if cmd.Flags().Lookup("server") != nil {
		t.Error("set-default-user should not have its own --server flag; it uses the global -s")
	}

	if cmd.Annotations["auth-mode"] != "none" {
		t.Errorf("auth-mode annotation = %v, want none", cmd.Annotations["auth-mode"])
	}
}
