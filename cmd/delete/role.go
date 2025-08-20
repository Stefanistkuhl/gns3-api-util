package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteRoleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role [role-name/id]",
		Short: "Delete a role",
		Long:  `Delete a role from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete role my-role
gns3util -s https://controller:3080 delete role 123e4567-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			roleID := args[0]
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

			utils.ExecuteAndPrint(cfg, "deleteRole", []string{roleID})
		},
	}

	return cmd
}
