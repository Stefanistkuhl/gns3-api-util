package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/ctlcmd"
	"github.com/spf13/cobra"
)

func NewCtlCmdGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ctl",
		Short: "gns3util control plane operations",
		Long:  `gns3util control plane operations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(ctlcmd.NewCreateCmd())
	cmd.AddCommand(ctlcmd.NewAuthCmd())
	cmd.AddCommand(ctlcmd.NewAddClusterCMD())
	cmd.AddCommand(ctlcmd.NewAddDiscoverCMD())
	return cmd
}
