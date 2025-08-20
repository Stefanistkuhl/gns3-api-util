package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [user-name/id]",
		Short: "Delete a user",
		Long:  `Delete a user from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete user john
gns3util -s https://controller:3080 delete user 123e4567-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			userID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(userID) {
				id, err := utils.ResolveID(cfg, "user", userID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				userID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteUser", []string{userID})
		},
	}

	return cmd
}
