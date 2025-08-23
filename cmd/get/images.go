package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetImagesCmd() *cobra.Command {
	var (
		imageType string
	)
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get the available images on the Server",
		Long:    `Get the available images on the Server`,
		Example: "gns3util -s https://controller:3080 image ls",
		Run: func(cmd *cobra.Command, args []string) {
			if imageType != "" && imageType != "qemu" && imageType != "ios" && imageType != "iou" {
				fmt.Println("The image type can only be qemu, ios, or iou")
				return
			}
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getImages", []string{imageType})
		},
	}
	cmd.Flags().StringVarP(&imageType, "image-type", "t", "", "What type of image to get (qemu/ios/iou)")
	return cmd
}

func NewGetImageCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [image-path]",
		Short:   "Get an image by path",
		Long:    `Get an image by path`,
		Example: "gns3util -s https://controller:3080 image info /path/to/image",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getImage", []string{path})
		},
	}
	return cmd
}
