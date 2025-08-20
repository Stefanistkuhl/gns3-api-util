package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetStatisticsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "statistics",
		Short: "Get the statistics of the GNS3 Server",
		Long:  `Get the statistics of the GNS3 Server`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getMe", nil)
		},
	}
	return cmd
}
