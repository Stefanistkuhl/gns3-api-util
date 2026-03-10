package cluster

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
)

func ApplyConfig(cfg Config) error {
	store, err := db.Init()
	if err != nil {
		return fmt.Errorf("db init: %w", err)
	}
	defer store.DB.Close()

	ctx := context.Background()

	for _, cluster := range cfg.Clusters {
		for _, node := range cluster.Nodes {
			url := fmt.Sprintf("%s://%s:%d", node.Protocol, node.Host, node.Port)
			if !utils.ValidateAndTestURL(ctx, url) {
				return fmt.Errorf("cannot connect to: %s", url)
			}
		}
	}

	tx, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Log the rollback error but don't override the original error
			fmt.Printf("Warning: failed to rollback transaction: %v\n", rollbackErr)
		}
	}()

	qtx := store.WithTx(tx)

	dbClusters, err := qtx.GetClusters(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get clusters: %w", err)
	}

	dbNodes, err := qtx.GetNodes(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get nodes: %w", err)
	}

	clusterByName := make(map[string]sqlc.Cluster)
	for _, c := range dbClusters {
		clusterByName[norm(c.Name)] = c
	}

	nodesByCluster := make(map[int64][]sqlc.Node)
	for _, n := range dbNodes {
		nodesByCluster[n.ClusterID] = append(nodesByCluster[n.ClusterID], n)
	}

	configClusterNames := make(map[string]bool)

	for _, cfgCluster := range cfg.Clusters {
		nname := norm(cfgCluster.Name)
		configClusterNames[nname] = true

		dbCluster, exists := clusterByName[nname]

		if !exists {
			created, err := qtx.CreateCluster(ctx, sqlc.CreateClusterParams{
				Name:        cfgCluster.Name,
				Description: toNullString(cfgCluster.Description),
			})
			if err != nil {
				return fmt.Errorf("create cluster %s: %w", cfgCluster.Name, err)
			}
			dbCluster = created
		} else {
			dbDesc := strings.TrimSpace(nullToString(dbCluster.Description))
			cfgDesc := strings.TrimSpace(cfgCluster.Description)
			if dbDesc != cfgDesc {
				err := qtx.UpdateClusterDescription(ctx, sqlc.UpdateClusterDescriptionParams{
					Description: toNullString(cfgDesc),
					ClusterID:   dbCluster.ClusterID,
				})
				if err != nil {
					return fmt.Errorf("update cluster %s desc: %w", cfgCluster.Name, err)
				}
			}
		}

		if err := syncNodes(ctx, qtx, cfg, cfgCluster, dbCluster.ClusterID, nodesByCluster[dbCluster.ClusterID]); err != nil {
			return err
		}
	}

	for _, dbCluster := range dbClusters {
		if !configClusterNames[norm(dbCluster.Name)] {
			nodes := nodesByCluster[dbCluster.ClusterID]
			for _, n := range nodes {
				if err := qtx.DeleteNode(ctx, n.NodeID); err != nil {
					return fmt.Errorf("delete node %d: %w", n.NodeID, err)
				}
			}
			if err := qtx.DeleteCluster(ctx, dbCluster.ClusterID); err != nil {
				return fmt.Errorf("delete cluster %s: %w", dbCluster.Name, err)
			}
		}
	}

	return tx.Commit()
}

func syncNodes(ctx context.Context, qtx *sqlc.Queries, cfg Config, cfgCluster Cluster, clusterID int64, dbNodes []sqlc.Node) error {
	dbNodeByKey := make(map[string]sqlc.Node)
	for _, n := range dbNodes {
		dbNodeByKey[nodeKey(n.Host, int(n.Port))] = n
	}

	configNodeKeys := make(map[string]bool)

	for _, cfgNode := range cfgCluster.Nodes {
		key := nodeKey(cfgNode.Host, cfgNode.Port)
		configNodeKeys[key] = true

		proto := defaultStr(cfgNode.Protocol, cfg.Settings.DefaultProtocol, "http")
		maxGroups := defaultInt(cfgNode.MaxGroups, cfg.Settings.DefaultMaxGroups)

		dbNode, exists := dbNodeByKey[key]

		if !exists {
			_, err := qtx.InsertNodeIntoCluster(ctx, sqlc.InsertNodeIntoClusterParams{
				ClusterID: clusterID,
				Protocol:  proto,
				Host:      cfgNode.Host,
				Port:      int64(cfgNode.Port),
				Weight:    int64(cfgNode.Weight),
				MaxGroups: toNullInt64(maxGroups),
				AuthUser:  cfgNode.User,
			})
			if err != nil {
				return fmt.Errorf("insert node %s: %w", key, err)
			}
		} else {
			needsUpdate := !equalStr(dbNode.Protocol, proto) ||
				dbNode.Weight != int64(cfgNode.Weight) ||
				nullToInt(dbNode.MaxGroups) != maxGroups ||
				!equalStr(dbNode.AuthUser, cfgNode.User)

			if needsUpdate {
				err := qtx.UpdateNode(ctx, sqlc.UpdateNodeParams{
					Protocol:  proto,
					AuthUser:  cfgNode.User,
					Weight:    int64(cfgNode.Weight),
					MaxGroups: toNullInt64(maxGroups),
					ClusterID: clusterID,
					Host:      dbNode.Host,
					Port:      dbNode.Port,
				})
				if err != nil {
					return fmt.Errorf("update node %s: %w", key, err)
				}
			}
		}
	}

	for _, dbNode := range dbNodes {
		key := nodeKey(dbNode.Host, int(dbNode.Port))
		if !configNodeKeys[key] {
			if err := qtx.DeleteNode(ctx, dbNode.NodeID); err != nil {
				return fmt.Errorf("delete node %s: %w", key, err)
			}
		}
	}

	return nil
}

func SyncConfigWithDB(ctx context.Context, cfg Config) (Config, bool, error) {
	store, err := db.Init()
	if err != nil {
		return cfg, false, fmt.Errorf("db init: %w", err)
	}
	defer store.DB.Close()

	var dbClusters []sqlc.Cluster
	var dbNodes []sqlc.Node

	err = store.ReadOnly(ctx, func(q *sqlc.Queries) error {
		var txErr error
		dbClusters, txErr = q.GetClusters(ctx)
		if txErr != nil && !errors.Is(txErr, sql.ErrNoRows) {
			return txErr
		}
		dbNodes, txErr = q.GetNodes(ctx)
		if txErr != nil && !errors.Is(txErr, sql.ErrNoRows) {
			return txErr
		}
		return nil
	})
	if err != nil {
		return cfg, false, fmt.Errorf("read db: %w", err)
	}

	return mergeConfigWithDB(cfg, dbClusters, dbNodes)
}

func CheckConfigWithDB(ctx context.Context, store *db.Store, cfg Config, verbose bool) (bool, error) {
	var dbClusters []sqlc.Cluster
	var dbNodes []sqlc.Node

	err := store.ReadOnly(ctx, func(q *sqlc.Queries) error {
		var txErr error
		dbClusters, txErr = q.GetClusters(ctx)
		if txErr != nil && !errors.Is(txErr, sql.ErrNoRows) {
			return txErr
		}
		dbNodes, txErr = q.GetNodes(ctx)
		if txErr != nil && !errors.Is(txErr, sql.ErrNoRows) {
			return txErr
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("read db: %w", err)
	}

	return compareConfig(cfg, dbClusters, dbNodes, verbose), nil
}

func PurgeConfig(cfg Config, ctx context.Context) error {
	store, err := db.Init()
	if err != nil {
		return err
	}
	nukeErr := store.NukeEverything(ctx)
	if nukeErr != nil {
		return fmt.Errorf("nuke: %w", nukeErr)
	}

	return nil
}

func compareConfig(cfg Config, dbClusters []sqlc.Cluster, dbNodes []sqlc.Node, verbose bool) bool {
	dbView := buildDBView(dbClusters, dbNodes)
	cfgView := buildCfgView(cfg)
	inSync := true

	logMismatch := func(format string, args ...any) {
		inSync = false
		if verbose {
			fmt.Printf("Mismatch: "+format+"\n", args...)
		}
	}

	for name := range cfgView {
		if _, found := dbView[name]; !found {
			logMismatch("cluster %q in config but not DB", name)
		}
	}
	for name := range dbView {
		if _, found := cfgView[name]; !found {
			logMismatch("cluster %q in DB but not config", name)
		}
	}

	for name, cv := range cfgView {
		dv, found := dbView[name]
		if !found {
			continue
		}

		if strings.TrimSpace(cv.Description) != strings.TrimSpace(dv.Description) {
			logMismatch("cluster %q description differs", name)
		}

		for key := range cv.Nodes {
			if _, f := dv.Nodes[key]; !f {
				logMismatch("cluster %q node %s in config but not DB", name, key)
			}
		}
		for key := range dv.Nodes {
			if _, f := cv.Nodes[key]; !f {
				logMismatch("cluster %q node %s in DB but not config", name, key)
			}
		}

		for key, cn := range cv.Nodes {
			dn, f := dv.Nodes[key]
			if !f {
				continue
			}

			if !equalStr(cn.Protocol, dn.Protocol) {
				logMismatch("cluster %q node %s protocol: cfg=%q db=%q", name, key, cn.Protocol, dn.Protocol)
			}
			if cn.Weight != dn.Weight {
				logMismatch("cluster %q node %s weight: cfg=%d db=%d", name, key, cn.Weight, dn.Weight)
			}
			if cn.MaxGroups != dn.MaxGroups {
				logMismatch("cluster %q node %s max_groups: cfg=%d db=%d", name, key, cn.MaxGroups, dn.MaxGroups)
			}
			if !equalStr(cn.User, dn.User) {
				logMismatch("cluster %q node %s user: cfg=%q db=%q", name, key, cn.User, dn.User)
			}
		}
	}

	return inSync
}

func mergeConfigWithDB(cfg Config, dbClusters []sqlc.Cluster, dbNodes []sqlc.Node) (Config, bool, error) {
	nodesByCluster := make(map[int64][]sqlc.Node)
	for _, n := range dbNodes {
		nodesByCluster[n.ClusterID] = append(nodesByCluster[n.ClusterID], n)
	}

	cfgView := buildCfgView(cfg)

	rebuilt := Config{
		Settings: cfg.Settings,
		Clusters: make([]Cluster, 0, len(dbClusters)),
	}

	for _, dbCluster := range dbClusters {
		nname := norm(dbCluster.Name)
		existing := cfgView[nname]

		desc := nullToString(dbCluster.Description)
		if desc == "" {
			desc = existing.Description
		}

		clusterNodes := nodesByCluster[dbCluster.ClusterID]
		sort.Slice(clusterNodes, func(i, j int) bool {
			if clusterNodes[i].Host != clusterNodes[j].Host {
				return clusterNodes[i].Host < clusterNodes[j].Host
			}
			return clusterNodes[i].Port < clusterNodes[j].Port
		})

		nodes := make([]Node, 0, len(clusterNodes))
		for _, dbNode := range clusterNodes {
			key := nodeKey(dbNode.Host, int(dbNode.Port))
			existingNode := existing.Nodes[key]

			proto := strings.ToLower(strings.TrimSpace(dbNode.Protocol))
			if proto == "" {
				proto = defaultStr(existingNode.Protocol, cfg.Settings.DefaultProtocol, "http")
			}

			maxGroups := nullToInt(dbNode.MaxGroups)
			if maxGroups == 0 {
				maxGroups = defaultInt(existingNode.MaxGroups, cfg.Settings.DefaultMaxGroups)
			}

			user := strings.TrimSpace(dbNode.AuthUser)
			if user == "" {
				user = existingNode.User
			}

			nodes = append(nodes, Node{
				Host:      strings.ToLower(strings.TrimSpace(dbNode.Host)),
				Port:      int(dbNode.Port),
				Protocol:  proto,
				Weight:    int(dbNode.Weight),
				MaxGroups: maxGroups,
				User:      user,
			})
		}

		rebuilt.Clusters = append(rebuilt.Clusters, Cluster{
			Name:        dbCluster.Name,
			Description: desc,
			Nodes:       nodes,
		})
	}

	sort.Slice(rebuilt.Clusters, func(i, j int) bool {
		return norm(rebuilt.Clusters[i].Name) < norm(rebuilt.Clusters[j].Name)
	})

	changed := !reflect.DeepEqual(normalizeConfig(cfg), normalizeConfig(rebuilt))
	return rebuilt, changed, nil
}

type clusterView struct {
	Description string
	Nodes       map[string]nodeView
}

type nodeView struct {
	Protocol  string
	Weight    int
	MaxGroups int
	User      string
}

func buildDBView(clusters []sqlc.Cluster, nodes []sqlc.Node) map[string]clusterView {
	res := make(map[string]clusterView, len(clusters))

	idToName := make(map[int64]string)
	for _, c := range clusters {
		nname := norm(c.Name)
		idToName[c.ClusterID] = nname
		res[nname] = clusterView{
			Description: nullToString(c.Description),
			Nodes:       make(map[string]nodeView),
		}
	}

	for _, n := range nodes {
		clusterName, ok := idToName[n.ClusterID]
		if !ok {
			continue
		}
		key := nodeKey(n.Host, int(n.Port))
		res[clusterName].Nodes[key] = nodeView{
			Protocol:  strings.ToLower(strings.TrimSpace(n.Protocol)),
			Weight:    int(n.Weight),
			MaxGroups: nullToInt(n.MaxGroups),
			User:      strings.TrimSpace(n.AuthUser),
		}
	}
	return res
}

func buildCfgView(cfg Config) map[string]clusterView {
	res := make(map[string]clusterView)

	for _, c := range cfg.Clusters {
		cv := clusterView{
			Description: strings.TrimSpace(c.Description),
			Nodes:       make(map[string]nodeView),
		}

		for _, n := range c.Nodes {
			key := nodeKey(n.Host, n.Port)
			cv.Nodes[key] = nodeView{
				Protocol:  strings.ToLower(defaultStr(n.Protocol, cfg.Settings.DefaultProtocol, "http")),
				Weight:    n.Weight,
				MaxGroups: defaultInt(n.MaxGroups, cfg.Settings.DefaultMaxGroups),
				User:      strings.TrimSpace(n.User),
			}
		}
		res[norm(c.Name)] = cv
	}
	return res
}

func normalizeConfig(cfg Config) Config {
	normalized := Config{
		Settings: cfg.Settings,
		Clusters: make([]Cluster, len(cfg.Clusters)),
	}

	for i, cl := range cfg.Clusters {
		nodes := make([]Node, len(cl.Nodes))
		for j, n := range cl.Nodes {
			nodes[j] = Node{
				Host:      strings.ToLower(strings.TrimSpace(n.Host)),
				Port:      n.Port,
				Protocol:  strings.ToLower(strings.TrimSpace(n.Protocol)),
				Weight:    n.Weight,
				MaxGroups: n.MaxGroups,
				User:      strings.TrimSpace(n.User),
			}
		}
		sort.Slice(nodes, func(a, b int) bool {
			if nodes[a].Host != nodes[b].Host {
				return nodes[a].Host < nodes[b].Host
			}
			return nodes[a].Port < nodes[b].Port
		})

		normalized.Clusters[i] = Cluster{
			Name:        strings.TrimSpace(cl.Name),
			Description: strings.TrimSpace(cl.Description),
			Nodes:       nodes,
		}
	}

	sort.Slice(normalized.Clusters, func(i, j int) bool {
		return norm(normalized.Clusters[i].Name) < norm(normalized.Clusters[j].Name)
	})

	return normalized
}

func nodeKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", strings.ToLower(strings.TrimSpace(host)), port)
}

func norm(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func equalStr(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func toNullString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	return sql.NullString{String: s, Valid: s != ""}
}

func nullToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func toNullInt64(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: i != 0}
}

func nullToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

func defaultStr(val, fallback, fallback2 string) string {
	val = strings.TrimSpace(val)
	if val != "" {
		return strings.ToLower(val)
	}
	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return strings.ToLower(fallback)
	}
	return strings.ToLower(fallback2)
}

func defaultInt(val, fallback int) int {
	if val != 0 {
		return val
	}
	return fallback
}
