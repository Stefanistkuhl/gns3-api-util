package cluster

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathUtils"
	"github.com/pelletier/go-toml/v2"
)

var ErrNoConfig = errors.New("no config for clusters")

func LoadClusterConfig() (Config, error) {
	var c Config
	dir, getDirErr := pathUtils.GetGNS3Dir()
	if getDirErr != nil {
		return c, getDirErr
	}
	path := filepath.Join(dir, "cluster_config.toml")
	f, readErr := os.ReadFile(path) // #nosec G304
	if readErr != nil {
		return c, ErrNoConfig
	}
	unmarshallErr := toml.Unmarshal(f, &c)
	if unmarshallErr != nil {
		return c, unmarshallErr
	}
	return c, nil
}

func EnsureConfigSyncedFromDB(ctx context.Context) (Config, bool, error) {
	cfg, err := LoadClusterConfig()
	if err != nil {
		if errors.Is(err, ErrNoConfig) {
			store, openErr := db.Init()
			if openErr != nil {
				return Config{}, false, fmt.Errorf("db open error: %w", openErr)
			}
			defer store.DB.Close()

			dbClusters, ferr := store.GetClusters(ctx)
			if ferr != nil && !errors.Is(ferr, sql.ErrNoRows) {
				return Config{}, false, fmt.Errorf("error fetching clusters: %w", ferr)
			}
			dbNodes, nerr := store.GetNodes(ctx)
			if nerr != nil && !errors.Is(nerr, sql.ErrNoRows) {
				return Config{}, false, fmt.Errorf("error fetching nodes: %w", nerr)
			}

			base := NewConfig()
			bootstrapped, _, mergeErr := mergeConfigWithDb(base, dbClusters, dbNodes)
			if mergeErr != nil {
				return Config{}, false, fmt.Errorf("merge config: %w", mergeErr)
			}

			if writeErr := WriteClusterConfig(bootstrapped); writeErr != nil {
				return Config{}, false, fmt.Errorf("failed to write new config: %w", writeErr)
			}
			return bootstrapped, true, nil
		}
		return Config{}, false, err
	}

	store, openErr := db.Init()
	if openErr != nil {
		return cfg, false, fmt.Errorf("db open error: %w", openErr)
	}
	defer store.DB.Close()

	inSync, cerr := CheckConfigWithDb(ctx, store, cfg, false)
	if cerr != nil {
		return cfg, false, cerr
	}
	if inSync {
		return cfg, false, nil
	}

	cfgNew, changed, serr := SyncConfigWithDb(ctx, cfg)
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
	dir, getDirErr := pathUtils.GetGNS3Dir()
	if getDirErr != nil {
		return getDirErr
	}
	path := filepath.Join(dir, "cluster_config.toml")
	res, marshallErr := toml.Marshal(&c)
	if marshallErr != nil {
		return marshallErr
	}
	return os.WriteFile(path, res, 0o600)
}
