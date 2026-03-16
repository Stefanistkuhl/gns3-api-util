package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetDrawingsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Aliases: []string{"list", "l"},
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

				projectDrawings, drawingsErr := utils.GetResourceWithContext(cfg, "getDrawings", ids, "project", "Project:")
				if drawingsErr != nil {
					return fmt.Errorf("error getting drawings: %w", drawingsErr)
				}

				utils.PrintResourceWithContext(projectDrawings, "Project:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getDrawings", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple projects")
	return cmd
}

func NewGetDrawingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [drawing-name/id]",
		Aliases: []string{"get", "i"},
		Short:   "Get a drawing within a project by name or id",
		Long:    `Get a drawing within a project by name or id`,
		Example: "gns3util -s https://controller:3080 drawing info my-project my-drawing",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			projectID := args[0]
			linkID := args[1]
			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getDrawing", []string{projectID, linkID})
			return nil
		},
	}
	return cmd
}
