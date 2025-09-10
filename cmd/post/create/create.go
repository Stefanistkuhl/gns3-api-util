package create

import (
	"github.com/spf13/cobra"
)

func NewCreateCmdGroup() *cobra.Command {
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create GNS3 resources (e.g., users, projects)",
		Long:  "Commands for creating resources on a GNS3v3 controller, such as users and projects. Use subcommands to specify what to create.",
		Example: `
  # Create a simple user
  gns3util -s https://controller:3080 create user -u alice -p secret

  # Create a user with additional attributes
  gns3util -s https://controller:3080 create user -u alice -p secret --email alice@example.com --full-name "Alice Doe" --is-active

  # Create using a raw JSON payload
  gns3util -s https://controller:3080 create user --use-json '{"username":"alice","password":"secret","is_active":true}'
        `,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	createCmd.AddCommand(NewCreateUserCmd())
	createCmd.AddCommand(NewCreateGroupCmd())
	createCmd.AddCommand(NewCreateRoleCmd())
	createCmd.AddCommand(NewCreateProjectCmd())
	createCmd.AddCommand(NewCreateTemplateCmd())
	createCmd.AddCommand(NewCreateACLCmd())
	createCmd.AddCommand(NewCreateNodeCmd())
	createCmd.AddCommand(NewCreateNodeFromTemplateCmd())
	createCmd.AddCommand(NewCreateQemuImageCmd())
	createCmd.AddCommand(NewCreateQemuDiskImageCmd())
	createCmd.AddCommand(NewCreateLinkCmd())
	createCmd.AddCommand(NewCreateDrawingCmd())
	createCmd.AddCommand(NewCreateSnapshotCmd())
	createCmd.AddCommand(NewCreateComputeCmd())
	createCmd.AddCommand(NewCreatePoolCmd())
	createCmd.AddCommand(NewCreateClassCmd())
	createCmd.AddCommand(NewCreateExerciseCmd())

	return createCmd
}
