package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/add"
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewPoolCmdGroup() *cobra.Command {
	poolCmd := &cobra.Command{
		Use:     "pool",
		Aliases: []string{"po"},
		Short:   "Resource pool operations",
		Long:    `Create, manage, and manipulate GNS3 resource pools.`,
	}

	poolCmd.AddCommand(create.NewCreatePoolCmd())

	poolCmd.AddCommand(get.NewGetPoolsCmd())
	poolCmd.AddCommand(get.NewGetPoolCmd())
	poolCmd.AddCommand(get.NewGetPoolResourcesCmd())

	poolCmd.AddCommand(update.NewUpdatePoolCmd())

	poolCmd.AddCommand(delete.NewDeletePoolCmd())
	poolCmd.AddCommand(delete.NewDeletePoolResourceCmd())

	poolCmd.AddCommand(add.NewAddToPoolCmd())

	return poolCmd
}
