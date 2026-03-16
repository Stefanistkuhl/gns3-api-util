package delete

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewDeleteImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [image-path]",
		Aliases: []string{"remove", "rm", "del"},
		Short:   "Delete an image",
		Long:    `Delete an image from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 image delete /path/to/image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			imageID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			utils.ExecuteAndPrint(cfg, "deleteImage", []string{imageID})
			return nil
		},
	}

	return cmd
}
