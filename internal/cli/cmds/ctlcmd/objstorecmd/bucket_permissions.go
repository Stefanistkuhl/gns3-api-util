package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type permissionRow struct {
	ID            string `json:"id"`
	PrincipalType string `json:"principal_type"`
	PrincipalID   string `json:"principal_id"`
	Permission    string `json:"permission"`
	GrantedBy     string `json:"granted_by"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

func (p *permissionRow) GetHeaders() []string {
	return []string{"ID", "PRINCIPAL TYPE", "PRINCIPAL ID", "PERMISSION", "GRANTED BY", "CREATED AT", "EXPIRES AT"}
}

func (p *permissionRow) GetRow() []string {
	return []string{
		p.ID,
		p.PrincipalType,
		p.PrincipalID,
		p.Permission,
		p.GrantedBy,
		p.CreatedAt,
		p.ExpiresAt,
	}
}

func resolveFilestoreAndToken(cfg *config.GlobalOptions) (*pathutils.ServiceEntry, string, error) {
	clusterName := cfg.Cluster
	keyPath, err := pathutils.ResolveKeyFilePath("")
	if err != nil {
		return nil, "", err
	}
	kf, err := pathutils.LoadGNS3KeysFile(keyPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load keys: %w", err)
	}

	var targetCluster *pathutils.ClusterEntry
	for i := range kf.Clusters {
		if kf.Clusters[i].Name == clusterName {
			targetCluster = &kf.Clusters[i]
			break
		}
	}
	if targetCluster == nil {
		return nil, "", fmt.Errorf("cluster %q not found", clusterName)
	}
	cfg.ClusterEntry.CaCert = targetCluster.CaCert

	fs, err := selectFilestore(targetCluster, "")
	if err != nil {
		return nil, "", err
	}
	return fs, cfg.ClusterEntry.Master.AccessToken, nil
}

func NewBucketPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucket-permissions",
		Short: "Manage access permissions on buckets",
		Long:  "Grant, revoke, and list delegated permissions on buckets.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newGrantBucketPermCmd(),
		newRevokeBucketPermCmd(),
		newListBucketPermsCmd(),
	)

	return cmd
}

func newGrantBucketPermCmd() *cobra.Command {
	var (
		principalType string
		principalID   string
		permission    string
		expiresAt     string
	)

	cmd := &cobra.Command{
		Use:   "grant <bucket-id>",
		Short: "Grant a permission on a bucket",
		Args:  cobra.ExactArgs(1),
		Example: `  gns3util object-store bucket-permissions grant <bucket-id> \
    --principal-type user --principal-id user123 --permission read`,
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

			entry, err := helpers.RunGrantBucketPermission(fs.URL, token, args[0], req, cfg)
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

func newRevokeBucketPermCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke <bucket-id> <permission-id>",
		Short:   "Revoke a permission entry from a bucket",
		Args:    cobra.ExactArgs(2),
		Example: `  gns3util object-store bucket-permissions revoke <bucket-id> <perm-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fs, token, err := resolveFilestoreAndToken(cfg)
			if err != nil {
				return err
			}

			resp, err := helpers.RunRevokeBucketPermission(fs.URL, token, args[0], args[1], cfg)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Permission %s deleted: %v\n", resp.ID, resp.Deleted)
			return nil
		},
	}

	return cmd
}

func newListBucketPermsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <bucket-id>",
		Short:   "List all permissions on a bucket",
		Args:    cobra.ExactArgs(1),
		Example: `  gns3util object-store bucket-permissions list <bucket-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fs, token, err := resolveFilestoreAndToken(cfg)
			if err != nil {
				return err
			}

			resp, err := helpers.RunListBucketPermissions(fs.URL, token, args[0], cfg)
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
