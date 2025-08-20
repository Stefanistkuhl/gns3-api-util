package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image [image-name/id]",
		Short: "Delete an image",
		Long:  `Delete an image from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 delete image my-image
gns3util -s https://controller:3080 delete image 123e4567-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			imageID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			utils.ExecuteAndPrint(cfg, "deleteImage", []string{imageID})
		},
	}

	return cmd
}
