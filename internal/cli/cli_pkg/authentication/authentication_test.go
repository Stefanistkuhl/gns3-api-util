package authentication

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathUtils"
	"github.com/0xveya/gns3util/pkg/api/schemas"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"http://example.com:8080", "example.com"},
		{"https://example.com:8443", "example.com"},
		{"example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeURL(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeURL(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSaveAuthData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gns3_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	keyFile := filepath.Join(tempDir, "gns3key")

	cfg := config.GlobalOptions{
		Server:   "http://example.com",
		KeyFile:  keyFile,
		Insecure: false,
	}

	token := schemas.Token{
		AccessToken: stringPtr("test_token"),
		TokenType:   stringPtr("Bearer"),
	}

	username := "testuser"

	err = SaveAuthData(cfg, token, username)
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	if _, statErr := os.Stat(keyFile); os.IsNotExist(statErr) {
		t.Error("Key file was not created")
	}

	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("Key file has wrong permissions: got %o, want %o", info.Mode().Perm(), 0o600)
	}

	keys, err := pathUtils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}

	key := keys[0]
	if key.ServerURL != cfg.Server {
		t.Errorf("ServerURL = %v, want %v", key.ServerURL, cfg.Server)
	}
	if key.User != username {
		t.Errorf("User = %v, want %v", key.User, username)
	}
	if key.AccessToken != *token.AccessToken {
		t.Errorf("AccessToken = %v, want %v", key.AccessToken, *token.AccessToken)
	}
	if key.TokenType != *token.TokenType {
		t.Errorf("TokenType = %v, want %v", key.TokenType, *token.TokenType)
	}
}

func TestSaveAuthDataUpdateExisting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gns3_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	keyFile := filepath.Join(tempDir, "gns3key")

	cfg := config.GlobalOptions{
		Server:   "http://example.com",
		KeyFile:  keyFile,
		Insecure: false,
	}

	token1 := schemas.Token{
		AccessToken: stringPtr("token1"),
		TokenType:   stringPtr("Bearer"),
	}

	err = SaveAuthData(cfg, token1, "user1")
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	token2 := schemas.Token{
		AccessToken: stringPtr("token2"),
		TokenType:   stringPtr("Bearer"),
	}

	err = SaveAuthData(cfg, token2, "user2")
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	keys, err := pathUtils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}

	key := keys[0]
	if key.AccessToken != "token2" {
		t.Errorf("AccessToken = %v, want %v", key.AccessToken, "token2")
	}
	if key.User != "user2" {
		t.Errorf("User = %v, want %v", key.User, "user2")
	}
}

func TestSaveAuthDataMultipleServers(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gns3_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	keyFile := filepath.Join(tempDir, "gns3key")

	cfg1 := config.GlobalOptions{
		Server:   "http://server1.com",
		KeyFile:  keyFile,
		Insecure: false,
	}
	token1 := schemas.Token{
		AccessToken: stringPtr("token1"),
		TokenType:   stringPtr("Bearer"),
	}
	err = SaveAuthData(cfg1, token1, "user1")
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	cfg2 := config.GlobalOptions{
		Server:   "http://server2.com",
		KeyFile:  keyFile,
		Insecure: false,
	}
	token2 := schemas.Token{
		AccessToken: stringPtr("token2"),
		TokenType:   stringPtr("Bearer"),
	}
	err = SaveAuthData(cfg2, token2, "user2")
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	keys, err := pathUtils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestLoadKeys(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gns3_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	t.Run("non-existent file", func(t *testing.T) {
		keys, err := LoadKeys(filepath.Join(tempDir, "nonexistent"))
		if err != nil {
			t.Errorf("LoadKeys() error = %v, want nil", err)
		}
		if keys != nil {
			t.Errorf("LoadKeys() = %v, want nil", keys)
		}
	})

	t.Run("existing file", func(t *testing.T) {
		keyFile := filepath.Join(tempDir, "gns3key")

		key := pathUtils.GNS3Key{
			ServerURL:   "http://example.com",
			User:        "testuser",
			AccessToken: "testtoken",
			TokenType:   "Bearer",
		}

		data, err := json.Marshal(key)
		if err != nil {
			t.Fatalf("Failed to marshal key: %v", err)
		}

		err = os.WriteFile(keyFile, append(data, '\n'), 0o600)
		if err != nil {
			t.Fatalf("Failed to write key file: %v", err)
		}

		keys, err := LoadKeys(keyFile)
		if err != nil {
			t.Errorf("LoadKeys() error = %v, want nil", err)
		}
		if len(keys) != 1 {
			t.Errorf("LoadKeys() = %v, want 1 key", len(keys))
		}
	})
}

func TestGetKeyForServer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gns3_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	keyFile := filepath.Join(tempDir, "gns3key")

	cfg := config.GlobalOptions{
		Server:   "http://example.com",
		KeyFile:  keyFile,
		Insecure: false,
	}

	t.Run("no keys file", func(t *testing.T) {
		token, err := GetKeyForServer(cfg)
		if err == nil {
			t.Error("GetKeyForServer() should return error when no keys file exists")
		}
		if token != "" {
			t.Errorf("GetKeyForServer() = %v, want empty string", token)
		}
	})

	t.Run("matching key exists", func(t *testing.T) {
		key := pathUtils.GNS3Key{
			ServerURL:   "http://example.com",
			User:        "testuser",
			AccessToken: "testtoken",
			TokenType:   "Bearer",
		}

		data, err := json.Marshal(key)
		if err != nil {
			t.Fatalf("Failed to marshal key: %v", err)
		}

		err = os.WriteFile(keyFile, append(data, '\n'), 0o600)
		if err != nil {
			t.Fatalf("Failed to write key file: %v", err)
		}

		token, err := GetKeyForServer(cfg)
		if err != nil {
			t.Errorf("GetKeyForServer() error = %v, want nil", err)
		}
		if token != "testtoken" {
			t.Errorf("GetKeyForServer() = %v, want testtoken", token)
		}
	})

	t.Run("no matching key", func(t *testing.T) {
		key := pathUtils.GNS3Key{
			ServerURL:   "http://different.com",
			User:        "testuser",
			AccessToken: "testtoken",
			TokenType:   "Bearer",
		}

		data, err := json.Marshal(key)
		if err != nil {
			t.Fatalf("Failed to marshal key: %v", err)
		}

		err = os.WriteFile(keyFile, append(data, '\n'), 0o600)
		if err != nil {
			t.Fatalf("Failed to write key file: %v", err)
		}

		token, err := GetKeyForServer(cfg)
		if err == nil {
			t.Error("GetKeyForServer() should return error when no matching key found")
		}
		if token != "" {
			t.Errorf("GetKeyForServer() = %v, want empty string", token)
		}
	})
}

func TestTryKeys(t *testing.T) {
	cfg := config.GlobalOptions{
		Server: "http://example.com",
	}

	keys := []pathUtils.GNS3Key{
		{ServerURL: "http://different.com", AccessToken: "token1"},
		{ServerURL: "https://example.com", AccessToken: "token2"},
		{ServerURL: "http://example.com:8080", AccessToken: "token3"},
	}

	for _, key := range keys {
		if normalizeURL(cfg.Server) == normalizeURL(key.ServerURL) {
			t.Logf("Found matching key: %s -> %s", key.ServerURL, cfg.Server)
			break
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
