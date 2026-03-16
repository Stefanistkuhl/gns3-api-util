package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewComputeCmdGroup() *cobra.Command {
	computeCmd := &cobra.Command{
		Use:     "compute",
		Aliases: []string{"comp"},
		Short:   "Compute operations",
		Long:    `Create, manage, and manipulate GNS3 computes.`,
	}

	computeCmd.AddCommand(create.NewCreateComputeCmd())

	computeCmd.AddCommand(get.NewGetComputesCmd())
	computeCmd.AddCommand(get.NewGetComputeCmd())
	computeCmd.AddCommand(get.NewGetComputeDockerImagesCmd())
	computeCmd.AddCommand(get.NewGetComputeVirtualboxVMSCmd())
	computeCmd.AddCommand(get.NewGetComputeVmWareVMSCmd())

	computeCmd.AddCommand(update.NewUpdateComputeCmd())

	computeCmd.AddCommand(delete.NewDeleteComputeCmd())

	return computeCmd
}
