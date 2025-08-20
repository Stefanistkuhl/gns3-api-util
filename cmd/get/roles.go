package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetRolesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "roles",
		Short: "Get the roles of the Server",
		Long:  `Get the roles of the Server`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getRoles", nil)
		},
	}
	return cmd
}

func NewGetRoleCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "role",
		Short: "Get a role by id or name",
		Long:  `Get a role by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "role", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getRole", []string{id})
		},
	}
	return cmd
}

func NewGetRolePrivsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "role-privileges",
		Short: "Get get the privileges of a role by id or name",
		Long:  `Get get the privileges of a role by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "role", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getRolePrivs", []string{id})
		},
	}
	return cmd
}
