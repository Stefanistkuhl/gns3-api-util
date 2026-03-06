package sync

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbpkg "github.com/0xveya/gns3util/internal/file-store/db"
	"github.com/0xveya/gns3util/internal/file-store/db/sqlc_file_store"
	pb "github.com/0xveya/gns3util/internal/shared/pb/master"
)

type MasterSyncClient interface {
	FullSync(
		ctx context.Context,
		req *pb.FullSyncRequest,
	) (*pb.FullSyncResponse, error)
}

type Service struct {
	DB       *dbpkg.Store
	Client   MasterSyncClient
	NodeName string
}

func NewService(
	db *dbpkg.Store,
	client MasterSyncClient,
	nodeName string,
) *Service {
	return &Service{
		DB:       db,
		Client:   client,
		NodeName: nodeName,
	}
}

func (s *Service) Run(ctx context.Context, interval time.Duration) error {
	if err := s.FullSync(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = s.FullSync(ctx)
		}
	}
}

func (s *Service) FullSync(ctx context.Context) error {
	resp, err := s.Client.FullSync(ctx, &pb.FullSyncRequest{
		NodeName: s.NodeName,
	})
	if err != nil {
		_ = s.DB.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
			return q.UpsertSyncState(ctx, sqlc_file_store.UpsertSyncStateParams{
				ReplicaName:           s.NodeName,
				LastFullSyncAt:        sql.NullTime{},
				LastIncrementalSyncAt: sql.NullTime{},
				LastSourceRevision:    0,
				LastStatus:            "error",
				LastError:             sql.NullString{String: err.Error(), Valid: true},
			})
		})
		return fmt.Errorf("full sync failed: %w", err)
	}

	now := time.Now().UTC()

	return s.DB.WithTx(ctx, func(q *sqlc_file_store.Queries) error {
		if err := q.DeleteAllUserPermissions(ctx); err != nil {
			return err
		}

		for _, p := range resp.Permissions {
			if err := q.InsertUserPermission(
				ctx,
				sqlc_file_store.InsertUserPermissionParams{
					UserID: p.UserId,
					Scope:  p.Scope,
				},
			); err != nil {
				return err
			}
		}

		if err := q.DeleteAllClusterNodes(ctx); err != nil {
			return err
		}

		for _, n := range resp.Nodes {
			if err := q.UpsertClusterNode(
				ctx,
				sqlc_file_store.UpsertClusterNodeParams{
					NodeID:          n.NodeId,
					NodeName:        n.NodeName,
					NodeKind:        n.NodeKind,
					ApiUrl:          nullString(n.ApiUrl),
					DrpcAddr:        nullString(n.DrpcAddr),
					AdvertiseAddr:   nullString(n.AdvertiseAddr),
					Status:          n.Status,
					LastHeartbeatAt: nullTime(n.LastHeartbeatAt),
				},
			); err != nil {
				return err
			}
		}

		if err := q.DeleteAllClusterKV(ctx); err != nil {
			return err
		}

		for _, kv := range resp.Kv {
			if err := q.UpsertClusterKV(
				ctx,
				sqlc_file_store.UpsertClusterKVParams{
					Key:     kv.Key,
					Value:   kv.Value,
					Version: kv.Version,
				},
			); err != nil {
				return err
			}
		}

		if err := q.DeleteExpiredRevokedTokens(ctx); err != nil {
			return err
		}

		for _, rt := range resp.RevokedTokens {
			if err := q.RevokeToken(
				ctx,
				sqlc_file_store.RevokeTokenParams{
					Jti:       rt.Jti,
					UserID:    rt.UserId,
					ExpiresAt: nullTime(rt.ExpiresAt),
				},
			); err != nil {
				return err
			}
		}

		return q.UpsertSyncState(ctx, sqlc_file_store.UpsertSyncStateParams{
			ReplicaName:           s.NodeName,
			LastFullSyncAt:        sql.NullTime{Time: now, Valid: true},
			LastIncrementalSyncAt: sql.NullTime{},
			LastSourceRevision:    resp.SourceRevision,
			LastStatus:            "ok",
			LastError:             sql.NullString{},
		})
	})
}

func nullString(v string) sql.NullString {
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: v,
		Valid:  true,
	}
}

func nullTime(v string) sql.NullTime {
	if v == "" {
		return sql.NullTime{}
	}

	parsed, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return sql.NullTime{}
	}

	return sql.NullTime{
		Time:  parsed,
		Valid: true,
	}
}
