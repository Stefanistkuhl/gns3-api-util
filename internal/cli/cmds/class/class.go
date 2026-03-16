package class

import (
	"github.com/spf13/cobra"
)

func NewClassCmdGroup() *cobra.Command {
	classCmd := &cobra.Command{
		Use:     "class",
		Aliases: []string{"c", "cls"},
		Short:   "Class operations",
		Long:    `Create, manage, and manipulate GNS3 classes.`,
	}

	classCmd.AddCommand(
		NewCreateClassCmd(),
		NewClassDeleteCmd(),
		NewClassLsCmd(),
	)

	return classCmd
}
