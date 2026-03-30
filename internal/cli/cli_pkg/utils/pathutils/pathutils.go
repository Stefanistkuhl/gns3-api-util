package pathutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
)

type ServiceType string

const (
	TypeGNS3Server       ServiceType = "gns3_server"
	TypeClusterMaster    ServiceType = "cluster_master"
	TypeClusterNode      ServiceType = "cluster_node"
	TypeClusterFileStore ServiceType = "cluster_filestore"
)

var typeMap = map[models.NodeType]ServiceType{
	models.NodeTypeMaster:    TypeClusterMaster,
	models.NodeTypeWorker:    TypeClusterNode,
	models.NodeTypeFilestore: TypeClusterFileStore,
}

type LegacyGNS3Key struct {
	ServerURL   string `json:"server_url"`
	User        string `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type KeyFileV2 struct {
	Version        int               `json:"version"`
	StandaloneGNS3 []GNS3ServerEntry `json:"standalone_gns3,omitempty"`
	Clusters       []ClusterEntry    `json:"clusters,omitempty"`
}

type GNS3ServerEntry struct {
	Name        string `json:"name,omitempty"`
	URL         string `json:"url"`
	User        string `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Default     bool   `json:"default,omitempty"`
}

type ClusterUser struct {
	User        string `json:"user"`
	AccessToken string `json:"access_token"`
	Default     bool   `json:"default,omitempty"`
}

type ClusterEntry struct {
	Name            string            `json:"name"`
	Master          ServiceEntry      `json:"master"`
	Nodes           []ServiceEntry    `json:"nodes,omitempty"`
	RootFingerprint string            `json:"root_fingerprint,omitempty"`
	CaCert          string            `json:"ca_cert,omitempty"`
	GNS3Servers     []GNS3ServerEntry `json:"gns3_servers,omitempty"`
	Users           []ClusterUser     `json:"users,omitempty"`
}

type ServiceEntry struct {
	Type        ServiceType `json:"type"`
	ID          string      `json:"node_id"`
	URL         string      `json:"url"`
	User        string      `json:"user,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	TokenType   string      `json:"token_type,omitempty"`
}

func ExpandPath(p string) (string, error) {
	if p == "" || p[0] != '~' {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if p == "~" {
		return home, nil
	}

	return filepath.Join(home, p[2:]), nil
}

func GetGNS3Dir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("could not detect home dir: %w", err)
	}

	gns3Dir := filepath.Join(home, ".gns3")
	info, err := os.Stat(gns3Dir)
	switch {
	case os.IsNotExist(err):
		err = os.MkdirAll(gns3Dir, 0o750)
		if err != nil {
			return "", fmt.Errorf("could not create %q: %w", gns3Dir, err)
		}
	case err != nil:
		return "", fmt.Errorf("could not stat %q: %w", gns3Dir, err)
	case !info.IsDir():
		return "", fmt.Errorf("%q already exists and is not a directory", gns3Dir)
	}

	return gns3Dir, nil
}

func LoadGNS3KeysFile(path string) (*KeyFileV2, error) {
	f, err := os.Open(path) // #nosec G304
	if os.IsNotExist(err) {
		return &KeyFileV2{Version: 2}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not open %q: %w", path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("failed to close file: %v", closeErr)
		}
	}()

	buf := make([]byte, 1)
	n, readErr := f.Read(buf)
	if readErr == io.EOF || n == 0 {
		return &KeyFileV2{Version: 2}, nil
	}
	if readErr != nil {
		return nil, fmt.Errorf("failed to read file: %w", readErr)
	}

	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		return nil, fmt.Errorf("failed to seek: %w", seekErr)
	}

	if buf[0] == '{' && isV2Format(f) {
		if _, seekErr := f.Seek(0, 0); seekErr != nil {
			return nil, seekErr
		}
		var kf KeyFileV2
		if decodeErr := json.NewDecoder(f).Decode(&kf); decodeErr != nil {
			return nil, fmt.Errorf("failed to decode V2: %w", decodeErr)
		}
		return &kf, nil
	}

	// If not V2, rewind and parse as legacy
	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		return nil, seekErr
	}

	kf, err := parseLegacyFormat(f)
	if err != nil {
		return nil, err
	}

	_ = SaveKeysFile(path, kf)

	return kf, nil
}

func parseLegacyFormat(r io.Reader) (*KeyFileV2, error) {
	kf := &KeyFileV2{Version: 2}
	dec := json.NewDecoder(r)

	for {
		var old LegacyGNS3Key
		if err := dec.Decode(&old); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode legacy JSON: %w", err)
		}

		kf.StandaloneGNS3 = append(kf.StandaloneGNS3, GNS3ServerEntry{
			URL:         old.ServerURL,
			User:        old.User,
			AccessToken: old.AccessToken,
			TokenType:   old.TokenType,
		})
	}

	return kf, nil
}

func isV2Format(f *os.File) bool {
	var peek struct {
		Version int `json:"version"`
	}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&peek); err != nil {
		return false
	}
	return peek.Version == 2
}

func SaveKeysFile(path string, kf *KeyFileV2) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open key file %q: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(kf); encErr != nil {
		return fmt.Errorf("failed to write key file: %w", encErr)
	}

	return nil
}

func ResolveKeyFilePath(cfgKeyFile string) (string, error) {
	if cfgKeyFile != "" {
		return ExpandPath(cfgKeyFile)
	}
	dir, err := GetGNS3Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gns3key"), nil
}

func (k *KeyFileV2) IsMasterPresent(masterURL string) (*ServiceEntry, bool) {
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Master.URL == masterURL {
			return &c.Master, true
		}
	}
	return nil, false
}

func (k *KeyFileV2) UpdateMaster(token, serverURL, user string) {
	entry := ServiceEntry{
		Type:        TypeClusterMaster,
		AccessToken: token,
		URL:         serverURL,
		User:        user,
	}
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Master.URL == serverURL {
			c.Master = entry
			return
		}
	}
}

func (k *KeyFileV2) CheckIfClusterExists(name string) bool {
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Name == name {
			return true
		}
	}
	return false
}

func (k *KeyFileV2) GetCACertForMaster(serverURL string) []byte {
	if k == nil {
		return nil
	}

	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Master.URL == serverURL {
			if c.CaCert == "" {
				return nil
			}
			return []byte(c.CaCert)
		}
	}
	return nil
}

func (k *KeyFileV2) IsNodePresentInCluster(node *models.NodeInfo, cluster *ClusterEntry) bool {
	for i := range cluster.Nodes {
		n := cluster.Nodes[i]
		if k.CompareSvcType(node.Type, n.Type) && n.URL == fmt.Sprintf("https://%s:%d", node.IP, node.APIPort) && node.ID == n.ID {
			return true
		}
	}
	return false
}

func (k *KeyFileV2) NodeTypeToSvcType(inp models.NodeType) ServiceType {
	if svc, ok := typeMap[inp]; ok {
		return svc
	}
	return "unknown_service"
}

func (k *KeyFileV2) CompareSvcType(node models.NodeType, svc ServiceType) bool {
	expectedSvc, ok := typeMap[node]
	if !ok {
		return false
	}

	return expectedSvc == svc
}

func (k *KeyFileV2) AddNodes(nodes []ServiceEntry, cluster *ClusterEntry) {
	totalExpected := len(nodes) + len(cluster.Nodes)
	nodeArr := make([]ServiceEntry, 0, totalExpected)
	nodeArr = append(nodeArr, nodes...)
	nodeArr = append(nodeArr, cluster.Nodes...)
	for i := range k.Clusters {
		if k.Clusters[i].Name == cluster.Name {
			k.Clusters[i].Nodes = nodeArr
		}
	}
}

func (k *KeyFileV2) RemoveNonExistantEntrys(clusterName string, discoveredIDs []string) []ServiceEntry {
	var removed []ServiceEntry

	alive := make(map[string]struct{})
	for _, id := range discoveredIDs {
		alive[id] = struct{}{}
	}

	for i := range k.Clusters {
		if k.Clusters[i].Name == clusterName {
			var kept []ServiceEntry
			for _, node := range k.Clusters[i].Nodes {
				if _, exists := alive[node.ID]; exists {
					kept = append(kept, node)
				} else {
					removed = append(removed, node)
				}
			}
			k.Clusters[i].Nodes = kept
		}
	}
	return removed
}

func (k *KeyFileV2) SyncNodes(freshNodes []ServiceEntry, cluster *ClusterEntry) {
	for i := range k.Clusters {
		if k.Clusters[i].Name == cluster.Name {
			k.Clusters[i].Nodes = freshNodes
			break
		}
	}
}

func (k *KeyFileV2) ListUsersForServer(serverURL string) []GNS3ServerEntry {
	norm := nwutils.NormalizeURL(serverURL)
	var out []GNS3ServerEntry
	for _, e := range k.StandaloneGNS3 {
		if nwutils.NormalizeURL(e.URL) == norm {
			out = append(out, e)
		}
	}
	return out
}

func (k *KeyFileV2) GetUserForServer(serverURL, name string) (*GNS3ServerEntry, bool) {
	norm := nwutils.NormalizeURL(serverURL)
	var first *GNS3ServerEntry
	for i := range k.StandaloneGNS3 {
		e := &k.StandaloneGNS3[i]
		if nwutils.NormalizeURL(e.URL) != norm {
			continue
		}
		if first == nil {
			first = e
		}
		if name == "" && e.Default {
			return e, true
		}
		if name != "" && e.User == name {
			return e, true
		}
	}
	if name == "" && first != nil {
		return first, true // implicit first-entry default
	}
	return nil, false
}

func (k *KeyFileV2) SetDefaultUserForServer(serverURL, userName string) error {
	norm := nwutils.NormalizeURL(serverURL)
	found := false
	for i := range k.StandaloneGNS3 {
		e := &k.StandaloneGNS3[i]
		if nwutils.NormalizeURL(e.URL) != norm {
			continue
		}
		if e.User == userName {
			e.Default = true
			found = true
		} else {
			e.Default = false
		}
	}
	if !found {
		return fmt.Errorf("user %q not found for server %s", userName, serverURL)
	}
	return nil
}

func (k *KeyFileV2) UpsertUserForServer(entry *GNS3ServerEntry) {
	norm := nwutils.NormalizeURL(entry.URL)
	isFirst := true
	for i := range k.StandaloneGNS3 {
		e := &k.StandaloneGNS3[i]
		if nwutils.NormalizeURL(e.URL) == norm {
			isFirst = false
			if e.User == entry.User {
				entry.Default = e.Default // preserve existing default flag
				k.StandaloneGNS3[i] = *entry
				return
			}
		}
	}
	if isFirst {
		entry.Default = true
	}
	k.StandaloneGNS3 = append(k.StandaloneGNS3, *entry)
}

func (k *KeyFileV2) ListClusterUsers(clusterName string) []ClusterUser {
	for i := range k.Clusters {
		if k.Clusters[i].Name == clusterName {
			return k.Clusters[i].Users
		}
	}
	return nil
}

func (k *KeyFileV2) GetClusterUser(clusterName, name string) (*ClusterUser, bool) {
	var first *ClusterUser
	for i := range k.Clusters {
		c := &k.Clusters[i]
		if c.Name != clusterName {
			continue
		}
		for j := range c.Users {
			u := &c.Users[j]
			if first == nil {
				first = u
			}
			if name == "" && u.Default {
				return u, true
			}
			if name != "" && u.User == name {
				return u, true
			}
		}
	}
	if name == "" && first != nil {
		return first, true
	}
	return nil, false
}

func (k *KeyFileV2) SetDefaultClusterUser(clusterName, userName string) error {
	for i := range k.Clusters {
		if k.Clusters[i].Name != clusterName {
			continue
		}
		found := false
		for j := range k.Clusters[i].Users {
			u := &k.Clusters[i].Users[j]
			if u.User == userName {
				u.Default = true
				found = true
			} else {
				u.Default = false
			}
		}
		if !found {
			return fmt.Errorf("user %q not found in cluster %q", userName, clusterName)
		}
		return nil
	}
	return fmt.Errorf("cluster %q not found", clusterName)
}

func (k *KeyFileV2) UpsertClusterUser(clusterName string, cu ClusterUser) {
	for i := range k.Clusters {
		if k.Clusters[i].Name != clusterName {
			continue
		}
		if len(k.Clusters[i].Users) == 0 {
			cu.Default = true
		}
		for j := range k.Clusters[i].Users {
			if k.Clusters[i].Users[j].User == cu.User {
				cu.Default = k.Clusters[i].Users[j].Default // preserve
				k.Clusters[i].Users[j] = cu
				return
			}
		}
		k.Clusters[i].Users = append(k.Clusters[i].Users, cu)
		return
	}
}

func (k *KeyFileV2) RemoveClusterByName(name string) {
	for i := range k.Clusters {
		cluster := k.Clusters[i]
		if cluster.Name == name {
			k.Clusters = append(k.Clusters[:i], k.Clusters[i+1:]...)
			return
		}
	}
}
