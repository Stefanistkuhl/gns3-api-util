package update

import (
	"github.com/spf13/cobra"
)

func NewUpdateCmdGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update GNS3 resources",
		Long:    "Update various GNS3 resources like users, groups, projects, nodes, etc.",
		Example: "gns3util -s https://controller:3080 update user [user-name/id] --username newname",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(
		NewUpdateIOULicenseCmd(),
		NewUpdateMeCmd(),
		NewUpdateUserCmd(),
		NewUpdateGroupCmd(),
		NewUpdateRoleCmd(),
		NewUpdateACECmd(),
		NewUpdateProjectCmd(),
		NewUpdateNodeCmd(),
		NewUpdateQemuDiskImageCmd(),
		NewUpdateLinkCmd(),
		NewUpdateDrawingCmd(),
		NewUpdateComputeCmd(),
		NewUpdatePoolCmd(),
		NewUpdateTemplateCmd(),
	)

	return cmd
}
