package keys

import (
	"path/filepath"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

const (
	keyFile = "device_key.pem"
)

func DefaultKeyPath() (string, error) {
	base, err := pathutils.GetGNS3Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, keyFile), nil
}
