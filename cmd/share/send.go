package share

import (
	"github.com/spf13/cobra"
)

func NewSendCmd() *cobra.Command {
	var sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send a device to another user",
		Long:  `Send a device to another user`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	return sendCmd
}
