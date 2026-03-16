package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetStatisticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "statistics",
		Aliases: []string{"stats", "st"},
		Short:   "Get the statistics of the GNS3 Server",
		Long:    `Get the statistics of the GNS3 Server`,
		Example: "gns3util -s https://controller:3080 get statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getStatistics", nil)
			return nil
		},
	}
	return cmd
}
