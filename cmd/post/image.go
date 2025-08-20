package post

import (
	"github.com/spf13/cobra"
)

func NewImageCmdGroup() *cobra.Command {
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Image operations",
		Long:  `Image operations for managing GNS3 images.`,
	}

	return imageCmd
}
