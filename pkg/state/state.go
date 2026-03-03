package state

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type StateManager struct {
	MasterClient *clientv3.Client
	LocalClient  *clientv3.Client
}

func NewStateManager(masterEndpoints, localEndpoints []string, tlsDir string) (*StateManager, error) {
	certFile := filepath.Join(tlsDir, "node.crt")
	keyFile := filepath.Join(tlsDir, "node.key")
	caFile := filepath.Join(tlsDir, "ca.crt")

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key: %w", err)
	}

	caData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA cert: %w", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caData)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	masterCli, err := clientv3.New(clientv3.Config{
		Endpoints:   masterEndpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	var localCli *clientv3.Client
	if len(localEndpoints) > 0 {
		localCli, err = clientv3.New(clientv3.Config{
			Endpoints:   localEndpoints,
			DialTimeout: 5 * time.Second,
			TLS:         tlsConfig,
		})
		if err != nil {
			masterCli.Close()
			return nil, fmt.Errorf("failed to connect to local replica: %w", err)
		}
	} else {
		localCli = masterCli
	}

	return &StateManager{
		MasterClient: masterCli,
		LocalClient:  localCli,
	}, nil
}

func (s *StateManager) Close() error {
	var err error
	if e := s.MasterClient.Close(); e != nil {
		err = e
	}
	if s.LocalClient != s.MasterClient {
		if e := s.LocalClient.Close(); e != nil {
			err = e
		}
	}
	return err
}

func (s *StateManager) PutUserPermissions(ctx context.Context, userID string, scopes []string) error {
	key := fmt.Sprintf("/auth/scopes/%s", userID)
	val := strings.Join(scopes, ",")

	_, err := s.MasterClient.Put(ctx, key, val)
	return err
}

func (s *StateManager) CheckPermission(ctx context.Context, userID string, requiredScope string) (bool, error) {
	key := fmt.Sprintf("/auth/scopes/%s", userID)
	resp, err := s.LocalClient.Get(ctx, key, clientv3.WithSerializable())
	if err != nil || len(resp.Kvs) == 0 {
		return false, err
	}

	return strings.Contains(string(resp.Kvs[0].Value), requiredScope), nil
}
