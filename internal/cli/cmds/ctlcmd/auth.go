package ctlcmd

import "github.com/spf13/cobra"

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication related operations",
		Long:  `Authentication related operations change this text later`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(NewAuthStoreCmd())
	return cmd
}

func NewAuthStoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Store auth credentials",
		Long:  `Store auth credentials the keyfile`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	return cmd
}

//auth status ts goes to master to check
// auth store to store key in keyfile or ig new keyfile bc i am too scared of migrating it
