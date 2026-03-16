package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetAclCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get the acl-rules of the GNS3 Server",
		Long:    `Get the acl-rules of the GNS3 Server`,
		Example: "gns3util -s https://controller:3080 acl ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getAcl", nil)
			return nil
		},
	}
	return cmd
}

func NewGetAceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [ace-id]",
		Aliases: []string{"get", "i"},
		Short:   "Get an ace by id",
		Long:    `Get an ace by id`,
		Example: "gns3util -s https://controller:3080 acl info ace-id",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getAce", []string{id})
			return nil
		},
	}
	return cmd
}

func NewGetAclEndpointsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "endpoints",
		Aliases: []string{"ends", "ep"},
		Short:   "Get the available endpoints for acl-rules",
		Long:    `Get the available endpoints for acl-rules`,
		Example: "gns3util -s https://controller:3080 acl endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getAclEndpoints", nil)
			return nil
		},
	}
	return cmd
}
