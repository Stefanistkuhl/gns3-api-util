package post

import (
	"github.com/spf13/cobra"
)

func NewSnapshotCmdGroup() *cobra.Command {
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot operations",
		Long:  `Snapshot operations for managing GNS3 snapshots.`,
	}

	return snapshotCmd
}
