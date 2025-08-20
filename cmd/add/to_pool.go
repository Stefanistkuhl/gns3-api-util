package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewAddToPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-pool [pool-name/id] [project-name/id]",
		Short: "Add a resource to a pool",
		Long:  `Add a resource (like a project) to a pool on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 add to-pool my-pool my-project
gns3util -s https://controller:3080 add to-pool 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			poolID := args[0]
			projectID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(poolID) {
				id, err := utils.ResolveID(cfg, "pool", poolID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				poolID = id
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
			}

			utils.ExecuteAndPrint(cfg, "addToPool", []string{poolID, projectID})
		},
	}

	return cmd
}
