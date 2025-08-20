package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeletePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [pool-name/id]",
		Short: "Delete a pool",
		Long:  `Delete a pool from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete pool my-pool
gns3util -s https://controller:3080 delete pool 123e4567-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			poolID := args[0]
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

			utils.ExecuteAndPrint(cfg, "deletePool", []string{poolID})
		},
	}

	return cmd
}
