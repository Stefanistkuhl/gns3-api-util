package post

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewControllerCmdGroup() *cobra.Command {
	controllerCmd := &cobra.Command{
		Use:   "controller",
		Short: "Controller operations",
		Long:  `Controller operations for managing the GNS3 server.`,
	}

	controllerCmd.AddCommand(
		NewReloadCmd(),
		NewShutdownCmd(),
	)

	return controllerCmd
}

func NewReloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reload",
		Short:   "Reload the controller",
		Long:    `Reload the GNS3 controller.`,
		Example: `gns3util -s https://controller:3080 post controller reload`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			utils.ExecuteAndPrint(cfg, "reloadController", nil)
		},
	}

	return cmd
}

func NewShutdownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shutdown",
		Short:   "Shutdown the server",
		Long:    `Shutdown the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post controller shutdown`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			utils.ExecuteAndPrint(cfg, "shutdownController", nil)
		},
	}

	return cmd
}
