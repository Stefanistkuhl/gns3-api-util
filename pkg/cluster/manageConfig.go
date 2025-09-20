package cluster

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func ApplyConfig(cfg Config) error {
	conn, openErr := db.InitIfNeeded()
	if openErr != nil {
		return fmt.Errorf("DB open error: %v", openErr)
	}

	defer conn.Close()

	var create db.CreateClustersAndNodes
	for _, cluster := range cfg.Clusters {
		var createCluster db.ClusterAndNodes
		c := db.ClusterName{
			Name: cluster.Name,
			Desc: sql.NullString{
				String: cluster.Description,
				Valid:  cluster.Description != "",
			},
		}
		createCluster.Cluster = c
		for _, node := range cluster.Nodes {
			n := db.NodeDataAll{
				User:      node.User,
				Protocol:  node.Protocol,
				Host:      node.Host,
				Port:      node.Port,
				Weight:    node.Weight,
				MaxGroups: node.MaxGroups,
			}
			if utils.ValidateAndTestUrl(fmt.Sprintf("%s://%s:%d", node.Protocol, node.Host, node.Port)) {
				createCluster.Nodes = append(createCluster.Nodes, n)
			} else {
				return fmt.Errorf("cant connect to: %s", fmt.Sprintf("%s://%s:%d", node.Protocol, node.Host, node.Port))
			}

		}
		create.Clusters = append(create.Clusters, createCluster)
	}
	createNeeded, err := BuildCreateDelta(create, conn)
	if err != nil {
		return fmt.Errorf("failed to get the diff of the existing elements in the db and config: %w", err)
	}
	createErr := CreateClusterAndNodes(createNeeded, conn)
	if createErr != nil {
		return createErr
	}

	if err := UpdateClustersAndNodes(cfg, conn); err != nil {
		return fmt.Errorf("update existing: %w", err)
	}

	return nil
}

func UpdateClustersAndNodes(cfg Config, conn *sql.DB) error {
	dbClusters, err := db.GetClusters(conn)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("load clusters: %w", err)
	}
	dbNodes, err := db.GetNodes(conn)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("load nodes: %w", err)
	}

	cByName := make(map[string]db.ClusterName, len(dbClusters))
	for _, c := range dbClusters {
		cByName[normName(c.Name)] = c
	}

	// Map (cluster_id, host:port) -> db node
	type key struct {
		ClusterID int
		HostPort  string
	}
	nByKey := make(map[key]db.NodeDataAll, len(dbNodes))
	for _, n := range dbNodes {
		nByKey[key{ClusterID: n.ClusterID, HostPort: nodeKey(n.Host, n.Port)}] = n
	}

	for _, cl := range cfg.Clusters {
		nname := normName(cl.Name)
		dbc, ok := cByName[nname]
		if !ok {
			continue
		}
		curDesc := ""
		if dbc.Desc.Valid {
			curDesc = strings.TrimSpace(dbc.Desc.String)
		}
		newDesc := strings.TrimSpace(cl.Description)
		if curDesc != newDesc {
			_, err := db.UpdateRows(conn,
				"UPDATE clusters SET description = ? WHERE cluster_id = ?",
				func() sql.NullString {
					if newDesc == "" {
						return sql.NullString{Valid: false}
					}
					return sql.NullString{String: newDesc, Valid: true}
				}(),
				dbc.Id,
			)
			if err != nil {
				return fmt.Errorf("update cluster desc %s: %w", cl.Name, err)
			}
		}

		for _, nn := range cl.Nodes {
			hostPort := nodeKey(nn.Host, nn.Port)
			dbn, ok := nByKey[key{ClusterID: dbc.Id, HostPort: hostPort}]
			if !ok {
				continue
			}

			proto := strings.ToLower(strings.TrimSpace(nn.Protocol))
			if proto == "" {
				proto = strings.ToLower(strings.TrimSpace(cfg.Settings.DefaultProtocol))
			}
			maxGroups := nn.MaxGroups
			if maxGroups == 0 {
				maxGroups = cfg.Settings.DefaultMaxGroups
			}
			user := strings.TrimSpace(nn.User)

			needProto := proto != strings.ToLower(strings.TrimSpace(dbn.Protocol))
			needWeight := nn.Weight != dbn.Weight
			needMax := maxGroups != dbn.MaxGroups
			needUser := user != strings.TrimSpace(dbn.User)

			if needProto || needWeight || needMax || needUser {
				_, err := db.UpdateRows(conn, `
UPDATE nodes SET protocol = ?, auth_user = ?, weight = ?, max_groups = ?
WHERE cluster_id = ? AND host = ? AND port = ?
`, proto, user, nn.Weight, maxGroups, dbc.Id, dbn.Host, dbn.Port)
				if err != nil {
					return fmt.Errorf("update node %s in cluster %s: %w", hostPort, cl.Name, err)
				}
			}
		}
	}

	return nil
}

func normName(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func nodeUniqKey(protocol, host string, port int) string {
	return fmt.Sprintf("%s|%s|%d",
		strings.ToLower(strings.TrimSpace(protocol)),
		strings.ToLower(strings.TrimSpace(host)),
		port,
	)
}

func CreateClusterAndNodes(create db.CreateClustersAndNodes, conn *sql.DB) error {
	var na []db.NodeDataAll
	var ca []db.ClusterName
	for _, cluster := range create.Clusters {
		ca = append(ca, cluster.Cluster)
	}
	clusters, createClustersErr := db.CreateClusters(conn, ca)
	if createClustersErr != nil {
		return createClustersErr
	}
	for i, cluster := range clusters {
		for j := range create.Clusters[i].Nodes {
			create.Clusters[i].Nodes[j].ClusterID = cluster.Id
			na = append(na, create.Clusters[i].Nodes[j])
		}
	}
	createNodesErr := db.InsertNodesIntoClusters(conn, na)
	if createNodesErr != nil {
		return createNodesErr
	}

	return nil
}

func BuildCreateDelta(create db.CreateClustersAndNodes, conn *sql.DB) (db.CreateClustersAndNodes, error) {
	dbClusters, err := db.GetClusters(conn)
	if err != nil && err != sql.ErrNoRows {
		return db.CreateClustersAndNodes{}, fmt.Errorf("load clusters: %w", err)
	}
	dbNodes, err := db.GetNodes(conn)
	if err != nil && err != sql.ErrNoRows {
		return db.CreateClustersAndNodes{}, fmt.Errorf("load nodes: %w", err)
	}

	existingClusters := make(map[string]db.ClusterName, len(dbClusters))
	for _, c := range dbClusters {
		existingClusters[normName(c.Name)] = c
	}

	existingNodes := make(map[int]map[string]struct{})
	for _, n := range dbNodes {
		if _, ok := existingNodes[n.ClusterID]; !ok {
			existingNodes[n.ClusterID] = make(map[string]struct{})
		}
		k := nodeUniqKey(n.Protocol, n.Host, n.Port)
		existingNodes[n.ClusterID][k] = struct{}{}
	}

	var out db.CreateClustersAndNodes

	for _, req := range create.Clusters {
		nname := normName(req.Cluster.Name)
		existing, clusterExists := existingClusters[nname]

		if !clusterExists {
			seen := make(map[string]struct{})
			var newNodes []db.NodeDataAll
			for _, n := range req.Nodes {
				key := nodeUniqKey(n.Protocol, n.Host, n.Port)
				if _, dup := seen[key]; dup {
					continue
				}
				seen[key] = struct{}{}
				newNodes = append(newNodes, n)
			}
			if len(newNodes) > 0 || true {
				out.Clusters = append(out.Clusters, db.ClusterAndNodes{
					Cluster: req.Cluster,
					Nodes:   newNodes,
				})
			}
			continue
		}

		exNodes := existingNodes[existing.Id]
		seenNew := make(map[string]struct{})
		var missing []db.NodeDataAll
		for _, n := range req.Nodes {
			key := nodeUniqKey(n.Protocol, n.Host, n.Port)
			if _, dup := seenNew[key]; dup {
				continue
			}
			seenNew[key] = struct{}{}
			if exNodes != nil {
				if _, ok := exNodes[key]; ok {
					continue
				}
			}
			missing = append(missing, n)
		}
		if len(missing) > 0 {
			out.Clusters = append(out.Clusters, db.ClusterAndNodes{
				Cluster: existing,
				Nodes:   missing,
			})
		}
	}

	return out, nil
}

func SyncConfigWithDb(cfg Config) (Config, bool, error) {
	conn, openErr := db.InitIfNeeded()
	if openErr != nil {
		return cfg, false, fmt.Errorf("DB open error: %v", openErr)
	}
	defer conn.Close()

	dbClusters, err := db.GetClusters(conn)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, false, nil
		}
		return cfg, false, fmt.Errorf("error fetching clusters: %w", err)
	}

	dbNodes, err := db.GetNodes(conn)
	if err != nil {
		if err == sql.ErrNoRows {
		} else {
			return cfg, false, fmt.Errorf("error fetching nodes: %w", err)
		}
	}

	updatedCfg, changed := MergeConfigWithDb(cfg, dbClusters, dbNodes)
	return updatedCfg, changed, nil
}

func CheckConfigWithDb(cfg Config, verbose bool) (bool, error) {
	conn, openErr := db.InitIfNeeded()
	if openErr != nil {
		if verbose {
			fmt.Printf("DB open error: %v\n", openErr)
		}
		return false, fmt.Errorf("db open error: %w", openErr)
	}
	defer conn.Close()

	dbClusters, err := db.GetClusters(conn)
	if err != nil {
		if err == sql.ErrNoRows {
			return len(cfg.Clusters) == 0, nil
		}
		if verbose {
			fmt.Printf("Error fetching clusters: %v\n", err)
		}
		return false, fmt.Errorf("error fetching clusters: %w", err)
	}

	dbNodes, err := db.GetNodes(conn)
	if err != nil {
		if err == sql.ErrNoRows {
		} else {
			if verbose {
				fmt.Printf("Error fetching nodes: %v\n", err)
			}
			return false, fmt.Errorf("error fetching nodes: %w", err)
		}
	}

	dbView := buildDbView(dbClusters, dbNodes)
	cfgView := buildCfgView(cfg)

	inSync := true

	for name := range cfgView {
		if _, found := dbView[name]; !found {
			if verbose {
				fmt.Printf("Mismatch: cluster %q exists in config but not in DB\n", name)
			}
			inSync = false
		}
	}
	for name := range dbView {
		if _, found := cfgView[name]; !found {
			if verbose {
				fmt.Printf("Mismatch: cluster %q exists in DB but not in config\n", name)
			}
			inSync = false
		}
	}

	for name, cv := range cfgView {
		dv, found := dbView[name]
		if !found {
			continue
		}

		cfgDesc := strings.TrimSpace(cv.Description)
		dbDesc := ""
		if dv.Description.Valid {
			dbDesc = strings.TrimSpace(dv.Description.String)
		}
		if cfgDesc != dbDesc {
			if verbose {
				fmt.Printf("Mismatch: cluster %q description differs. cfg=%q db=%q\n", name, cfgDesc, dbDesc)
			}
			inSync = false
		}

		for key := range cv.Nodes {
			if _, f := dv.Nodes[key]; !f {
				if verbose {
					fmt.Printf("Mismatch: cluster %q node %s exists in config but not in DB\n", name, key)
				}
				inSync = false
			}
		}
		for key := range dv.Nodes {
			if _, f := cv.Nodes[key]; !f {
				if verbose {
					fmt.Printf("Mismatch: cluster %q node %s exists in DB but not in config\n", name, key)
				}
				inSync = false
			}
		}

		for key, cn := range cv.Nodes {
			dn, f := dv.Nodes[key]
			if !f {
				continue
			}
			dbProto := dn.Protocol
			if strings.TrimSpace(dbProto) == "" {
				dbProto = cfg.Settings.DefaultProtocol
			}
			if !equalStr(cn.Protocol, dbProto) {
				if verbose {
					fmt.Printf("Mismatch: cluster %q node %s protocol differs. cfg=%q db(effective)=%q\n", name, key, cn.Protocol, dbProto)
				}
				inSync = false
			}
			if cn.Weight != dn.Weight {
				if verbose {
					fmt.Printf("Mismatch: cluster %q node %s weight differs. cfg=%d db=%d\n", name, key, cn.Weight, dn.Weight)
				}
				inSync = false
			}
			dbMax := dn.MaxGroups
			if dbMax == 0 {
				dbMax = cfg.Settings.DefaultMaxGroups
			}
			if cn.MaxGroups != dbMax {
				if verbose {
					fmt.Printf("Mismatch: cluster %q node %s max_groups differs. cfg=%d db(effective)=%d\n", name, key, cn.MaxGroups, dbMax)
				}
				inSync = false
			}
			if !equalStr(cn.User, dn.User) {
				if verbose {
					fmt.Printf("Mismatch: cluster %q node %s user differs. cfg=%q db=%q\n", name, key, cn.User, dn.User)
				}
				inSync = false
			}
		}
	}

	return inSync, nil
}

func MergeConfigWithDb(
	cfg Config,
	dbClusters []db.ClusterName,
	dbNodes []db.NodeDataAll,
) (Config, bool) {
	changed := false

	cfgView := make(map[string]cfgClusterView)
	for _, c := range cfg.Clusters {
		nname := norm(c.Name)
		cv, ok := cfgView[nname]
		if !ok {
			cv = cfgClusterView{
				Description: strings.TrimSpace(c.Description),
				Nodes:       make(map[string]cfgNode),
			}
		} else if cv.Description == "" && strings.TrimSpace(c.Description) != "" {
			cv.Description = strings.TrimSpace(c.Description)
		}
		for _, n := range c.Nodes {
			proto := n.Protocol
			if proto == "" {
				proto = cfg.Settings.DefaultProtocol
			}
			maxGroups := n.MaxGroups
			if maxGroups == 0 {
				maxGroups = cfg.Settings.DefaultMaxGroups
			}
			key := nodeKey(n.Host, n.Port)
			if _, exists := cv.Nodes[key]; exists {
				changed = true
				continue
			}
			cv.Nodes[key] = cfgNode{
				Protocol:  strings.ToLower(proto),
				Weight:    n.Weight,
				MaxGroups: maxGroups,
				User:      n.User,
			}
		}
		cfgView[nname] = cv
	}

	idToName := make(map[int]string, len(dbClusters))
	for _, cdb := range dbClusters {
		nname := norm(cdb.Name)
		idToName[cdb.Id] = nname
		cv, exists := cfgView[nname]
		if !exists {
			cv = cfgClusterView{
				Description: strings.TrimSpace(cdb.Desc.String),
				Nodes:       make(map[string]cfgNode),
			}
			cfgView[nname] = cv
			changed = true
		} else if cv.Description == "" && cdb.Desc.Valid {
			cv.Description = strings.TrimSpace(cdb.Desc.String)
			cfgView[nname] = cv
			changed = true
		}
	}

	for _, n := range dbNodes {
		nname, ok := idToName[n.ClusterID]
		if !ok {
			continue
		}
		cv := cfgView[nname]
		if cv.Nodes == nil {
			cv.Nodes = make(map[string]cfgNode)
		}
		key := nodeKey(n.Host, n.Port)
		proto := strings.TrimSpace(n.Protocol)
		if proto == "" {
			proto = cfg.Settings.DefaultProtocol
		}
		dbNodeValue := cfgNode{
			Protocol: strings.ToLower(proto),
			Weight:   n.Weight,
			MaxGroups: func() int {
				if n.MaxGroups == 0 {
					return cfg.Settings.DefaultMaxGroups
				}
				return n.MaxGroups
			}(),
			User: n.User,
		}
		if existing, exists := cv.Nodes[key]; !exists {
			cv.Nodes[key] = dbNodeValue
			cfgView[nname] = cv
			changed = true
		} else if !equalStr(existing.Protocol, dbNodeValue.Protocol) ||
			existing.Weight != dbNodeValue.Weight ||
			existing.MaxGroups != dbNodeValue.MaxGroups ||
			existing.User != dbNodeValue.User {
			cv.Nodes[key] = dbNodeValue
			cfgView[nname] = cv
			changed = true
		}
	}

	rebuilt := Config{
		Settings: cfg.Settings,
		Clusters: make([]Cluster, 0, len(cfgView)),
	}

	type kv struct {
		name string
		cv   cfgClusterView
	}
	all := make([]kv, 0, len(cfgView))
	for nname, cv := range cfgView {
		all = append(all, kv{name: nname, cv: cv})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].name < all[j].name })

	for _, it := range all {
		nname, cv := it.name, it.cv
		cl := Cluster{
			Name:        nname,
			Description: cv.Description,
			Nodes:       make([]Node, 0, len(cv.Nodes)),
		}
		nkeys := make([]string, 0, len(cv.Nodes))
		for k := range cv.Nodes {
			nkeys = append(nkeys, k)
		}
		sort.Strings(nkeys)
		for _, key := range nkeys {
			nn := cv.Nodes[key]
			host, portStr, _ := strings.Cut(key, ":")
			pi, _ := strconv.Atoi(portStr)
			cl.Nodes = append(cl.Nodes, Node{
				Host:      host,
				Port:      pi,
				Protocol:  nn.Protocol,
				Weight:    nn.Weight,
				MaxGroups: nn.MaxGroups,
				User:      nn.User,
			})
		}
		rebuilt.Clusters = append(rebuilt.Clusters, cl)
	}

	return rebuilt, changed
}

type dbClusterView struct {
	Description sql.NullString
	Nodes       map[string]dbNode
}

type dbNode struct {
	Protocol  string
	Weight    int
	MaxGroups int
	User      string
}

type cfgClusterView struct {
	Description string
	Nodes       map[string]cfgNode
}

type cfgNode struct {
	Protocol  string
	Weight    int
	MaxGroups int
	User      string
}

func buildDbView(clusters []db.ClusterName, nodes []db.NodeDataAll) map[string]dbClusterView {
	idToMeta := make(map[int]db.ClusterName, len(clusters))
	for _, c := range clusters {
		idToMeta[c.Id] = c
	}
	res := make(map[string]dbClusterView, len(clusters))
	for _, c := range clusters {
		res[norm(c.Name)] = dbClusterView{
			Description: c.Desc,
			Nodes:       make(map[string]dbNode),
		}
	}
	for _, n := range nodes {
		meta, ok := idToMeta[n.ClusterID]
		name := "unknown"
		if ok {
			name = norm(meta.Name)
		}
		entry := res[name]
		if entry.Nodes == nil {
			entry.Nodes = make(map[string]dbNode)
		}

		key := nodeKey(n.Host, n.Port)

		entry.Nodes[key] = dbNode{
			Protocol:  strings.ToLower(n.Protocol),
			Weight:    n.Weight,
			MaxGroups: n.MaxGroups,
			User:      n.User,
		}
		res[name] = entry
	}
	return res
}

func buildCfgView(cfg Config) map[string]cfgClusterView {
	res := make(map[string]cfgClusterView)

	for _, c := range cfg.Clusters {
		nname := norm(c.Name)

		cv, ok := res[nname]
		if !ok {
			cv = cfgClusterView{
				Description: strings.TrimSpace(c.Description),
				Nodes:       make(map[string]cfgNode),
			}
		} else {
			if cv.Description == "" && strings.TrimSpace(c.Description) != "" {
				cv.Description = strings.TrimSpace(c.Description)
			}
		}

		for _, n := range c.Nodes {
			proto := n.Protocol
			if proto == "" {
				proto = cfg.Settings.DefaultProtocol
			}
			maxGroups := n.MaxGroups
			if maxGroups == 0 {
				maxGroups = cfg.Settings.DefaultMaxGroups
			}

			key := nodeKey(n.Host, n.Port)
			if _, exists := cv.Nodes[key]; !exists {
				cv.Nodes[key] = cfgNode{
					Protocol:  strings.ToLower(proto),
					Weight:    n.Weight,
					MaxGroups: maxGroups,
					User:      n.User,
				}
			}
		}

		res[nname] = cv
	}

	return res
}

func nodeKey(host string, port int) string {
	return fmt.Sprintf("%s:%d",
		strings.ToLower(strings.TrimSpace(host)),
		port,
	)
}

func norm(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func equalStr(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
