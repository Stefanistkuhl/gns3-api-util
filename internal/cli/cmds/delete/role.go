package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteRoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [role-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a role",
		Long:    `Delete a role from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 role delete my-role",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(roleID) {
				id, err := utils.ResolveID(cfg, "role", roleID, nil)
				if err != nil {
					return err
				}
				roleID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteRole", []string{roleID})
			return nil
		},
	}

	return cmd
}
