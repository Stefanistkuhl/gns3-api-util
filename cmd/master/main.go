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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/0xveya/gns3util/internal/master/handlers"
	clusteraccess "github.com/0xveya/gns3util/internal/shared/cluster_access"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/otel"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/grandcat/zeroconf"
	"github.com/riandyrn/otelchi"
	clientv3 "go.etcd.io/etcd/client/v3"
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

	store, err := state.NewStateManager(
		[]string{"localhost:" + strconv.Itoa(cfg.EtcdPort)},
		nil,
		cfg.TLSDir,
	)
	if err != nil {
		logger.Error("Failed to connect to etcd", "err", err)
		return
	}

	if bootstrapErr := bootstrapIfNeeded(ctx, store.MasterClient); bootstrapErr != nil {
		logger.Error("Failed to bootstrap auth", "err", bootstrapErr)
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
		errChan <- drpcServer.Serve(ctx, drpcListener)
	}()

	for range 2 {
		if err := <-errChan; err != nil {
			logger.Error("Server error", "err", err)
		}
	}
}

func setupRouter(r chi.Router, master *handlers.Master, otelEnabled bool) {
	if !otelEnabled {
		r.Use(chimiddleware.Logger)
	} else {
		r.Use(otelchi.Middleware(fmt.Sprintf("%s-api", cfg.AppName)))

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ww := chimiddleware.NewWrapResponseWriter(w, req.ProtoMajor)
				next.ServeHTTP(ww, req)

				logger.InfoContext(req.Context(), "HTTP Request",
					"method", req.Method,
					"path", req.URL.Path,
					"status", ww.Status(),
				)
			})
		})
	}
	r.Use(chimiddleware.Recoverer)

	r.Get("/healthz", commonhandlers.HandleHealthz)

	r.Post("/auth/token", master.HandleCreateToken)
	r.Post("/auth/grant", master.HandleGrantAccess)
	r.Post("/auth/revoke", master.HandleRevokeAccess)
	r.Post("/cluster/join", master.HandleJoinCluster)
}

func bootstrapIfNeeded(ctx context.Context, cli *clientv3.Client) error {
	authResp, err := cli.AuthStatus(ctx)
	if err != nil {
		return err
	}

	if authResp.Enabled {
		logger.Info("Auth already enabled, skipping bootstrap")
		return nil
	}

	logger.Info("Bootstrapping RBAC...")
	return state.BootstrapMasterAuth(ctx, cli)
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
			CommonName:   "admin",
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
