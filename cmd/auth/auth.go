package auth

import (
	"github.com/spf13/cobra"
)

func NewAuthCmdGroup() *cobra.Command {
	var authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  `Authentication commands`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	authCmd.AddCommand(NewAuthStatusCmd())
	authCmd.AddCommand(NewAuthLoginCmd())

	return authCmd
}
