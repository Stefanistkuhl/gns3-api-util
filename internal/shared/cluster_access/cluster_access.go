package clusteraccess

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type ClientConfig struct {
	ServerURL string `toml:"server_url"`
	CACert    string `toml:"ca_cert"`
	UserCert  string `toml:"user_cert"`
	UserKey   string `toml:"user_key"`
}

func CreateClusterAcessConfig(url string, caCert, userCert, userKey []byte) *ClientConfig {
	return &ClientConfig{
		ServerURL: url,
		CACert:    string(caCert),
		UserCert:  string(userCert),
		UserKey:   string(userKey),
	}
}

func (c *ClientConfig) WriteAccessConfig(path string) error {
	dat, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, dat, 0o600)
}

func (c *ClientConfig) CheckIfAccessConfigExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
