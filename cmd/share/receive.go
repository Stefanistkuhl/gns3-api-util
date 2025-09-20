package share

import (
	"github.com/spf13/cobra"
)

func NewReceiveCmd() *cobra.Command {
	var receiveCmd = &cobra.Command{
		Use:   "receive",
		Short: "Receive a device from another user",
		Long:  `Receive a device from another user`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	return receiveCmd
}
