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
	// The first entry for a server must be auto-defaulted.
	if !key.Default {
		t.Error("First saved user should be marked as default")
	}
}

// TestSaveAuthDataUpdateExisting checks that logging in again as the SAME user
// updates the token in-place rather than creating a second entry.
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

	// First login for user1.
	err = SaveAuthData(cfg, token1, "user1")
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	token2 := schemas.Token{
		AccessToken: stringPtr("token2"),
		TokenType:   stringPtr("Bearer"),
	}

	// Second login for the same user1 should update, not append.
	err = SaveAuthData(cfg, token2, "user1")
	if err != nil {
		t.Fatalf("SaveAuthData() error = %v", err)
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(kf.StandaloneGNS3) != 1 {
		t.Fatalf("Expected 1 key after updating same user, got %d", len(kf.StandaloneGNS3))
	}

	key := kf.StandaloneGNS3[0]
	if key.AccessToken != "token2" {
		t.Errorf("AccessToken = %v, want token2", key.AccessToken)
	}
	if key.User != "user1" {
		t.Errorf("User = %v, want user1", key.User)
	}
}

// TestSaveAuthDataMultiUser checks that different users on the same server each
// get their own entry. The first user should be marked as default.
func TestSaveAuthDataMultiUser(t *testing.T) {
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
		Server:  "http://example.com",
		KeyFile: keyFile,
	}

	err = SaveAuthData(cfg, schemas.Token{AccessToken: stringPtr("tok-alice"), TokenType: stringPtr("Bearer")}, "alice")
	if err != nil {
		t.Fatalf("SaveAuthData alice: %v", err)
	}
	err = SaveAuthData(cfg, schemas.Token{AccessToken: stringPtr("tok-bob"), TokenType: stringPtr("Bearer")}, "bob")
	if err != nil {
		t.Fatalf("SaveAuthData bob: %v", err)
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("Failed to load keys: %v", err)
	}

	if len(kf.StandaloneGNS3) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(kf.StandaloneGNS3))
	}

	// alice was first — she must be the default.
	aliceEntry, ok := kf.GetUserForServer("http://example.com", "alice")
	if !ok {
		t.Fatal("alice not found")
	}
	if !aliceEntry.Default {
		t.Error("alice should be the default user")
	}
	if aliceEntry.AccessToken != "tok-alice" {
		t.Errorf("alice token = %v, want tok-alice", aliceEntry.AccessToken)
	}

	bobEntry, ok := kf.GetUserForServer("http://example.com", "bob")
	if !ok {
		t.Fatal("bob not found")
	}
	if bobEntry.Default {
		t.Error("bob should not be the default user")
	}

	// Default lookup (empty name) should return alice.
	def, ok := kf.GetUserForServer("http://example.com", "")
	if !ok {
		t.Fatal("default lookup failed")
	}
	if def.User != "alice" {
		t.Errorf("default user = %v, want alice", def.User)
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

func writeKeyFile(t *testing.T, path string, kf *pathutils.KeyFileV2) {
	t.Helper()
	data, err := json.Marshal(kf)
	if err != nil {
		t.Fatalf("Failed to marshal keys: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
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

	t.Run("no keys file", func(t *testing.T) {
		_ = os.Remove(keyFile)
		cfg := &config.GlobalOptions{Server: "http://example.com", KeyFile: keyFile}
		token, err := GetKeyForServer(cfg)
		if err == nil {
			t.Error("GetKeyForServer() should return error when no keys file exists")
		}
		if token != "" {
			t.Errorf("GetKeyForServer() = %v, want empty string", token)
		}
	})

	t.Run("matching key exists — returns default", func(t *testing.T) {
		kf := &pathutils.KeyFileV2{Version: 2}
		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, pathutils.GNS3ServerEntry{
			URL:         "http://example.com",
			User:        "testuser",
			AccessToken: "testtoken",
			TokenType:   "Bearer",
			Default:     true,
		})
		writeKeyFile(t, keyFile, kf)

		cfg := &config.GlobalOptions{Server: "http://example.com", KeyFile: keyFile}
		token, err := GetKeyForServer(cfg)
		if err != nil {
			t.Errorf("GetKeyForServer() error = %v, want nil", err)
		}
		if token != "testtoken" {
			t.Errorf("GetKeyForServer() = %v, want testtoken", token)
		}
	})

	t.Run("named user selected", func(t *testing.T) {
		kf := &pathutils.KeyFileV2{Version: 2}
		kf.StandaloneGNS3 = []pathutils.GNS3ServerEntry{
			{URL: "http://example.com", User: "alice", AccessToken: "tok-alice", TokenType: "Bearer", Default: true},
			{URL: "http://example.com", User: "bob", AccessToken: "tok-bob", TokenType: "Bearer"},
		}
		writeKeyFile(t, keyFile, kf)

		cfgBob := &config.GlobalOptions{Server: "http://example.com", KeyFile: keyFile, User: "bob"}
		token, err := GetKeyForServer(cfgBob)
		if err != nil {
			t.Errorf("GetKeyForServer() error = %v, want nil", err)
		}
		if token != "tok-bob" {
			t.Errorf("GetKeyForServer() = %v, want tok-bob", token)
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
		writeKeyFile(t, keyFile, kf)

		cfg := &config.GlobalOptions{Server: "http://example.com", KeyFile: keyFile}
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
