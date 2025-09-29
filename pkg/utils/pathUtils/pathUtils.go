package pathUtils

import (
	"encoding/json"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"io"
	"os"
	"path/filepath"
)

type GNS3Key struct {
	ServerURL   string `json:"server_url"`
	User        string `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func ExpandPath(p string) (string, error) {
	if p == "" || p[0] != '~' {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if p == "~" {
		return home, nil
	}

	return filepath.Join(home, p[2:]), nil
}

func GetGNS3Dir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("could not detect home dir: %w", err)
	}

	gns3Dir := filepath.Join(home, ".gns3")
	info, err := os.Stat(gns3Dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(gns3Dir, 0o755); err != nil {
			return "", fmt.Errorf("could not create %q: %w", gns3Dir, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("could not stat %q: %w", gns3Dir, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("%q already exists and is not a directory", gns3Dir)
	}

	return gns3Dir, nil
}

func LoadGNS3KeysFile(path string) ([]GNS3Key, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not open %q: %w", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}()

	var keys []GNS3Key
	dec := json.NewDecoder(f)
	for {
		var k GNS3Key
		if err := dec.Decode(&k); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode JSON in %q: %w", path, err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}
