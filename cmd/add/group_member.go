package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewAddGroupMemberCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-member [group-name/id] [user-name/id]",
		Short: "Add a user to a group",
		Long:  `Add a user to a group on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 add group-member my-group my-user
gns3util -s https://controller:3080 add group-member 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
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
