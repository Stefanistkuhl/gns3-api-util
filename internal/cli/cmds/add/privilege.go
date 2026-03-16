package add

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewAddPrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "privilege [role-name/id] [privilege-name/id]",
		Aliases: []string{"add", "join", "ap"},
		Short:   "Add a privilege to a role",
		Long:    `Add a privilege to a role on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 role privilege my-role my-privilege",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleID := args[0]
			privilegeID := args[1]
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

			if !utils.IsValidUUIDv4(privilegeID) {
				return fmt.Errorf("privilege ID must be a valid UUID")
			}

			utils.ExecuteAndPrint(cfg, "addPrivilege", []string{roleID, privilegeID})
			return nil
		},
	}

	return cmd
}
