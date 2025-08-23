package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeletePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [pool-name/id]",
		Short:   "Delete a pool",
		Long:    `Delete a pool from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 pool delete my-pool",
		Args:    cobra.ExactArgs(1),
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
