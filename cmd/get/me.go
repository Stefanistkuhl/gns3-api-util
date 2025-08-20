package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetMeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "me",
		Short: "Display info about the currenly logged in user on the GNS3 Server",
		Long:  `Display info about the currenly logged in user on the GNS3 Server`,
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
