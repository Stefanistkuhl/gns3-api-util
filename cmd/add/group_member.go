package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewAddGroupMemberCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-member [group-name/id] [user-name/id]",
		Short:   "Add a user to a group",
		Long:    `Add a user to a group on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 group add-member my-group my-user",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			groupID := args[0]
			userID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(groupID) {
				id, err := utils.ResolveID(cfg, "group", groupID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				groupID = id
			}

			if !utils.IsValidUUIDv4(userID) {
				id, err := utils.ResolveID(cfg, "user", userID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				userID = id
			}

			utils.ExecuteAndPrint(cfg, "addGroupMember", []string{groupID, userID})
		},
	}

	return cmd
}
