package auth

import (
	"github.com/spf13/cobra"
)

func NewAuthCmdGroup() *cobra.Command {
	authCmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"a"},
		Short:   "Authentication commands",
		Long:    `Authentication commands`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	authCmd.AddCommand(NewAuthStatusCmd())
	authCmd.AddCommand(NewAuthLoginCmd())

	return authCmd
}
