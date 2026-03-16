package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetIouLicenseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "iou-license",
		Aliases: []string{"iou", "license"},
		Short:   "Get the iou-license of the GNS3 Server",
		Long:    `Get the iou-license of the GNS3 Server`,
		Example: "gns3util -s https://controller:3080 get iou-license",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getIouLicense", nil)
			return nil
		},
	}
	return cmd
}
