package post

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUserAuthenticateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user-authenticate",
		Short:   "Authenticate as a user",
		Long:    `Authenticate as a user on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post user-authenticate`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			utils.ExecuteAndPrint(cfg, "userAuthenticate", nil)
		},
	}

	return cmd
}
