package state

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
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
	TLSDir     string
}

func StartEmbedded(cfg *EtcdConfig) (*InternalState, error) {
	ec := embed.NewConfig()
	ec.Dir = cfg.DataDir
	ec.Name = cfg.Name

	lpurl, _ := url.Parse("https://0.0.0.0:" + cfg.PeerPort)
	lcurl, _ := url.Parse("https://0.0.0.0:" + cfg.ClientPort)
	acurl, _ := url.Parse("https://localhost:" + cfg.ClientPort)
	apurl, _ := url.Parse("https://localhost:" + cfg.PeerPort)

	ec.ListenPeerUrls = []url.URL{*lpurl}
	ec.ListenClientUrls = []url.URL{*lcurl}
	ec.AdvertiseClientUrls = []url.URL{*acurl}
	ec.AdvertisePeerUrls = []url.URL{*apurl}

	ec.PeerTLSInfo = transport.TLSInfo{
		CertFile:       filepath.Join(cfg.TLSDir, "node.crt"),
		KeyFile:        filepath.Join(cfg.TLSDir, "node.key"),
		TrustedCAFile:  filepath.Join(cfg.TLSDir, "ca.crt"),
		ClientCertAuth: true,
	}

	ec.ClientTLSInfo = transport.TLSInfo{
		CertFile:       filepath.Join(cfg.TLSDir, "node.crt"),
		KeyFile:        filepath.Join(cfg.TLSDir, "node.key"),
		TrustedCAFile:  filepath.Join(cfg.TLSDir, "ca.crt"),
		ClientCertAuth: true,
	}

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

func StartMaster(dataDir, tlsDir string) (*InternalState, error) {
	return StartEmbedded(&EtcdConfig{
		DataDir:    dataDir,
		ClientPort: "2379",
		PeerPort:   "2380",
		Name:       "master",
		Cluster:    "master=https://localhost:2380",
		State:      "new",
		TLSDir:     tlsDir,
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

	_, err = client.RoleAdd(ctx, "root")
	if err != nil {
		return fmt.Errorf("create root role: %w", err)
	}

	_, err = client.UserAddWithOptions(ctx, "root", "", &clientv3.UserAddOptions{NoPassword: true})
	if err != nil {
		return fmt.Errorf("create root user: %w", err)
	}

	_, err = client.UserGrantRole(ctx, "root", "root")
	if err != nil {
		return fmt.Errorf("grant root role to root user: %w", err)
	}

	_, err = client.RoleAdd(ctx, "cluster-node")
	if err != nil {
		return fmt.Errorf("create cluster-node role: %w", err)
	}

	_, err = client.RoleGrantPermission(
		ctx, "cluster-node", "\x00", "\x00", clientv3.PermissionType(clientv3.PermReadWrite),
	)
	if err != nil {
		return fmt.Errorf("grant permissions to cluster-node: %w", err)
	}

	_, err = client.UserAddWithOptions(ctx, "filestore", "", &clientv3.UserAddOptions{NoPassword: true})
	if err != nil {
		return fmt.Errorf("create filestore user: %w", err)
	}

	_, err = client.UserGrantRole(ctx, "filestore", "cluster-node")
	if err != nil {
		return fmt.Errorf("grant cluster-node role to filestore: %w", err)
	}

	_, err = client.AuthEnable(ctx)
	if err != nil {
		return fmt.Errorf("enable auth: %w", err)
	}
	log.Println("Authentication enabled with mTLS!")

	return nil
}

func StartFileStore(dataDir, initialCluster, tlsDir, nodeName string) (*InternalState, error) {
	memberDir := filepath.Join(dataDir, "member")
	_, err := os.Stat(memberDir)
	hasData := err == nil

	stateStr := "existing"

	if hasData {
		log.Println("Found existing etcd data, booting directly...")
		initialCluster = ""
	} else {
		log.Println("No local data found, booting etcd from cluster join state...")
	}

	return StartEmbedded(&EtcdConfig{
		DataDir:    dataDir,
		ClientPort: "2479",
		PeerPort:   "2480",
		Name:       nodeName,
		Cluster:    initialCluster,
		State:      stateStr,
		TLSDir:     tlsDir,
	})
}
