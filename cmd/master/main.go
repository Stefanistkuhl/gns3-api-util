package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/0xveya/gns3util/internal/master/handlers"
	"github.com/0xveya/gns3util/internal/master/rpc"
	clusteraccess "github.com/0xveya/gns3util/internal/shared/cluster_access"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/otel"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
	"github.com/go-chi/chi/v5"

	pb "github.com/0xveya/gns3util/internal/shared/pb/master"
	"github.com/grandcat/zeroconf"
	"go.opentelemetry.io/otel/log/global"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type MasterConfig struct {
	TLSDir         string `env:"MASTER_TLS_DIR" type:"string" default:"/data/master/tls/"`
	APIPort        int    `env:"MASTER_API_PORT" type:"port" default:"8443"`
	DrpcPort       int    `env:"MASTER_STORE_DRPC_PORT" type:"port" default:"2748"`
	APIListenAddr  string `env:"MASTER_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	EtcdPort       int    `env:"MASTER_ETCD_API_PORT" type:"port" default:"2379"`
	EtcdListenAddr string `env:"MASTER_ETCD_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	PrivKeyStr     string `env:"CLUSTER_PRIV_KEY" type:"string"`
	TLSSubject     string `env:"MASTER_TLS_SUBJ" type:"string" default:"/CN=root"`
	DataDir        string `env:"MASTER_DATA_DIR" type:"string" default:"/data/master/etcd/"`
	EnableMDNS     bool   `env:"MASTER_ENABLE_MDNS" type:"bool" default:"true"`
	OTELEndpoint   string `env:"OTEL_ENDPOINT" type:"string" default:""`
	AppName        string `env:"APP_NAME" type:"string" default:"gns3util-cluster"`
	AdvertiseAddr  string `env:"MASTER_ADVERTISE_ADDR" type:"string" default:"localhost"`
}

var (
	logger *slog.Logger
	cfg    MasterConfig
)

func init() {
	if err := env.LoadConfig(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg.AppName += "-master"

	shutdown, err := otel.Init(context.Background(), cfg.AppName, cfg.OTELEndpoint)
	if err != nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("prefix", cfg.AppName)
		logger.Error("Failed to initialize telemetry", "err", err)
	} else {
		defer func() { _ = shutdown(context.Background()) }()
	}

	otlpEnabled := cfg.OTELEndpoint != ""

	stdoutHandler := slog.NewJSONHandler(os.Stdout, nil)
	var handler slog.Handler = stdoutHandler

	if otlpEnabled {
		otlpLogger := global.GetLoggerProvider().Logger(cfg.AppName)
		otlpHandler := otel.NewOTLPHandler(otlpLogger)
		handler = otlpHandler
	}

	logger = slog.New(handler).With("prefix", cfg.AppName)

	certPath := filepath.Join(cfg.TLSDir, "node.crt")
	keyPath := filepath.Join(cfg.TLSDir, "node.key")
	caPath := filepath.Join(cfg.TLSDir, "ca.crt")

	cm := &certs.CertManager{}

	if cfg.PrivKeyStr != "" {
		logger.Info("Loaded private key from environment")
	} else {
		logger.Warn("No private key in environment, generating a temporary one!")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err == nil {
		cm.SetCertificate(&cert)
	} else {
		logger.Info("No local CA certificate found, generating self-signed Ed25519 CA...")

		certPEM, keyPEM, genCertErr := generateSelfSignedCert(cfg.TLSSubject)
		if genCertErr != nil {
			logger.Error("Failed to generate cert", "err", genCertErr)
			return
		}

		if mkdirErr := os.MkdirAll(cfg.TLSDir, 0o700); mkdirErr != nil {
			logger.Error("Failed to create tls dir", "err", mkdirErr)
			return
		}
		certErr := os.WriteFile(certPath, certPEM, 0o600)
		if certErr != nil {
			logger.Error("Failed to write node.crt", "err", certErr)
			return
		}
		keyErr := os.WriteFile(keyPath, keyPEM, 0o600)
		if keyErr != nil {
			logger.Error("Failed to write node.key", "err", keyErr)
			return
		}
		caErr := os.WriteFile(caPath, certPEM, 0o600)
		if caErr != nil {
			logger.Error("Failed to write ca.crt", "err", caErr)
			return
		}

		cert, err = tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			logger.Error("Failed to load generated cert", "err", err)
			return
		}
		cm.SetCertificate(&cert)
		logger.Info("Generated self-signed certificate with subject", "subject", cfg.TLSSubject)

		caX509, parseCertErr := x509.ParseCertificate(cert.Certificate[0])
		if parseCertErr != nil {
			logger.Error("failed to parse CA", "err", parseCertErr)
			return
		}
		caPrivKey, ok := cert.PrivateKey.(ed25519.PrivateKey)
		if !ok {
			logger.Error("failed to parse CA", "err", parseCertErr)
			return
		}
		adminCertPEM, adminKeyPEM, genAdminCertErr := generateAdminCert(caX509, caPrivKey)
		if genAdminCertErr != nil {
			logger.Error("Failed to generate admin cert", "err", genAdminCertErr)
			return
		}
		accessConfig := clusteraccess.CreateClusterAcessConfig(
			fmt.Sprintf("https://%s:%d", nwutils.GetFirstNonLoopbackIP(), cfg.APIPort),
			certPEM,
			adminCertPEM,
			adminKeyPEM,
		)
		writeErr := accessConfig.WriteAccessConfig(filepath.Join(cfg.TLSDir, "cluster_access.toml"))
		if writeErr != nil {
			logger.Error("Failed to write cluster_access.toml", "err", writeErr)
			return
		}
	}

	etcdState, startEtcdErr := state.StartMaster(cfg.DataDir, cfg.TLSDir, handler, fmt.Sprintf("%s-etcd", cfg.AppName), cfg.AdvertiseAddr)
	if startEtcdErr != nil {
		logger.Error("Failed to start etcd", "err", startEtcdErr)
		return
	}

	store, storeErr := state.NewMasterStateManager(
		[]string{fmt.Sprintf("https://localhost:%d", cfg.EtcdPort)},
		cfg.TLSDir,
	)
	if storeErr != nil {
		logger.Error("Failed to connect to etcd", "err", storeErr)
		return
	}
	defer store.Close()

	if bootstrapErr := state.BootstrapMasterAuth(ctx, store.MasterClient); bootstrapErr != nil {
		logger.Error("Failed to bootstrap etcd auth", "bootstrapErr", bootstrapErr)
		return
	}

	var idMgr *auth.IdentityManager

	if cfg.PrivKeyStr == "" {
		pubKey, privKey, genKeyErr := auth.GenerateKeyPair()
		if genKeyErr != nil {
			logger.Error("Failed to generate keys", "err", genKeyErr)
			return
		}
		logger.Info("Generated new key pair")
		logger.Info("CLUSTER_PUB_KEY", "key", auth.EncodePublicKey(pubKey))
		logger.Info("CLUSTER_PRIV_KEY", "key", auth.EncodePrivateKey(privKey))

		var idMgrErr error
		idMgr, idMgrErr = auth.NewIdentityManager(privKey)
		if idMgrErr != nil {
			logger.Error("Failed to create identity manager", "err", idMgrErr)
			return
		}
	} else {
		privKey, decodePrivKeyErr := auth.DecodePrivateKey(cfg.PrivKeyStr)
		if decodePrivKeyErr != nil {
			logger.Error("Failed to decode private key", "err", decodePrivKeyErr)
		}
		var idMgrErr error
		idMgr, idMgrErr = auth.NewIdentityManager(privKey)
		if idMgrErr != nil {
			logger.Error("Failed to create identity manager", "err", idMgrErr)
		}
	}

	master := &handlers.Master{
		IDMgr:  idMgr,
		Store:  store,
		TLSDir: cfg.TLSDir,
		Logger: logger,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")
		cancel()
		etcdState.Server.Close()
		closeErr := store.Close()
		if closeErr != nil {
			logger.Error("Failed to close state manager", "err", closeErr)
		}
		os.Exit(0)
	}()

	m := drpcmux.New()
	syncSvc := rpc.NewSyncService(store)
	if rpcErr := pb.DRPCRegisterMasterSyncService(m, syncSvc); rpcErr != nil {
		logger.Error("Failed to register sync service", "err", rpcErr)
		return
	}
	r := chi.NewRouter()
	setupRouter(r, master, otlpEnabled)

	tlsConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.APIPort),
		Handler:           r,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
	}

	drpcServer := drpcserver.New(m)

	var lis net.ListenConfig
	drpcListener, err := lis.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cfg.APIListenAddr, cfg.DrpcPort))
	if err != nil {
		logger.Error("Failed to listen for drpc", "err", err)
		return
	}

	caPath = filepath.Join(cfg.TLSDir, "ca.crt") //#nosec G304
	var caData []byte
	var readErr error
	caData, readErr = os.ReadFile(caPath) //#nosec G304
	if readErr != nil {
		logger.Error("Failed to read CA for dRPC TLS", "err", readErr)
		return
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caData) {
		logger.Error("Failed to append CA cert for dRPC TLS")
		return
	}

	drpcTLSConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
		ClientAuth:     tls.RequireAndVerifyClientCert,
		ClientCAs:      caPool,
		MinVersion:     tls.VersionTLS13,
	}

	tlsDRPCListener := tls.NewListener(drpcListener, drpcTLSConfig)
	host, hostNameErr := os.Hostname()
	if hostNameErr != nil {
		logger.Error("Failed to get hostname", "err", hostNameErr)
		return
	}
	if cfg.EnableMDNS {
		mdnsServer, err := zeroconf.Register(
			host,
			"_gns3util_master_api._tcp",
			"local.",
			cfg.APIPort,
			[]string{"info=gns3util master api"},
			nwutils.GetActiveMulticastInterfaces(),
		)
		if err != nil {
			logger.Error("Failed to start mdns server", "err", err)
			return
		}
		defer mdnsServer.Shutdown()

		logger.Info("Registering mDNS service",
			"name", host,
			"service", "_gns3util_master_api._tcp",
			"port", cfg.APIPort,
		)
		logger.Info("Starting mdns discovery server")
	}

	logger.Info("Starting Master Node", "port", cfg.APIPort)
	logger.Info("Starting drpc server", "port", cfg.DrpcPort)

	errChan := make(chan error, 2)

	go func() {
		errChan <- server.ListenAndServeTLS("", "")
	}()

	go func() {
		errChan <- drpcServer.Serve(ctx, tlsDRPCListener)
	}()

	for range 2 {
		if err := <-errChan; err != nil {
			logger.Error("Server error", "err", err)
		}
	}
}

func setupRouter(r chi.Router, master *handlers.Master, otelEnabled bool) {
	middleware.SetupCommonMiddleware(r, otelEnabled, cfg.AppName, logger)

	r.Get("/healthz", commonhandlers.HandleHealthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/token", master.HandleCreateToken)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(master.IDMgr))

				r.Post("/grant", master.HandleGrantAccess)
				r.Post("/revoke", master.HandleRevokeAccess)
				r.Get("/status", master.HandleAuthStatus)
			})
		})

		r.Route("/cluster", func(r chi.Router) {
			r.Post("/join", master.HandleJoinCluster)
			r.Post("/join/filestore", master.HandleJoinFilestore)
		})
	})
}

func generateSelfSignedCert(subject string) (certPEM, keyPEM []byte, err error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: strings.TrimPrefix(subject, "/CN=")},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{"localhost"},
		IPAddresses:           nwutils.GetLocalIPs(),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, pubKey, privKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	return certPEM, keyPEM, nil
}

func generateAdminCert(caCert *x509.Certificate, caPrivKey any) (certPEM, keyPEM []byte, err error) {
	adminPubKey, adminPrivKey, genAdminKeyErr := ed25519.GenerateKey(rand.Reader)
	if genAdminKeyErr != nil {
		return nil, nil, genAdminKeyErr
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "root",
			Organization: []string{"system:masters"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, adminPubKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyBytes, err := x509.MarshalPKCS8PrivateKey(adminPrivKey)
	if err != nil {
		return nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	return certPEM, keyPEM, nil
}
