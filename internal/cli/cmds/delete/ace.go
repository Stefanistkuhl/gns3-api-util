package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteACECmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [ace-id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete an ACE",
		Long:    `Delete an Access Control Entry (ACE) from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 acl delete ace-id",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			aceID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(aceID) {
				return fmt.Errorf("ACE ID must be a valid UUID")
			}

			utils.ExecuteAndPrint(cfg, "deleteACE", []string{aceID})
			return nil
		},
	}

	return cmd
}
