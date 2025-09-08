package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
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
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [role-name/id]",
		Short:   "Get a role by id or name",
		Long:    `Get a role by id or name`,
		Example: "gns3util -s https://controller:3080 role info my-role",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [role-name/id] when --fuzzy is not set")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if multi && !useFuzzy {
				return fmt.Errorf("the --multi (-m) flag can only be used together with --fuzzy (-f)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getRoles", "name", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "role", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getRole", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a role")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple roles")
	return cmd
}

func NewGetRolePrivsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     "privileges [role-name/id]",
		Short:   "Get the privileges of a role by id or name",
		Long:    `Get the privileges of a role by id or name`,
		Example: "gns3util -s https://controller:3080 role privileges my-role",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [role-name/id] when --fuzzy is not set")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if multi && !useFuzzy {
				return fmt.Errorf("the --multi (-m) flag can only be used together with --fuzzy (-f)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getRoles", "name", multi, "role", "Role:")
				ids, err := fuzzy.FuzzyInfoIDs(params)
				if err != nil {
					fmt.Println(err)
					return
				}

				rolePrivs, err := utils.GetResourceWithContext(cfg, "getRolePrivs", ids, "role", "Role:")
				if err != nil {
					fmt.Printf("Error getting role privileges: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(rolePrivs, "Role:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "role", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getRolePrivs", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a role")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple roles")
	return cmd
}
