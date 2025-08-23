package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeletePruneImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prune",
		Short:   "Prune unused images",
		Long:    `Delete unused images from the GNS3 server to free up disk space.`,
		Example: `gns3util -s https://controller:3080 image prune`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			utils.ExecuteAndPrint(cfg, "deletePruneImages", nil)
		},
	}

	return cmd
}
