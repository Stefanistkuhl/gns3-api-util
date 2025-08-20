package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewAclCmdGroup() *cobra.Command {
	aclCmd := &cobra.Command{
		Use:   "acl",
		Short: "ACL operations",
		Long:  `Create, manage, and manipulate GNS3 ACL rules.`,
	}

	// Create subcommands
	aclCmd.AddCommand(create.NewCreateACLCmd())

	// Get subcommands
	aclCmd.AddCommand(get.NewGetAclCmd())
	aclCmd.AddCommand(get.NewGetAceCmd())
	aclCmd.AddCommand(get.NewGetAclEndpointsCmd())

	// Update subcommands
	aclCmd.AddCommand(update.NewUpdateACECmd())

	// Delete subcommands
	aclCmd.AddCommand(delete.NewDeleteACECmd())

	return aclCmd
}
