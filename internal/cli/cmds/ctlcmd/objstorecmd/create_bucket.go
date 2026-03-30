package objstorecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func NewCreateBucketCmd() *cobra.Command {
	var (
		isPublic       bool
		requiredScopes string
		fileStoreName  string
	)

	cmd := &cobra.Command{
		Use:   "create-bucket <name>",
		Short: "Create a new bucket in the filestore",
		Long: `Create a new bucket in the cluster filestore.

The authenticated user becomes the bucket owner automatically. Only the owner
can delete the bucket. Use --public to make the bucket readable without a token,
and --required-scopes to gate access to a specific JWT scope string.

Requires ACTION_WRITE:RESOURCE_FILES scope on the cluster.`,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Example: `  gns3util ctl object-store create-bucket my-bucket -c mycluster
  gns3util ctl obj create-bucket shared --public -c mycluster
  gns3util ctl obj create-bucket private --required-scopes "write:files" -c mycluster`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			fs, token, err := resolveFilestoreAndToken(cfg)
			if err != nil {
				return err
			}

			resp, err := helpers.RunCreateBucket(fs.URL, token, args[0], isPublic, requiredScopes, cfg)
			if err != nil {
				return err
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return fmt.Errorf("printer setup failed: %w", err)
			}

			return printer.PrintObj([]utils.TableRecord{&BucketRow{
				FilestoreID: fs.ID,
				BucketID:    resp.BucketID,
				Name:        resp.Name,
				IsPublic:    resp.IsPublic,
				CreatedAt:   resp.CreatedAt,
			}}, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&isPublic, "public", false, "Make the bucket publicly readable without a token")
	cmd.Flags().StringVar(&requiredScopes, "required-scopes", "", "JWT scope string required to access files in this bucket")
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID (optional when only one filestore exists)")

	return cmd
}
