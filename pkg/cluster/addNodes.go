package cluster

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

type AddNodeOptions struct {
	Servers   []string
	Weight    int
	MaxGroups int
	Username  string
	Password  string
	ClusterID int
}

func RunAddNode(server string, opts *AddNodeOptions, cmd *cobra.Command) (db.NodeData, error) {
	cfg, _ := config.GetGlobalOptionsFromContext(cmd.Context())
	u := utils.ValidateUrlWithReturn(server)
	if u == nil {
		return db.NodeData{}, fmt.Errorf("invalid server URL: %s", server)
	}

	cfg.Server = server
	_, status, reqErr := utils.CallClient(cfg, "getMe", nil, nil)
	if reqErr != nil || status != 200 {
		return db.NodeData{}, fmt.Errorf("failed to query node %s: %v", server, reqErr)
	}
	port, toIntErr := strconv.Atoi(u.Port())
	if toIntErr != nil {
		return db.NodeData{}, fmt.Errorf("failed to convert port to an int")
	}

	return db.NodeData{
		Protocol:  u.Scheme,
		Host:      u.Hostname(),
		Port:      port,
		Weight:    opts.Weight,
		MaxGroups: opts.MaxGroups,
		User:      opts.Username,
	}, nil
}

func RunAddNodes(opts *AddNodeOptions, cmd *cobra.Command) ([]db.NodeData, error) {
	if len(opts.Servers) == 0 {
		fmt.Printf("%s\n", colorUtils.Info("No servers provided, entering interactive mode..."))

		cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
		if err != nil {
		}

		// Load servers from keyfile
		keys, err := authentication.LoadKeys(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load keyfile: %w", err)
		}

		if len(keys) == 0 {
			return nil, fmt.Errorf("no servers found in keyfile. Please use 'auth login' to add servers")
		}

		// Create fuzzy picker options
		serverOptions := make([]string, len(keys))
		serverMap := make(map[string]string)

		for i, key := range keys {
			plainName := fmt.Sprintf("%-30s (%s)", key.ServerURL, key.User)
			serverOptions[i] = plainName
			serverMap[plainName] = key.ServerURL
		}

		selectedServers := fuzzy.NewFuzzyFinderWithTitle(serverOptions, true, "Select servers to add to cluster:")

		if len(selectedServers) == 0 {
			fmt.Printf("%s\n", colorUtils.Warning("No servers selected"))
			return nil, nil
		}

		// Convert selected display names to server URLs
		for _, displayName := range selectedServers {
			if serverURL, ok := serverMap[displayName]; ok {
				opts.Servers = append(opts.Servers, serverURL)
			}
		}

		fmt.Printf("\n%s\n", colorUtils.Info("Selected servers:"))
		for _, server := range opts.Servers {
			fmt.Printf("  %s %s\n", colorUtils.Seperator("•"), colorUtils.Highlight(server))
		}
	}

	if len(opts.Servers) == 1 {
		n, err := RunAddNode(opts.Servers[0], opts, cmd)
		if err != nil {
			return nil, err
		}
		return []db.NodeData{n}, nil
	}

	var nodes []db.NodeData
	var errs []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, server := range opts.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			n, err := RunAddNode(s, opts, cmd)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			nodes = append(nodes, n)
		}(server)
	}
	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("some nodes failed: %v", errs)
	}

	return nodes, nil
}

func ValidateClusterAndCreds(clusterName string, opts *AddNodeOptions, cmd *cobra.Command) {
	viper.SetEnvPrefix("GNS3")
	viper.AutomaticEnv()

	_ = viper.BindPFlag("user", cmd.Flags().Lookup("user"))
	_ = viper.BindPFlag("password", cmd.Flags().Lookup("password"))

	dbConn, err := db.InitIfNeeded()
	if err != nil {
		fmt.Printf("%s failed to init db: %v\n", messageUtils.ErrorMsg("Error"), err)
		os.Exit(1)
	}
	defer dbConn.Close()

	clusters, fetchErr := db.QueryRows(dbConn,
		"SELECT cluster_id, name, description FROM clusters WHERE name = ? LIMIT 1",
		func(rows *sql.Rows) (db.ClusterName, error) {
			var c db.ClusterName
			err := rows.Scan(&c.Id, &c.Name, &c.Desc)
			return c, err
		},
		clusterName,
	)
	if fetchErr != nil {
		if fetchErr == sql.ErrNoRows || len(clusters) == 0 {
			fmt.Printf("%s cluster %s not found\n", messageUtils.ErrorMsg("Error"), clusterName)
			os.Exit(1)
		}
		fmt.Printf("%s failed to query cluster: %v\n", messageUtils.ErrorMsg("Error"), fetchErr)
		os.Exit(1)
	}

	opts.ClusterID = clusters[0].Id

	if !cmd.Flags().Changed("user") {
		opts.Username = viper.GetString("user")
	}
	if !cmd.Flags().Changed("password") {
		opts.Password = viper.GetString("password")
	}

	if opts.Weight < 0 || opts.Weight > 10 {
		fmt.Printf("%s --weight must be between 0 and 10\n", messageUtils.ErrorMsg("Error"))
		os.Exit(1)
	}
	if opts.MaxGroups < 0 {
		fmt.Printf("%s --max-groups must be > 0\n", messageUtils.ErrorMsg("Error"))
		os.Exit(1)
	}
}
