package class

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/class"
	"github.com/stefanistkuhl/gns3util/pkg/utils/errorUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/server"
)

var interactive bool

func NewCreateClassCmd() *cobra.Command {
	var createClassCmd = &cobra.Command{
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
			serverUrl, _ := cmd.InheritedFlags().GetString("server")
			cluster, _ := cmd.Flags().GetString("cluster")
			filePath, _ := cmd.Flags().GetString("file")

			if serverUrl != "" && cluster != "" {
				return errorUtils.FormatError("cannot specify both --cluster and --server")
			}
			if serverUrl == "" && cluster == "" {
				return errorUtils.FormatError("either --cluster or --server must be specified")
			}
			if filePath == "" && !interactive {
				return errorUtils.FormatError("either --file or --interactive must be specified")
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

	serverUrl, _ := cmd.InheritedFlags().GetString("server")

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
			fmt.Printf("%v\n", err)
			return err
		}
	} else {
		var err error
		classData, err = class.LoadClassFromFile(filePath)
		if err != nil {
			fmt.Printf("%v\n", err)
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
		cfg.Server = serverUrl
		urlObj := utils.ValidateUrlWithReturn(cfg.Server)
		user, getUserErr := utils.GetUserInKeyFileForUrl(cfg)
		if getUserErr != nil {
			return getUserErr
		}
		clusterName = fmt.Sprintf("%s%s", urlObj.Hostname(), "_single_node_cluster")
		port, convErr := strconv.Atoi(urlObj.Port())
		if convErr != nil {
			err := errorUtils.WrapError(convErr, "failed to convert port to int")
			fmt.Printf("%v\n", err)
			return err
		}
		nodeData.Protocol = urlObj.Scheme
		nodeData.Host = urlObj.Hostname()
		nodeData.Port = port
		nodeData.Weight = 10
		nodeData.MaxGroups = 0
		nodeData.User = user
	}

	conn, err := db.InitIfNeeded()
	if err != nil {
		err = errorUtils.WrapError(err, "failed to initialize database")
		fmt.Printf("%v\n", err)
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	err = db.CheckIfCluterExists(clusterName)
	if err != nil {
		if err == db.ErrClusterExists {
			clusterExists = true
		} else if err != db.ErrNoDb {
			err = errorUtils.WrapError(err, "failed to check if cluster exists")
			fmt.Printf("%v\n", err)
			return err
		}
	}

	var info []db.ClusterName
	var clusterID int

	if !clusterExists && noCluster {
		info, err = db.CreateClusters(conn, []db.ClusterName{{Name: clusterName}})
		if err != nil {
			err = errorUtils.WrapError(err, "failed to create cluster")
			fmt.Printf("%v\n", err)
			return err
		}
		clusterID = info[0].Id
	} else {
		clusters, err := db.GetClusters(conn)
		if err != nil {
			err = errorUtils.WrapError(err, "failed to get existing clusters")
			fmt.Printf("%v\n", err)
			return err
		}
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				clusterID = cluster.Id
				break
			}
		}
		if clusterID == 0 {
			err = errorUtils.WrapError(fmt.Errorf("cluster %s exists but ID not found", clusterName), "cluster ID not found")
			fmt.Printf("%v\n", err)
			return err
		}
	}

	err = db.CheckIfNodeExists(conn, clusterID, nodeData.Host, nodeData.Port)
	if err != nil {
		if err == db.ErrNodeExists {
			nodeExists = true
		} else {
			err = errorUtils.WrapError(err, "failed to check if node exists")
			fmt.Printf("%v\n", err)
			return err
		}
	}
	var insertedNodes []db.NodeDataAll
	if noCluster {
		if !nodeExists {
			if _, err := db.InsertNodes(clusterID, []db.NodeData{nodeData}); err != nil {
				err = errorUtils.WrapError(err, "failed to create node")
				fmt.Printf("%v\n", err)
				return err
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
		updatedCfg, changed, syncErr := cluster.SyncConfigWithDb(currentCfg)
		if syncErr != nil {
			return syncErr
		}
		if changed {
			if err := cluster.WriteClusterConfig(updatedCfg); err != nil {
				return err
			}
		}

		allNodes, getNodesErr := db.GetNodes(conn)
		if getNodesErr != nil {
			getNodesErr = errorUtils.WrapError(getNodesErr, "failed to get nodes for cluster")
			fmt.Printf("%v\n", getNodesErr)
			return getNodesErr
		}
		filteredNodes := []db.NodeDataAll{}
		for _, node := range allNodes {
			if node.ClusterID == clusterID {
				filteredNodes = append(filteredNodes, node)
			}
		}
		insertedNodes = filteredNodes
	} else {
		allNodes, getNodesErr := db.GetNodes(conn)
		if getNodesErr != nil {
			getNodesErr = errorUtils.WrapError(getNodesErr, "failed to get nodes for cluster")
			fmt.Printf("%v\n", getNodesErr)
			return getNodesErr
		}
		filteredNodes := []db.NodeDataAll{}
		for _, node := range allNodes {
			if node.ClusterID == clusterID {
				filteredNodes = append(filteredNodes, node)
			}
		}
		insertedNodes = filteredNodes
	}

	success, err := class.CreateClass(cfg, clusterID, classData, insertedNodes)
	if err != nil {
		err = errorUtils.WrapError(err, "failed to create class")
		fmt.Printf("%v\n", err)
		return err
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
