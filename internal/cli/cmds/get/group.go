package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetGroupCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [group-name/id]",
		Aliases: []string{"get", "i"},
		Short:   "Get a group by id or name",
		Long:    `Get a group by id or name`,
		Example: "gns3util -s https://controller:3080 group info my-group",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [group-name/id] when --fuzzy is not set")
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
				params := fuzzy.NewFuzzyInfoParams(cfg, "getGroups", "name", multi)
				err = fuzzy.FuzzyInfo(params)
				if err != nil {
					return err
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "group", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getGroup", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a group")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple groups")
	return cmd
}

func NewGetGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get all groups of the Server",
		Long:    `Get all groups of the Server`,
		Example: "gns3util -s https://controller:3080 group ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getGroups", nil)
			return nil
		},
	}
	return cmd
}

func NewGetGroupMembersCmd() *cobra.Command {
	var multi bool
	var useFuzzy bool
	cmd := &cobra.Command{
		Use:     "members [group-name/id]",
		Aliases: []string{"users", "mb"},
		Short:   "Get the members of a group by id or name",
		Long:    `Get the members of a group by id or name`,
		Example: "gns3util -s https://controller:3080 group members my-group",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [group-name/id] when --fuzzy is not set")
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
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getGroups", "name", multi, "group", "Group:")
				ids, fuzzyErr := fuzzy.FuzzyInfoIDs(params)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				groupMembers, membersErr := utils.GetResourceWithContext(cfg, "getGroupMembers", ids, "group", "Group:")
				if membersErr != nil {
					return fmt.Errorf("error getting group members: %w", membersErr)
				}

				utils.PrintResourceWithContext(groupMembers, "Group:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(id) {
					id, err = utils.ResolveID(cfg, "group", id, nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getGroupMembers", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a group")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple groups")
	return cmd
}
