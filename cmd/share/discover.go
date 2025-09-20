package share

import (
	"github.com/spf13/cobra"
)

func NewDiscoverCmd() *cobra.Command {
	var discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "Discover devices that are shared with you",
		Long:  `Discover devices that are shared with you`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	return discoverCmd
}
