package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetGroupCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "group",
		Short: "Get a group by id or name",
		Long:  `Get a group by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "group", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getGroup", []string{id})
		},
	}
	return cmd
}

func NewGetGroupsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "groups",
		Short: "Get all groups of the Server",
		Long:  `Get all groups of the Server`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getGroups", nil)
		},
	}
	return cmd
}

func NewGetGroupMembersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "group-members",
		Short: "Get the members of a group by id or name",
		Long:  `Get the members of a group by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "group", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getGroupMembers", []string{id})
		},
	}
	return cmd
}
