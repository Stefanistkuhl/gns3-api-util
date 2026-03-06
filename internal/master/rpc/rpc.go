package rpc

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/0xveya/gns3util/internal/shared/pb/master"
	"github.com/0xveya/gns3util/pkg/state"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type SyncService struct {
	Store *state.StateManager
}

func NewSyncService(store *state.StateManager) *SyncService {
	return &SyncService{Store: store}
}

func (s *SyncService) FullSync(
	ctx context.Context,
	req *pb.FullSyncRequest,
) (*pb.FullSyncResponse, error) {
	permissions, err := s.loadPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}

	nodes, err := s.loadNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("load nodes: %w", err)
	}

	kv, err := s.loadClusterKV(ctx)
	if err != nil {
		return nil, fmt.Errorf("load cluster kv: %w", err)
	}

	revokedTokens, err := s.loadRevokedTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("load revoked tokens: %w", err)
	}

	return &pb.FullSyncResponse{
		SourceRevision: 0,
		Permissions:    permissions,
		Nodes:          nodes,
		Kv:             kv,
		RevokedTokens:  revokedTokens,
	}, nil
}

func (s *SyncService) loadPermissions(
	ctx context.Context,
) ([]*pb.PermissionRecord, error) {
	resp, err := s.Store.MasterClient.Get(
		ctx,
		"/auth/scopes/",
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, err
	}

	var out []*pb.PermissionRecord

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		userID := strings.TrimPrefix(key, "/auth/scopes/")
		if userID == "" {
			continue
		}

		scopeCSV := string(kv.Value)
		if scopeCSV == "" {
			continue
		}

		for scope := range strings.SplitSeq(scopeCSV, ",") {
			scope = strings.TrimSpace(scope)
			if scope == "" {
				continue
			}

			out = append(out, &pb.PermissionRecord{
				UserId: userID,
				Scope:  scope,
			})
		}
	}

	return out, nil
}

func (s *SyncService) loadNodes(
	ctx context.Context,
) ([]*pb.ClusterNodeRecord, error) {
	resp, err := s.Store.MasterClient.Get(
		ctx,
		"/cluster/nodes/",
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, err
	}

	var out []*pb.ClusterNodeRecord
	for _, kv := range resp.Kvs {
		nodeID := strings.TrimPrefix(string(kv.Key), "/cluster/nodes/")
		if nodeID == "" {
			continue
		}

		// For now store raw value as api_url-ish payload if you have not
		// normalized this yet. Replace with proper JSON parsing later if needed.
		out = append(out, &pb.ClusterNodeRecord{
			NodeId: nodeID,
			Status: "unknown",
			ApiUrl: string(kv.Value),
		})
	}

	return out, nil
}

func (s *SyncService) loadClusterKV(
	ctx context.Context,
) ([]*pb.ClusterKVRecord, error) {
	resp, err := s.Store.MasterClient.Get(
		ctx,
		"/config/",
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, err
	}

	var out []*pb.ClusterKVRecord
	for _, kv := range resp.Kvs {
		out = append(out, &pb.ClusterKVRecord{
			Key:     string(kv.Key),
			Value:   string(kv.Value),
			Version: kv.Version,
		})
	}

	return out, nil
}

func (s *SyncService) loadRevokedTokens(
	ctx context.Context,
) ([]*pb.RevokedTokenRecord, error) {
	resp, err := s.Store.MasterClient.Get(
		ctx,
		"/auth/revoked/",
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, err
	}

	var out []*pb.RevokedTokenRecord
	for _, kv := range resp.Kvs {
		jti := strings.TrimPrefix(string(kv.Key), "/auth/revoked/")
		if jti == "" {
			continue
		}

		out = append(out, &pb.RevokedTokenRecord{
			Jti: jti,
		})
	}

	return out, nil
}
