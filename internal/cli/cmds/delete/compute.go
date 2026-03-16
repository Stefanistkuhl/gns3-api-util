package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [compute-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a compute",
		Long:    `Delete a compute from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 compute delete my-compute",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			computeID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(computeID) {
				id, err := utils.ResolveID(cfg, "compute", computeID, nil)
				if err != nil {
					return err
				}
				computeID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteCompute", []string{computeID})
			return nil
		},
	}

	return cmd
}
