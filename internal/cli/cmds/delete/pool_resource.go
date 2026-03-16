package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeletePoolResourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource [pool-name/id] [resource-id]",
		Aliases: []string{"res", "pr"},
		Short:   "Delete a resource from a pool",
		Long:    `Delete a resource from a pool on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 pool resource my-pool resource-id",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			poolID := args[0]
			resourceID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(poolID) {
				id, err := utils.ResolveID(cfg, "pool", poolID, nil)
				if err != nil {
					return err
				}
				poolID = id
			}

			if !utils.IsValidUUIDv4(resourceID) {
				return fmt.Errorf("resource ID must be a valid UUID")
			}

			utils.ExecuteAndPrint(cfg, "deletePoolResource", []string{poolID, resourceID})
			return nil
		},
	}

	return cmd
}
