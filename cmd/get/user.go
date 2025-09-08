package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetUserCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [user-name/id]",
		Short:   "Get a user by id or name",
		Long:    `Get a user by id or name`,
		Example: "gns3util -s https://controller:3080 user my-user",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}

			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [user-name/id] when --fuzzy is not set")
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
				fmt.Printf("failed to get global options: %v\n", err)
				return
			}

			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getUsers", "username", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "user", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getUser", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Enable fuzzy search mode for interactive selection")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Allow selecting multiple items (requires --fuzzy)")
	return cmd
}

func NewGetUsersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get all users",
		Long:    `Get all users`,
		Example: "gns3util -s https://controller:3080 user ls",
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
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     "group-membership [user-name/id]",
		Short:   "Get the group memberships of a user by id or name",
		Long:    `Get the group memberships of a user by id or name`,
		Example: "gns3util -s https://controller:3080 user group-membership my-user",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}

			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [user-name/id] when --fuzzy is not set")
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
				return
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getUsers", "username", multi, "user", "User:")
				ids, err := fuzzy.FuzzyInfoIDs(params)
				if err != nil {
					fmt.Println(err)
					return
				}

				userMemberships, err := utils.GetResourceWithContext(cfg, "getGroupMemberships", ids, "user", "User:")
				if err != nil {
					fmt.Printf("Error getting group memberships: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(userMemberships, "User:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "user", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getGroupMemberships", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Enable fuzzy search mode for interactive selection")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Allow selecting multiple items (requires --fuzzy)")
	return cmd
}
