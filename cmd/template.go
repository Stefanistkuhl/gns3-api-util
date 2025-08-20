package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewTemplateCmdGroup() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Template operations",
		Long:  `Create, manage, and manipulate GNS3 templates.`,
	}

	// Create subcommands
	templateCmd.AddCommand(create.NewCreateTemplateCmd())

	// Get subcommands
	templateCmd.AddCommand(get.NewGetTemplatesCmd())
	templateCmd.AddCommand(get.NewGetTemplateCmd())

	// Post subcommands
	templateCmd.AddCommand(post.NewDuplicateTemplateCmd())

	// Update subcommands
	templateCmd.AddCommand(update.NewUpdateTemplateCmd())

	// Delete subcommands
	templateCmd.AddCommand(delete.NewDeleteTemplateCmd())

	return templateCmd
}
