package sharecmd

import (
	"github.com/spf13/cobra"
)

func NewDevicesCmd() *cobra.Command {
	var devicesCmd = &cobra.Command{
		Use:   "devices",
		Short: "List all devices that are shared with you",
		Long:  `List all devices that are shared with you`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	return devicesCmd
}
