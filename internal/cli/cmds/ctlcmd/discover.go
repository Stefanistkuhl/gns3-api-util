package ctlcmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/ctlhelpers"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/mdns"
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
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return err
			}

			if err := ctlhelpers.DiscoverAndSyncNodes(cmd.Context(), cfg); err != nil {
				return err
			}

			status := struct {
				Cluster string `json:"cluster"`
				Action  string `json:"action"`
				Status  string `json:"status"`
			}{
				Cluster: cfg.Cluster,
				Action:  "sync_nodes",
				Status:  "success",
			}

			body, _ := json.Marshal(status)
			utils.PrintOutput(body, cfg)
			return nil
		},
	}
	return cmd
}
