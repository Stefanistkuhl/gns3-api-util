package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewTemplateCmdGroup() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"tmpl", "t"},
		Short:   "Template operations",
		Long:    `Create, manage, and manipulate GNS3 templates.`,
	}

	templateCmd.AddCommand(create.NewCreateTemplateCmd())

	templateCmd.AddCommand(get.NewGetTemplatesCmd())
	templateCmd.AddCommand(get.NewGetTemplateCmd())

	templateCmd.AddCommand(post.NewDuplicateTemplateCmd())

	templateCmd.AddCommand(update.NewUpdateTemplateCmd())

	templateCmd.AddCommand(delete.NewDeleteTemplateCmd())

	return templateCmd
}
