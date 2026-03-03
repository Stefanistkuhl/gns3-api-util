package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type StateManager struct {
	MasterClient *clientv3.Client
	LocalClient  *clientv3.Client
}

func NewStateManager(masterEndpoints, localEndpoints []string, username, password string) (*StateManager, error) {
	masterCli, err := clientv3.New(clientv3.Config{
		Endpoints:   masterEndpoints,
		DialTimeout: 5 * time.Second,
		Username:    username,
		Password:    password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	var localCli *clientv3.Client
	if len(localEndpoints) > 0 {
		localCli, err = clientv3.New(clientv3.Config{
			Endpoints:   localEndpoints,
			DialTimeout: 5 * time.Second,
			Username:    username,
			Password:    password,
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

func (s *StateManager) PutCertificate(ctx context.Context, certPEM, keyPEM []byte) error {
	_, err := s.MasterClient.Put(ctx, "/tls/cert", string(certPEM))
	if err != nil {
		return err
	}
	_, err = s.MasterClient.Put(ctx, "/tls/key", string(keyPEM))
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

func (s *StateManager) GetCertificate(ctx context.Context) (certPEM, keyPEM []byte, err error) {
	certResp, err := s.LocalClient.Get(ctx, "/tls/cert", clientv3.WithSerializable())
	if err != nil {
		return nil, nil, err
	}
	if len(certResp.Kvs) == 0 {
		return nil, nil, fmt.Errorf("certificate not found in etcd")
	}

	keyResp, err := s.LocalClient.Get(ctx, "/tls/key", clientv3.WithSerializable())
	if err != nil {
		return nil, nil, err
	}
	if len(keyResp.Kvs) == 0 {
		return nil, nil, fmt.Errorf("key not found in etcd")
	}

	return certResp.Kvs[0].Value, keyResp.Kvs[0].Value, nil
}

func (s *StateManager) WatchCertificate(ctx context.Context) <-chan struct{} {
	notify := make(chan struct{}, 1)

	go func() {
		defer close(notify)
		watchChan := s.MasterClient.Watch(ctx, "/tls/", clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				return
			case <-watchChan:
				select {
				case notify <- struct{}{}:
				default:
				}
			}
		}
	}()

	return notify
}
