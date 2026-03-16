package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/sharecmd"
	"github.com/spf13/cobra"
)

func NewShareCmdGroup() *cobra.Command {
	shareCmd := &cobra.Command{
		Use:     "share",
		Aliases: []string{"sh"},
		Short:   "share operations",
		Long:    `Share your configuration in the lan with other users.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Server is optional for share commands
			return nil
		},
	}
	shareCmd.AddCommand(sharecmd.NewReceiveCmd())
	shareCmd.AddCommand(sharecmd.NewSendCmd())

	return shareCmd
}
