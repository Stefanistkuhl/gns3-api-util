package add

import (
	"github.com/spf13/cobra"
)

func NewAddCmdGroup() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add GNS3 resources",
		Long:  `Add various GNS3 resources like users to groups, resources to pools, etc.`,
	}

	addCmd.AddCommand(
		NewAddGroupMemberCmd(),
		NewAddToPoolCmd(),
		NewAddPrivilegeCmd(),
	)

	return addCmd
}
