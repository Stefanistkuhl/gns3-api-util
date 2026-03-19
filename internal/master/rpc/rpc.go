package rpc

import (
	"context"
	"fmt"

	pb "github.com/0xveya/gns3util/internal/shared/pb/master"
	"github.com/0xveya/gns3util/pkg/state"
	statepb "github.com/0xveya/gns3util/pkg/state/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SyncService struct {
	Store *state.StateManager
}

func NewSyncService(store *state.StateManager) *SyncService {
	return &SyncService{Store: store}
}

func (s *SyncService) CheckPermission(
	ctx context.Context,
	req *pb.CheckPermissionRequest,
) (*pb.CheckPermissionResponse, error) {
	revoked, err := s.Store.IsTokenRevoked(ctx, req.Jti)
	if err != nil {
		return nil, fmt.Errorf("check revoked token: %w", err)
	}
	if revoked {
		return &pb.CheckPermissionResponse{Allowed: false}, nil
	}

	allowed, err := s.Store.CheckPermission(ctx, req.UserId, req.Scope)
	if err != nil {
		return nil, fmt.Errorf("check permission: %w", err)
	}

	return &pb.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *SyncService) RegisterJob(
	ctx context.Context,
	req *pb.RegisterJobRequest,
) (*pb.RegisterJobResponse, error) {
	jobDef := &statepb.JobDefinition{
		Name:         req.JobName,
		NodeId:       req.NodeId,
		Interval:     req.Interval,
		Description:  req.Description,
		RegisteredAt: timestamppb.Now(),
	}

	if err := s.Store.RegisterJob(ctx, jobDef); err != nil {
		return nil, fmt.Errorf("register job: %w", err)
	}

	return &pb.RegisterJobResponse{Success: true}, nil
}
