package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewACLCmdGroup() *cobra.Command {
	aclCmd := &cobra.Command{
		Use:     "acl",
		Aliases: []string{"ac"},
		Short:   "ACL operations",
		Long:    `Create, manage, and manipulate GNS3 ACL rules.`,
	}

	aclCmd.AddCommand(create.NewCreateACLCmd())

	aclCmd.AddCommand(get.NewGetAclCmd())
	aclCmd.AddCommand(get.NewGetAceCmd())
	aclCmd.AddCommand(get.NewGetAclEndpointsCmd())

	aclCmd.AddCommand(update.NewUpdateACECmd())

	aclCmd.AddCommand(delete.NewDeleteACECmd())

	return aclCmd
}
