package rbaccmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/models"
)

type roleRow struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Scopes      string `json:"scopes"`
	UpdatedAt   string `json:"updated_at"`
}

func (r *roleRow) GetHeaders() []string {
	return []string{"NAME", "DESCRIPTION", "SCOPES", "UPDATED AT"}
}

func (r *roleRow) GetRow() []string {
	return []string{r.Name, r.Description, r.Scopes, r.UpdatedAt}
}

func scopesString(scopes []models.ScopeInfo) string {
	parts := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if isSuperuserScope(s) {
			parts = append(parts, "superuser")
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s", s.Action, s.Resource))
	}
	return strings.Join(parts, ", ")
}

func isSuperuserScope(s models.ScopeInfo) bool {
	action := strings.ToLower(s.Action)
	resource := strings.ToLower(s.Resource)
	isAdminAction := action == "superuser" || action == "admin" || action == "action_admin"
	isGlobal := resource == "" || resource == "unspecified" || resource == "resource_unspecified"
	return isAdminAction && isGlobal
}

type clientFactory func(cfg *config.GlobalOptions) (rbacClient, error)

func NewRolesCmd() *cobra.Command {
	return newRolesCmdWithFactory(newMasterClient)
}

func newRolesCmdWithFactory(factory clientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roles",
		Short: "Manage control-plane roles in etcd",
		Long:  "List, inspect, create, update, and delete role definitions stored in etcd.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	queryGroup := &cobra.Group{ID: "query", Title: "Query Commands:"}
	actionGroup := &cobra.Group{ID: "action", Title: "Action Commands:"}
	cmd.AddGroup(queryGroup, actionGroup)

	listCmd := newListRolesCmd(factory)
	listCmd.GroupID = "query"
	getCmd := newGetRoleCmd(factory)
	getCmd.GroupID = "query"
	createCmd := newCreateRoleCmd(factory)
	createCmd.GroupID = "action"
	updateCmd := newUpdateRoleCmd(factory)
	updateCmd.GroupID = "action"
	deleteCmd := newDeleteRoleCmd(factory)
	deleteCmd.GroupID = "action"

	cmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)
	return cmd
}

func newListRolesCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:         "list",
		Short:       "List all roles in etcd",
		Example:     `  gns3util ctl roles list`,
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

			resp, err := client.ListRoles(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list roles: %w", err)
			}

			if len(resp.Roles) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No roles found.")
				return nil
			}

			var records []utils.TableRecord
			for _, r := range resp.Roles {
				records = append(records, &roleRow{
					Name:        r.Name,
					Description: r.Description,
					Scopes:      scopesString(r.Scopes),
					UpdatedAt:   r.UpdatedAt,
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

func newGetRoleCmd(factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:         "get <role-name>",
		Short:       "Get a role's details from etcd",
		Args:        cobra.ExactArgs(1),
		Example:     `  gns3util ctl roles get admin`,
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

			role, err := client.GetRole(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to get role: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&roleRow{
				Name:        role.Name,
				Description: role.Description,
				Scopes:      scopesString(role.Scopes),
				UpdatedAt:   role.UpdatedAt,
			}}, cmd.OutOrStdout())
		},
	}
}

func newCreateRoleCmd(factory clientFactory) *cobra.Command {
	var (
		description string
		scopesStr   string
	)

	cmd := &cobra.Command{
		Use:   "create <role-name>",
		Short: "Create a new role in etcd",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util ctl roles create viewer \
    --description "Read-only access" \
    --scopes "read:files,read:metrics"`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			scopes, err := parseScopeStr(scopesStr)
			if err != nil {
				return err
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			role, err := client.CreateRole(cmd.Context(), models.CreateRoleRequest{
				Name:        args[0],
				Description: description,
				Scopes:      scopes,
			})
			if err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&roleRow{
				Name:        role.Name,
				Description: role.Description,
				Scopes:      scopesString(role.Scopes),
				UpdatedAt:   role.UpdatedAt,
			}}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringVar(&scopesStr, "scopes", "", `Comma-separated action:resource pairs (e.g. "read:files,write:backups")`)
	return cmd
}

func newUpdateRoleCmd(factory clientFactory) *cobra.Command {
	var (
		description string
		scopesStr   string
	)

	cmd := &cobra.Command{
		Use:   "update <role-name>",
		Short: "Replace a role's description and scopes in etcd",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util ctl roles update viewer \
    --description "Updated read-only" \
    --scopes "read:files"`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			scopes, err := parseScopeStr(scopesStr)
			if err != nil {
				return err
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			role, err := client.UpdateRole(cmd.Context(), args[0], models.UpdateRoleRequest{
				Description: description,
				Scopes:      scopes,
			})
			if err != nil {
				return fmt.Errorf("failed to update role: %w", err)
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}
			return printer.PrintObj([]utils.TableRecord{&roleRow{
				Name:        role.Name,
				Description: role.Description,
				Scopes:      scopesString(role.Scopes),
				UpdatedAt:   role.UpdatedAt,
			}}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&scopesStr, "scopes", "", `Comma-separated action:resource pairs`)
	return cmd
}

func newDeleteRoleCmd(factory clientFactory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:         "delete <role-name>",
		Short:       "Delete a role from etcd",
		Args:        cobra.ExactArgs(1),
		Example:     `  gns3util ctl roles delete viewer --force`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !force {
				msg := fmt.Sprintf("Delete role %q from etcd? All users with this role will lose access.", args[0])
				if !utils.ConfirmPrompt(msg, false) {
					return fmt.Errorf("operation cancelled")
				}
			}

			client, err := factory(cfg)
			if err != nil {
				return err
			}

			if err := client.DeleteRole(cmd.Context(), args[0]); err != nil {
				return fmt.Errorf("failed to delete role: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Role %q deleted.\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func parseScopeStr(raw string) ([]models.ScopeInfo, error) {
	if raw == "" {
		return nil, nil
	}
	var scopes []models.ScopeInfo
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if lower == "superuser" || lower == "admin" {
			scopes = append(scopes, models.ScopeInfo{Action: "superuser", Resource: ""})
			continue
		}
		before, after, ok := strings.Cut(part, ":")
		if !ok {
			return nil, fmt.Errorf("invalid scope %q: expected action:resource (or use 'superuser' for full access)", part)
		}
		scopes = append(scopes, models.ScopeInfo{
			Action:   strings.TrimSpace(before),
			Resource: strings.TrimSpace(after),
		})
	}
	return scopes, nil
}
