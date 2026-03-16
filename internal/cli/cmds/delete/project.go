package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [project-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a project",
		Long:    `Delete a project from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project delete my-project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					return err
				}
				projectID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteProject", []string{projectID})
			return nil
		},
	}

	return cmd
}
