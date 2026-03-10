package class

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/class"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/server"
	"github.com/0xveya/gns3util/pkg/api/schemas"
)

var interactive bool

func NewCreateClassCmd() *cobra.Command {
	createClassCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a class with students and groups",
		Long: `Create a class with students and groups. This command can either:
- Create a class from a JSON file
- Launch an interactive web interface for class creation

The class structure includes:
- A main class group
- Student groups within the class
- Students assigned to both the class group and their respective student groups`,
		Example: `
  # Create class from JSON file
  gns3util -s https://controller:3080 class create --file class.json

  # Launch interactive class creation
  gns3util -s https://controller:3080 class create --interactive
		`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.InheritedFlags().GetString("server")
			cluster, _ := cmd.Flags().GetString("cluster")
			filePath, _ := cmd.Flags().GetString("file")

			if serverURL != "" && cluster != "" {
				return fmt.Errorf("cannot specify both --cluster and --server")
			}
			if serverURL == "" && cluster == "" {
				return fmt.Errorf("either --cluster or --server must be specified")
			}
			if filePath == "" && !interactive {
				return fmt.Errorf("either --file or --interactive must be specified")
			}
			return nil
		},
		RunE: runCreateClass,
	}

	createClassCmd.Flags().String("file", "", "JSON file containing class data")
	createClassCmd.Flags().BoolVar(&interactive, "interactive", false, "Launch interactive web interface for class creation")
	createClassCmd.Flags().Int("port", 8080, "Port for interactive web interface")
	createClassCmd.Flags().String("host", "localhost", "Host for interactive web interface")
	createClassCmd.Flags().StringP("cluster", "c", "", "Cluster name")

	return createClassCmd
}

func runCreateClass(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	serverURL, _ := cmd.InheritedFlags().GetString("server")

	cfg, _ := config.GetGlobalOptionsFromContext(cmd.Context())

	filePath, _ := cmd.Flags().GetString("file")
	className, _ := cmd.Flags().GetString("name")
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	clusterName, _ := cmd.Flags().GetString("cluster")

	var classData schemas.Class

	if interactive {
		var err error
		classData, err = server.StartInteractiveServer(host, port)
		if err != nil {
			return err
		}
	} else {
		var err error
		classData, err = class.LoadClassFromFile(filePath)
		if err != nil {
			return err
		}
	}

	if className != "" {
		classData.Name = className
	}

	clusterExists := false
	nodeExists := false
	nodeData := db.NodeData{}
	var noCluster bool
	if cfg.Server != "" && clusterName == "" {
		noCluster = true
	}
	if noCluster {
		cfg.Server = serverURL
		urlObj := utils.ValidateURLWithReturn(cfg.Server)
		user, getUserErr := utils.GetUserInKeyFileForURL(cfg)
		if getUserErr != nil {
			return getUserErr
		}
		clusterName = fmt.Sprintf("%s%s", urlObj.Hostname(), "_single_node_cluster")
		port, convErr := strconv.Atoi(urlObj.Port())
		if convErr != nil {
			return fmt.Errorf("failed to convert port to int: %w", convErr)
		}
		nodeData.Host = urlObj.Hostname()
		nodeData.Port = port
		nodeData.Weight = 10
		nodeData.MaxGroups = 0
		nodeData.User = user
	}

	store, err := db.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	ctx := context.Background()
	tx, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
			}
		}
	}()

	qtx := store.WithTx(tx)

	exists, err := qtx.CheckIfClusterExists(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}
	if exists == 1 {
		clusterExists = true
	}

	var clusterID int

	if !clusterExists && noCluster {
		clusterData, createErr := qtx.CreateCluster(ctx, sqlc.CreateClusterParams{Name: clusterName})
		if createErr != nil {
			return fmt.Errorf("failed to create cluster: %w", createErr)
		}
		clusterID = int(clusterData.ClusterID)
	} else {
		clusters, getErr := qtx.GetClusters(ctx)
		if getErr != nil {
			return fmt.Errorf("failed to get existing clusters: %w", getErr)
		}
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				clusterID = int(cluster.ClusterID)
				break
			}
		}
		if clusterID == 0 {
			return fmt.Errorf("cluster ID not found: cluster %s exists but ID not found", clusterName)
		}
	}

	nodeExistsRes, err := qtx.CheckIfNodeExists(ctx, sqlc.CheckIfNodeExistsParams{ClusterID: int64(clusterID), Host: nodeData.Host, Port: int64(nodeData.Port)})
	if err != nil {
		return fmt.Errorf("failed to check if node exists: %w", err)
	}
	if nodeExistsRes == 1 {
		nodeExists = true
	}
	var insertedNodes []sqlc.Node
	if noCluster {
		if !nodeExists {
			var maxGroups sql.NullInt64
			if nodeData.MaxGroups == 0 {
				maxGroups = sql.NullInt64{Int64: 0, Valid: false}
			} else {
				maxGroups = sql.NullInt64{Int64: int64(nodeData.MaxGroups), Valid: true}
			}
			sqlcNodeData := sqlc.InsertNodeParams{ClusterID: int64(clusterID), Host: nodeData.Host, Port: int64(nodeData.Port), Weight: int64(nodeData.Weight), MaxGroups: maxGroups, AuthUser: nodeData.User}
			if insertErr := qtx.InsertNode(ctx, sqlcNodeData); insertErr != nil {
				return fmt.Errorf("failed to create node: %w", insertErr)
			}
		}

		currentCfg, cfgErr := cluster.LoadClusterConfig()
		if cfgErr != nil {
			if errors.Is(cfgErr, cluster.ErrNoConfig) {
				currentCfg = cluster.NewConfig()
			} else {
				return cfgErr
			}
		}
		updatedCfg, changed, syncErr := cluster.SyncConfigWithDB(cmd.Context(), currentCfg)
		if syncErr != nil {
			return syncErr
		}
		if changed {
			if writeErr := cluster.WriteClusterConfig(updatedCfg); writeErr != nil {
				return writeErr
			}
		}

		allNodes, getNodesErr := qtx.GetNodes(ctx)
		if getNodesErr != nil {
			return fmt.Errorf("failed to get nodes for cluster: %w", getNodesErr)
		}
		filteredNodes := []sqlc.Node{}
		for _, node := range allNodes {
			if int(node.ClusterID) == clusterID {
				filteredNodes = append(filteredNodes, node)
			}
		}
		insertedNodes = filteredNodes
	} else {
		allNodes, getNodesErr := qtx.GetNodes(ctx)
		if getNodesErr != nil {
			return fmt.Errorf("failed to get nodes for cluster: %w", getNodesErr)
		}
		filteredNodes := []sqlc.Node{}
		for _, node := range allNodes {
			if int(node.ClusterID) == clusterID {
				filteredNodes = append(filteredNodes, node)
			}
		}
		insertedNodes = filteredNodes
	}

	var nodes []db.NodeDataAll
	for _, node := range insertedNodes {
		nodes = append(nodes, db.NodeDataAll{ClusterID: int(node.ClusterID), Host: node.Host, Port: int(node.Port), Weight: int(node.Weight), MaxGroups: int(node.MaxGroups.Int64)})
	}
	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}
	success, err := class.CreateClass(cfg, clusterID, classData, nodes)
	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	if success {
		fmt.Printf("%v Created class %v\n",
			messageUtils.SuccessMsg("Created class"),
			messageUtils.Bold(classData.Name))
	} else {
		fmt.Printf("%v Class creation failed\n", messageUtils.ErrorMsg("Class creation failed"))
		return fmt.Errorf("class creation failed")
	}

	return nil
}
