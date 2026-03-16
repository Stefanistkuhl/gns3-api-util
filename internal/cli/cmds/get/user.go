package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetUserCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [user-name/id]",
		Aliases: []string{"get", "i"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getUsers", "username", multi)
				err = fuzzy.FuzzyInfo(params)
				if err != nil {
					return err
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "user", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getUser", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Enable fuzzy search mode for interactive selection")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Allow selecting multiple items (requires --fuzzy)")
	return cmd
}

func NewGetUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get all users",
		Long:    `Get all users`,
		Example: "gns3util -s https://controller:3080 user ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getUsers", nil)
			return nil
		},
	}
	return cmd
}

func NewGetGroupMembershipsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "group-membership [user-name/id]",
		Aliases: []string{"groups", "gm"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getUsers", "username", multi, "user", "User:")
				ids, fuzzyErr := fuzzy.FuzzyInfoIDs(params)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				userMemberships, membershipErr := utils.GetResourceWithContext(cfg, "getGroupMemberships", ids, "user", "User:")
				if membershipErr != nil {
					return fmt.Errorf("error getting group memberships: %w", membershipErr)
				}

				utils.PrintResourceWithContext(userMemberships, "User:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "user", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getGroupMemberships", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Enable fuzzy search mode for interactive selection")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Allow selecting multiple items (requires --fuzzy)")
	return cmd
}
