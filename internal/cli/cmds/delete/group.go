package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [group-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a group",
		Long:    `Delete a group from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 group delete my-group",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(groupID) {
				id, err := utils.ResolveID(cfg, "group", groupID, nil)
				if err != nil {
					return err
				}
				groupID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteGroup", []string{groupID})
			return nil
		},
	}

	return cmd
}
