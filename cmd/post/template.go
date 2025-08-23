package post

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewTemplateCmdGroup() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Template operations",
		Long:  `Template operations for managing GNS3 templates.`,
	}

	templateCmd.AddCommand(
		NewDuplicateTemplateCmd(),
	)

	return templateCmd
}

func NewDuplicateTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "duplicate [template-name/id]",
		Short:   "Duplicate a template",
		Long:    `Duplicate a template on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 template duplicate my-template",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			templateID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(templateID) {
				id, err := utils.ResolveID(cfg, "template", templateID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				templateID = id
			}

			utils.ExecuteAndPrint(cfg, "duplicateTemplate", []string{templateID})
		},
	}

	return cmd
}
