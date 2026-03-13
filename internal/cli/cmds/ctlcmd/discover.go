package ctlcmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/mdns"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
)

func NewAddDiscoverCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Subcommand for discovering clusters and their servcices",
		Long:  `Used to discover clusters and their services.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(NewDiscoverClusterCmd())
	cmd.AddCommand(NewDiscoverNodesCmd())
	return cmd
}

func NewDiscoverClusterCmd() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "discover a cluster master using mdns",
		Long:  `Discover a cluster master using mdns. This doenst check for authenticity of the master in anway and just displays avaliable masters on the network.`,
		Annotations: map[string]string{
			"auth-mode": "none",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			peers, err := mdns.BrowseMasters(cmd.Context(), timeout)
			if err != nil {
				return fmt.Errorf("failed to discover masters: %w", err)
			}

			if len(peers) == 0 {
				fmt.Println("No cluster masters discovered")
				return nil
			}

			body, err := json.Marshal(peers)
			if err != nil {
				return fmt.Errorf("failed to marshal discovery results: %w", err)
			}

			utils.PrintOutput(body, cfg)
			return nil
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 5*time.Second,
		"Duration to listen for mDNS responses")

	carapace.Gen(cmd)
	return cmd
}

func NewDiscoverNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "discover cluster nodes in the cluster",
		Long:  `Disscover nodes in the cluster to populate the keyfile only works with --cluster and requires authentication to the cluster master.`,
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			settings := api.NewSettings(
				api.WithBaseURLV2(cfg.ClusterEntry.Master.URL+"/api/v1"),
				api.WithToken(cfg.ClusterEntry.Master.AccessToken),
				api.WithVerify(!cfg.Insecure),
				api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
			)
			client := api.NewClientV2(settings)

			resp, err := client.GetNodes(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get auth status: %w", err)
			}
			body, err := json.Marshal(resp)
			if err != nil {
				return fmt.Errorf("failed to marshal discovery results: %w", err)
			}
			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}
			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}
			nodesToAdd := make([]pathutils.ServiceEntry, 0, len(resp.Nodes))
			for i := range resp.Nodes {
				node := &resp.Nodes[i]
				if !kf.IsNodePresentInCluster(node, cfg.ClusterEntry) {
					svc := pathutils.ServiceEntry{
						Type: kf.NodeTypeToSvcType(node.Type),
						URL:  fmt.Sprintf("https://%s:%d", node.IP, node.APIPort),
						ID:   node.ID,
					}
					nodesToAdd = append(nodesToAdd, svc)
				}
			}
			kf.AddNodes(nodesToAdd, cfg.ClusterEntry)
			saveErr := pathutils.SaveKeysFile(keyFilePath, kf)
			if saveErr != nil {
				return fmt.Errorf("failed to save key file: %w", saveErr)
			}

			utils.PrintOutput(body, cfg)

			return nil
		},
	}
	return cmd
}
