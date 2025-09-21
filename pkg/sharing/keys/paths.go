package keys

import (
	"path/filepath"

	"github.com/stefanistkuhl/gns3util/pkg/utils/pathUtils"
)

const (
	appDirName = "gns3util"
	keyFile    = "device_key.pem"
)

func DefaultKeyPath() (string, error) {
	base, err := pathUtils.GetGNS3Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, keyFile), nil
}
