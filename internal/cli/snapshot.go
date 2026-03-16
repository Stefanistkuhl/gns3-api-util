package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/delete"
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post/create"
	"github.com/spf13/cobra"
)

func NewSnapshotCmdGroup() *cobra.Command {
	snapshotCmd := &cobra.Command{
		Use:     "snapshot",
		Aliases: []string{"snap", "sn"},
		Short:   "Snapshot operations",
		Long:    `Create, manage, and manipulate GNS3 snapshots.`,
	}

	snapshotCmd.AddCommand(create.NewCreateSnapshotCmd())

	snapshotCmd.AddCommand(get.NewGetSnapshotsCmd())

	snapshotCmd.AddCommand(delete.NewDeleteSnapshotCmd())

	return snapshotCmd
}
