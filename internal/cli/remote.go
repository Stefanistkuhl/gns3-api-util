package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/remote/install"
	"github.com/0xveya/gns3util/internal/cli/cmds/remote/uninstall"
	"github.com/spf13/cobra"
)

func NewRemoteCmdGroup() *cobra.Command {
	remoteCmd := &cobra.Command{
		Use:     "remote",
		Aliases: []string{"rem"},
		Short:   "remote openrations via SSH",
		Long:    `Any actions that arent over the API and instead run over SSH directly on the server`,
		RunE:    func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	remoteCmd.AddCommand(install.NewInstallCmd())
	remoteCmd.AddCommand(uninstall.NewUninstallCmdGroup())
	return remoteCmd
}
