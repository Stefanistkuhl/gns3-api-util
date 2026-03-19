package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	backgroundjobs "github.com/0xveya/gns3util/internal/file-store/backroundjobs"
	dbpkg "github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/fs"
	"github.com/0xveya/gns3util/internal/file-store/handlers"
	filerpc "github.com/0xveya/gns3util/internal/file-store/rpc"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/metrics"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/otel"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	"github.com/quic-go/quic-go/http3"
	"go.opentelemetry.io/otel/log/global"
	_ "modernc.org/sqlite"

	"github.com/0xveya/gns3util/pkg/web/middleware"
)

type FilestoreConfig struct {
	NodeName                   string        `env:"FILE_STORE_NODE_NAME" type:"string" default:"filestore-1"`
	Port                       int           `env:"FILE_STORE_API_PORT" type:"port" default:"443"`
	DrpcPort                   int           `env:"FILE_STORE_DRPC_PORT" type:"port" default:"2749"`
	ListenAddr                 string        `env:"FILE_STORE_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	AdvertiseAddr              string        `env:"FILE_STORE_ADVERTISE_ADDR" type:"string" default:"localhost"`
	MasterAPIURL               string        `env:"MASTER_API_URL" type:"url" default:"https://localhost:8443"`
	MasterDRPC                 string        `env:"MASTER_DRPC_ADDR" type:"string" default:"localhost:2748"`
	ClusterPubKey              string        `env:"CLUSTER_PUB_KEY" type:"string" required:"true"`
	TLSDir                     string        `env:"FILE_STORE_TLS_DIR" type:"string" default:"/data/filestore/tls/"`
	OTELEndpoint               string        `env:"OTEL_ENDPOINT" type:"string" default:""`
	AppName                    string        `env:"APP_NAME" type:"string" default:"gns3util-cluster"`
	JoinToken                  string        `env:"JOIN_TOKEN" type:"string" default:""`
	DBPath                     string        `env:"FILE_STORE_DB_PATH" type:"string" default:"/data/sqlite/file-store.db"`
	DataDir                    string        `env:"FILE_STORE_DATA_PATH" type:"string" default:"/data/storrage/"`
	VacuumInterval             time.Duration `env:"FILE_STORE_VACUUM_INTERVAL" type:"duration" default:"60m"`
	TombstoneRemoveInterval    time.Duration `env:"FILE_STORE_TOMBSTONE_EXPIRY" type:"duration" default:"1440m"`
	PendingFileRemoveInterval  time.Duration `env:"FILE_STORE_PENDING_EXPIRY" type:"duration" default:"360m"`
	NoWriteSinceUploadInterval time.Duration `env:"FILE_STORE_STALLED_UPLOAD_EXPIRY" type:"duration" default:"120m"`
	MetricsEnabled             bool          `env:"METRICS_ENABLED" type:"bool" default:"true"`
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

	metricsMgr := metrics.New(metrics.Config{
		Enabled: cfg.MetricsEnabled,
	})

	httpMetrics := metrics.NewHTTPMetrics()
	drpcMetrics := metrics.NewDRPCMetrics()
	storageMetrics := metrics.NewStorageMetrics()
	jobMetrics := metrics.NewJobMetrics()

	metricsMgr.Register(
		httpMetrics.RequestsTotal,
		httpMetrics.RequestDuration,
		httpMetrics.InFlightRequests,
		drpcMetrics.RequestsTotal,
		drpcMetrics.RequestLatency,
		drpcMetrics.ErrorsTotal,
		storageMetrics.FileWritesTotal,
		storageMetrics.FileWriteBytes,
		storageMetrics.FileWriteDuration,
		storageMetrics.FileReadDuration,
		storageMetrics.ActiveUploads,
		jobMetrics.RunsTotal,
		jobMetrics.RunDuration,
		jobMetrics.ItemsProcessed,
		jobMetrics.ErrorsTotal,
	)

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

	mkdirErr := os.MkdirAll(cfg.DataDir, 0o750)
	if mkdirErr != nil {
		logger.Error("Failed to create data directory", "err", mkdirErr)
		return
	}
	fsState, createDirsErr := fs.CreateDirStructure(cfg.DataDir)
	if createDirsErr != nil {
		logger.Error("Failed to create need directorys to store files", "err", createDirsErr)
		return
	}

	// TODO: make this configurable once syncing is implemented
	dbStore, err := dbpkg.NewStore(cfg.DBPath, true)
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

	var masterClient *filerpc.MasterSyncClient
	for i := range 5 {
		masterClient, err = filerpc.NewMasterSyncClient(ctx, cfg.MasterDRPC, cfg.TLSDir)
		if err == nil {
			break
		}
		logger.Warn("Waiting for Master dRPC...", "attempt", i+1, "err", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Error("Failed to create master client after retries", "err", err)
		return
	}
	defer masterClient.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		sig := <-sigChan
		logger.Info("Shutdown signal received", "signal", sig)
		cancel()
	}()

	vacuumJob := &backgroundjobs.DBVacuumJob{
		Store:                      dbStore,
		Interval:                   cfg.VacuumInterval,
		Logger:                     logger,
		TombstoneRemoveInterval:    cfg.TombstoneRemoveInterval,
		PendingFileRemoveInterval:  cfg.PendingFileRemoveInterval,
		NoWriteSinceUploadInterval: cfg.NoWriteSinceUploadInterval,
		Dirs:                       fsState,
	}

	registerJobsWithMaster(ctx, masterClient, vacuumJob)

	r := chi.NewRouter()
	setupRouter(r, idMgr, masterClient, dbStore, otlpEnabled, fsState, metricsMgr, vacuumJob)

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

	g, groupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("Starting HTTP/TLS server", "addr", tcpServer.Addr)
		if err := tcpServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("tcp server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("Starting HTTP/3 server", "addr", h3Server.Addr)
		if err := h3Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("h3 server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("Starting DB vaccum job")
		vacuumJob.Run(groupCtx)
		return nil
	})

	g.Go(func() error {
		<-groupCtx.Done()
		logger.Info("Shutting down filestore services")

		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer shutdownCancel()

		if err := tcpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("TCP server shutdown error", "err", err)
		}

		if h3Server != nil {
			if err := h3Server.Close(); err != nil {
				logger.Error("HTTP/3 server close error", "err", err)
			}
		}

		if masterClient != nil {
			if err := masterClient.Close(); err != nil {
				logger.Error("Master client close error", "err", err)
			}
		}

		if dbStore != nil && dbStore.DB != nil {
			if err := dbStore.DB.Close(); err != nil {
				logger.Error("DB close error", "err", err)
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("Execution stopped with error", "err", err)
	} else {
		logger.Info("Filestore exited cleanly")
	}
}

func setupRouter(
	r chi.Router,
	idMgr *auth.IdentityManager,
	masterClient *filerpc.MasterSyncClient,
	store *dbpkg.Store,
	otelEnabled bool,
	dirs fs.Dirs,
	metricsMgr *metrics.Manager,
	vacuumJob *backgroundjobs.DBVacuumJob,
) {
	middleware.SetupCommonMiddleware(r, otelEnabled, cfg.AppName, logger)

	r.NotFound(commonhandlers.Handle404)
	r.MethodNotAllowed(commonhandlers.Handle405)
	r.Get("/healthz", commonhandlers.HandleHealthz)

	jobRunners := map[string]handlers.JobRunnerFunc{
		"db-vacuum": func(ctx context.Context, invokedBy backgroundjobs.Invocator) (any, error) {
			return vacuumJob.ExecuteIteration(ctx, invokedBy)
		},
	}

	fileStoreHandlers := &handlers.FilestoreHandlers{
		Store:      store,
		Logger:     logger,
		Dirs:       &dirs,
		JobRunners: jobRunners,
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(idMgr))

		r.Route("/buckets", func(r chi.Router) {
			r.With(middleware.RequireScopeRemote(masterClient, "files:read")).
				Get("/", fileStoreHandlers.ListBuckets)
			r.With(middleware.RequireScopeRemote(masterClient, "files:write")).
				Post("/", fileStoreHandlers.CreateBucket)

			r.Route("/{bucket_id}", func(r chi.Router) {
				r.Use(middleware.BucketAccessMiddleware(store, logger))

				r.With(middleware.RequireScopeRemote(masterClient, "files:write")).
					Post("/files", fileStoreHandlers.HandleInitUpload)

				r.Get("/files", fileStoreHandlers.ListBucketFiles)

				r.Delete("/files", fileStoreHandlers.DeleteBucket)
			})
		})

		r.Route("/files", func(r chi.Router) {
			r.With(middleware.RequireScopeRemote(masterClient, "files:write")).
				Post("/", fileStoreHandlers.HandleInitUpload)

			r.Route("/{file_uuid}", func(r chi.Router) {
				r.Use(middleware.FileAccessMiddleware(store, logger))

				r.With(middleware.RequireScopeRemote(masterClient, "files:read")).
					Get("/", fileStoreHandlers.DownloadFileHandler)

				r.Get("/status", fileStoreHandlers.GetUploadStatus)

				r.With(middleware.RequireScopeRemote(masterClient, "files:write")).
					Put("/content", fileStoreHandlers.HandleStreamUpload)

				r.With(middleware.RequireScopeRemote(masterClient, "files:write")).
					Delete("/", fileStoreHandlers.DeleteFile)

				r.With(middleware.RequireScopeRemote(masterClient, "files:write")).
					Post("/public-token", fileStoreHandlers.GeneratePublicToken)
			})
		})

		r.Route("/public", func(r chi.Router) {
			r.Get("/files/{bucket_id}/{token}", fileStoreHandlers.PublicFileHandler)
		})

		r.Route("/jobs", func(r chi.Router) {
			r.With(middleware.RequireScopeRemote(masterClient, "execute:jobs")).
				Post("/{job_name}/run", fileStoreHandlers.RunJob)
		})
	})
	if metricsMgr != nil && metricsMgr.Enabled && metricsMgr.Registry != nil {
		r.With(
			middleware.AuthMiddleware(idMgr),
			middleware.RequireScopeRemote(masterClient, "metrics:read"),
		).Get("/metrics", promhttp.HandlerFor(metricsMgr.Registry, promhttp.HandlerOpts{}).ServeHTTP)
	}
}

func registerJobsWithMaster(ctx context.Context, client *filerpc.MasterSyncClient, vacuumJob *backgroundjobs.DBVacuumJob) {
	if err := client.RegisterJob(
		ctx,
		"db-vacuum",
		cfg.NodeName,
		vacuumJob.Interval.String(),
		"Cleans up expired files, stalled uploads, orphaned blobs, and temporary files",
	); err != nil {
		logger.Warn("Failed to register db-vacuum job with master", "err", err)
	} else {
		logger.Info("Registered db-vacuum job with master")
	}
}

func bootstrapCertificates(
	masterURL string,
	csrPEM []byte,
	certPath, caPath, token, masterCACert string,
	ctx context.Context,
) error {
	advertiseIP := nwutils.GetFirstNonLoopbackIP()

	if cfg.Port < 0 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", cfg.Port)
	}
	port := uint32(cfg.Port)
	joinReq := models.JoinFilestoreRequest{
		CSRPEM:  string(csrPEM),
		ID:      cfg.NodeName,
		IP:      advertiseIP,
		APIPort: port,
	}

	payload, err := json.Marshal(joinReq)
	if err != nil {
		return fmt.Errorf("failed to marshal join request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/api/v1/cluster/join/filestore", masterURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

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
		Timeout: 15 * time.Second,
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("master rejected join (status %d): %s", resp.StatusCode, string(body))
	}

	var result models.JoinFilestoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if err := os.WriteFile(certPath, []byte(result.NodeCert), 0o600); err != nil {
		return fmt.Errorf("failed to save cert: %w", err)
	}
	if err := os.WriteFile(caPath, []byte(result.CACert), 0o600); err != nil {
		return fmt.Errorf("failed to save CA: %w", err)
	}

	return nil
}
