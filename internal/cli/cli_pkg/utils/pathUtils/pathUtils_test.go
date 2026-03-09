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
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	t.Run("non-existent file", func(t *testing.T) {
		keys, err := LoadGNS3KeysFile(filepath.Join(tempDir, "nonexistent"))
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if len(keys) != 0 {
			t.Errorf("LoadGNS3KeysFile() = %v, want empty slice", keys)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty")
		err := os.WriteFile(emptyFile, []byte{}, 0o600)
		if err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		keys, err := LoadGNS3KeysFile(emptyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if len(keys) != 0 {
			t.Errorf("LoadGNS3KeysFile() = %v, want empty slice", keys)
		}
	})

	t.Run("single key", func(t *testing.T) {
		keyFile := filepath.Join(tempDir, "single_key")
		key := GNS3Key{
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

		keys, err := LoadGNS3KeysFile(keyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if len(keys) != 1 {
			t.Errorf("LoadGNS3KeysFile() = %v, want 1 key", len(keys))
		}

		loadedKey := keys[0]
		if loadedKey.ServerURL != key.ServerURL {
			t.Errorf("ServerURL = %v, want %v", loadedKey.ServerURL, key.ServerURL)
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

	t.Run("multiple keys", func(t *testing.T) {
		keyFile := filepath.Join(tempDir, "multiple_keys")

		keys := []GNS3Key{
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
			data, err := json.Marshal(key)
			if err != nil {
				t.Fatalf("Failed to marshal key: %v", err)
			}
			fileData = append(fileData, data...)
			fileData = append(fileData, '\n')
		}

		err := os.WriteFile(keyFile, fileData, 0o600)
		if err != nil {
			t.Fatalf("Failed to write key file: %v", err)
		}

		loadedKeys, err := LoadGNS3KeysFile(keyFile)
		if err != nil {
			t.Errorf("LoadGNS3KeysFile() error = %v, want nil", err)
		}
		if len(loadedKeys) != 2 {
			t.Errorf("LoadGNS3KeysFile() = %v, want 2 keys", len(loadedKeys))
		}

		for i, expectedKey := range keys {
			if loadedKeys[i].ServerURL != expectedKey.ServerURL {
				t.Errorf("Key %d ServerURL = %v, want %v", i, loadedKeys[i].ServerURL, expectedKey.ServerURL)
			}
			if loadedKeys[i].User != expectedKey.User {
				t.Errorf("Key %d User = %v, want %v", i, loadedKeys[i].User, expectedKey.User)
			}
			if loadedKeys[i].AccessToken != expectedKey.AccessToken {
				t.Errorf("Key %d AccessToken = %v, want %v", i, loadedKeys[i].AccessToken, expectedKey.AccessToken)
			}
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid")
		err := os.WriteFile(invalidFile, []byte("invalid json\n"), 0o600)
		if err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}

		keys, err := LoadGNS3KeysFile(invalidFile)
		if err == nil {
			t.Error("LoadGNS3KeysFile() should return error for invalid JSON")
		}
		if keys != nil {
			t.Errorf("LoadGNS3KeysFile() = %v, want nil", keys)
		}
	})

	t.Run("partial JSON", func(t *testing.T) {
		partialFile := filepath.Join(tempDir, "partial")
		content := `{"server_url": "http://example.com"}` + "\n" + `invalid json` + "\n"
		err := os.WriteFile(partialFile, []byte(content), 0o600)
		if err != nil {
			t.Fatalf("Failed to write partial file: %v", err)
		}

		keys, err := LoadGNS3KeysFile(partialFile)
		if err == nil {
			t.Error("LoadGNS3KeysFile() should return error for partial JSON")
		}
		if keys != nil {
			t.Errorf("LoadGNS3KeysFile() = %v, want nil", keys)
		}
	})
}

func TestGNS3KeySerialization(t *testing.T) {
	key := GNS3Key{
		ServerURL:   "http://example.com",
		User:        "testuser",
		AccessToken: "testtoken",
		TokenType:   "Bearer",
	}

	data, err := json.Marshal(key)
	if err != nil {
		t.Fatalf("Failed to marshal GNS3Key: %v", err)
	}

	var unmarshaled GNS3Key
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal GNS3Key: %v", err)
	}

	if unmarshaled.ServerURL != key.ServerURL {
		t.Errorf("ServerURL = %v, want %v", unmarshaled.ServerURL, key.ServerURL)
	}
	if unmarshaled.User != key.User {
		t.Errorf("User = %v, want %v", unmarshaled.User, key.User)
	}
	if unmarshaled.AccessToken != key.AccessToken {
		t.Errorf("AccessToken = %v, want %v", unmarshaled.AccessToken, key.AccessToken)
	}
	if unmarshaled.TokenType != key.TokenType {
		t.Errorf("TokenType = %v, want %v", unmarshaled.TokenType, key.TokenType)
	}
}

func TestGNS3KeyEmptyFields(t *testing.T) {
	key := GNS3Key{}

	data, err := json.Marshal(key)
	if err != nil {
		t.Fatalf("Failed to marshal empty GNS3Key: %v", err)
	}

	var unmarshaled GNS3Key
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty GNS3Key: %v", err)
	}

	if unmarshaled.ServerURL != "" {
		t.Errorf("ServerURL = %v, want empty string", unmarshaled.ServerURL)
	}
	if unmarshaled.User != "" {
		t.Errorf("User = %v, want empty string", unmarshaled.User)
	}
	if unmarshaled.AccessToken != "" {
		t.Errorf("AccessToken = %v, want empty string", unmarshaled.AccessToken)
	}
	if unmarshaled.TokenType != "" {
		t.Errorf("TokenType = %v, want empty string", unmarshaled.TokenType)
	}
}
