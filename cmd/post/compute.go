package post

import (
	"github.com/spf13/cobra"
)

func NewComputeCmdGroup() *cobra.Command {
	computeCmd := &cobra.Command{
		Use:   "compute",
		Short: "Compute operations",
		Long:  `Compute operations for managing GNS3 computes.`,
	}

	return computeCmd
}
