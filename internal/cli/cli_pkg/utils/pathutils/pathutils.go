package pathutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

type ServiceType string

const (
	TypeGNS3Server       ServiceType = "gns3_server"
	TypeClusterMaster    ServiceType = "cluster_master"
	TypeClusterNode      ServiceType = "cluster_node"
	TypeClusterFileStore ServiceType = "cluster_filestore"
)

type LegacyGNS3Key struct {
	ServerURL   string `json:"server_url"`
	User        string `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type KeyFileV2 struct {
	Version        int               `json:"version"`
	StandaloneGNS3 []GNS3ServerEntry `json:"standalone_gns3,omitempty"`
	Clusters       []ClusterEntry    `json:"clusters,omitempty"`
}

type GNS3ServerEntry struct {
	Name        string `json:"name,omitempty"`
	URL         string `json:"url"`
	User        string `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type ClusterEntry struct {
	Name   string         `json:"name"`
	Master ServiceEntry   `json:"master"`
	Nodes  []ServiceEntry `json:"nodes,omitempty"`

	GNS3Servers []GNS3ServerEntry `json:"gns3_servers,omitempty"`
}

type ServiceEntry struct {
	Type        ServiceType `json:"type"`
	URL         string      `json:"url"`
	User        string      `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
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
	switch {
	case os.IsNotExist(err):
		err = os.MkdirAll(gns3Dir, 0o750)
		if err != nil {
			return "", fmt.Errorf("could not create %q: %w", gns3Dir, err)
		}
	case err != nil:
		return "", fmt.Errorf("could not stat %q: %w", gns3Dir, err)
	case !info.IsDir():
		return "", fmt.Errorf("%q already exists and is not a directory", gns3Dir)
	}

	return gns3Dir, nil
}

func LoadGNS3KeysFile(path string) (*KeyFileV2, error) {
	f, err := os.Open(path) // #nosec G304
	if os.IsNotExist(err) {
		return &KeyFileV2{Version: 2}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not open %q: %w", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("failed to close file: %v", closeErr)
		}
	}()

	buf := make([]byte, 1)
	n, readErr := f.Read(buf)
	if readErr == io.EOF || n == 0 {
		return &KeyFileV2{Version: 2}, nil
	}
	if readErr != nil {
		return nil, fmt.Errorf("failed to read file: %w", readErr)
	}

	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		return nil, fmt.Errorf("failed to seek: %w", seekErr)
	}

	if buf[0] == '{' && isV2Format(f) {
		if _, seekErr := f.Seek(0, 0); seekErr != nil {
			return nil, seekErr
		}
		var kf KeyFileV2
		if decodeErr := json.NewDecoder(f).Decode(&kf); decodeErr != nil {
			return nil, fmt.Errorf("failed to decode V2: %w", decodeErr)
		}
		return &kf, nil
	}

	// If not V2, rewind and parse as legacy
	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		return nil, seekErr
	}

	kf, err := parseLegacyFormat(f)
	if err != nil {
		return nil, err
	}

	_ = SaveKeysFile(path, kf)

	return kf, nil
}

func parseLegacyFormat(r io.Reader) (*KeyFileV2, error) {
	kf := &KeyFileV2{Version: 2}
	dec := json.NewDecoder(r)

	for {
		var old LegacyGNS3Key
		if err := dec.Decode(&old); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode legacy JSON: %w", err)
		}

		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, GNS3ServerEntry{
			URL:         old.ServerURL,
			User:        old.User,
			AccessToken: old.AccessToken,
			TokenType:   old.TokenType,
		})
	}

	return kf, nil
}

func isV2Format(f *os.File) bool {
	var peek struct {
		Version int `json:"version"`
	}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&peek); err != nil {
		return false
	}
	return peek.Version == 2
}

func SaveKeysFile(path string, kf *KeyFileV2) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open key file %q: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(kf); encErr != nil {
		return fmt.Errorf("failed to write key file: %w", encErr)
	}

	return nil
}

func ResolveKeyFilePath(cfgKeyFile string) (string, error) {
	if cfgKeyFile != "" {
		return ExpandPath(cfgKeyFile)
	}
	dir, err := GetGNS3Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gns3key"), nil
}

func (k *KeyFileV2) IsMasterPresent(masterURL string) (*ServiceEntry, bool) {
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Master.URL == masterURL {
			return &c.Master, true
		}
	}
	return nil, false
}

func (k *KeyFileV2) UpdateMaster(token, serverURL, user string) {
	entry := ServiceEntry{
		Type:        TypeClusterMaster,
		AccessToken: token,
		URL:         serverURL,
		User:        user,
	}
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Master.URL == serverURL {
			c.Master = entry
			return
		}
	}
}

func (k *KeyFileV2) CheckIfClusterExists(name string) bool {
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Name == name {
			return true
		}
	}
	return false
}
