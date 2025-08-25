package get

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetPrivilegesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "privileges-all",
		Short:   "Get all privileges",
		Long:    `Get all privileges from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 role privileges-all`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				cmd.PrintErrf("failed to get global options: %v\n", err)
				return
			}

			utils.ExecuteAndPrint(cfg, "getPrivileges", nil)
		},
	}

	return cmd
}
