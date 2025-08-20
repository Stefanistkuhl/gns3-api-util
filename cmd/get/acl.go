package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetAclCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "acl",
		Short: "Get the acl-rules of the GNS3 Server",
		Long:  `Get the acl-rules of the GNS3 Server`,
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
		Use:   "ace",
		Short: "Get a ace by id",
		Long:  `Get a ace by id`,
		Args:  cobra.ExactArgs(1),
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
		Use:   "acl-endpoints",
		Short: "Get the avaliable endpoints for acl-rules",
		Long:  `Get the avaliable endpoints for acl-rules`,
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
