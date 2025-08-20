package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/add"
	"github.com/stefanistkuhl/gns3util/cmd/delete"
	"github.com/stefanistkuhl/gns3util/cmd/get"
	"github.com/stefanistkuhl/gns3util/cmd/post/create"
	"github.com/stefanistkuhl/gns3util/cmd/put/update"
)

func NewPoolCmdGroup() *cobra.Command {
	poolCmd := &cobra.Command{
		Use:   "pool",
		Short: "Resource pool operations",
		Long:  `Create, manage, and manipulate GNS3 resource pools.`,
	}

	// Create subcommands
	poolCmd.AddCommand(create.NewCreatePoolCmd())

	// Get subcommands
	poolCmd.AddCommand(get.NewGetPoolsCmd())
	poolCmd.AddCommand(get.NewGetPoolCmd())
	poolCmd.AddCommand(get.NewGetPoolResourcesCmd())

	// Update subcommands
	poolCmd.AddCommand(update.NewUpdatePoolCmd())

	// Delete subcommands
	poolCmd.AddCommand(delete.NewDeletePoolCmd())
	poolCmd.AddCommand(delete.NewDeletePoolResourceCmd())

	// Add subcommands
	poolCmd.AddCommand(add.NewAddToPoolCmd())

	return poolCmd
}
