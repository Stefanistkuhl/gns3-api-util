package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/remote/install"
	"github.com/stefanistkuhl/gns3util/cmd/remote/uninstall"
)

func NewRemoteCmdGroup() *cobra.Command {
	remoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "remote openrations via SSH",
		Long:  `Any actions that arent over the API and instead run over SSH directly on the server`,
		Run:   func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
	}
	remoteCmd.AddCommand(install.NewInstallCmd())
	remoteCmd.AddCommand(uninstall.NewUninstallCmdGroup())
	return remoteCmd
}
