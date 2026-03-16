package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeletePruneImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prune",
		Aliases: []string{"clean"},
		Short:   "Prune unused images",
		Long:    `Delete unused images from the GNS3 server to free up disk space.`,
		Example: `gns3util -s https://controller:3080 image prune`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			utils.ExecuteAndPrint(cfg, "deletePruneImages", nil)
			return nil
		},
	}

	return cmd
}
