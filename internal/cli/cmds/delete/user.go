package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [user-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a user",
		Long:    `Delete a user from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 user delete my-user",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(userID) {
				id, err := utils.ResolveID(cfg, "user", userID, nil)
				if err != nil {
					return err
				}
				userID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteUser", []string{userID})
			return nil
		},
	}

	return cmd
}
