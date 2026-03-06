package state

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/0xveya/gns3util/pkg/otel"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type InternalState struct {
	Server *embed.Etcd
}

type EtcdConfig struct {
	DataDir       string
	ClientPort    string
	PeerPort      string
	Name          string
	Cluster       string
	State         string
	TLSDir        string
	AdvertiseAddr string
}

func StartEmbedded(cfg *EtcdConfig, otlpHandler slog.Handler, otelName string) (*InternalState, error) {
	ec := embed.NewConfig()
	ec.Dir = cfg.DataDir
	ec.Name = cfg.Name

	advAddr := cfg.AdvertiseAddr
	if advAddr == "" {
		advAddr = "localhost"
	}

	etcdLogger := slog.New(otlpHandler).With("prefix", otelName)

	core := &otel.SlogCore{
		Handler: etcdLogger.Handler(),
		Level:   zap.NewAtomicLevelAt(zapcore.InfoLevel),
	}
	ec.ZapLoggerBuilder = embed.NewZapLoggerBuilder(zap.New(core))
	ec.LogLevel = "info"

	lpurl, _ := url.Parse("https://0.0.0.0:" + cfg.PeerPort)
	lcurl, _ := url.Parse("https://0.0.0.0:" + cfg.ClientPort)
	acurl, _ := url.Parse(fmt.Sprintf("https://%s:%s", advAddr, cfg.ClientPort))
	apurl, _ := url.Parse(fmt.Sprintf("https://%s:%s", advAddr, cfg.PeerPort))

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
		etcdLogger.Info("Etcd node ready",
			"name", cfg.Name,
			"port", cfg.ClientPort,
		)
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		return nil, fmt.Errorf("etcd startup timeout")
	}

	return &InternalState{Server: e}, nil
}

func StartMaster(dataDir, tlsDir string, otlpHandler slog.Handler, otelName, advertiseAddr string) (*InternalState, error) {
	if advertiseAddr == "" {
		advertiseAddr = "localhost"
	}

	masterClusterStr := fmt.Sprintf("master=https://%s:2380", advertiseAddr)

	return StartEmbedded(&EtcdConfig{
		DataDir:       dataDir,
		ClientPort:    "2379",
		PeerPort:      "2380",
		Name:          "master",
		Cluster:       masterClusterStr,
		State:         "new",
		TLSDir:        tlsDir,
		AdvertiseAddr: advertiseAddr,
	}, otlpHandler, otelName)
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

	return nil
}

func StartFileStore(dataDir, initialCluster, tlsDir, nodeName string, otlpHandler slog.Handler, otelName, advertiseAddr string) (*InternalState, error) {
	logger := slog.New(otlpHandler).With("prefix", nodeName+"-etcd")

	memberDir := filepath.Join(dataDir, "member")
	_, err := os.Stat(memberDir)
	hasData := err == nil

	stateStr := "existing"

	if hasData {
		logger.Info("Found existing etcd data, booting directly", "member_dir", memberDir)
		initialCluster = ""
	} else {
		logger.Info("No local data found, booting etcd from cluster join state", "initial_cluster", initialCluster)
	}

	return StartEmbedded(&EtcdConfig{
		DataDir:       dataDir,
		ClientPort:    "2479",
		PeerPort:      "2480",
		Name:          nodeName,
		Cluster:       initialCluster,
		State:         stateStr,
		TLSDir:        tlsDir,
		AdvertiseAddr: advertiseAddr,
	}, otlpHandler, otelName)
}
