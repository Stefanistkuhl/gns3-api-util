package class

import (
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

  # Create class with specific name
  gns3util -s https://controller:3080 class create --file class.json --name "CS101"
		`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("cluster") && cmd.Flags().Changed("server") {
				return fmt.Errorf("cannot specify both --cluster and --server")
			}
			server, _ := cmd.InheritedFlags().GetString("server")
			cluster, _ := cmd.Flags().GetString("cluster")
			if server == "" && cluster != "" {
				return nil
			}
			return nil
		},
		RunE: runCreateClass,
	}

	createClassCmd.Flags().String("file", "", "JSON file containing class data")
	createClassCmd.Flags().Bool("interactive", false, "Launch interactive web interface for class creation")
	createClassCmd.Flags().String("name", "", "Override class name from file")
	createClassCmd.Flags().Int("port", 8080, "Port for interactive web interface")
	createClassCmd.Flags().String("host", "localhost", "Host for interactive web interface")
	createClassCmd.Flags().StringP("cluster", "c", "", "Cluster name")

	return createClassCmd
}

func runCreateClass(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
	}

	filePath, _ := cmd.Flags().GetString("file")
	interactive, _ := cmd.Flags().GetBool("interactive")
	className, _ := cmd.Flags().GetString("name")
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	clusterName, _ := cmd.Flags().GetString("cluster")

	if filePath == "" && !interactive {
		err := errorUtils.FormatError("either --file or --interactive must be specified")
		fmt.Printf("%v\n", err)
		return err
	}

	if filePath != "" && interactive {
		err := errorUtils.FormatError("cannot specify both --file and --interactive")
		fmt.Printf("%v\n", err)
		return err
	}

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
	noCluster := false
	if cfg.Server != "" && clusterName == "" {
		noCluster = true
	}
	if noCluster {
		urlObj := utils.ValidateUrlWithReturn(cfg.Server)
		user, getUserErr := utils.GetUserInKeyFileForUrl(cfg)
		if getUserErr != nil {
			return err
		}
		clusterName = fmt.Sprintf("%s%s", urlObj.Hostname(), "_single_node_cluster")
		port, convErr := strconv.Atoi(urlObj.Port())
		if convErr != nil {
			err = errorUtils.WrapError(convErr, "failed to convert port to int")
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
	defer conn.Close()

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
	insertedNodes := []db.NodeDataAll{}
	if noCluster {
		if !nodeExists {
			insertedNodes, err = db.InsertNodes(clusterID, []db.NodeData{nodeData})
			if err != nil {
				err = errorUtils.WrapError(err, "failed to create node")
				fmt.Printf("%v\n", err)
				return err
			}
			config, getConfigErr := cluster.LoadClusterConfig()
			if getConfigErr != nil {
				return getConfigErr
			}
			configNew, _, syncConfiErr := cluster.SyncConfigWithDb(config)
			if syncConfiErr != nil {
				return syncConfiErr
			}
			writeCfgErr := cluster.WriteClusterConfig(configNew)
			if writeCfgErr != nil {
				return writeCfgErr
			}
		}
	} else {
		insertedNodes, err = db.GetNodes(conn)
		if err != nil {
			err = errorUtils.WrapError(err, "failed to get nodes for cluster")
			fmt.Printf("%v\n", err)
			return err
		}
		filteredNodes := []db.NodeDataAll{}
		for _, node := range insertedNodes {
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
