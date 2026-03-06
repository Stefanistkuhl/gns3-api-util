package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	dbpkg "github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/handlers"
	filerpc "github.com/0xveya/gns3util/internal/file-store/rpc"
	syncsvc "github.com/0xveya/gns3util/internal/file-store/sync"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/otel"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"

	"github.com/0xveya/gns3util/pkg/web/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quic-go/quic-go/http3"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel/log/global"
	_ "modernc.org/sqlite"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type FilestoreConfig struct {
	NodeName      string `env:"FILE_STORE_NODE_NAME" type:"string" default:"filestore-1"`
	Port          int    `env:"FILE_STORE_API_PORT" type:"port" default:"443"`
	DrpcPort      int    `env:"FILE_STORE_DRPC_PORT" type:"port" default:"2749"`
	ListenAddr    string `env:"FILE_STORE_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	AdvertiseAddr string `env:"FILE_STORE_ADVERTISE_ADDR" type:"string" default:"localhost"`
	MasterAPIURL  string `env:"MASTER_API_URL" type:"url" default:"https://localhost:8443"`
	MasterDRPC    string `env:"MASTER_DRPC_ADDR" type:"string" default:"localhost:2748"`
	ClusterPubKey string `env:"CLUSTER_PUB_KEY" type:"string" required:"true"`
	TLSDir        string `env:"FILE_STORE_TLS_DIR" type:"string" default:"/data/filestore/tls/"`
	OTELEndpoint  string `env:"OTEL_ENDPOINT" type:"string" default:""`
	AppName       string `env:"APP_NAME" type:"string" default:"gns3util-cluster"`
	SyncInterval  int    `env:"FILE_STORE_SYNC_INTERVAL_SEC" type:"int" default:"15"`
	JoinToken     string `env:"JOIN_TOKEN" type:"string" default:""`
	DB_PATH       string `env:"FILE_STORE_DB_PATH" type:"string" default:"/data/sqlite/file-store.db"`
}

var (
	logger *slog.Logger
	cfg    FilestoreConfig
)

func init() {
	if err := env.LoadConfig(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg.AppName += "-filestore"

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
		handler = otel.NewOTLPHandler(otlpLogger)
	}

	logger = slog.New(handler).With("prefix", cfg.AppName)

	certPath := filepath.Join(cfg.TLSDir, "node.crt")
	keyPath := filepath.Join(cfg.TLSDir, "node.key")
	caPath := filepath.Join(cfg.TLSDir, "ca.crt")

	if _, statErr := os.Stat(certPath); statErr != nil {
		logger.Info("Certificates not found locally. Bootstrapping from Master...")

		csrPEM, genErr := certs.GenerateNodeKeyAndCSR(cfg.TLSDir, cfg.NodeName, []string{"localhost", cfg.AdvertiseAddr})
		if genErr != nil {
			logger.Error("Failed to generate CSR", "err", genErr)
			return
		}

		bootstrapErr := bootstrapCertificates(cfg.MasterAPIURL, csrPEM, certPath, caPath, cfg.JoinToken, "", ctx)
		if bootstrapErr != nil {
			logger.Error("Failed to bootstrap certificates from Master", "err", bootstrapErr)
			return
		}
		logger.Info("Successfully bootstrapped certificates")
	}

	dbStore, err := dbpkg.NewStore(cfg.DB_PATH)
	if err != nil {
		logger.Error("Failed to open sqlite store", "err", err)
		return
	}
	defer dbStore.DB.Close()

	idMgr, err := auth.NewIdentityManagerFromPubKey(cfg.ClusterPubKey)
	if err != nil {
		logger.Error("Failed to create identity manager", "err", err)
		return
	}

	cm := &certs.CertManager{}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		logger.Error("Failed to load TLS certs for HTTP server", "err", err)
		return
	}
	cm.SetCertificate(&cert)

	syncClient, err := filerpc.NewMasterSyncClient(ctx, cfg.MasterDRPC, cfg.TLSDir)
	if err != nil {
		logger.Error("Failed to create master sync client", "err", err)
		return
	}
	defer syncClient.Close()

	syncService := syncsvc.NewService(dbStore, syncClient, cfg.NodeName)
	go func() {
		if svcErr := syncService.Run(
			ctx,
			time.Duration(cfg.SyncInterval)*time.Second,
		); svcErr != nil {
			logger.Error("Sync service stopped", "err", svcErr)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")
		cancel()
		os.Exit(0)
	}()

	m := drpcmux.New()
	r := chi.NewRouter()
	setupRouter(r, idMgr, dbStore, otlpEnabled)

	tlsConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
	}

	apiAddr := fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.Port)

	tcpServer := &http.Server{
		Addr:              apiAddr,
		Handler:           r,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
	}

	h3Server := &http3.Server{
		Addr:      apiAddr,
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	drpcServer := drpcserver.New(m)

	var lis net.ListenConfig
	drpcListener, err := lis.Listen(
		ctx,
		"tcp",
		fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.DrpcPort),
	)
	if err != nil {
		logger.Error("Failed to listen for drpc", "err", err)
		return
	}

	logger.Info("Starting HTTP/3 server", "port", cfg.Port)
	logger.Info("Starting drpc server", "port", cfg.DrpcPort)

	errChan := make(chan error, 3)

	go func() {
		errChan <- tcpServer.ListenAndServeTLS("", "")
	}()

	go func() {
		errChan <- h3Server.ListenAndServe()
	}()

	go func() {
		errChan <- drpcServer.Serve(ctx, drpcListener)
	}()

	if err := <-errChan; err != nil {
		logger.Error("Server error", "err", err)
	}
}

func setupRouter(
	r chi.Router,
	idMgr *auth.IdentityManager,
	store *dbpkg.Store,
	otelEnabled bool,
) {
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

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(idMgr))

			r.With(middleware.RequireScope(store, "files:read")).
				Get("/files", handlers.DownloadFileHandler)

			r.With(middleware.RequireScope(store, "files:write")).
				Post("/files", handlers.UploadFileHandler)
		})
	})
}

func bootstrapCertificates(masterURL string, csrPEM []byte, certPath, caPath, token, masterCACert string, ctx context.Context) error {
	payload, _ := json.Marshal(map[string]string{"csr": string(csrPEM)})

	reqURL := masterURL + "/cluster/join/filestore"
	req, _ := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	tlsConfig := &tls.Config{}

	if masterCACert != "" {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM([]byte(masterCACert)) {
			return fmt.Errorf("failed to parse provided MASTER_CA_CERT")
		}
		tlsConfig.RootCAs = caPool
	} else {
		tlsConfig.InsecureSkipVerify = true
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("master rejected join request: status %d", resp.StatusCode)
	}

	var result struct {
		NodeCert string `json:"node_cert"`
		CACert   string `json:"ca_cert"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if err := os.WriteFile(certPath, []byte(result.NodeCert), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(caPath, []byte(result.CACert), 0o600); err != nil {
		return err
	}

	return nil
}
