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
	isLearner    bool
}

func NewMasterStateManager(
	endpoints []string,
	tlsDir string,
) (*StateManager, error) {
	cleanDir := filepath.Clean(tlsDir)
	certFile := filepath.Join(cleanDir, "node.crt")
	keyFile := filepath.Join(cleanDir, "node.key")
	caFile := filepath.Join(cleanDir, "ca.crt")

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

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	return &StateManager{
		MasterClient: cli,
		LocalClient:  cli,
		isLearner:    false,
	}, nil
}

func NewLearnerStateManager(
	masterEndpoints []string,
	localSocketPath string,
	tlsDir string,
) (*StateManager, error) {
	cleanDir := filepath.Clean(tlsDir)
	certFile := filepath.Join(cleanDir, "node.crt")
	keyFile := filepath.Join(cleanDir, "node.key")
	caFile := filepath.Join(cleanDir, "ca.crt")

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

	masterTLS := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	masterCli, err := clientv3.New(clientv3.Config{
		Endpoints:   masterEndpoints,
		DialTimeout: 5 * time.Second,
		TLS:         masterTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	localCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"unix://" + localSocketPath},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		_ = masterCli.Close()
		return nil, fmt.Errorf("failed to connect to local socket: %w", err)
	}

	return &StateManager{
		MasterClient: masterCli,
		LocalClient:  localCli,
		isLearner:    true,
	}, nil
}

func (s *StateManager) Close() error {
	var err error

	if s.MasterClient != nil {
		if e := s.MasterClient.Close(); e != nil {
			err = e
		}
	}

	if s.LocalClient != nil && s.LocalClient != s.MasterClient {
		if e := s.LocalClient.Close(); e != nil && err == nil {
			err = e
		}
	}

	return err
}

func (s *StateManager) PutUserPermissions(
	ctx context.Context,
	userID string,
	scopes []string,
) error {
	key := fmt.Sprintf("/auth/scopes/%s", userID)
	val := strings.Join(scopes, ",")

	_, err := s.MasterClient.Put(ctx, key, val)
	return err
}

func (s *StateManager) CheckPermission(
	ctx context.Context,
	userID,
	requiredScope string,
) (bool, error) {
	key := fmt.Sprintf("/auth/scopes/%s", userID)

	var (
		resp *clientv3.GetResponse
		err  error
	)

	if s.isLearner {
		resp, err = s.LocalClient.Get(ctx, key, clientv3.WithSerializable())
	} else {
		resp, err = s.LocalClient.Get(ctx, key)
	}

	if err != nil {
		return false, err
	}

	if len(resp.Kvs) == 0 {
		return false, nil
	}

	currentScopes := strings.SplitSeq(string(resp.Kvs[0].Value), ",")
	for scope := range currentScopes {
		if strings.TrimSpace(scope) == requiredScope {
			return true, nil
		}
	}

	return false, nil
}

func (s *StateManager) GetUserScopes(
	ctx context.Context,
	userID string,
) (string, error) {
	key := fmt.Sprintf("/auth/scopes/%s", userID)

	var (
		resp *clientv3.GetResponse
		err  error
	)

	if s.isLearner {
		resp, err = s.LocalClient.Get(ctx, key, clientv3.WithSerializable())
	} else {
		resp, err = s.LocalClient.Get(ctx, key)
	}

	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("user not found")
	}

	return string(resp.Kvs[0].Value), nil
}
