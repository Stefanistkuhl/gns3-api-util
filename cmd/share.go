package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/share"
)

func NewShareCmdGroup() *cobra.Command {
	shareCmd := &cobra.Command{
		Use:   "share",
		Short: "share operations",
		Long:  `Share your configuration in the lan with other users.`,
	}
	shareCmd.AddCommand(share.NewDevicesCmd())
	shareCmd.AddCommand(share.NewDiscoverCmd())
	shareCmd.AddCommand(share.NewReceiveCmd())
	shareCmd.AddCommand(share.NewSendCmd())

	return shareCmd
}
