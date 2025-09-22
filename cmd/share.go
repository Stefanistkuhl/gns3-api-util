package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/sharecmd"
)

func NewShareCmdGroup() *cobra.Command {
	shareCmd := &cobra.Command{
		Use:   "share",
		Short: "share operations",
		Long:  `Share your configuration in the lan with other users.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if server == "" {

			}
			return nil
		},
	}
	shareCmd.AddCommand(sharecmd.NewReceiveCmd())
	shareCmd.AddCommand(sharecmd.NewSendCmd())

	return shareCmd
}
