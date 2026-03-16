package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

func NewGetLinksCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Aliases: []string{"list", "l"},
		Short:   "Get the links within a project by name or id",
		Long:    `Get the links within a project by name or id`,
		Example: "gns3util -s https://controller:3080 link ls my-project",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [project-name/id] when --fuzzy is not set")
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
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", multi, "project", "Project:")
				ids, fuzzyErr := fuzzy.FuzzyInfoIDs(params)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				projectLinks, projectErr := utils.GetResourceWithContext(cfg, "getLinks", ids, "project", "Project:")
				if projectErr != nil {
					return fmt.Errorf("error getting links: %w", projectErr)
				}

				utils.PrintResourceWithContext(projectLinks, "Project:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getLinks", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple projects")
	return cmd
}

func NewGetLinkCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [link-name/id]",
		Aliases: []string{"get", "i"},
		Short:   "Get a link within a project by name or id",
		Long:    `Get a link within a project by name or id`,
		Example: "gns3util -s https://controller:3080 link info my-project my-link",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [link-name/id] when --fuzzy is not set")
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
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, fuzzyErr := fuzzy.FuzzyInfoIDs(projectParams)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				if len(projectIDs) == 0 {
					return fmt.Errorf("no project selected")
				}

				rawData, _, apiErr := utils.CallClient(cfg, "getLinks", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting links: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var linkIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if linkID := value.Get("link_id"); linkID.Exists() {
						linkIDs = append(linkIDs, linkID.String())
					}
					return true
				})

				if len(linkIDs) == 0 {
					return fmt.Errorf("no links found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(linkIDs, multi)

				for _, linkID := range results {
					utils.ExecuteAndPrint(cfg, "getLink", []string{projectIDs[0], linkID})
				}
			} else {
				projectID := args[0]
				linkID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getLink", []string{projectID, linkID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and link")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple links")
	return cmd
}

func NewGetLinkIfaceCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "iface [project-name/id] [link-name/id]",
		Short:   "Get the interface of a link within a project by name or id for Cloud or NAT devices.",
		Long:    `Return iface info for links to Cloud or NAT devices for a link in a project by id or name.`,
		Example: "gns3util -s https://controller:3080 link iface my-project my-link",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [link-name/id] when --fuzzy is not set")
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
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, fuzzyErr := fuzzy.FuzzyInfoIDs(projectParams)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				if len(projectIDs) == 0 {
					return fmt.Errorf("no project selected")
				}

				rawData, _, apiErr := utils.CallClient(cfg, "getLinks", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting links: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var linkIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if linkID := value.Get("link_id"); linkID.Exists() {
						linkIDs = append(linkIDs, linkID.String())
					}
					return true
				})

				if len(linkIDs) == 0 {
					return fmt.Errorf("no links found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(linkIDs, multi)

				for _, linkID := range results {
					utils.ExecuteAndPrint(cfg, "getLinkIface", []string{projectIDs[0], linkID})
				}
			} else {
				projectID := args[0]
				linkID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getLinkIface", []string{projectID, linkID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and link")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple links")
	return cmd
}

func NewGetLinkFiltersCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "filters [project-name/id] [link-name/id]",
		Short:   "Get the filters for a link within a project by name or id",
		Long:    `Get the filters for a link within a project by name or id`,
		Example: "gns3util -s https://controller:3080 link filters my-project my-link",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [link-name/id] when --fuzzy is not set")
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
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, fuzzyErr := fuzzy.FuzzyInfoIDs(projectParams)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				if len(projectIDs) == 0 {
					return fmt.Errorf("no project selected")
				}

				rawData, _, apiErr := utils.CallClient(cfg, "getLinks", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting links: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var linkIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if linkID := value.Get("link_id"); linkID.Exists() {
						linkIDs = append(linkIDs, linkID.String())
					}
					return true
				})

				if len(linkIDs) == 0 {
					return fmt.Errorf("no links found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(linkIDs, multi)

				for _, linkID := range results {
					utils.ExecuteAndPrint(cfg, "getLinkFilters", []string{projectIDs[0], linkID})
				}
			} else {
				projectID := args[0]
				linkID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getLinkFilters", []string{projectID, linkID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and link")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple links")
	return cmd
}
