package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetRolesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get the roles of the Server",
		Long:    `Get the roles of the Server`,
		Example: "gns3util -s https://controller:3080 role ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getRoles", nil)
			return nil
		},
	}
	return cmd
}

func NewGetRoleCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [role-name/id]",
		Aliases: []string{"get", "i"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getRoles", "name", multi)
				err = fuzzy.FuzzyInfo(params)
				if err != nil {
					return err
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "role", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getRole", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a role")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple roles")
	return cmd
}

func NewGetRolePrivsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "privileges [role-name/id]",
		Aliases: []string{"privs", "perms"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getRoles", "name", multi, "role", "Role:")
				ids, fuzzyErr := fuzzy.FuzzyInfoIDs(params)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				rolePrivs, privsErr := utils.GetResourceWithContext(cfg, "getRolePrivs", ids, "role", "Role:")
				if privsErr != nil {
					return fmt.Errorf("error getting role privileges: %w", privsErr)
				}

				utils.PrintResourceWithContext(rolePrivs, "Role:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "role", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getRolePrivs", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a role")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple roles")
	return cmd
}
