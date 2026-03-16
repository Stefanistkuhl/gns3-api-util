package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

func NewGetNodesCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Aliases: []string{"list", "l"},
		Short:   "Get the nodes within a project by name or id",
		Long:    `Get the nodes within a project by name or id`,
		Example: "gns3util -s https://controller:3080 node ls my-project",
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

				projectNodes, projectErr := utils.GetResourceWithContext(cfg, "getNodes", ids, "project", "Project:")
				if projectErr != nil {
					return fmt.Errorf("error getting nodes: %w", projectErr)
				}

				utils.PrintResourceWithContext(projectNodes, "Project:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodes", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple projects")
	return cmd
}

func NewGetNodeCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [node-name/id]",
		Aliases: []string{"get", "i"},
		Short:   "Get a node in a project by name or id",
		Long:    `Get a node in a project by name or id`,
		Example: "gns3util -s https://controller:3080 node info my-project my-node",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [node-name/id] when --fuzzy is not set")
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

				rawData, _, apiErr := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting nodes: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					return fmt.Errorf("no nodes found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(nodeIDs, multi)

				for _, nodeID := range results {
					utils.ExecuteAndPrint(cfg, "getNode", []string{projectIDs[0], nodeID})
				}
			} else {
				projectID := args[0]
				nodeID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getNode", []string{projectID, nodeID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}

func NewGetNodeLinksCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "links [project-name/id] [node-name/id]",
		Short:   "Get links of a given node in a project by id or name",
		Long:    `Get links of a given node in a project by id or name`,
		Example: "gns3util -s https://controller:3080 node links my-project my-node",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [node-name/id] when --fuzzy is not set")
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

				rawData, _, apiErr := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting nodes: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					return fmt.Errorf("no nodes found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(nodeIDs, multi)

				for _, nodeID := range results {
					utils.ExecuteAndPrint(cfg, "getNodeLinks", []string{projectIDs[0], nodeID})
				}
			} else {
				projectID := args[0]
				nodeID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodeLinks", []string{projectID, nodeID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}

func NewGetNodesAutoIdlePCCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "auto-idle-pc [project-name/id] [node-name/id]",
		Short:   "Get the auto-idle-pc of a node in a project by id or name",
		Long:    `Get the auto-idle-pc of a node in a project by id or name`,
		Example: "gns3util -s https://controller:3080 node auto-idle-pc my-project my-node",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [node-name/id] when --fuzzy is not set")
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

				rawData, _, apiErr := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting nodes: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					return fmt.Errorf("no nodes found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(nodeIDs, multi)

				for _, nodeID := range results {
					utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePc", []string{projectIDs[0], nodeID})
				}
			} else {
				projectID := args[0]
				nodeID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePc", []string{projectID, nodeID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}

func NewGetNodesAutoIdlePCProposalsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "auto-idle-pc-proposals [project-name/id] [node-name/id]",
		Short:   "Get the auto-idle-pc-proposals of a node in a project by id or name",
		Long:    `Get the auto-idle-pc-proposals of a node in a project by id or name`,
		Example: "gns3util -s https://controller:3080 node auto-idle-pc-proposals my-project my-node",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 2 {
					return fmt.Errorf("at most 2 positional args allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args [project-name/id] [node-name/id] when --fuzzy is not set")
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

				rawData, _, apiErr := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if apiErr != nil {
					return fmt.Errorf("error getting nodes: %w", apiErr)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					return fmt.Errorf("no nodes found in selected project")
				}

				results := fuzzy.NewFuzzyFinder(nodeIDs, multi)

				for _, nodeID := range results {
					utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePcProposals", []string{projectIDs[0], nodeID})
				}
			} else {
				projectID := args[0]
				nodeID := args[1]
				if !utils.IsValidUUIDv4(args[0]) {
					projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePcProposals", []string{projectID, nodeID})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}
