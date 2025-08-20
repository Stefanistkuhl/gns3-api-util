package post

import (
	"github.com/spf13/cobra"
)

func NewPostCmdGroup() *cobra.Command {
	postCmd := &cobra.Command{
		Use:   "post",
		Short: "Misc post commands",
		Long:  `Miscellaneous POST commands for GNS3 operations.`,
	}

	postCmd.AddCommand(
		NewCheckVersionCmd(),
		NewUserAuthenticateCmd(),
		NewControllerCmdGroup(),
		NewComputeCmdGroup(),
		NewImageCmdGroup(),
		NewLinkCmdGroup(),
		NewNodeCmdGroup(),
		NewProjectCmdGroup(),
		NewTemplateCmdGroup(),
		NewSnapshotCmdGroup(),
	)

	return postCmd
}
