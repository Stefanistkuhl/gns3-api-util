package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewProjectCmdGroup() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"proj", "p"},
		Short:   "Project operations",
		Long:    `Create, manage, and manipulate GNS3 projects.`,
	}

	projectCmd.AddCommand(create.NewCreateProjectCmd())

	projectCmd.AddCommand(get.NewGetProjectCmd())
	projectCmd.AddCommand(get.NewGetProjectsCmd())
	projectCmd.AddCommand(get.NewGetProjectExportCmd())
	projectCmd.AddCommand(get.NewGetProjectFileCmd())
	projectCmd.AddCommand(get.NewGetProjectLockedCmd())
	projectCmd.AddCommand(get.NewGetProjectStatsCmd())

	projectCmd.AddCommand(post.NewProjectCloseCmd())
	projectCmd.AddCommand(post.NewProjectDuplicateCmd())
	projectCmd.AddCommand(post.NewProjectImportCmd())
	projectCmd.AddCommand(post.NewProjectLoadCmd())
	projectCmd.AddCommand(post.NewProjectLockCmd())
	projectCmd.AddCommand(post.NewProjectOpenCmd())
	projectCmd.AddCommand(post.NewProjectUnlockCmd())
	projectCmd.AddCommand(post.NewProjectWriteFileCmd())
	projectCmd.AddCommand(post.NewProjectStartCaptureCmd())

	projectCmd.AddCommand(update.NewUpdateProjectCmd())

	projectCmd.AddCommand(delete.NewDeleteProjectCmd())

	return projectCmd
}
