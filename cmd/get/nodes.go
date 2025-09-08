package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/tidwall/gjson"
)

func NewGetNodesCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", multi, "project", "Project:")
				ids, err := fuzzy.FuzzyInfoIDs(params)
				if err != nil {
					fmt.Println(err)
					return
				}

				projectNodes, err := utils.GetResourceWithContext(cfg, "getNodes", ids, "project", "Project:")
				if err != nil {
					fmt.Printf("Error getting nodes: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(projectNodes, "Project:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodes", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple projects")
	return cmd
}

func NewGetNodeCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [node-name/id]",
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, err := fuzzy.FuzzyInfoIDs(projectParams)
				if err != nil {
					fmt.Println(err)
					return
				}

				if len(projectIDs) == 0 {
					fmt.Println("No project selected")
					return
				}

				rawData, _, err := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if err != nil {
					fmt.Printf("Error getting nodes: %v\n", err)
					return
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					fmt.Println("Expected array response")
					return
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					fmt.Println("No nodes found in selected project")
					return
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
						fmt.Println(err)
						return
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getNode", []string{projectID, nodeID})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}

func NewGetNodeLinksCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, err := fuzzy.FuzzyInfoIDs(projectParams)
				if err != nil {
					fmt.Println(err)
					return
				}

				if len(projectIDs) == 0 {
					fmt.Println("No project selected")
					return
				}

				rawData, _, err := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if err != nil {
					fmt.Printf("Error getting nodes: %v\n", err)
					return
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					fmt.Println("Expected array response")
					return
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					fmt.Println("No nodes found in selected project")
					return
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
						fmt.Println(err)
						return
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodeLinks", []string{projectID, nodeID})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}

func NewGetNodesAutoIdlePCCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, err := fuzzy.FuzzyInfoIDs(projectParams)
				if err != nil {
					fmt.Println(err)
					return
				}

				if len(projectIDs) == 0 {
					fmt.Println("No project selected")
					return
				}

				rawData, _, err := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if err != nil {
					fmt.Printf("Error getting nodes: %v\n", err)
					return
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					fmt.Println("Expected array response")
					return
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					fmt.Println("No nodes found in selected project")
					return
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
						fmt.Println(err)
						return
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePc", []string{projectID, nodeID})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}

func NewGetNodesAutoIdlePCProposalsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				projectParams := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getProjects", "name", false, "project", "Project:")
				projectIDs, err := fuzzy.FuzzyInfoIDs(projectParams)
				if err != nil {
					fmt.Println(err)
					return
				}

				if len(projectIDs) == 0 {
					fmt.Println("No project selected")
					return
				}

				rawData, _, err := utils.CallClient(cfg, "getNodes", []string{projectIDs[0]}, nil)
				if err != nil {
					fmt.Printf("Error getting nodes: %v\n", err)
					return
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					fmt.Println("Expected array response")
					return
				}

				var nodeIDs []string

				result.ForEach(func(_, value gjson.Result) bool {
					if nodeID := value.Get("node_id"); nodeID.Exists() {
						nodeIDs = append(nodeIDs, nodeID.String())
					}
					return true
				})

				if len(nodeIDs) == 0 {
					fmt.Println("No nodes found in selected project")
					return
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
						fmt.Println(err)
						return
					}
				}
				if !utils.IsValidUUIDv4(args[1]) {
					nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePcProposals", []string{projectID, nodeID})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project and node")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple nodes")
	return cmd
}
