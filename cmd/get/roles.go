package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetRolesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get the roles of the Server",
		Long:    `Get the roles of the Server`,
		Example: "gns3util -s https://controller:3080 role ls",
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
		Use:     utils.ListSingleElementCmdName + " [role-name/id]",
		Short:   "Get a role by id or name",
		Long:    `Get a role by id or name`,
		Example: "gns3util -s https://controller:3080 role info my-role",
		Args:    cobra.ExactArgs(1),
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
		Use:     "privileges [role-name/id]",
		Short:   "Get the privileges of a role by id or name",
		Long:    `Get the privileges of a role by id or name`,
		Example: "gns3util -s https://controller:3080 role privileges my-role",
		Args:    cobra.ExactArgs(1),
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
