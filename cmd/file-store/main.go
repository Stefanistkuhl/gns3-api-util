package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xveya/gns3util/internal/file-store/handlers"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/utils/nwutils"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"
	"github.com/0xveya/gns3util/pkg/web/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/quic-go/quic-go/http3"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type FilestoreConfig struct {
	Port          int    `env:"FILE_STORE_API_PORT" type:"port" default:"443"`
	DrpcPort      int    `env:"FILE_STORE_DRPC_PORT" type:"port" default:"2749"`
	ListenAddr    string `env:"FILE_STORE_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	MasterAPIURL  string `env:"MASTER_API_URL" type:"url" default:"http://localhost:8443"`
	EtcdPort      int    `env:"FILE_STORE_ETCD_API_PORT" type:"port" default:"2479"`
	JoinToken     string `env:"JOIN_TOKEN" type:"secret" required:"true"`
	ClusterPubKey string `env:"CLUSTER_PUB_KEY" type:"string" required:"true"`
}

var cfg FilestoreConfig

func init() {
	if err := env.LoadConfig(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	masterGRPCURL := nwutils.ConvertMasterAPIURL(cfg.MasterAPIURL)

	etcdState, err := state.StartFileStore(ctx, "./filestore.etcd", cfg.MasterAPIURL)
	if err != nil {
		log.Fatalf("Failed to join cluster: %v", err)
	}

	store, err := state.NewStateManager(
		[]string{masterGRPCURL},
		[]string{"localhost:2379"},
		"joiner", cfg.JoinToken,
	)
	if err != nil {
		log.Fatalf("Failed to connect to local etcd: %v", err)
	}

	idMgr, err := auth.NewIdentityManagerFromPubKey(cfg.ClusterPubKey)
	if err != nil {
		log.Fatalf("Failed to create identity manager: %v", err)
	}

	cm := &certs.CertManager{}

	for i := range 60 {
		err = cm.LoadFromEtcd(ctx, store)
		if err == nil {
			break
		}
		log.Printf("Waiting for certificate... (%d/60)", i+1)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	notify := store.WatchCertificate(ctx)
	cm.WatchEtcd(ctx, store, notify)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		etcdState.Server.Close()
		store.Close()
		os.Exit(0)
	}()

	m := drpcmux.New()
	r := chi.NewRouter()
	setupRouter(r, idMgr, store)

	tlsConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
	}

	h3Server := &http3.Server{
		Addr:      fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.Port),
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	drpcServer := drpcserver.New(m)

	drpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.DrpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for drpc: %v", err)
	}

	log.Printf("Starting HTTP/3 server on :%d", cfg.Port)
	log.Printf("Starting drpc server on :%d", cfg.DrpcPort)

	errChan := make(chan error, 2)

	go func() {
		errChan <- h3Server.ListenAndServe()
	}()

	go func() {
		errChan <- drpcServer.Serve(ctx, drpcListener)
	}()

	if err := <-errChan; err != nil {
		log.Printf("Server error: %v", err)
	}
}

func setupRouter(r chi.Router, idMgr *auth.IdentityManager, store *state.StateManager) {
	r.Use(chimiddleware.Logger)
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
				w.Write([]byte("future proofing"))
			})
		})
	})
}
