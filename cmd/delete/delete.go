package delete

import (
	"github.com/spf13/cobra"
)

func NewDeleteCmdGroup() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete GNS3 resources",
		Long:  `Delete various GNS3 resources like users, projects, templates, etc.`,
	}

	deleteCmd.AddCommand(
		NewDeletePruneImagesCmd(),
		NewDeleteUserCmd(),
		NewDeleteComputeCmd(),
		NewDeleteProjectCmd(),
		NewDeleteTemplateCmd(),
		NewDeleteImageCmd(),
		NewDeleteACECmd(),
		NewDeleteRoleCmd(),
		NewDeleteGroupCmd(),
		NewDeletePoolCmd(),
		NewDeletePoolResourceCmd(),
		NewDeleteLinkCmd(),
		NewDeleteNodeCmd(),
		NewDeleteDrawingCmd(),
		NewDeleteRolePrivilegeCmd(),
		NewDeleteUserFromGroupCmd(),
		NewDeleteSnapshotCmd(),
	)

	return deleteCmd
}
