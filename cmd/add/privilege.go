package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewAddPrivilegeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "privilege [role-name/id] [privilege-name/id]",
		Short: "Add a privilege to a role",
		Long:  `Add a privilege to a role on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 add privilege my-role my-privilege
gns3util -s https://controller:3080 add privilege 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
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

			utils.ExecuteAndPrint(cfg, "addPrivilege", []string{roleID, privilegeID})
		},
	}

	return cmd
}
