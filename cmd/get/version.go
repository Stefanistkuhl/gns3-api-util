package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetVersionCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "version",
		Short:   "Get the version of the GNS3 Server",
		Long:    `Get the version of the GNS3 Server`,
		Example: "gns3util -s https://controller:3080 get version",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getVersion", nil)
		},
	}
	return cmd
}
