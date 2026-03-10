package rpc

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/0xveya/gns3util/internal/shared/pb/master"
	"github.com/0xveya/gns3util/pkg/state"
)

type SyncService struct {
	Store *state.StateManager
}

func NewSyncService(store *state.StateManager) *SyncService {
	return &SyncService{Store: store}
}

func (s *SyncService) CheckPermission(
	ctx context.Context,
	req *pb.PermissionCheckRequest,
) (*pb.PermissionCheckResponse, error) {
	revokedKey := "/auth/revoked/" + req.Jti
	resp, err := s.Store.MasterClient.Get(ctx, revokedKey)
	if err != nil {
		return nil, fmt.Errorf("check revoked token: %w", err)
	}
	if len(resp.Kvs) > 0 {
		return &pb.PermissionCheckResponse{Allowed: false}, nil
	}

	scopeKey := "/auth/scopes/" + req.UserId
	scopeResp, err := s.Store.MasterClient.Get(ctx, scopeKey)
	if err != nil {
		return nil, fmt.Errorf("get user scopes: %w", err)
	}

	if len(scopeResp.Kvs) == 0 {
		return &pb.PermissionCheckResponse{Allowed: false}, nil
	}

	scopesCSV := string(scopeResp.Kvs[0].Value)
	for scope := range strings.SplitSeq(scopesCSV, ",") {
		scope = strings.TrimSpace(scope)
		if scope == req.Scope {
			return &pb.PermissionCheckResponse{Allowed: true}, nil
		}
	}

	return &pb.PermissionCheckResponse{Allowed: false}, nil
}
