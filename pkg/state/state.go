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

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	sharedpb "github.com/0xveya/gns3util/internal/shared/pb"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"github.com/0xveya/gns3util/pkg/utils/globals"
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

func (s *StateManager) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	resp, err := s.LocalClient.Get(ctx, fmt.Sprintf("/auth/revoked/%s", jti))
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) > 0, nil
}

func (s *StateManager) RevokeToken(ctx context.Context, jti string) error {
	_, err := s.MasterClient.Put(ctx, fmt.Sprintf("/auth/revoked/%s", jti), "1")
	return err
}

func (s *StateManager) PutUserPermissions(ctx context.Context, userID string, roleNames []string, denyScopes []*pb.Scope) error {
	msg := &pb.UserPermissions{
		UserId:     userID,
		RoleNames:  roleNames,
		UpdatedAt:  timestamppb.Now(),
		DenyScopes: denyScopes,
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

func (s *StateManager) CheckPermission(
	ctx context.Context,
	userID string,
	requiredAction sharedpb.Action,
	requiredResource sharedpb.Resource,
) (bool, error) {
	userPerms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	if userPerms == nil || userPerms.UserId == "" {
		return false, nil
	}
	effectiveScopes, err := s.effectiveScopesForRoles(ctx, userPerms.RoleNames)
	if err != nil {
		return false, err
	}

	requiredScope := &pb.Scope{
		Action:   requiredAction,
		Resource: requiredResource,
	}

	for _, scope := range userPerms.DenyScopes {
		if scopesMatch(scope, requiredScope) {
			return false, nil
		}
	}

	for _, scope := range effectiveScopes {
		if scopesMatch(scope, requiredScope) {
			return true, nil
		}
	}

	return false, nil
}

func scopesMatch(have, need *pb.Scope) bool {
	if have == nil || need == nil {
		return false
	}

	if have.Action == need.Action && (have.Resource == need.Resource || have.Resource == sharedpb.Resource_RESOURCE_UNSPECIFIED) {
		return true
	}

	if have.Action == sharedpb.Action_ACTION_ADMIN {
		return have.Resource == need.Resource || have.Resource == sharedpb.Resource_RESOURCE_UNSPECIFIED
	}

	return false
}

func (s *StateManager) GetRole(ctx context.Context, roleName string) (*pb.Role, error) {
	resp, err := s.MasterClient.Get(ctx, s.roleKey(roleName))
	if err != nil {
		return nil, fmt.Errorf("failed to get role from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("role %s not found", roleName)
	}

	var role pb.Role
	err = proto.Unmarshal(resp.Kvs[0].Value, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal role: %w", err)
	}

	return &role, nil
}

func (s *StateManager) GetUserScopes(ctx context.Context, userID string) ([]*pb.Scope, error) {
	perms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.effectiveScopesForRoles(ctx, perms.RoleNames)
}

func (s *StateManager) GetUserDenyScopes(ctx context.Context, userID string) ([]*pb.Scope, error) {
	perms, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return perms.DenyScopes, nil
}

func (s *StateManager) effectiveScopesForRoles(ctx context.Context, roleNames []string) ([]*pb.Scope, error) {
	effectiveScopes := make([]*pb.Scope, 0)

	if slices.Contains(roleNames, globals.RoleAdmin) {
		effectiveScopes = append(effectiveScopes, &pb.Scope{
			Action:   sharedpb.Action_ACTION_ADMIN,
			Resource: sharedpb.Resource_RESOURCE_UNSPECIFIED,
		})
	}

	for _, roleName := range roleNames {
		if roleName == globals.RoleAdmin {
			continue
		}
		role, err := s.GetRole(ctx, roleName)
		if err != nil {
			continue
		}
		effectiveScopes = append(effectiveScopes, role.Scopes...)
	}

	return effectiveScopes, nil
}

func (s *StateManager) roleKey(roleName string) string {
	return fmt.Sprintf("%s/roles/%s", RootPrefix, roleName)
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

func (s *StateManager) RegisterNodeTxn(ctx context.Context, node *pb.Node, roleNames []string) (bool, error) {
	nodeKey := s.nodeKey(node.Id)
	userKey := s.userKey(node.Id)

	nodeData, err := proto.Marshal(node)
	if err != nil {
		return false, err
	}

	perms := &pb.UserPermissions{
		UserId:     node.Id,
		RoleNames:  roleNames,
		UpdatedAt:  timestamppb.Now(),
		DenyScopes: nil,
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

func (s *StateManager) HasEffectivePermission(
	ctx context.Context,
	userID string,
	requiredAction sharedpb.Action,
	requiredResource sharedpb.Resource,
) (bool, error) {
	return s.CheckPermission(ctx, userID, requiredAction, requiredResource)
}

func (s *StateManager) CreateUser(ctx context.Context, userID string, roleNames []string, denyScopes []*pb.Scope) error {
	msg := &pb.UserPermissions{
		UserId:     userID,
		RoleNames:  roleNames,
		UpdatedAt:  timestamppb.Now(),
		DenyScopes: denyScopes,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal user permissions: %w", err)
	}
	_, err = s.MasterClient.Put(ctx, s.userKey(userID), string(data))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *StateManager) CreateRole(ctx context.Context, role *pb.Role) error {
	if role.CreatedAt == nil {
		role.CreatedAt = timestamppb.Now()
	}
	role.UpdatedAt = timestamppb.Now()

	data, err := proto.Marshal(role)
	if err != nil {
		return fmt.Errorf("failed to marshal role: %w", err)
	}

	_, err = s.MasterClient.Put(ctx, s.roleKey(role.Name), string(data))
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

func (s *StateManager) UpdateRole(ctx context.Context, role *pb.Role) error {
	role.UpdatedAt = timestamppb.Now()

	data, err := proto.Marshal(role)
	if err != nil {
		return fmt.Errorf("failed to marshal role: %w", err)
	}

	_, err = s.MasterClient.Put(ctx, s.roleKey(role.Name), string(data))
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

func (s *StateManager) DeleteRole(ctx context.Context, roleName string) error {
	_, err := s.MasterClient.Delete(ctx, s.roleKey(roleName))
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (s *StateManager) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.MasterClient.Delete(ctx, s.userKey(userID))
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *StateManager) ListUsers(ctx context.Context) ([]*pb.UserPermissions, error) {
	prefix := fmt.Sprintf("%s/auth/scopes/", RootPrefix)
	resp, err := s.LocalClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*pb.UserPermissions, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		u := &pb.UserPermissions{}
		if err := proto.Unmarshal(kv.Value, u); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *StateManager) ListRoles(ctx context.Context) ([]*pb.Role, error) {
	prefix := fmt.Sprintf("%s/roles/", RootPrefix)
	resp, err := s.LocalClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	roles := make([]*pb.Role, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		r := &pb.Role{}
		if err := proto.Unmarshal(kv.Value, r); err != nil {
			continue
		}
		roles = append(roles, r)
	}
	return roles, nil
}

// Job management

func (s *StateManager) jobKey(nodeID, jobName string) string {
	return fmt.Sprintf("%s/jobs/%s/%s", RootPrefix, nodeID, jobName)
}

func (s *StateManager) jobRunKey(runID string) string {
	return fmt.Sprintf("%s/job-runs/%s", RootPrefix, runID)
}

func (s *StateManager) RegisterJob(ctx context.Context, job *pb.JobDefinition) error {
	data, err := proto.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job definition: %w", err)
	}

	_, err = s.MasterClient.Put(ctx, s.jobKey(job.NodeId, job.Name), string(data))
	return err
}

func (s *StateManager) GetRegisteredJobs(ctx context.Context) ([]*pb.JobDefinition, error) {
	resp, err := s.LocalClient.Get(ctx, fmt.Sprintf("%s/jobs/", RootPrefix), clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	jobs := make([]*pb.JobDefinition, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		job := &pb.JobDefinition{}
		if err := proto.Unmarshal(kv.Value, job); err != nil {
			return nil, fmt.Errorf("failed to unmarshal job definition: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *StateManager) PutJobRun(ctx context.Context, run *pb.JobRun) error {
	data, err := proto.Marshal(run)
	if err != nil {
		return fmt.Errorf("failed to marshal job run: %w", err)
	}

	_, err = s.MasterClient.Put(ctx, s.jobRunKey(run.RunId), string(data))
	return err
}

func (s *StateManager) GetJobRun(ctx context.Context, runID string) (*pb.JobRun, error) {
	resp, err := s.LocalClient.Get(ctx, s.jobRunKey(runID))
	if err != nil {
		return nil, fmt.Errorf("failed to get job run: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("job run %s not found", runID)
	}

	run := &pb.JobRun{}
	if err := proto.Unmarshal(resp.Kvs[0].Value, run); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job run: %w", err)
	}

	return run, nil
}

func (s *StateManager) ListJobRuns(ctx context.Context) ([]*pb.JobRun, error) {
	resp, err := s.LocalClient.Get(ctx, fmt.Sprintf("%s/job-runs/", RootPrefix), clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list job runs: %w", err)
	}

	runs := make([]*pb.JobRun, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		run := &pb.JobRun{}
		if err := proto.Unmarshal(kv.Value, run); err != nil {
			return nil, fmt.Errorf("failed to unmarshal job run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}
