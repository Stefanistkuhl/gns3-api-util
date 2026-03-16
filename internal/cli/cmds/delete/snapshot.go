package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [project-name/id] [snapshot-name/id]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete a snapshot from a project",
		Long:    `Delete a snapshot from a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project snapshot delete my-project my-snapshot",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			snapshotID := args[1]
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

			if !utils.IsValidUUIDv4(snapshotID) {
				return fmt.Errorf("snapshot ID must be a valid UUID")
			}

			utils.ExecuteAndPrint(cfg, "deleteSnapshot", []string{projectID, snapshotID})
			return nil
		},
	}

	return cmd
}
