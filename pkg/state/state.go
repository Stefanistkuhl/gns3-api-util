package state

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/0xveya/gns3util/pkg/state/pb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StateManager struct {
	MasterClient *clientv3.Client
	LocalClient  *clientv3.Client
}

const RootPrefix = "/gns3/v1"

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

func (s *StateManager) PutUserPermissions(ctx context.Context, userID string, scopes []string) error {
	msg := &pb.UserPermissions{
		UserId:    userID,
		Scopes:    scopes,
		UpdatedAt: timestamppb.Now(),
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = s.MasterClient.Put(ctx, s.userKey(userID), string(data))
	return err
}

func (s *StateManager) GetUserPermissions(ctx context.Context, userID string) (*pb.UserPermissions, error) {
	resp, err := s.LocalClient.Get(ctx, s.userKey(userID))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	perms := &pb.UserPermissions{}
	if err := proto.Unmarshal(resp.Kvs[0].Value, perms); err != nil {
		return nil, fmt.Errorf("failed to decode proto: %w", err)
	}

	return perms, nil
}

func (s *StateManager) CheckPermission(ctx context.Context, userID, requiredScope string) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	if slices.Contains(perms.Scopes, requiredScope) {
		return true, nil
	}
	return false, nil
}

func (s *StateManager) GetUserScopes(ctx context.Context, userID string) (string, error) {
	perms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return "", err
	}
	return strings.Join(perms.Scopes, ","), nil
}

func (s *StateManager) userKey(userID string) string {
	return fmt.Sprintf("%s/auth/scopes/%s", RootPrefix, userID)
}

func (s *StateManager) nodeKey(nodeID string) string {
	return fmt.Sprintf("%s/nodes/%s", RootPrefix, nodeID)
}

func (s *StateManager) PutNode(ctx context.Context, node *pb.Node) error {
	data, err := proto.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	_, err = s.MasterClient.Put(ctx, s.nodeKey(node.Id), string(data))
	return err
}

func (s *StateManager) RegisterNodeTxn(ctx context.Context, node *pb.Node, scopes []string) (bool, error) {
	nodeKey := s.nodeKey(node.Id)
	userKey := s.userKey(node.Id)

	nodeData, err := proto.Marshal(node)
	if err != nil {
		return false, err
	}

	perms := &pb.UserPermissions{
		UserId:    node.Id,
		Scopes:    scopes,
		UpdatedAt: timestamppb.Now(),
	}
	permData, err := proto.Marshal(perms)
	if err != nil {
		return false, err
	}

	txn := s.MasterClient.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(nodeKey), "=", 0)).
		Then(
			clientv3.OpPut(nodeKey, string(nodeData)),
			clientv3.OpPut(userKey, string(permData)),
		).
		Else(
			clientv3.OpGet(nodeKey),
		)

	resp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	return resp.Succeeded, nil
}

func (s *StateManager) UpdateNodeHealth(ctx context.Context, nodeID string, healthy bool) error {
	resp, getNodeErr := s.LocalClient.Get(ctx, s.nodeKey(nodeID))
	if getNodeErr != nil {
		return fmt.Errorf("failed to get node: %w", getNodeErr)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node := &pb.Node{}
	if err := proto.Unmarshal(resp.Kvs[0].Value, node); err != nil {
		return fmt.Errorf("failed to unmarshal node: %w", err)
	}

	node.LastSeen = timestamppb.Now()
	if healthy {
		node.Status = pb.NodeStatus_NODE_STATUS_ONLINE
	} else {
		node.Status = pb.NodeStatus_NODE_STATUS_OFFLINE
	}

	data, marshalErr := proto.Marshal(node)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal node: %w", marshalErr)
	}

	_, err := s.MasterClient.Put(ctx, s.nodeKey(nodeID), string(data))
	return err
}

func (s *StateManager) GetNodes(ctx context.Context) ([]*pb.Node, error) {
	resp, err := s.LocalClient.Get(ctx, fmt.Sprintf("%s/nodes/", RootPrefix), clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	nodes := make([]*pb.Node, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		node := &pb.Node{}
		if err := proto.Unmarshal(kv.Value, node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node: %w", err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}
