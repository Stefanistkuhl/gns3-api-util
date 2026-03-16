package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeletePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [pool-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a pool",
		Long:    `Delete a pool from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 pool delete my-pool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			poolID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(poolID) {
				id, err := utils.ResolveID(cfg, "pool", poolID, nil)
				if err != nil {
					return err
				}
				poolID = id
			}

			utils.ExecuteAndPrint(cfg, "deletePool", []string{poolID})
			return nil
		},
	}

	return cmd
}
