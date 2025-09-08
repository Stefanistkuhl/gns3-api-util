package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetGroupCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [group-name/id]",
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getGroups", "name", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "group", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getGroup", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a group")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple groups")
	return cmd
}

func NewGetGroupsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get all groups of the Server",
		Long:    `Get all groups of the Server`,
		Example: "gns3util -s https://controller:3080 group ls",
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
	var multi bool
	var useFuzzy bool
	var cmd = &cobra.Command{
		Use:     "members [group-name/id]",
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getGroups", "name", multi, "group", "Group:")
				ids, err := fuzzy.FuzzyInfoIDs(params)
				if err != nil {
					fmt.Println(err)
					return
				}

				groupMembers, err := utils.GetResourceWithContext(cfg, "getGroupMembers", ids, "group", "Group:")
				if err != nil {
					fmt.Printf("Error getting group members: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(groupMembers, "Group:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(id) {
					id, err = utils.ResolveID(cfg, "group", id, nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getGroupMembers", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a group")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple groups")
	return cmd
}
