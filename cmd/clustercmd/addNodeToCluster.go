package clustercmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewAddNodeCmd() *cobra.Command {
	opts := &cluster.AddNodeOptions{}
	cmd := &cobra.Command{
		Use:   "add-node [cluster-name]",
		Short: "Add a single node to a cluster",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("cluster name missing. Usage: %s", cmd.UseLine())
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			valErr := cluster.ValidateClusterAndCreds(args[0], opts, cmd)
			if valErr != nil {
				return valErr
			}
			if len(opts.Servers) == 0 {
				fmt.Println(messageUtils.InfoMsg("No servers provided, will enter interactive mode."))
				return nil
			}
			if len(opts.Servers) > 1 {
				return fmt.Errorf("add-node only supports a single --server. Use add-nodes for multiple. ")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, err := cluster.RunAddNodes(opts, cmd)
			if err != nil {
				return fmt.Errorf("failed to add node: %w", err)
			}
			if nodes == nil {
				return fmt.Errorf("no nodes added")
			}

			insertedNodes, err := db.InsertNodes(opts.ClusterID, nodes)
			if err != nil {
				return fmt.Errorf("failed to insert node: %w", err)
			}
			for _, node := range insertedNodes {
				fmt.Printf("Inserted node %s:%d with ID: %d\n", node.Host, node.Port, node.ID)
			}
			cfg, cfgErr := cluster.LoadClusterConfig()
			if cfgErr != nil {
				if cfgErr == cluster.ErrNoConfig {
					cfg = cluster.NewConfig()
				} else {
					return fmt.Errorf("failed to load config: %w", cfgErr)
				}
			}
			cfg, changed, syncErr := cluster.SyncConfigWithDb(cfg)
			if syncErr != nil {
				return fmt.Errorf("failed to sync config with db: %w", syncErr)
			}
			if changed {
				if err := cluster.WriteClusterConfig(cfg); err != nil {
					return fmt.Errorf("failed to write synced config: %w", err)
				}
			}
			return nil
		},
	}
	addCommonFlags(cmd, opts)
	return cmd
}

func NewAddNodesCmd() *cobra.Command {
	opts := &cluster.AddNodeOptions{}
	cmd := &cobra.Command{
		Use:   "add-nodes [cluster-name]",
		Short: "Add multiple nodes to a cluster",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("cluster name missing. Usage: %s", cmd.UseLine())
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			valErr := cluster.ValidateClusterAndCreds(args[0], opts, cmd)
			if valErr != nil {
				return valErr
			}
			if len(opts.Servers) == 0 {
				fmt.Println(messageUtils.InfoMsg("No servers provided, will enter interactive mode."))
				return nil
			}
			if len(opts.Servers) == 1 {
				return fmt.Errorf("add-nodes requires at least 2 --server entries. Use add-node for one. ")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, err := cluster.RunAddNodes(opts, cmd)
			if err != nil {
				return fmt.Errorf("failed to add nodes: %w", err)
			}
			if nodes == nil {
				return fmt.Errorf("no nodes added")
			}

			insertedNodes, err := db.InsertNodes(opts.ClusterID, nodes)
			if err != nil {
				return fmt.Errorf("DB insert error: %w", err)
			}
			for _, node := range insertedNodes {
				fmt.Printf("Inserted node %s:%d with ID: %d\n", node.Host, node.Port, node.ID)
			}

			cfg, cfgErr := cluster.LoadClusterConfig()
			if cfgErr != nil {
				if cfgErr == cluster.ErrNoConfig {
					cfg = cluster.NewConfig()
				} else {
					return fmt.Errorf("failed to load config: %w", cfgErr)
				}
			}
			cfg, changed, syncErr := cluster.SyncConfigWithDb(cfg)
			if syncErr != nil {
				return fmt.Errorf("failed to sync config with db: %w", syncErr)
			}
			if changed {
				if err := cluster.WriteClusterConfig(cfg); err != nil {
					return fmt.Errorf("failed to write synced config: %w", err)
				}
			}
			return nil
		},
	}
	addCommonFlags(cmd, opts)
	return cmd
}

func addCommonFlags(cmd *cobra.Command, opts *cluster.AddNodeOptions) {
	cmd.Flags().StringSliceVarP(&opts.Servers, "server", "s", nil, "Server(s) to add")
	cmd.Flags().IntVarP(&opts.Weight, "weight", "w", 5, "Weight to assign to node(s) (0–10, default 5)")
	cmd.Flags().IntVarP(&opts.MaxGroups, "max-groups", "g", 3, "Maximum groups per node (default 3)")
	cmd.Flags().StringVarP(&opts.Username, "user", "u", "", "User to log in as (env: GNS3_USER)")
	cmd.Flags().StringVarP(&opts.Password, "password", "p", "", "Password to use (env: GNS3_PASSWORD)")
}
