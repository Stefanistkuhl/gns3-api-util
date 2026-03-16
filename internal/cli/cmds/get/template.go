package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get all templates of the Server",
		Long:    `Get all templates of the Server`,
		Example: "gns3util -s https://controller:3080 template ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getTemplates", nil)
			return nil
		},
	}
	return cmd
}

func NewGetTemplateCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [template-name/id]",
		Aliases: []string{"get", "i"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getTemplates", "template_id", multi)
				err = fuzzy.FuzzyInfo(params)
				if err != nil {
					return err
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "template", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getTemplate", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a template")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple templates")
	return cmd
}
