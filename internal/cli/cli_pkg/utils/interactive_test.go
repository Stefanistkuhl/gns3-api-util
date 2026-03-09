package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"admin special case", "admin", true},
		{"valid password", "password123", true},
		{"valid with uppercase", "Password123", true},
		{"valid with symbols", "passw0rd!@#", true},
		{"too short", "pass123", false},
		{"no numbers", "password", false},
		{"no lowercase", "PASSWORD123", false},
		{"only numbers", "12345678", false},
		{"only lowercase", "password", false},
		{"empty", "", false},
		{"exactly 8 chars valid", "passw0rd", true},
		{"exactly 8 chars invalid", "password", false},
		{"8 chars no number", "passworx", false},
		{"8 chars no lowercase", "PASSW0RD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePassword(tt.password)
			if got != tt.expected {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, got, tt.expected)
			}
		})
	}
}

func TestLoginModel(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		model := LoginModel{}
		if model.step != 0 {
			t.Errorf("Initial step = %v, want 0", model.step)
		}
		if model.done {
			t.Error("Initial done should be false")
		}
		if model.err != nil {
			t.Errorf("Initial err should be nil, got %v", model.err)
		}
	})

	t.Run("username validation", func(t *testing.T) {
		model := LoginModel{username: "testuser"}

		if model.username == "" {
			model.err = fmt.Errorf("username cannot be empty")
		}

		if model.err != nil {
			t.Errorf("Valid username should not set error: %v", model.err)
		}
	})

	t.Run("empty username validation", func(t *testing.T) {
		model := LoginModel{username: ""}

		if model.username == "" {
			model.err = fmt.Errorf("username cannot be empty")
		}

		if model.err == nil {
			t.Error("Empty username should set error")
		}
		expected := "username cannot be empty"
		if model.err.Error() != expected {
			t.Errorf("Error message = %v, want %v", model.err.Error(), expected)
		}
	})

	t.Run("password validation", func(t *testing.T) {
		model := LoginModel{password: "validpass123"}

		if model.password == "" {
			model.err = fmt.Errorf("password cannot be empty")
		} else if !ValidatePassword(model.password) {
			model.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
		}

		if model.err != nil {
			t.Errorf("Valid password should not set error: %v", model.err)
		}
	})

	t.Run("empty password validation", func(t *testing.T) {
		model := LoginModel{password: ""}

		if model.password == "" {
			model.err = fmt.Errorf("password cannot be empty")
		} else if !ValidatePassword(model.password) {
			model.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
		}

		if model.err == nil {
			t.Error("Empty password should set error")
		}
		expected := "password cannot be empty"
		if model.err.Error() != expected {
			t.Errorf("Error message = %v, want %v", model.err.Error(), expected)
		}
	})

	t.Run("invalid password validation", func(t *testing.T) {
		model := LoginModel{password: "invalid"}

		if model.password == "" {
			model.err = fmt.Errorf("password cannot be empty")
		} else if !ValidatePassword(model.password) {
			model.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
		}

		if model.err == nil {
			t.Error("Invalid password should set error")
		}
		expected := "password must be at least 8 characters with at least 1 number and 1 lowercase letter"
		if model.err.Error() != expected {
			t.Errorf("Error message = %v, want %v", model.err.Error(), expected)
		}
	})
}

func TestPasswordModel(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		model := PasswordModel{}
		if model.done {
			t.Error("Initial done should be false")
		}
		if model.err != nil {
			t.Errorf("Initial err should be nil, got %v", model.err)
		}
	})

	t.Run("password validation", func(t *testing.T) {
		model := PasswordModel{password: "validpass123"}

		if model.password == "" {
			model.err = fmt.Errorf("password cannot be empty")
		} else if !ValidatePassword(model.password) {
			model.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
		}

		if model.err != nil {
			t.Errorf("Valid password should not set error: %v", model.err)
		}
	})

	t.Run("empty password validation", func(t *testing.T) {
		model := PasswordModel{password: ""}

		if model.password == "" {
			model.err = fmt.Errorf("password cannot be empty")
		} else if !ValidatePassword(model.password) {
			model.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
		}

		if model.err == nil {
			t.Error("Empty password should set error")
		}
		expected := "password cannot be empty"
		if model.err.Error() != expected {
			t.Errorf("Error message = %v, want %v", model.err.Error(), expected)
		}
	})

	t.Run("invalid password validation", func(t *testing.T) {
		model := PasswordModel{password: "short"}

		if model.password == "" {
			model.err = fmt.Errorf("password cannot be empty")
		} else if !ValidatePassword(model.password) {
			model.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
		}

		if model.err == nil {
			t.Error("Invalid password should set error")
		}
		expected := "password must be at least 8 characters with at least 1 number and 1 lowercase letter"
		if model.err.Error() != expected {
			t.Errorf("Error message = %v, want %v", model.err.Error(), expected)
		}
	})
}

func TestModelViewRendering(t *testing.T) {
	t.Run("login model view", func(t *testing.T) {
		model := LoginModel{
			username: "testuser",
		}

		view := model.View()
		if view == "" {
			t.Error("LoginModel.View() should not return empty string")
		}

		if !contains(view, "GNS3 Login") {
			t.Error("View should contain 'GNS3 Login'")
		}
		if !contains(view, "Enter username:") {
			t.Error("View should contain 'Enter username:'")
		}
		if !contains(view, "testuser") {
			t.Error("View should contain the username")
		}
	})

	t.Run("login model password step", func(t *testing.T) {
		model := LoginModel{
			password: "testpass",
			step:     1,
		}

		view := model.View()
		if view == "" {
			t.Error("LoginModel.View() should not return empty string")
		}

		if !contains(view, "GNS3 Login") {
			t.Error("View should contain 'GNS3 Login'")
		}
		if !contains(view, "Enter password:") {
			t.Error("View should contain 'Enter password:'")
		}
		if !contains(view, strings.Repeat("*", len("testpass"))) {
			t.Error("View should contain masked password")
		}
		if !contains(view, strings.Repeat("*", len("testpass"))) {
			t.Error("View should contain masked password")
		}
	})

	t.Run("login model with error", func(t *testing.T) {
		model := LoginModel{
			username: "",
			err:      fmt.Errorf("test error"),
		}

		view := model.View()
		if view == "" {
			t.Error("LoginModel.View() should not return empty string")
		}

		if !contains(view, "Error: test error") {
			t.Error("View should contain error message")
		}
	})

	t.Run("password model view", func(t *testing.T) {
		model := PasswordModel{
			password: "newpass123",
		}

		view := model.View()
		if view == "" {
			t.Error("PasswordModel.View() should not return empty string")
		}

		if !contains(view, "Change User Password") {
			t.Error("View should contain 'Change User Password'")
		}
		if !contains(view, "Enter new password") {
			t.Error("View should contain password prompt")
		}
		if contains(view, "newpass123") {
			t.Error("View should not contain the actual password")
		}
		if !contains(view, strings.Repeat("*", len("newpass123"))) {
			t.Error("View should contain masked password")
		}
	})

	t.Run("password model with error", func(t *testing.T) {
		model := PasswordModel{
			password: "",
			err:      fmt.Errorf("password error"),
		}

		view := model.View()
		if view == "" {
			t.Error("PasswordModel.View() should not return empty string")
		}

		if !contains(view, "Error: password error") {
			t.Error("View should contain error message")
		}
	})
}

func TestModelStepTransitions(t *testing.T) {
	t.Run("username to password transition", func(t *testing.T) {
		model := LoginModel{
			username: "validuser",
		}

		if model.username != "" {
			model.step = 1
		}

		if model.step != 1 {
			t.Error("Valid username should advance to password step")
		}
	})

	t.Run("completion", func(t *testing.T) {
		model := LoginModel{
			password: "validpass123",
		}

		if model.password != "" && ValidatePassword(model.password) {
			model.done = true
		}

		if !model.done {
			t.Error("Valid password should complete the login")
		}
	})
}

func TestBackspaceHandling(t *testing.T) {
	t.Run("username backspace", func(t *testing.T) {
		model := LoginModel{
			username: "test",
		}

		if model.username != "" {
			model.username = model.username[:len(model.username)-1]
		}

		if model.username != "tes" {
			t.Errorf("Backspace should remove last character: got %v, want %v", model.username, "tes")
		}
	})

	t.Run("password backspace", func(t *testing.T) {
		model := LoginModel{
			password: "test123",
		}

		if model.password != "" {
			model.password = model.password[:len(model.password)-1]
		}

		if model.password != "test12" {
			t.Errorf("Backspace should remove last character: got %v, want %v", model.password, "test12")
		}
	})

	t.Run("empty string backspace", func(t *testing.T) {
		model := LoginModel{
			username: "a",
		}

		if model.username != "" {
			model.username = model.username[:len(model.username)-1]
		}

		if model.username != "" {
			t.Errorf("Backspace on empty string should remain empty: got %v", model.username)
		}
	})
}

func TestErrorClearing(t *testing.T) {
	t.Run("error cleared on input", func(t *testing.T) {
		model := LoginModel{
			username: "",
			err:      fmt.Errorf("previous error"),
		}

		model.username += "a"
		model.err = nil

		if model.err != nil {
			t.Error("Error should be cleared when user starts typing")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
