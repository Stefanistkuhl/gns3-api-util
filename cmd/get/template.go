package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
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
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [template-name/id]",
		Short:   "Get a template by id or name",
		Long:    `Get a template by id or name`,
		Example: "gns3util -s https://controller:3080 template info my-template",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "template", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getTemplate", []string{id})
		},
	}
	return cmd
}
