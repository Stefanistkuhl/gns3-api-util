package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [template-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a template",
		Long:    `Delete a template from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 template delete my-template",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(templateID) {
				id, err := utils.ResolveID(cfg, "template", templateID, nil)
				if err != nil {
					return err
				}
				templateID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteTemplate", []string{templateID})
			return nil
		},
	}

	return cmd
}
