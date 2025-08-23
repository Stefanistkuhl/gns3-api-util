package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteACECmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [ace-id]",
		Short:   "Delete an ACE",
		Long:    `Delete an Access Control Entry (ACE) from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 acl delete ace-id",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			aceID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(aceID) {
				fmt.Println("ACE ID must be a valid UUID")
				return
			}

			utils.ExecuteAndPrint(cfg, "deleteACE", []string{aceID})
		},
	}

	return cmd
}
