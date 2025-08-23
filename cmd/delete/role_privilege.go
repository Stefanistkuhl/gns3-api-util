package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteRolePrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "role-privilege [role-name/id] [privilege-id]",
		Short:   "Delete a privilege from a role",
		Long:    `Delete a privilege from a role on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 role role-privilege my-role privilege-id",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			roleID := args[0]
			privilegeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(roleID) {
				id, err := utils.ResolveID(cfg, "role", roleID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				roleID = id
			}

			if !utils.IsValidUUIDv4(privilegeID) {
				fmt.Println("Privilege ID must be a valid UUID")
				return
			}

			utils.ExecuteAndPrint(cfg, "deleteRolePrivilege", []string{roleID, privilegeID})
		},
	}

	return cmd
}
