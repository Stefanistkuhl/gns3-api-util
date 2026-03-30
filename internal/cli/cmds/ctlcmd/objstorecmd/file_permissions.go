package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func NewFilePermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file-permissions",
		Short: "Manage access permissions on files",
		Long:  "Grant, revoke, and list delegated permissions on individual files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newGrantFilePermCmd(),
		newRevokeFilePermCmd(),
		newListFilePermsCmd(),
	)

	return cmd
}

func newGrantFilePermCmd() *cobra.Command {
	var (
		principalType string
		principalID   string
		permission    string
		expiresAt     string
	)

	cmd := &cobra.Command{
		Use:   "grant <file-uuid>",
		Short: "Grant a permission on a file",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util object-store file-permissions grant <file-uuid> \
    --principal-type role --principal-id viewer --permission read`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fs, token, err := resolveFilestoreAndToken(cfg)
			if err != nil {
				return err
			}

			req := models.GrantPermissionRequest{
				PrincipalType: principalType,
				PrincipalID:   principalID,
				Permission:    permission,
				ExpiresAt:     expiresAt,
			}

			entry, err := helpers.RunGrantFilePermission(fs.URL, token, args[0], req, cfg)
			if err != nil {
				return err
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed: %w", err)
			}

			return printer.PrintObj([]utils.TableRecord{&permissionRow{
				ID:            entry.ID,
				PrincipalType: entry.PrincipalType,
				PrincipalID:   entry.PrincipalID,
				Permission:    entry.Permission,
				GrantedBy:     entry.GrantedBy,
				CreatedAt:     entry.CreatedAt,
				ExpiresAt:     entry.ExpiresAt,
			}}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&principalType, "principal-type", "", "Principal type: user, group, or role (required)")
	cmd.Flags().StringVar(&principalID, "principal-id", "", "ID of the user/group/role to grant access to (required)")
	cmd.Flags().StringVar(&permission, "permission", "", "Permission level: read, write, or admin (required)")
	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "Optional expiry timestamp (RFC3339)")
	_ = cmd.MarkFlagRequired("principal-type")
	_ = cmd.MarkFlagRequired("principal-id")
	_ = cmd.MarkFlagRequired("permission")

	return cmd
}

func newRevokeFilePermCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke <file-uuid> <permission-id>",
		Short:   "Revoke a permission entry from a file",
		Args:    cobra.ExactArgs(2),
		Example: `  gns3util object-store file-permissions revoke <file-uuid> <perm-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fs, token, err := resolveFilestoreAndToken(cfg)
			if err != nil {
				return err
			}

			resp, err := helpers.RunRevokeFilePermission(fs.URL, token, args[0], args[1], cfg)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Permission %s deleted: %v\n", resp.ID, resp.Deleted)
			return nil
		},
	}

	return cmd
}

func newListFilePermsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <file-uuid>",
		Short:   "List all permissions on a file",
		Args:    cobra.ExactArgs(1),
		Example: `  gns3util object-store file-permissions list <file-uuid>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fs, token, err := resolveFilestoreAndToken(cfg)
			if err != nil {
				return err
			}

			resp, err := helpers.RunListFilePermissions(fs.URL, token, args[0], cfg)
			if err != nil {
				return err
			}

			var records []utils.TableRecord
			for _, e := range resp.Permissions {
				records = append(records, &permissionRow{
					ID:            e.ID,
					PrincipalType: e.PrincipalType,
					PrincipalID:   e.PrincipalID,
					Permission:    e.Permission,
					GrantedBy:     e.GrantedBy,
					CreatedAt:     e.CreatedAt,
					ExpiresAt:     e.ExpiresAt,
				})
			}

			if len(records) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No permissions found.")
				return nil
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed: %w", err)
			}

			return printer.PrintObj(records, cmd.OutOrStdout())
		},
	}

	return cmd
}
