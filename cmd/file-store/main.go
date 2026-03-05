package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	certPath := filepath.Join(cfg.TLSDir, "node.crt")
	keyPath := filepath.Join(cfg.TLSDir, "node.key")
	caPath := filepath.Join(cfg.TLSDir, "ca.crt")

	initialCluster := ""

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Println("No local certificates found. Generating CSR and joining cluster...")

		csrPEM, err := certs.GenerateNodeKeyAndCSR(cfg.TLSDir, "filestore")
		if err != nil {
			log.Printf("Failed to generate CSR: %v", err)
			return
		}

		peerURL := fmt.Sprintf("https://%s:%d", cfg.AdvertiseAddr, 2480)
		reqBody, _ := json.Marshal(map[string]any{
			"name":      "filestore",
			"peer_urls": []string{peerURL},
			"csr_pem":   csrPEM,
		})

		joinURL := cfg.MasterAPIURL + "/cluster/join"
		req, err := http.NewRequestWithContext(ctx, "POST", joinURL, bytes.NewReader(reqBody))
		if err != nil {
			log.Printf("Failed to create join request: %v", err)
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
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to call master join API: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Master rejected join request: %s", string(body))
			return
		}

		var joinResp struct {
			MemberID uint64 `json:"member_id"`
			Cluster  string `json:"cluster"`
			CertPEM  []byte `json:"cert_pem"`
			CACert   []byte `json:"ca_cert"`
		}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&joinResp); decodeErr != nil {
			log.Printf("Failed to decode join response: %v", decodeErr)
			return
		}

		if writeCertErr := os.WriteFile(certPath, joinResp.CertPEM, 0o600); writeCertErr != nil {
			log.Printf("Failed to write node.crt: %v", err)
			return
		}
		if writeCaErr := os.WriteFile(caPath, joinResp.CACert, 0o600); writeCaErr != nil {
			log.Printf("Failed to write ca.crt: %v", writeCaErr)
			return
		}

		initialCluster = joinResp.Cluster

		log.Printf("Successfully joined cluster! Assigned Member ID: %d", joinResp.MemberID)
	}
	etcdState, startEtcdErr := state.StartFileStore(cfg.DataDir, initialCluster, cfg.TLSDir, cfg.NodeName)
	if startEtcdErr != nil {
		log.Printf("Failed to join cluster: %v", startEtcdErr)
		return
	}

	store, newStateErr := state.NewStateManager(
		[]string{masterGRPCURL},
		[]string{"localhost:2379"},
		cfg.TLSDir,
	)
	if newStateErr != nil {
		log.Printf("Failed to connect to local etcd: %v", newStateErr)
		return
	}

	idMgr, err := auth.NewIdentityManagerFromPubKey(cfg.ClusterPubKey)
	if err != nil {
		log.Printf("Failed to create identity manager: %v", err)
		return
	}

	cm := &certs.CertManager{}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Printf("Failed to load TLS certs for HTTP server: %v", err)
		return
	}
	cm.SetCertificate(&cert)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		etcdState.Server.Close()
		closeErr := store.Close()
		if closeErr != nil {
			log.Fatalf("Failed to close state manager: %v", closeErr)
		}
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

	var lis net.ListenConfig
	drpcListener, err := lis.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cfg.ListenAddr, cfg.DrpcPort))
	if err != nil {
		log.Printf("Failed to listen for drpc: %v", err)
		return
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
				_, writeErr := w.Write([]byte("future proofing"))
				if writeErr != nil {
					log.Printf("Failed to write response: %v", writeErr)
					return
				}
			})
		})
	})
}
