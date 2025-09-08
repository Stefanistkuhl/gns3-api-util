package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetTemplatesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get all templates of the Server",
		Long:    `Get all templates of the Server`,
		Example: "gns3util -s https://controller:3080 template ls",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getTemplates", nil)
		},
	}
	return cmd
}

func NewGetTemplateCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [template-name/id]",
		Short:   "Get a template by id or name",
		Long:    `Get a template by id or name`,
		Example: "gns3util -s https://controller:3080 template info my-template",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [template-name/id] when --fuzzy is not set")
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
				params := fuzzy.NewFuzzyInfoParams(cfg, "getTemplates", "template_id", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "template", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getTemplate", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a template")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple templates")
	return cmd
}
