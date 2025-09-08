package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetDrawingsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Short:   "Get the drawings within a project by name or id",
		Long:    `Get the drawings within a project by name or id`,
		Example: "gns3util -s https://controller:3080 drawing ls my-project",
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

				projectDrawings, err := utils.GetResourceWithContext(cfg, "getDrawings", ids, "project", "Project:")
				if err != nil {
					fmt.Printf("Error getting drawings: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(projectDrawings, "Project:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getDrawings", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple projects")
	return cmd
}

func NewGetDrawingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [drawing-name/id]",
		Short:   "Get a drawing within a project by name or id",
		Long:    `Get a drawing within a project by name or id`,
		Example: "gns3util -s https://controller:3080 drawing info my-project my-drawing",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}

			projectID := args[0]
			linkID := args[1]
			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getDrawing", []string{projectID, linkID})

		},
	}
	return cmd
}
