package ctlcmd

import (
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources in the control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	createCmd.AddCommand(NewCreateUserCmd())
	createCmd.AddCommand(NewCreateTokenCmd())

	return createCmd
}
