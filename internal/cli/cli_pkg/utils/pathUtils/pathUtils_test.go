package pathUtils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"no tilde", "/absolute/path", "/absolute/path", false},
		{"just tilde", "~", home, false},
		{"tilde with path", "~/Documents", filepath.Join(home, "Documents"), false},
		{"empty string", "", "", false},
		{"relative path", "relative/path", "relative/path", false},
		{"tilde in middle", "/path/~/middle", "/path/~/middle", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ExpandPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetGNS3Dir(t *testing.T) {
	dir, err := GetGNS3Dir()
	if err != nil {
		t.Fatalf("GetGNS3Dir() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Failed to stat GNS3 dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("GetGNS3Dir() should return a directory")
	}

	expected := ".gns3"
	if filepath.Base(dir) != expected {
		t.Errorf("GetGNS3Dir() basename = %v, want %v", filepath.Base(dir), expected)
	}
}

func TestLoadGNS3KeysFile(t *testing.T) {
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
		kf, err := LoadGNS3KeysFile(filepath.Join(tempDir, "nonexistent"))
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if kf == nil {
			t.Fatal("LoadGNS3KeysFile() should return non-nil KeyFileV2")
		}
		if len(kf.StandaloneGNS3) != 0 {
			t.Errorf("LoadGNS3KeysFile() = %v, want empty slice", kf.StandaloneGNS3)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty")
		err := os.WriteFile(emptyFile, []byte{}, 0o600)
		if err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		kf, err := LoadGNS3KeysFile(emptyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if kf == nil {
			t.Fatal("LoadGNS3KeysFile() should return non-nil KeyFileV2")
		}
		if len(kf.StandaloneGNS3) != 0 {
			t.Errorf("LoadGNS3KeysFile() = %v, want empty slice", kf.StandaloneGNS3)
		}
	})

	t.Run("legacy single key", func(t *testing.T) {
		keyFile := filepath.Join(tempDir, "single_key_legacy")
		key := LegacyGNS3Key{
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

		kf, err := LoadGNS3KeysFile(keyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if kf == nil {
			t.Fatal("LoadGNS3KeysFile() returned nil")
		}
		if len(kf.StandaloneGNS3) != 1 {
			t.Fatalf("LoadGNS3KeysFile() = %v, want 1 key", len(kf.StandaloneGNS3))
		}

		loadedKey := kf.StandaloneGNS3[0]
		if loadedKey.URL != key.ServerURL {
			t.Errorf("URL = %v, want %v", loadedKey.URL, key.ServerURL)
		}
		if loadedKey.User != key.User {
			t.Errorf("User = %v, want %v", loadedKey.User, key.User)
		}
		if loadedKey.AccessToken != key.AccessToken {
			t.Errorf("AccessToken = %v, want %v", loadedKey.AccessToken, key.AccessToken)
		}
		if loadedKey.TokenType != key.TokenType {
			t.Errorf("TokenType = %v, want %v", loadedKey.TokenType, key.TokenType)
		}
	})

	t.Run("legacy multiple keys", func(t *testing.T) {
		keyFile := filepath.Join(tempDir, "multiple_keys_legacy")

		keys := []LegacyGNS3Key{
			{
				ServerURL:   "http://server1.com",
				User:        "user1",
				AccessToken: "token1",
				TokenType:   "Bearer",
			},
			{
				ServerURL:   "http://server2.com",
				User:        "user2",
				AccessToken: "token2",
				TokenType:   "Bearer",
			},
		}

		fileData := make([]byte, 0, len(keys)*100)
		for _, key := range keys {
			data, mErr := json.Marshal(key)
			if mErr != nil {
				t.Fatalf("Failed to marshal key: %v", mErr)
			}
			fileData = append(fileData, data...)
			fileData = append(fileData, '\n')
		}

		err := os.WriteFile(keyFile, fileData, 0o600)
		if err != nil {
			t.Fatalf("Failed to write key file: %v", err)
		}

		kf, err := LoadGNS3KeysFile(keyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if kf == nil {
			t.Fatal("LoadGNS3KeysFile() returned nil")
		}
		if len(kf.StandaloneGNS3) != 2 {
			t.Fatalf("LoadGNS3KeysFile() = %v, want 2 keys", len(kf.StandaloneGNS3))
		}

		for i, expectedKey := range keys {
			if kf.StandaloneGNS3[i].URL != expectedKey.ServerURL {
				t.Errorf("Key %d URL = %v, want %v", i, kf.StandaloneGNS3[i].URL, expectedKey.ServerURL)
			}
			if kf.StandaloneGNS3[i].User != expectedKey.User {
				t.Errorf("Key %d User = %v, want %v", i, kf.StandaloneGNS3[i].User, expectedKey.User)
			}
			if kf.StandaloneGNS3[i].AccessToken != expectedKey.AccessToken {
				t.Errorf("Key %d AccessToken = %v, want %v", i, kf.StandaloneGNS3[i].AccessToken, expectedKey.AccessToken)
			}
		}
	})

	t.Run("V2 format", func(t *testing.T) {
		keyFile := filepath.Join(tempDir, "v2_key")
		kf := &KeyFileV2{Version: 2}
		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, GNS3ServerEntry{
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

		loadedKf, err := LoadGNS3KeysFile(keyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if loadedKf == nil {
			t.Fatal("LoadGNS3KeysFile() returned nil")
		}
		if loadedKf.Version != 2 {
			t.Errorf("Version = %v, want 2", loadedKf.Version)
		}
		if len(loadedKf.StandaloneGNS3) != 1 {
			t.Fatalf("Expected 1 key, got %d", len(loadedKf.StandaloneGNS3))
		}
		if loadedKf.StandaloneGNS3[0].URL != "http://example.com" {
			t.Errorf("URL = %v, want http://example.com", loadedKf.StandaloneGNS3[0].URL)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid")
		err := os.WriteFile(invalidFile, []byte("invalid json\n"), 0o600)
		if err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}

		kf, err := LoadGNS3KeysFile(invalidFile)
		if err == nil {
			t.Error("LoadGNS3KeysFile() should return error for invalid JSON")
		}
		if kf != nil {
			t.Errorf("LoadGNS3KeysFile() = %v, want nil", kf)
		}
	})
}

func TestSaveKeysFile(t *testing.T) {
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

	kf := &KeyFileV2{Version: 2}
	kf.StandaloneGNS3 = append(kf.StandaloneGNS3, GNS3ServerEntry{
		URL:         "http://example.com",
		User:        "testuser",
		AccessToken: "testtoken",
		TokenType:   "Bearer",
	})

	err = SaveKeysFile(keyFile, kf)
	if err != nil {
		t.Fatalf("SaveKeysFile() error = %v", err)
	}

	loadedKf, err := LoadGNS3KeysFile(keyFile)
	if err != nil {
		t.Fatalf("LoadGNS3KeysFile() error = %v", err)
	}

	if loadedKf.Version != 2 {
		t.Errorf("Version = %v, want 2", loadedKf.Version)
	}
	if len(loadedKf.StandaloneGNS3) != 1 {
		t.Errorf("Expected 1 key, got %d", len(loadedKf.StandaloneGNS3))
	}
	if loadedKf.StandaloneGNS3[0].URL != "http://example.com" {
		t.Errorf("URL = %v, want http://example.com", loadedKf.StandaloneGNS3[0].URL)
	}
}

func TestKeyFileV2Serialization(t *testing.T) {
	kf := &KeyFileV2{Version: 2}
	kf.StandaloneGNS3 = append(kf.StandaloneGNS3, GNS3ServerEntry{
		URL:         "http://example.com",
		User:        "testuser",
		AccessToken: "testtoken",
		TokenType:   "Bearer",
	})
	kf.Clusters = append(kf.Clusters, ClusterEntry{
		Name: "test-cluster",
		Master: ServiceEntry{
			URL:         "http://master.example.com",
			User:        "admin",
			AccessToken: "cluster-token",
			TokenType:   "Bearer",
		},
	})

	data, err := json.Marshal(kf)
	if err != nil {
		t.Fatalf("Failed to marshal KeyFileV2: %v", err)
	}

	var unmarshaled KeyFileV2
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal KeyFileV2: %v", err)
	}

	if unmarshaled.Version != 2 {
		t.Errorf("Version = %v, want 2", unmarshaled.Version)
	}
	if len(unmarshaled.StandaloneGNS3) != 1 {
		t.Errorf("StandaloneGNS3 length = %v, want 1", len(unmarshaled.StandaloneGNS3))
	}
	if len(unmarshaled.Clusters) != 1 {
		t.Errorf("Clusters length = %v, want 1", len(unmarshaled.Clusters))
	}
	if unmarshaled.StandaloneGNS3[0].URL != "http://example.com" {
		t.Errorf("URL = %v, want http://example.com", unmarshaled.StandaloneGNS3[0].URL)
	}
	if unmarshaled.Clusters[0].Name != "test-cluster" {
		t.Errorf("Cluster name = %v, want test-cluster", unmarshaled.Clusters[0].Name)
	}
}
