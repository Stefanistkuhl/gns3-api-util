package jobscmd

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/pkg/api"
)

func newMasterClient(cfg *config.GlobalOptions) (*api.ClientV2, error) {
	if cfg.ClusterEntry == nil {
		return nil, fmt.Errorf("no cluster entry found, use --cluster to specify a cluster")
	}

	serverURL := cfg.ClusterEntry.Master.URL
	token := cfg.ClusterEntry.Master.AccessToken

	if serverURL == "" {
		return nil, fmt.Errorf("master URL not configured for cluster")
	}
	if token == "" {
		return nil, fmt.Errorf("no access token for master (run 'ctl auth status' first)")
	}

	settings := api.NewSettings(
		api.WithBaseURLV2(serverURL+"/api/v1"),
		api.WithToken(token),
		api.WithVerify(!cfg.Insecure),
		api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
	)

	return api.NewClientV2(&settings), nil
}
