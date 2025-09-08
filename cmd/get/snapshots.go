package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetSnapshotsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Short:   "Get the snapshots within a project by name or id",
		Long:    `Get the snapshots within a project by name or id`,
		Example: "gns3util -s https://controller:3080 snapshot ls my-project",
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

				snapshots, err := utils.GetResourceWithContext(cfg, "getSnapshots", ids, "project", "Project:")
				if err != nil {
					fmt.Printf("Error getting snapshots: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(snapshots, "Project:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "project", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getSnapshots", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get snapshots from multiple projects")
	return cmd
}
