package add

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewAddToPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "to-pool [pool-name/id] [project-name/id]",
		Aliases: []string{"add", "join", "atp"},
		Short:   "Add a resource to a pool",
		Long:    `Add a resource (like a project) to a pool on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 pool to-pool my-pool my-project",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			poolID := args[0]
			projectID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(poolID) {
				id, err := utils.ResolveID(cfg, "pool", poolID, nil)
				if err != nil {
					return err
				}
				poolID = id
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					return err
				}
				projectID = id
			}

			utils.ExecuteAndPrint(cfg, "addToPool", []string{poolID, projectID})
			return nil
		},
	}

	return cmd
}
