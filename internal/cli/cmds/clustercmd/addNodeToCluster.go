package clustercmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/colorUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/spf13/cobra"
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

			clusterName := args[0]
			usesID := false
			found := false
			clusterID, err := strconv.Atoi(clusterName)
			if err == nil {
				usesID = true
			}
			store, storeErr := db.Init()
			if storeErr != nil {
				return fmt.Errorf(" failed to open database connection %w", storeErr)
			}
			ctx := context.Background()
			clusters, err := store.GetClusters(ctx)
			if err != nil {
				return fmt.Errorf("failed to get clusters: %w", err)
			}
			for _, c := range clusters {
				if !usesID {
					if c.Name == clusterName {
						opts.ClusterID = int(c.ClusterID)
						found = true
						break
					}
				} else {
					if int(c.ClusterID) == clusterID {
						opts.ClusterID = int(c.ClusterID)
						found = true
						break
					}
				}
			}
			if !found {
				fmt.Printf("%s\n", colorUtils.Error("Cluster not found"))
				return nil
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

			store, err := db.Init()
			if err != nil {
				return fmt.Errorf("failed to init db: %w", err)
			}
			ctx := context.Background()
			tx, err := store.DB.BeginTx(context.Background(), nil)
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer func() {
				if err != nil {
					rollbackErr := tx.Rollback()
					if rollbackErr != nil {
						fmt.Printf("failed to rollback transaction: %v", rollbackErr)
					}
				}
				_ = tx.Commit()
			}()
			qtx := store.WithTx(tx)

			var insertedNodes []sqlc.Node
			for _, node := range nodes {
				var maxGroups sql.NullInt64
				if node.MaxGroups == 0 {
					maxGroups = sql.NullInt64{Int64: 0, Valid: false}
				} else {
					maxGroups = sql.NullInt64{Int64: int64(node.MaxGroups), Valid: true}
				}
				nodeData := sqlc.InsertNodeIntoClusterParams{
					ClusterID: int64(opts.ClusterID),
					Protocol:  node.Protocol,
					Host:      node.Host,
					Port:      int64(node.Port),
					Weight:    int64(node.Weight),
					MaxGroups: maxGroups,
					AuthUser:  node.User,
				}
				nodeDat, insertErr := qtx.InsertNodeIntoCluster(ctx, nodeData)
				if insertErr != nil {
					return fmt.Errorf("failed to insert node: %w", insertErr)
				}
				insertedNodes = append(insertedNodes, nodeDat)
			}
			if err != nil {
				return fmt.Errorf("failed to insert node: %w", err)
			}
			for _, node := range insertedNodes {
				fmt.Printf("Inserted node %s:%d with ID: %d\n", node.Host, node.Port, node.NodeID)
			}
			cfg, cfgErr := cluster.LoadClusterConfig()
			if cfgErr != nil {
				if errors.Is(cfgErr, cluster.ErrNoConfig) {
					cfg = cluster.NewConfig()
				} else {
					return fmt.Errorf("failed to load config: %w", cfgErr)
				}
			}
			cfg, changed, syncErr := cluster.SyncConfigWithDB(cmd.Context(), cfg)
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
			if len(opts.Servers) > 1 {
				return fmt.Errorf("add-node only supports a single --server. Use add-nodes for multiple. ")
			}

			clusterName := args[0]
			usesID := false
			found := false
			clusterID, err := strconv.Atoi(clusterName)
			if err == nil {
				usesID = true
			}
			store, storeErr := db.Init()
			if storeErr != nil {
				return fmt.Errorf(" failed to open database connection %w", storeErr)
			}
			ctx := context.Background()
			clusters, err := store.GetClusters(ctx)
			if err != nil {
				return fmt.Errorf("failed to get clusters: %w", err)
			}
			for _, c := range clusters {
				if !usesID {
					if c.Name == clusterName {
						opts.ClusterID = int(c.ClusterID)
						found = true
						break
					}
				} else {
					if int(c.ClusterID) == clusterID {
						opts.ClusterID = int(c.ClusterID)
						found = true
						break
					}
				}
			}
			if !found {
				fmt.Printf("%s\n", colorUtils.Error("Cluster not found"))
				return nil
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

			store, err := db.Init()
			if err != nil {
				return fmt.Errorf("failed to init db: %w", err)
			}
			ctx := context.Background()
			tx, err := store.DB.BeginTx(context.Background(), nil)
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer func() {
				if err != nil {
					rollbackErr := tx.Rollback()
					if rollbackErr != nil {
						fmt.Printf("failed to rollback transaction: %v", rollbackErr)
					}
				}
				_ = tx.Commit()
			}()
			qtx := store.WithTx(tx)
			var insertedNodes []sqlc.Node
			for _, node := range nodes {
				var maxGroups sql.NullInt64
				if node.MaxGroups == 0 {
					maxGroups = sql.NullInt64{Int64: 0, Valid: false}
				} else {
					maxGroups = sql.NullInt64{Int64: int64(node.MaxGroups), Valid: true}
				}
				nodeData := sqlc.InsertNodeIntoClusterParams{
					ClusterID: int64(opts.ClusterID),
					Protocol:  node.Protocol,
					Host:      node.Host,
					Port:      int64(node.Port),
					Weight:    int64(node.Weight),
					MaxGroups: maxGroups,
					AuthUser:  node.User,
				}
				nodeDat, insertErr := qtx.InsertNodeIntoCluster(ctx, nodeData)
				if insertErr != nil {
					return fmt.Errorf("failed to insert node: %w", insertErr)
				}
				insertedNodes = append(insertedNodes, nodeDat)
			}
			for _, node := range insertedNodes {
				fmt.Printf("Inserted node %s:%d with ID: %d\n", node.Host, node.Port, node.NodeID)
			}

			cfg, cfgErr := cluster.LoadClusterConfig()
			if cfgErr != nil {
				if errors.Is(cfgErr, cluster.ErrNoConfig) {
					cfg = cluster.NewConfig()
				} else {
					return fmt.Errorf("failed to load config: %w", cfgErr)
				}
			}
			cfg, changed, syncErr := cluster.SyncConfigWithDB(cmd.Context(), cfg)
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
