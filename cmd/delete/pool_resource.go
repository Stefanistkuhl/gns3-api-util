package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeletePoolResourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-resource [pool-name/id] [resource-id]",
		Short: "Delete a resource from a pool",
		Long:  `Delete a resource from a pool on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete pool-resource my-pool 123e4567-e89b-12d3-a456-426614174000
gns3util -s https://controller:3080 delete pool-resource 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			poolID := args[0]
			resourceID := args[1]
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

			if !utils.IsValidUUIDv4(resourceID) {
				fmt.Println("Resource ID must be a valid UUID")
				return
			}

			utils.ExecuteAndPrint(cfg, "deletePoolResource", []string{poolID, resourceID})
		},
	}

	return cmd
}
