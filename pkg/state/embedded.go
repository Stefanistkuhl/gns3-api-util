package state

import (
	"context"
	"fmt"
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
	Server     *embed.Etcd
	Client     *clientv3.Client
	SocketPath string
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
	UseUnixSocket bool
	SocketDir     string
}

func StartEmbedded(
	cfg *EtcdConfig,
	otlpHandler slog.Handler,
	otelName string,
) (*InternalState, error) {
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

	lpURL, _ := url.Parse("https://0.0.0.0:" + cfg.PeerPort)
	apURL, _ := url.Parse(fmt.Sprintf("https://%s:%s", advAddr, cfg.PeerPort))
	ec.ListenPeerUrls = []url.URL{*lpURL}
	ec.AdvertisePeerUrls = []url.URL{*apURL}

	var socketPath string

	if cfg.UseUnixSocket {
		if cfg.SocketDir == "" {
			cfg.SocketDir = filepath.Join(cfg.DataDir, "run")
		}

		if err := os.MkdirAll(cfg.SocketDir, 0o750); err != nil {
			return nil, fmt.Errorf("failed to create socket directory: %w", err)
		}

		socketPath = filepath.Join(cfg.SocketDir, "etcd.sock")

		if err := os.RemoveAll(socketPath); err != nil {
			return nil, fmt.Errorf("failed to clean old socket: %w", err)
		}

		unixURL, err := url.Parse("unix://" + socketPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse unix socket url: %w", err)
		}
		ec.ListenClientUrls = []url.URL{*unixURL}

		// etcd still wants advertise-client-urls to look like host:port.
		// This is only to satisfy config validation; the filestore app will
		// use the unix socket for local access.
		acURL, err := url.Parse(fmt.Sprintf("https://127.0.0.1:%s", cfg.ClientPort))
		if err != nil {
			return nil, fmt.Errorf("failed to parse advertise client url: %w", err)
		}
		ec.AdvertiseClientUrls = []url.URL{*acURL}
	} else {
		lcURL, _ := url.Parse("https://0.0.0.0:" + cfg.ClientPort)
		acURL, _ := url.Parse(
			fmt.Sprintf("https://%s:%s", advAddr, cfg.ClientPort))
		ec.ListenClientUrls = []url.URL{*lcURL}
		ec.AdvertiseClientUrls = []url.URL{*acURL}
	}

	ec.PeerTLSInfo = transport.TLSInfo{
		CertFile:       filepath.Join(cfg.TLSDir, "node.crt"),
		KeyFile:        filepath.Join(cfg.TLSDir, "node.key"),
		TrustedCAFile:  filepath.Join(cfg.TLSDir, "ca.crt"),
		ClientCertAuth: true,
	}

	if !cfg.UseUnixSocket {
		ec.ClientTLSInfo = transport.TLSInfo{
			CertFile:       filepath.Join(cfg.TLSDir, "node.crt"),
			KeyFile:        filepath.Join(cfg.TLSDir, "node.key"),
			TrustedCAFile:  filepath.Join(cfg.TLSDir, "ca.crt"),
			ClientCertAuth: true,
		}
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
		etcdLogger.Info(
			"Etcd node ready",
			"name", cfg.Name,
			"socket", socketPath,
			"peer_port", cfg.PeerPort,
			"client_port", cfg.ClientPort,
		)
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		return nil, fmt.Errorf("etcd startup timeout")
	}

	var localCli *clientv3.Client
	if cfg.UseUnixSocket {
		localCli, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{"unix://" + socketPath},
			DialTimeout: 2 * time.Second,
		})
		if err != nil {
			e.Server.Stop()
			return nil, fmt.Errorf("failed to create local client: %w", err)
		}
	}

	return &InternalState{
		Server:     e,
		Client:     localCli,
		SocketPath: socketPath,
	}, nil
}

func StartMaster(
	dataDir,
	tlsDir string,
	otlpHandler slog.Handler,
	otelName,
	advertiseAddr string,
) (*InternalState, error) {
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
		UseUnixSocket: false,
	}, otlpHandler, otelName)
}

func StartFileStore(
	dataDir,
	initialCluster,
	tlsDir,
	nodeName string,
	otlpHandler slog.Handler,
	otelName,
	advertiseAddr string,
) (*InternalState, error) {
	logger := slog.New(otlpHandler).With("prefix", nodeName+"-etcd")

	memberDir := filepath.Join(dataDir, "member")
	_, err := os.Stat(memberDir)
	hasData := err == nil

	stateStr := "existing"

	if hasData {
		logger.Info(
			"Found existing etcd data, booting directly",
			"member_dir", memberDir,
		)
		initialCluster = ""
	} else {
		logger.Info(
			"No local data found, booting etcd from cluster join state",
			"initial_cluster", initialCluster,
		)
	}

	socketDir := filepath.Join(dataDir, "run")

	return StartEmbedded(&EtcdConfig{
		DataDir:       dataDir,
		ClientPort:    "2479",
		PeerPort:      "2480",
		Name:          nodeName,
		Cluster:       initialCluster,
		State:         stateStr,
		TLSDir:        tlsDir,
		AdvertiseAddr: advertiseAddr,
		UseUnixSocket: true,
		SocketDir:     socketDir,
	}, otlpHandler, otelName)
}

func BootstrapMasterAuth(ctx context.Context, client *clientv3.Client) error {
	authResp, err := client.AuthStatus(ctx)
	if err != nil {
		return fmt.Errorf("check auth status: %w", err)
	}
	if authResp.Enabled {
		return nil
	}

	_, err = client.RoleAdd(ctx, "root")
	if err != nil {
		return fmt.Errorf("create root role: %w", err)
	}

	_, err = client.UserAddWithOptions(
		ctx,
		"root",
		"",
		&clientv3.UserAddOptions{NoPassword: true},
	)
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
		ctx,
		"cluster-node",
		"\x00",
		"\x00",
		clientv3.PermissionType(clientv3.PermReadWrite),
	)
	if err != nil {
		return fmt.Errorf("grant permissions to cluster-node: %w", err)
	}

	_, err = client.UserAddWithOptions(
		ctx,
		"filestore",
		"",
		&clientv3.UserAddOptions{NoPassword: true},
	)
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

func (s *InternalState) Close() error {
	if s == nil {
		return nil
	}

	if s.Client != nil {
		_ = s.Client.Close()
	}

	if s.Server != nil {
		s.Server.Server.Stop()
	}

	if s.SocketPath != "" {
		_ = os.RemoveAll(s.SocketPath)
	}

	return nil
}
