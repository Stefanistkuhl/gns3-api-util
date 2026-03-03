package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/0xveya/gns3util/internal/master/handlers"
	"github.com/0xveya/gns3util/pkg/env"
	"github.com/0xveya/gns3util/pkg/state"
	"github.com/0xveya/gns3util/pkg/web/auth"
	"github.com/0xveya/gns3util/pkg/web/certs"
	commonhandlers "github.com/0xveya/gns3util/pkg/web/common_handlers"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	clientv3 "go.etcd.io/etcd/client/v3"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type MasterConfig struct {
	RootPassword   string `env:"ROOT_PASSWORD" type:"secret" required:"true"`
	APIPort        int    `env:"MASTER_API_PORT" type:"port" default:"8443"`
	DrpcPort       int    `env:"MASTER_STORE_DRPC_PORT" type:"port" default:"2748"`
	APIListenAddr  string `env:"MASTER_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	EtcdPort       int    `env:"MASTER_ETCD_API_PORT" type:"port" default:"2379"`
	EtcdListenAddr string `env:"MASTER_ETCD_API_LISTEN_ADDR" type:"listen" default:"0.0.0.0"`
	PrivKeyStr     string `env:"CLUSTER_PRIV_KEY" type:"string"`
	TLSSubject     string `env:"MASTER_TLS_SUBJ" type:"string" default:"/CN=localhost"`
	TLSCertFile    string `env:"MASTER_TLS_CERT_FILE" type:"string"`
	TLSKeyFile     string `env:"MASTER_TLS_KEY_FILE" type:"string"`
}

var cfg MasterConfig

func init() {
	if err := env.LoadConfig(&cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	etcdState, err := state.StartMaster("./master.etcd")
	if err != nil {
		log.Fatalf("Failed to start etcd: %v", err)
	}

	store, err := state.NewStateManager(
		[]string{"localhost:" + strconv.Itoa(cfg.EtcdPort)},
		nil,
		"root", cfg.RootPassword,
	)
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}

	if err := bootstrapIfNeeded(ctx, store.MasterClient); err != nil {
		log.Fatalf("Failed to bootstrap auth: %v", err)
	}

	var idMgr *auth.IdentityManager

	if cfg.PrivKeyStr == "" {
		pubKey, privKey, err := auth.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate keys: %v", err)
		}
		log.Printf("Generated new key pair")
		log.Printf("CLUSTER_PUB_KEY=%s", auth.EncodePublicKey(pubKey))
		log.Printf("CLUSTER_PRIV_KEY=%s", auth.EncodePrivateKey(privKey))

		idMgr, err = auth.NewIdentityManager(privKey)
		if err != nil {
			log.Fatalf("Failed to create identity manager: %v", err)
		}
	} else {
		privKey, err := auth.DecodePrivateKey(cfg.PrivKeyStr)
		if err != nil {
			log.Fatalf("Failed to decode private key: %v", err)
		}
		idMgr, err = auth.NewIdentityManager(privKey)
		if err != nil {
			log.Fatalf("Failed to create identity manager: %v", err)
		}
	}

	master := &handlers.Master{
		IDMgr: idMgr,
		Store: store,
	}

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

	cm := &certs.CertManager{}

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS cert: %v", err)
		}
		cm.SetCertificate(&cert)
	} else {
		err = cm.LoadFromEtcd(ctx, store)
		if err != nil {
			log.Println("No certificate in etcd, generating self-signed...")
			certPEM, keyPEM, err := generateSelfSignedCert(cfg.TLSSubject)
			if err != nil {
				log.Fatalf("Failed to generate cert: %v", err)
			}
			if err := store.PutCertificate(ctx, certPEM, keyPEM); err != nil {
				log.Printf("Warning: failed to upload cert to etcd: %v", err)
			}
			if err := cm.LoadFromEtcd(ctx, store); err != nil {
				log.Fatalf("Failed to load generated cert: %v", err)
			}
			log.Printf("Generated self-signed certificate with subject: %s", cfg.TLSSubject)
		}
	}

	notify := store.WatchCertificate(ctx)
	cm.WatchEtcd(ctx, store, notify)

	m := drpcmux.New()
	r := chi.NewRouter()
	setupRouter(r, master)

	tlsConfig := &tls.Config{
		GetCertificate: cm.GetCertificate,
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", cfg.APIPort),
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	drpcServer := drpcserver.New(m)

	drpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.APIListenAddr, cfg.DrpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for drpc: %v", err)
	}

	log.Printf("Starting Master Node on :%d", cfg.APIPort)
	log.Printf("Starting drpc server on :%d", cfg.DrpcPort)

	errChan := make(chan error, 2)

	go func() {
		errChan <- server.ListenAndServeTLS("", "")
	}()

	go func() {
		errChan <- drpcServer.Serve(ctx, drpcListener)
	}()

	for range 2 {
		if err := <-errChan; err != nil {
			log.Printf("Server error: %v", err)
		}
	}
}

func setupRouter(r chi.Router, master *handlers.Master) {
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/healthz", commonhandlers.HandleHealthz)

	r.Post("/auth/token", master.HandleCreateToken)
	r.Post("/auth/grant", master.HandleGrantAccess)
	r.Post("/auth/revoke", master.HandleRevokeAccess)
	r.Post("/certs/upload", master.HandleUploadCert)
	r.Get("/certs", master.HandleGetCert)
	r.Post("/cluster/join", master.HandleJoinCluster)
}

func bootstrapIfNeeded(ctx context.Context, cli *clientv3.Client) error {
	authResp, err := cli.AuthStatus(ctx)
	if err != nil {
		return err
	}

	if authResp.Enabled {
		log.Println("Auth already enabled, skipping bootstrap")
		return nil
	}

	log.Println("Bootstrapping RBAC...")
	return state.BootstrapMasterAuth(ctx, cli)
}

func generateSelfSignedCert(subject string) (certPEM, keyPEM []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: strings.TrimPrefix(subject, "/CN=")},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPEM, keyPEM, nil
}
