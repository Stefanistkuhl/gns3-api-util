package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [compute-name/id]",
		Short: "Delete a compute",
		Long:  `Delete a compute from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete compute my-compute
gns3util -s https://controller:3080 delete compute 123e4567-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			computeID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(computeID) {
				id, err := utils.ResolveID(cfg, "compute", computeID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				computeID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteCompute", []string{computeID})
		},
	}

	return cmd
}
