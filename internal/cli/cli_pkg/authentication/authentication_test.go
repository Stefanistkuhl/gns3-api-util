package authentication

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api/schemas"
)

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

	cfg := &config.GlobalOptions{
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
		t.Fatal("Key file was not created")
	}

	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("Key file has wrong permissions: got %o, want %o", info.Mode().Perm(), 0o600)
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(kf.StandaloneGNS3) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(kf.StandaloneGNS3))
	}

	key := kf.StandaloneGNS3[0]
	if key.URL != cfg.Server {
		t.Errorf("URL = %v, want %v", key.URL, cfg.Server)
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

	cfg := &config.GlobalOptions{
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

	kf, err := pathutils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(kf.StandaloneGNS3) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(kf.StandaloneGNS3))
	}

	key := kf.StandaloneGNS3[0]
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

	cfg1 := &config.GlobalOptions{
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

	cfg2 := &config.GlobalOptions{
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

	kf, err := pathutils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(kf.StandaloneGNS3) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(kf.StandaloneGNS3))
	}
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

	cfg := &config.GlobalOptions{
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
		kf := &pathutils.KeyFileV2{Version: 2}
		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, pathutils.GNS3ServerEntry{
			URL:         "http://example.com",
			User:        "testuser",
			AccessToken: "testtoken",
			TokenType:   "Bearer",
		})

		data, err := json.Marshal(kf)
		if err != nil {
			t.Fatalf("Failed to marshal keys: %v", err)
		}

		err = os.WriteFile(keyFile, data, 0o600)
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
		kf := &pathutils.KeyFileV2{Version: 2}
		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, pathutils.GNS3ServerEntry{
			URL:         "http://different.com",
			User:        "testuser",
			AccessToken: "testtoken",
			TokenType:   "Bearer",
		})

		data, err := json.Marshal(kf)
		if err != nil {
			t.Fatalf("Failed to marshal keys: %v", err)
		}

		err = os.WriteFile(keyFile, data, 0o600)
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

func stringPtr(s string) *string {
	return &s
}
