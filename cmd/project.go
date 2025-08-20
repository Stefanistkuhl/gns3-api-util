package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewProjectCmdGroup() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Project operations",
		Long:  `Create, manage, and manipulate GNS3 projects.`,
	}

	// Create subcommands
	projectCmd.AddCommand(create.NewCreateProjectCmd())

	// Get subcommands
	projectCmd.AddCommand(get.NewGetProjectCmd())
	projectCmd.AddCommand(get.NewGetProjectsCmd())
	projectCmd.AddCommand(get.NewGetProjectExportCmd())
	projectCmd.AddCommand(get.NewGetProjectFileCmd())
	projectCmd.AddCommand(get.NewGetProjectLockedCmd())
	projectCmd.AddCommand(get.NewGetProjectStatsCmd())

	// Post subcommands
	projectCmd.AddCommand(post.NewProjectCloseCmd())
	projectCmd.AddCommand(post.NewProjectDuplicateCmd())
	projectCmd.AddCommand(post.NewProjectImportCmd())
	projectCmd.AddCommand(post.NewProjectLoadCmd())
	projectCmd.AddCommand(post.NewProjectLockCmd())
	projectCmd.AddCommand(post.NewProjectOpenCmd())
	projectCmd.AddCommand(post.NewProjectUnlockCmd())
	projectCmd.AddCommand(post.NewProjectWriteFileCmd())
	projectCmd.AddCommand(post.NewProjectStartCaptureCmd())

	// Update subcommands
	projectCmd.AddCommand(update.NewUpdateProjectCmd())

	// Delete subcommands
	projectCmd.AddCommand(delete.NewDeleteProjectCmd())

	return projectCmd
}
