package class

import (
	"github.com/spf13/cobra"
)

func NewClassCmdGroup() *cobra.Command {
	var classCmd = &cobra.Command{
		Use:   "class",
		Short: "Class operations",
		Long:  `Create, manage, and manipulate GNS3 classes.`,
	}

	classCmd.AddCommand(
		NewClassCreateCmd(),
		NewClassDeleteCmd(),
	)

	return classCmd
}
