package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/handlers"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/otel"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quic-go/quic-go/http3"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel/log/global"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type FilestoreConfig struct {
	NodeName      string `env:"FILE_STORE_NODE_NAME" type:"string" default:"filestore-1"`
	Port          int    `env:"FILE_STORE_API_PORT" type:"port" default:"443"`
	DrpcPort      int    `env:"FILE_STORE_DRPC_PORT" type:"port" default:"2749"`
	ListenAddr    string `env:"FILE_STORE_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	AdvertiseAddr string `env:"FILE_STORE_ADVERTISE_ADDR" type:"string" default:"localhost"`
	MasterAPIURL  string `env:"MASTER_API_URL" type:"url" default:"http://localhost:8443"`
	EtcdPort      int    `env:"FILE_STORE_ETCD_API_PORT" type:"port" default:"2479"`
	JoinToken     string `env:"JOIN_TOKEN" type:"secret" required:"true"`
	ClusterPubKey string `env:"CLUSTER_PUB_KEY" type:"string" required:"true"`
	TLSDir        string `env:"FILE_STORE_TLS_DIR" type:"string" default:"/data/filestore/tls/"`
	DataDir       string `env:"FILE_STORE_DATA_DIR" type:"string" default:"/data/filestore/etcd/"`
	OTELEndpoint  string `env:"OTEL_ENDPOINT" type:"string" default:""`
	AppName       string `env:"APP_NAME" type:"string" default:"gns3util-cluster"`
}

var logger *slog.Logger
var cfg FilestoreConfig

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
		otlpHandler := otel.NewOTLPHandler(otlpLogger)
		handler = otlpHandler
	}

	logger = slog.New(handler).With("prefix", cfg.AppName)

	masterGRPCURL := nwutils.ConvertMasterAPIURL(cfg.MasterAPIURL)

	certPath := filepath.Join(cfg.TLSDir, "node.crt")
	keyPath := filepath.Join(cfg.TLSDir, "node.key")
	caPath := filepath.Join(cfg.TLSDir, "ca.crt")

	initialCluster := ""

	if _, statErr := os.Stat(certPath); os.IsNotExist(statErr) {
		logger.Info("No local certificates found. Generating CSR and joining cluster...")

		csrPEM, genErr := certs.GenerateNodeKeyAndCSR(cfg.TLSDir, "filestore")
		if genErr != nil {
			logger.Error("Failed to generate CSR", "err", genErr)
			return
		}

		peerURL := fmt.Sprintf("https://%s:%d", cfg.AdvertiseAddr, 2480)
		reqBody, _ := json.Marshal(map[string]any{
			"name":      cfg.NodeName,
			"peer_urls": []string{peerURL},
			"csr_pem":   csrPEM,
		})

		joinURL := cfg.MasterAPIURL + "/cluster/join"
		req, reqErr := http.NewRequestWithContext(ctx, "POST", joinURL, bytes.NewReader(reqBody))
		if reqErr != nil {
			logger.Error("Failed to create join request", "err", reqErr)
			return
		}

		req.Header.Set("Authorization", "Bearer "+cfg.JoinToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402
			},
		}
		resp, respErr := client.Do(req)
		if respErr != nil {
			logger.Error("Failed to call master join API", "err", respErr)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			logger.Error("Master rejected join request", "err", string(body))
			return
		}

		var joinResp struct {
			MemberID uint64 `json:"member_id"`
			Cluster  string `json:"cluster"`
			CertPEM  []byte `json:"cert_pem"`
			CACert   []byte `json:"ca_cert"`
		}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&joinResp); decodeErr != nil {
			logger.Error("Failed to decode join response", "err", decodeErr)
			return
		}

		if writeCertErr := os.WriteFile(certPath, joinResp.CertPEM, 0o600); writeCertErr != nil {
			logger.Error("Failed to write node.crt", "err", writeCertErr)
			return
		}
		if writeCaErr := os.WriteFile(caPath, joinResp.CACert, 0o600); writeCaErr != nil {
			logger.Error("Failed to write ca.crt", "err", writeCaErr)
			return
		}

		initialCluster = joinResp.Cluster

		logger.Info("Successfully joined cluster! Assigned Member ID", "member_id", joinResp.MemberID)
	}
	etcdState, startEtcdErr := state.StartFileStore(cfg.DataDir, initialCluster, cfg.TLSDir, cfg.NodeName, handler, fmt.Sprintf("%s-etcd", cfg.AppName))
	if startEtcdErr != nil {
		logger.Error("Failed to join cluster", "err", startEtcdErr)
		return
	}

	store, newStateErr := state.NewStateManager(
		[]string{masterGRPCURL},
		[]string{"localhost:2379"},
		cfg.TLSDir,
	)
	if newStateErr != nil {
		logger.Error("Failed to connect to local etcd", "err", newStateErr)
		return
	}

	idMgr, idMgrErr := auth.NewIdentityManagerFromPubKey(cfg.ClusterPubKey)
	if idMgrErr != nil {
		logger.Error("Failed to create identity manager", "err", idMgrErr)
		return
	}

	cm := &certs.CertManager{}

	cert, loadCertErr := tls.LoadX509KeyPair(certPath, keyPath)
	if loadCertErr != nil {
		logger.Error("Failed to load TLS certs for HTTP server", "err", loadCertErr)
		return
	}
	cm.SetCertificate(&cert)

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
	setupRouter(r, idMgr, store, otlpEnabled)

	tlsConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
	}

	h3Server := &http3.Server{
		Addr:      fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.Port),
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	drpcServer := drpcserver.New(m)

	var lis net.ListenConfig
	drpcListener, err := lis.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.DrpcPort))
	if err != nil {
		logger.Error("Failed to listen for drpc", "err", err)
		return
	}

	logger.Info("Starting HTTP/3 server", "port", cfg.Port)
	logger.Info("Starting drpc server", "port", cfg.DrpcPort)

	errChan := make(chan error, 2)

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

func setupRouter(r chi.Router, idMgr *auth.IdentityManager, store *state.StateManager, otelEnabled bool) {
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
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(idMgr))

				r.With(middleware.RequireScope(store, "files:read")).
					Get("/files", handlers.UploadFileHandler)

				r.With(middleware.RequireScope(store, "files:write")).
					Post("/files", handlers.UploadFileHandler)
			})
		})

		r.Route("/v2", func(r chi.Router) {
			r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
				_, writeErr := w.Write([]byte("future proofing"))
				if writeErr != nil {
					logger.Error("Failed to write response", "err", writeErr)
					return
				}
			})
		})
	})
}
