package ctlcmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/models"
)

type AuthStatusResult struct {
	Server string   `json:"server"`
	User   string   `json:"user"`
	Roles  []string `json:"roles"`
	Status string   `json:"status"`
}

func (a AuthStatusResult) GetHeaders() []string {
	return []string{"SERVER", "USER", "ROLES", "STATUS"}
}

func (a AuthStatusResult) GetRow() []string {
	return []string{
		a.Server,
		a.User,
		strings.Join(a.Roles, ", "),
		a.Status,
	}
}

type AuthPermissionsResult struct {
	Server      string             `json:"server"`
	User        string             `json:"user"`
	Roles       []string           `json:"roles"`
	Permissions []models.ScopeInfo `json:"permissions"`
	DenyScopes  []models.ScopeInfo `json:"deny_scopes,omitempty"`
	Status      string             `json:"status"`
}

func (a *AuthPermissionsResult) GetHeaders() []string {
	return []string{"SERVER", "USER", "ROLES", "PERMISSIONS", "STATUS"}
}

func (a *AuthPermissionsResult) GetRow() []string {
	return []string{
		a.Server,
		a.User,
		strings.Join(a.Roles, ", "),
		formatScopeInfos(a.Permissions),
		a.Status,
	}
}

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication related operations for clusters",
		Long:  `Authentication related operations for cluster management.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(NewAuthStatusCmd())
	cmd.AddCommand(NewAuthPermsCmd())
	cmd.AddCommand(NewSetDefaultClusterUserCmd())
	cmd.AddCommand(NewAddClusterUserCmd())
	return cmd
}

func formatScopeInfos(scopes []models.ScopeInfo) string {
	if len(scopes) == 0 {
		return ""
	}

	parts := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		parts = append(parts, scope.Action+":"+scope.Resource)
	}
	return strings.Join(parts, ", ")
}

func resolveAuthStatusInputs(cmd *cobra.Command) (cfg *config.GlobalOptions, serverURL, token string, caCert []byte, err error) {
	cfg, err = config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("failed to get global options: %w", err)
	}

	if envErr := viper.BindEnv("token", "GNS3_TOKEN"); envErr != nil {
		return nil, "", "", nil, fmt.Errorf("failed to bind token environment variable: %w", envErr)
	}

	keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("failed to resolve key file path: %w", err)
	}

	kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("failed to load keys: %w", err)
	}

	if cfg.Cluster != "" && cfg.ClusterEntry != nil {
		token = cfg.ClusterEntry.Master.AccessToken
		if token == "" {
			token = viper.GetString("token")
		}
		if token == "" {
			return nil, "", "", nil, fmt.Errorf("no access token for cluster %q in key file", cfg.Cluster)
		}

		return cfg, cfg.ClusterEntry.Master.URL, token, []byte(cfg.ClusterEntry.CaCert), nil
	}

	if cfg.Server == "" {
		return nil, "", "", nil, fmt.Errorf("either --server or --cluster must be specified")
	}

	serverURL = cfg.Server
	token = ""
	if cfg.User != "" {
		entry, ok := kf.GetUserForServer(serverURL, cfg.User)
		if !ok {
			return nil, "", "", nil, fmt.Errorf("user %q not found for %q in key file", cfg.User, serverURL)
		}
		token = entry.AccessToken
	}
	if token == "" {
		token = viper.GetString("token")
	}
	if token == "" {
		entry, ok := kf.GetUserForServer(serverURL, "")
		if !ok {
			return nil, "", "", nil, fmt.Errorf("no token found for %q in key file (set GNS3_TOKEN or add cluster first)", serverURL)
		}
		token = entry.AccessToken
	}

	return cfg, serverURL, token, nil, nil
}

func NewAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check authentication status against a cluster master",
		Long: `Check authentication status against a cluster master node.
Use --server to test a master URL before adding to a cluster.
Use --cluster to test against an existing cluster's master from the keyfile.
If GNS3_TOKEN environment variable isn't set, the token from the keyfile will be used.`,
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, serverURL, token, caCert, err := resolveAuthStatusInputs(cmd)
			if err != nil {
				return err
			}

			settings := api.NewSettings(
				api.WithBaseURLV2(serverURL+"/api/v1"),
				api.WithToken(token),
				api.WithVerify(!cfg.Insecure),
				api.WithCA(caCert),
			)
			client := api.NewClientV2(&settings)

			resp, err := client.GetAuthStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get auth status: %w", err)
			}

			result := AuthStatusResult{
				Server: serverURL,
				User:   resp.User,
				Roles:  resp.Roles,
				Status: "Authenticated",
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(&result, cmd.OutOrStdout())
		},
	}

	return cmd
}

func NewAuthPermsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perms",
		Short: "Show the effective permissions for the current cluster auth context",
		Long: `Show the effective permissions resolved by the cluster master for the current user.
Use -o json or -o yaml to inspect the permission object directly.`,
		Annotations: map[string]string{
			"auth-mode": "flexible",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, serverURL, token, caCert, err := resolveAuthStatusInputs(cmd)
			if err != nil {
				return err
			}

			settings := api.NewSettings(
				api.WithBaseURLV2(serverURL+"/api/v1"),
				api.WithToken(token),
				api.WithVerify(!cfg.Insecure),
				api.WithCA(caCert),
			)
			client := api.NewClientV2(&settings)

			resp, err := client.GetAuthStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get auth status: %w", err)
			}

			result := AuthPermissionsResult{
				Server:      serverURL,
				User:        resp.User,
				Roles:       resp.Roles,
				Permissions: resp.Permissions,
				DenyScopes:  resp.DenyScopes,
				Status:      "Authenticated",
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(result, cmd.OutOrStdout())
		},
	}

	return cmd
}
