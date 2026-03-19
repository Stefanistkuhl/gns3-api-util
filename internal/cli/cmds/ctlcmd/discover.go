package ctlcmd

import (
	"fmt"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/ctlhelpers"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/mdns"
)

type DiscoveredMasterRow struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

func (d DiscoveredMasterRow) GetHeaders() []string {
	return []string{"HOSTNAME", "ADDRESS", "PORT"}
}

func (d DiscoveredMasterRow) GetRow() []string {
	return []string{
		d.Name,
		d.Address,
		fmt.Sprintf("%d", d.Port),
	}
}

type NodeSyncResult struct {
	Cluster string `json:"cluster"`
	Action  string `json:"action"`
	Status  string `json:"status"`
}

func (n NodeSyncResult) GetHeaders() []string {
	return []string{"CLUSTER", "ACTION", "STATUS"}
}

func (n NodeSyncResult) GetRow() []string {
	return []string{
		n.Cluster,
		n.Action,
		n.Status,
	}
}

func NewAddDiscoverCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Subcommand for discovering clusters and their services",
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
		Long:  `Discover a cluster master using mdns. This doesn't check for authenticity of the master in any way and just displays available masters on the network.`,
		Annotations: map[string]string{
			"auth-mode": "none",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Listening for mDNS responses for %v...\n", timeout)

			peers, err := mdns.BrowseMasters(cmd.Context(), timeout)
			if err != nil {
				return fmt.Errorf("failed to discover masters: %w", err)
			}

			if len(peers) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No cluster masters discovered")
				return nil
			}

			var records []utils.TableRecord
			for _, p := range peers {
				records = append(records, DiscoveredMasterRow{
					Name:    p.Instance,
					Address: p.Address,
					Port:    p.Port,
				})
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(records, cmd.OutOrStdout())
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

			fmt.Fprintf(cmd.ErrOrStderr(), "Syncing nodes for cluster %q...\n", cfg.Cluster)

			if syncErr := ctlhelpers.DiscoverAndSyncNodes(cmd.Context(), cfg); syncErr != nil {
				return syncErr
			}

			result := NodeSyncResult{
				Cluster: cfg.Cluster,
				Action:  "sync_nodes",
				Status:  "success",
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(result, cmd.OutOrStdout())
		},
	}
	return cmd
}
