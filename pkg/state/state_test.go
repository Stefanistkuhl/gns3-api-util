package state

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	sharedpb "github.com/0xveya/gns3util/internal/shared/pb"
	"github.com/0xveya/gns3util/pkg/state/pb"
	"github.com/0xveya/gns3util/pkg/utils/globals"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func TestCheckPermissionPrecedence(t *testing.T) {
	manager := newTestStateManager(t)
	ctx := context.Background()

	if err := manager.CreateRole(ctx, &pb.Role{
		Name: "file-reader",
		Scopes: []*pb.Scope{
			{Action: sharedpb.Action_ACTION_READ, Resource: sharedpb.Resource_RESOURCE_FILES},
		},
	}); err != nil {
		t.Fatalf("CreateRole(file-reader) failed: %v", err)
	}

	if err := manager.CreateRole(ctx, &pb.Role{
		Name: "job-writer",
		Scopes: []*pb.Scope{
			{Action: sharedpb.Action_ACTION_WRITE, Resource: sharedpb.Resource_RESOURCE_UNSPECIFIED},
		},
	}); err != nil {
		t.Fatalf("CreateRole(job-writer) failed: %v", err)
	}

	tests := []struct {
		name           string
		userID         string
		roles          []string
		denyScopes     []*pb.Scope
		requiredAction sharedpb.Action
		requiredRes    sharedpb.Resource
		wantAllowed    bool
	}{
		{
			name:   "deny overrides matching role allow",
			userID: "alice",
			roles:  []string{"file-reader"},
			denyScopes: []*pb.Scope{
				{Action: sharedpb.Action_ACTION_READ, Resource: sharedpb.Resource_RESOURCE_FILES},
			},
			requiredAction: sharedpb.Action_ACTION_READ,
			requiredRes:    sharedpb.Resource_RESOURCE_FILES,
			wantAllowed:    false,
		},
		{
			name:   "deny overrides builtin admin",
			userID: "admin-denied",
			roles:  []string{globals.RoleAdmin},
			denyScopes: []*pb.Scope{
				{Action: sharedpb.Action_ACTION_READ, Resource: sharedpb.Resource_RESOURCE_FILES},
			},
			requiredAction: sharedpb.Action_ACTION_READ,
			requiredRes:    sharedpb.Resource_RESOURCE_FILES,
			wantAllowed:    false,
		},
		{
			name:   "admin still allowed outside denied scope",
			userID: "admin-denied-other-scope",
			roles:  []string{globals.RoleAdmin},
			denyScopes: []*pb.Scope{
				{Action: sharedpb.Action_ACTION_READ, Resource: sharedpb.Resource_RESOURCE_FILES},
			},
			requiredAction: sharedpb.Action_ACTION_WRITE,
			requiredRes:    sharedpb.Resource_RESOURCE_FILES,
			wantAllowed:    true,
		},
		{
			name:           "wildcard resource allow matches specific resource",
			userID:         "wildcard-allow",
			roles:          []string{"job-writer"},
			requiredAction: sharedpb.Action_ACTION_WRITE,
			requiredRes:    sharedpb.Resource_RESOURCE_JOBS,
			wantAllowed:    true,
		},
		{
			name:   "wildcard resource deny matches specific resource",
			userID: "wildcard-deny",
			roles:  []string{"job-writer"},
			denyScopes: []*pb.Scope{
				{Action: sharedpb.Action_ACTION_WRITE, Resource: sharedpb.Resource_RESOURCE_UNSPECIFIED},
			},
			requiredAction: sharedpb.Action_ACTION_WRITE,
			requiredRes:    sharedpb.Resource_RESOURCE_JOBS,
			wantAllowed:    false,
		},
		{
			name:           "missing role fails closed without error",
			userID:         "ghost-role-user",
			roles:          []string{"missing-role"},
			requiredAction: sharedpb.Action_ACTION_READ,
			requiredRes:    sharedpb.Resource_RESOURCE_FILES,
			wantAllowed:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := manager.CreateUser(ctx, tt.userID, tt.roles, tt.denyScopes); err != nil {
				t.Fatalf("CreateUser(%s) failed: %v", tt.userID, err)
			}

			allowed, err := manager.CheckPermission(ctx, tt.userID, tt.requiredAction, tt.requiredRes)
			if err != nil {
				t.Fatalf("CheckPermission(%s) failed: %v", tt.userID, err)
			}
			if allowed != tt.wantAllowed {
				t.Fatalf("CheckPermission(%s) = %v, want %v", tt.userID, allowed, tt.wantAllowed)
			}
		})
	}
}

func newTestStateManager(t *testing.T) *StateManager {
	t.Helper()

	clientPort := freePort(t)
	peerPort := freePort(t)

	cfg := embed.NewConfig()
	cfg.Dir = t.TempDir()
	cfg.Name = "test-node"
	cfg.LogLevel = "error"

	peerURL := mustParseURL(t, fmt.Sprintf("http://127.0.0.1:%d", peerPort))
	clientURL := mustParseURL(t, fmt.Sprintf("http://127.0.0.1:%d", clientPort))
	cfg.ListenPeerUrls = []url.URL{peerURL}
	cfg.AdvertisePeerUrls = []url.URL{peerURL}
	cfg.ListenClientUrls = []url.URL{clientURL}
	cfg.AdvertiseClientUrls = []url.URL{clientURL}
	cfg.InitialCluster = fmt.Sprintf("%s=%s", cfg.Name, peerURL.String())
	cfg.ClusterState = "new"

	etcd, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("StartEtcd failed: %v", err)
	}

	select {
	case <-etcd.Server.ReadyNotify():
	case <-time.After(30 * time.Second):
		etcd.Close()
		t.Fatal("embedded etcd startup timeout")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{clientURL.String()},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		etcd.Close()
		t.Fatalf("clientv3.New failed: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
		etcd.Close()
	})

	return &StateManager{
		MasterClient: client,
		LocalClient:  client,
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		_ = listener.Close()
		t.Fatal("listener did not expose a TCP address")
	}
	addr := tcpAddr.Port
	if closeErr := listener.Close(); closeErr != nil {
		t.Fatalf("failed to close port allocator listener: %v", closeErr)
	}

	return addr
}

func mustParseURL(t *testing.T, raw string) url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q) failed: %v", raw, err)
	}
	return *parsed
}
