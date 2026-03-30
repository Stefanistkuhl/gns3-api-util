package rbaccmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/models"
)

type tokenRow struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

func (t *tokenRow) GetHeaders() []string { return []string{"USER ID", "TOKEN"} }
func (t *tokenRow) GetRow() []string     { return []string{t.UserID, t.Token} }

type userRow struct {
	UserID      string             `json:"user_id"`
	Roles       string             `json:"roles"`
	Permissions []models.ScopeInfo `json:"permissions,omitempty"`
	DenyScopes  []models.ScopeInfo `json:"deny_scopes,omitempty"`
	UpdatedAt   string             `json:"updated_at"`
}

func (u *userRow) GetHeaders() []string {
	return []string{"USER ID", "ROLES", "PERMISSIONS", "UPDATED AT"}
}

func (u *userRow) GetRow() []string {
	return []string{u.UserID, u.Roles, formatPermissions(u.Permissions), u.UpdatedAt}
}

func formatPermissions(scopes []models.ScopeInfo) string {
	if len(scopes) == 0 {
		return ""
	}

	parts := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		parts = append(parts, scope.Action+":"+scope.Resource)
	}
	return strings.Join(parts, ", ")
}

func NewUsersCmd() *cobra.Command {
	return newUsersCmdWithFactory(newMasterClient)
}

func newUsersCmdWithFactory(factory clientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage control-plane users in etcd",
		Long:  "List, inspect, and delete user permission records stored in etcd.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	queryGroup := &cobra.Group{ID: "query", Title: "Query Commands:"}
	actionGroup := &cobra.Group{ID: "action", Title: "Action Commands:"}
	cmd.AddGroup(queryGroup, actionGroup)

	listCmd := newListUsersCmd(factory)
	listCmd.GroupID = "query"
	getCmd := newGetUserCmd(factory)
	getCmd.GroupID = "query"
	deleteCmd := newDeleteUserCmd(factory)
	deleteCmd.GroupID = "action"
	createCmd := newCreateUserCmd(factory)
	createCmd.GroupID = "action"
	assignRoleCmd := newAssignRoleCmd(factory)
	assignRoleCmd.GroupID = "action"
	genTokenCmd := newGenTokenCmd(factory)
	genTokenCmd.GroupID = "action"
	revokeTokenCmd := newRevokeTokenCmd(factory)
	revokeTokenCmd.GroupID = "action"

	cmd.AddCommand(listCmd, getCmd, deleteCmd, createCmd, assignRoleCmd, genTokenCmd, revokeTokenCmd)
	return cmd
}

func newListUsersCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all users in etcd",
		Example: `  gns3util ctl users list
  gns3util ctl users list -o json`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			resp, err := client.ListUsers(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			if len(resp.Users) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No users found.")
				return nil
			}

			var records []utils.TableRecord
			for _, u := range resp.Users {
				records = append(records, &userRow{
					UserID:      u.UserID,
					Roles:       strings.Join(u.RoleNames, ", "),
					Permissions: u.Permissions,
					DenyScopes:  u.DenyScopes,
					UpdatedAt:   u.UpdatedAt,
				})
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj(records, cmd.OutOrStdout())
		},
	}
}

func newGetUserCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <user-id>",
		Short: "Get a user's roles from etcd",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util ctl users get alice
  gns3util ctl users get alice -o json`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			u, err := client.GetUser(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&userRow{
				UserID:      u.UserID,
				Roles:       strings.Join(u.RoleNames, ", "),
				Permissions: u.Permissions,
				DenyScopes:  u.DenyScopes,
				UpdatedAt:   u.UpdatedAt,
			}}, cmd.OutOrStdout())
		},
	}
}

func newDeleteUserCmd(factory clientFactory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <user-id>",
		Short: "Delete a user's permission record from etcd",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util ctl users delete alice
  gns3util ctl users delete alice --force`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !force {
				msg := fmt.Sprintf("Delete user %q from etcd? This cannot be undone.", args[0])
				if !utils.ConfirmPrompt(msg, false) {
					return fmt.Errorf("operation cancelled")
				}
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			if err := client.DeleteUser(cmd.Context(), args[0]); err != nil {
				return fmt.Errorf("failed to delete user: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "User %q deleted.\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func newCreateUserCmd(factory clientFactory) *cobra.Command {
	var roles []string
	var denyScopeArgs []string

	cmd := &cobra.Command{
		Use:   "create <user-id>",
		Short: "Create a user in etcd and assign roles",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util ctl users create alice -c cluster
  gns3util ctl users create bob --role admin -c cluster`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			req := models.CreateUserRequest{
				Name:  args[0],
				Roles: roles,
			}
			denyScopes, err := parseScopeArgs(denyScopeArgs)
			if err != nil {
				return err
			}
			req.DenyScopes = denyScopes

			resp, err := client.CreateUser(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&userRow{
				UserID:      resp.UserID,
				Roles:       strings.Join(resp.RoleNames, ", "),
				Permissions: resp.Permissions,
				DenyScopes:  resp.DenyScopes,
				UpdatedAt:   resp.UpdatedAt,
			}}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringArrayVarP(&roles, "roles", "r", []string{}, "Roles to assign to the user (can specify multiple times)")
	cmd.Flags().StringArrayVar(&denyScopeArgs, "deny-scope", []string{}, "Deny scope to assign (format: action:resource, repeatable)")
	return cmd
}

func newAssignRoleCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "assign-role <user-id> <role-name>",
		Short: "Assign a role to an existing user",
		Args:  cobra.ExactArgs(2),
		Example: `  gns3util ctl users assign-role alice viewer
  gns3util ctl users assign-role bob admin -c cluster`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			resp, err := client.AssignRole(cmd.Context(), args[0], models.AssignRoleRequest{Role: args[1]})
			if err != nil {
				return fmt.Errorf("failed to assign role: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&userRow{
				UserID:      resp.UserID,
				Roles:       strings.Join(resp.RoleNames, ", "),
				Permissions: resp.Permissions,
				DenyScopes:  resp.DenyScopes,
				UpdatedAt:   resp.UpdatedAt,
			}}, cmd.OutOrStdout())
		},
	}
}

func parseScopeArgs(rawScopes []string) ([]models.ScopeInfo, error) {
	scopes := make([]models.ScopeInfo, 0, len(rawScopes))
	for _, raw := range rawScopes {
		parts := strings.SplitN(strings.TrimSpace(raw), ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid scope %q: expected action:resource", raw)
		}
		scopes = append(scopes, models.ScopeInfo{
			Action:   strings.ToLower(parts[0]),
			Resource: strings.ToLower(parts[1]),
		})
	}
	return scopes, nil
}

func newGenTokenCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "gen-token <user-id>",
		Short: "Generate a JWT token for a user",
		Long: `Mint a new JWT for a user via the REST API using your current admin session.
Unlike 'ctl create token', this does not require cluster_access.toml or a TLS
client certificate — any authenticated admin can call it.`,
		Args: cobra.ExactArgs(1),
		Example: `  gns3util ctl users gen-token alice
  gns3util ctl users gen-token alice -c cluster -o json`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			token, err := client.GenerateUserToken(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to generate token: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&tokenRow{
				UserID: args[0],
				Token:  token,
			}}, cmd.OutOrStdout())
		},
	}
}

func newRevokeTokenCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-token <jwt>",
		Short: "Revoke a JWT token",
		Long: `Revoke a token by pasting its raw JWT string.
The JTI is extracted from the token payload automatically — no manual decoding
needed. Once revoked the token is rejected on all future requests even if it
has not yet expired.`,
		Args: cobra.ExactArgs(1),
		Example: `  gns3util ctl users revoke-token eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9...
  gns3util ctl users revoke-token $TOKEN -c cluster`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			jti, err := extractJTI(args[0])
			if err != nil {
				return fmt.Errorf("invalid token: %w", err)
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			if err := client.RevokeToken(cmd.Context(), models.RevokeTokenRequest{JTI: jti}); err != nil {
				return fmt.Errorf("failed to revoke token: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Token (jti: %q) revoked.\n", jti)
			return nil
		},
	}
}

// extractJTI decodes a raw JWT string without verifying its signature and
// returns the value of the "jti" claim embedded in the payload.
func extractJTI(tokenStr string) (string, error) {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("expected 3 dot-separated parts, got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to base64-decode payload: %w", err)
	}

	var claims struct {
		JTI string `json:"jti"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse payload JSON: %w", err)
	}

	if claims.JTI == "" {
		return "", fmt.Errorf("token does not contain a jti claim")
	}

	return claims.JTI, nil
}
