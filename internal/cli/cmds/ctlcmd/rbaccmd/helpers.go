package rbaccmd

import (
	"context"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/models"
)

type rbacClient interface {
	ListRoles(ctx context.Context) (*models.ListRolesResponse, error)
	GetRole(ctx context.Context, roleName string) (*models.RoleInfo, error)
	CreateRole(ctx context.Context, req models.CreateRoleRequest) (*models.RoleInfo, error)
	UpdateRole(ctx context.Context, roleName string, req models.UpdateRoleRequest) (*models.RoleInfo, error)
	DeleteRole(ctx context.Context, roleName string) error

	ListUsers(ctx context.Context) (*models.ListUsersResponse, error)
	GetUser(ctx context.Context, userID string) (*models.UserInfo, error)
	DeleteUser(ctx context.Context, userID string) error
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserInfo, error)
	AssignRole(ctx context.Context, userID string, req models.AssignRoleRequest) (*models.UserInfo, error)
	GenerateUserToken(ctx context.Context, userID string) (string, error)

	RevokeToken(ctx context.Context, req models.RevokeTokenRequest) error
}

func newMasterClient(cfg *config.GlobalOptions) (rbacClient, error) {
	if cfg.ClusterEntry == nil {
		return nil, fmt.Errorf("no cluster entry found, use --cluster to specify a cluster")
	}

	serverURL := cfg.ClusterEntry.Master.URL
	token := cfg.ClusterEntry.Master.AccessToken

	if serverURL == "" {
		return nil, fmt.Errorf("master URL not configured for cluster")
	}
	if token == "" {
		return nil, fmt.Errorf("no access token for master (run 'ctl auth status' first)")
	}

	settings := api.NewSettings(
		api.WithBaseURLV2(serverURL+"/api/v1"),
		api.WithToken(token),
		api.WithVerify(!cfg.Insecure),
		api.WithCA([]byte(cfg.ClusterEntry.CaCert)),
	)

	return api.NewClientV2(&settings), nil
}
