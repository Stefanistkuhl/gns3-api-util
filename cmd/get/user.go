package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetUserCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Get a user by id or name",
		Long:  `Get a user by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "user", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getUser", []string{id})
		},
	}
	return cmd
}

func NewGetUsersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "users",
		Short: "Get a user by id or name",
		Long:  `Get a user by id or name`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getUsers", nil)
		},
	}
	return cmd
}

func NewGetGroupMembershipsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "group-membership",
		Short: "Get the group memberships of a user by id or name",
		Long:  `Get the group memberships of a user by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "user", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getGroupMemberships", []string{id})
		},
	}
	return cmd
}
