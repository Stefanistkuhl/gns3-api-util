package cluster

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/colorUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

type AddNodeOptions struct {
	Servers     []string
	ServerUsers map[string]string
	Weight      int
	MaxGroups   int
	Username    string
	Password    string
	ClusterID   int
}

func RunAddNode(server string, opts *AddNodeOptions, cmd *cobra.Command) (db.NodeData, error) {
	cfg, _ := config.GetGlobalOptionsFromContext(cmd.Context())
	u := utils.ValidateURLWithReturn(server)
	if u == nil {
		return db.NodeData{}, fmt.Errorf("invalid server URL: %s", server)
	}

	cfg.Server = server
	_, status, reqErr := utils.CallClient(cfg, "getMe", nil, nil)
	if reqErr != nil || status != 200 {
		return db.NodeData{}, fmt.Errorf("failed to query node %s: %w", server, reqErr)
	}
	port, toIntErr := strconv.Atoi(u.Port())
	if toIntErr != nil {
		return db.NodeData{}, fmt.Errorf("failed to convert port to an int")
	}

	username := strings.TrimSpace(opts.Username)
	if username == "" && opts.ServerUsers != nil {
		if mappedUser, ok := opts.ServerUsers[server]; ok {
			username = strings.TrimSpace(mappedUser)
		}
	}

	return db.NodeData{
		Protocol:  u.Scheme,
		Host:      u.Hostname(),
		Port:      port,
		Weight:    opts.Weight,
		MaxGroups: opts.MaxGroups,
		User:      username,
	}, nil
}

func RunAddNodes(opts *AddNodeOptions, cmd *cobra.Command) ([]db.NodeData, error) {
	if len(opts.Servers) == 0 {
		fmt.Printf("%s\n", colorUtils.Info("No servers provided, entering interactive mode..."))

		cfg, _ := config.GetGlobalOptionsFromContext(cmd.Context())

		keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
		if err != nil {
			return nil, err
		}

		kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load keyfile: %w", err)
		}

		if len(kf.StandaloneGNS3) == 0 {
			return nil, fmt.Errorf("no servers found in keyfile. Please use 'auth login' to add servers")
		}

		if len(kf.StandaloneGNS3) == 0 {
			return nil, fmt.Errorf("no servers found in keyfile. Please use 'auth login' to add servers")
		}

		serverOptions := make([]string, len(kf.StandaloneGNS3))
		serverMap := make(map[string]string)
		serverUserMap := make(map[string]string)

		for i, entry := range kf.StandaloneGNS3 {
			plainName := fmt.Sprintf("%-30s (%s)", entry.URL, entry.User)
			serverOptions[i] = plainName
			serverMap[plainName] = entry.URL
			serverUserMap[entry.URL] = entry.User
		}

		selectedServers := fuzzy.NewFuzzyFinderWithTitle(serverOptions, true, "Select servers to add to cluster:")

		if len(selectedServers) == 0 {
			fmt.Printf("%s\n", colorUtils.Warning("No servers selected"))
			return nil, nil
		}

		if opts.ServerUsers == nil {
			opts.ServerUsers = make(map[string]string)
		}
		for _, displayName := range selectedServers {
			if serverURL, ok := serverMap[displayName]; ok {
				opts.Servers = append(opts.Servers, serverURL)
				if user, ok := serverUserMap[serverURL]; ok && strings.TrimSpace(user) != "" {
					opts.ServerUsers[serverURL] = user
				}
			}
		}

		if strings.TrimSpace(opts.Username) == "" && len(opts.ServerUsers) == 1 {
			for _, user := range opts.ServerUsers {
				opts.Username = user
				break
			}
		}

		fmt.Printf("\n%s\n", colorUtils.Info("Selected servers:"))
		for _, server := range opts.Servers {
			fmt.Printf("  %s %s\n", colorUtils.Separator("•"), colorUtils.Highlight(server))
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

func ValidateClusterAndCreds(clusterName string, opts *AddNodeOptions, cmd *cobra.Command) error {
	_ = viper.BindPFlag("user", cmd.Flags().Lookup("user"))
	_ = viper.BindPFlag("password", cmd.Flags().Lookup("password"))

	store, err := db.Init()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}

	cluster, fetchErr := store.CheckClusterExistsWithData(cmd.Context(), clusterName)
	if fetchErr != nil {
		if errors.Is(fetchErr, sql.ErrNoRows) {
			return fmt.Errorf("cluster %s not found", clusterName)
		}
		return fmt.Errorf("failed to query cluster: %w", fetchErr)
	}

	opts.ClusterID = int(cluster.ClusterID)

	if !cmd.Flags().Changed("user") {
		opts.Username = viper.GetString("user")
	}
	if !cmd.Flags().Changed("password") {
		opts.Password = viper.GetString("password")
	}

	if opts.Weight < 0 || opts.Weight > 10 {
		return fmt.Errorf("--weight must be between 0 and 10")
	}
	if opts.MaxGroups < 0 {
		return fmt.Errorf("--max-groups must be > 0")
	}
	return nil
}
