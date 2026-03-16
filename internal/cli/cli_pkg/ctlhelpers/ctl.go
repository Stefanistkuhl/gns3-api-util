package ctlhelpers

import (
	"context"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
)

func DiscoverAndSyncNodes(ctx context.Context, cfg *config.GlobalOptions) error {
	settings := api.NewSettings(
		api.WithBaseURLV2(cfg.ClusterEntry.Master.URL+"/api/v1"),
		api.WithToken(cfg.ClusterEntry.Master.AccessToken),
		api.WithVerify(!cfg.Insecure),
		api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
	)
	client := api.NewClientV2(&settings)

	resp, err := client.GetNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch nodes from master: %w", err)
	}

	keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to resolve key file path: %w", err)
	}
	kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to load keys: %w", err)
	}

	freshNodes := make([]pathutils.ServiceEntry, 0, len(resp.Nodes))
	for _, node := range resp.Nodes {
		freshNodes = append(freshNodes, pathutils.ServiceEntry{
			Type: kf.NodeTypeToSvcType(node.Type),
			URL:  fmt.Sprintf("https://%s:%d", node.IP, node.APIPort),
			ID:   node.ID,
		})
	}

	kf.SyncNodes(freshNodes, cfg.ClusterEntry)
	if err := pathutils.SaveKeysFile(keyFilePath, kf); err != nil {
		return fmt.Errorf("failed to save updated key file: %w", err)
	}

	return nil
}
