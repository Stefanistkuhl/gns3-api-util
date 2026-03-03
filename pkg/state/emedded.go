package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

type InternalState struct {
	Server *embed.Etcd
}

type EtcdConfig struct {
	DataDir    string
	ClientPort string
	PeerPort   string
	Name       string
	Cluster    string
	State      string
}

func StartEmbedded(cfg *EtcdConfig) (*InternalState, error) {
	ec := embed.NewConfig()
	ec.Dir = cfg.DataDir
	ec.Name = cfg.Name

	lpurl, _ := url.Parse("http://0.0.0.0:" + cfg.PeerPort)
	lcurl, _ := url.Parse("http://0.0.0.0:" + cfg.ClientPort)
	acurl, _ := url.Parse("http://localhost:" + cfg.ClientPort)
	apurl, _ := url.Parse("http://localhost:" + cfg.PeerPort)

	ec.ListenPeerUrls = []url.URL{*lpurl}
	ec.ListenClientUrls = []url.URL{*lcurl}
	ec.AdvertiseClientUrls = []url.URL{*acurl}
	ec.AdvertisePeerUrls = []url.URL{*apurl}

	ec.InitialCluster = cfg.Cluster
	ec.ClusterState = cfg.State
	ec.InitialClusterToken = "gns3util-cluster"

	e, err := embed.StartEtcd(ec)
	if err != nil {
		return nil, err
	}

	select {
	case <-e.Server.ReadyNotify():
		log.Printf("Etcd node '%s' ready on :%s", cfg.Name, cfg.ClientPort)
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		return nil, fmt.Errorf("etcd startup timeout")
	}

	return &InternalState{Server: e}, nil
}

func StartMaster(dataDir string) (*InternalState, error) {
	return StartEmbedded(&EtcdConfig{
		DataDir:    dataDir,
		ClientPort: "2379",
		PeerPort:   "2380",
		Name:       "master",
		Cluster:    "master=http://localhost:2380",
		State:      "new",
	})
}

func BootstrapMasterAuth(ctx context.Context, client *clientv3.Client) error {
	authResp, err := client.AuthStatus(ctx)
	if err != nil {
		return fmt.Errorf("check auth status: %w", err)
	}
	if authResp.Enabled {
		log.Println("Auth already enabled, skipping bootstrap")
		return nil
	}

	rootPassword := os.Getenv("ROOT_PASSWORD")
	if rootPassword == "" {
		return fmt.Errorf("ROOT_PASSWORD env var required")
	}

	joinToken := os.Getenv("JOIN_TOKEN")
	if joinToken == "" {
		return fmt.Errorf("JOIN_TOKEN env var required")
	}

	_, err = client.RoleAdd(ctx, "root")
	if err != nil {
		return fmt.Errorf("create root role: %w", err)
	}
	log.Println("Created root role")

	_, err = client.UserAdd(ctx, "root", rootPassword)
	if err != nil {
		return fmt.Errorf("create root user: %w", err)
	}
	log.Println("Created root user")

	_, err = client.UserGrantRole(ctx, "root", "root")
	if err != nil {
		return fmt.Errorf("grant root role: %w", err)
	}
	log.Println("Granted root role to root user")

	_, err = client.AuthEnable(ctx)
	if err != nil {
		return fmt.Errorf("enable auth: %w", err)
	}
	log.Println("Authentication enabled")

	rootAuthClient, err := clientv3.New(clientv3.Config{
		Endpoints:   client.Endpoints(),
		DialTimeout: 5 * time.Second,
		Username:    "root",
		Password:    rootPassword,
	})
	if err != nil {
		return fmt.Errorf("create authenticated client: %w", err)
	}
	defer rootAuthClient.Close()

	_, err = rootAuthClient.RoleAdd(ctx, "cluster-node")
	if err != nil {
		return fmt.Errorf("create cluster-node role: %w", err)
	}
	log.Println("Created cluster-node role")

	_, err = rootAuthClient.RoleGrantPermission(
		ctx,
		"cluster-node",
		"\x00",
		"\x00",
		clientv3.PermissionType(clientv3.PermReadWrite),
	)
	if err != nil {
		return fmt.Errorf("grant permissions to cluster-node: %w", err)
	}
	log.Println("Granted readwrite permissions to cluster-node role")

	_, err = rootAuthClient.UserAdd(ctx, "joiner", joinToken)
	if err != nil {
		return fmt.Errorf("create joiner user: %w", err)
	}
	log.Println("Created joiner user")

	_, err = rootAuthClient.UserGrantRole(ctx, "joiner", "cluster-node")
	if err != nil {
		return fmt.Errorf("grant cluster-node role to joiner: %w", err)
	}
	log.Println("Granted cluster-node role to joiner")

	return nil
}

func StartFileStore(ctx context.Context, dataDir, masterAPIURL string) (*InternalState, error) {
	memberDir := filepath.Join(dataDir, "member")
	_, err := os.Stat(memberDir)
	hasData := err == nil

	var cluster string
	var stateStr string

	if hasData {
		log.Println("Found existing etcd data, skipping join API call and booting directly...")
		cluster = ""
		stateStr = "existing"
	} else {
		log.Println("No local data found. Calling Master API to join cluster...")

		joinToken := os.Getenv("JOIN_TOKEN")
		if joinToken == "" {
			return nil, fmt.Errorf("JOIN_TOKEN env var required")
		}

		peerURL := "http://localhost:2480"
		reqBody, _ := json.Marshal(map[string]any{
			"name":      "filestore",
			"peer_urls": []string{peerURL},
		})

		req, err := http.NewRequestWithContext(ctx, "POST", masterAPIURL+"/cluster/join", bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+joinToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to call master join api: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("master rejected join request: %s", string(body))
		}

		var joinResp struct {
			MemberID uint64 `json:"member_id"`
			Cluster  string `json:"cluster"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&joinResp); err != nil {
			return nil, fmt.Errorf("failed to decode join response: %w", err)
		}

		log.Printf("Added to cluster by Master. Assigned Member ID: %d", joinResp.MemberID)

		cluster = joinResp.Cluster
		stateStr = "existing"
	}

	return StartEmbedded(&EtcdConfig{
		DataDir:    dataDir,
		ClientPort: "2479",
		PeerPort:   "2480",
		Name:       "filestore",
		Cluster:    cluster,
		State:      stateStr,
	})
}
