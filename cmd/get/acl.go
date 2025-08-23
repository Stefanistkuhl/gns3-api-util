package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetAclCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get the acl-rules of the GNS3 Server",
		Long:    `Get the acl-rules of the GNS3 Server`,
		Example: "gns3util -s https://controller:3080 acl ls",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getAcl", nil)
		},
	}
	return cmd
}

func NewGetAceCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [ace-id]",
		Short:   "Get an ace by id",
		Long:    `Get an ace by id`,
		Example: "gns3util -s https://controller:3080 acl info ace-id",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getAce", []string{id})
		},
	}
	return cmd
}

func NewGetAclEndpointsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "endpoints",
		Short:   "Get the available endpoints for acl-rules",
		Long:    `Get the available endpoints for acl-rules`,
		Example: "gns3util -s https://controller:3080 acl endpoints",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getAclEndpoints", nil)
		},
	}
	return cmd
}
