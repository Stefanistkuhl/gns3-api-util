package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [template-name/id]",
		Short:   "Delete a template",
		Long:    `Delete a template from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 template delete my-template",
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

			utils.ExecuteAndPrint(cfg, "deleteTemplate", []string{templateID})
		},
	}

	return cmd
}
