package cluster

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

var ErrNoConfig = errors.New("no config for clusters")

func LoadClusterConfig() (Config, error) {
	var c Config
	dir, getDirErr := utils.GetGNS3Dir()
	if getDirErr != nil {
		return c, getDirErr
	}
	path := filepath.Join(dir, "cluster_config.toml")
	f, readErr := os.ReadFile(path)
	if readErr != nil {
		return c, ErrNoConfig
	}
	unmarshallErr := toml.Unmarshal(f, &c)
	if unmarshallErr != nil {
		return c, unmarshallErr
	}
	return c, nil
}

func EnsureConfigSyncedFromDB() (Config, bool, error) {
	cfg, err := LoadClusterConfig()
	if err != nil {
		if errors.Is(err, ErrNoConfig) {
			conn, openErr := db.InitIfNeeded()
			if openErr != nil {
				return Config{}, false, fmt.Errorf("db open error: %w", openErr)
			}
			defer conn.Close()

			dbClusters, ferr := db.GetClusters(conn)
			if ferr != nil && ferr != sql.ErrNoRows {
				return Config{}, false, fmt.Errorf("error fetching clusters: %w", ferr)
			}
			dbNodes, nerr := db.GetNodes(conn)
			if nerr != nil && nerr != sql.ErrNoRows {
				return Config{}, false, fmt.Errorf("error fetching nodes: %w", nerr)
			}

			base := NewConfig()
			bootstrapped, _ := MergeConfigWithDb(base, dbClusters, dbNodes)

			if err := WriteClusterConfig(bootstrapped); err != nil {
				return Config{}, false, fmt.Errorf("failed to write new config: %w", err)
			}
			return bootstrapped, true, nil
		}
		return Config{}, false, err
	}

	inSync, cerr := CheckConfigWithDb(cfg, false)
	if cerr != nil {
		return cfg, false, cerr
	}
	if inSync {
		return cfg, false, nil
	}

	cfgNew, changed, serr := SyncConfigWithDb(cfg)
	if serr != nil {
		return cfg, false, serr
	}
	if changed {
		if err := WriteClusterConfig(cfgNew); err != nil {
			return cfg, false, fmt.Errorf("failed to write synced config: %w", err)
		}
	}
	return cfgNew, changed, nil
}

func WriteClusterConfig(c Config) error {
	for i := range c.Clusters {
		if len(c.Clusters[i].Nodes) == 0 {
			c.Clusters[i].Nodes = nil
		}
	}
	dir, getDirErr := utils.GetGNS3Dir()
	if getDirErr != nil {
		return getDirErr
	}
	path := filepath.Join(dir, "cluster_config.toml")
	res, marshallErr := toml.Marshal(&c)
	if marshallErr != nil {
		return marshallErr
	}
	if err := os.WriteFile(path, res, 0o644); err != nil {
		return err
	}
	return nil
}
