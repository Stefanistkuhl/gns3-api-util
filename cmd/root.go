package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/auth"
	"github.com/stefanistkuhl/gns3util/cmd/class"
	"github.com/stefanistkuhl/gns3util/cmd/exercise"
	"github.com/stefanistkuhl/gns3util/pkg/config"
)

var (
	server   string
	keyFile  string
	insecure bool
	raw      bool
)

var Foo bool

var rootCmd = &cobra.Command{
	Use:   "gns3util",
	Short: "A utility for GNS3v3",
	Long:  `A utility for GNS3v3 for managing GNS3v3 projects and devices.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		opts := config.GlobalOptions{
			Server:   server,
			Insecure: insecure,
			KeyFile:  keyFile,
			Raw:      raw,
		}
		ctx := config.WithGlobalOptions(cmd.Context(), opts)
		cmd.SetContext(ctx)
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	cobra.OnFinalize()
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "GNS3v3 Server URL (required)")
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key-file", "k", "", "Set a location for a keyfile to use")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "i", false, "Ignore unsigned SSL-Certificates")
	rootCmd.PersistentFlags().BoolVarP(&raw, "raw", "", false, "Output all data in raw json")
	_ = rootCmd.MarkPersistentFlagRequired("server")

	rootCmd.AddCommand(auth.NewAuthCmdGroup())

	rootCmd.AddCommand(class.NewClassCmdGroup())
	rootCmd.AddCommand(exercise.NewExerciseCmdGroup())

	rootCmd.AddCommand(NewProjectCmdGroup())
	rootCmd.AddCommand(NewNodeCmdGroup())
	rootCmd.AddCommand(NewLinkCmdGroup())
	rootCmd.AddCommand(NewDrawingCmdGroup())
	rootCmd.AddCommand(NewTemplateCmdGroup())
	rootCmd.AddCommand(NewComputeCmdGroup())
	rootCmd.AddCommand(NewApplianceCmdGroup())
	rootCmd.AddCommand(NewImageCmdGroup())
	rootCmd.AddCommand(NewSymbolCmdGroup())

	rootCmd.AddCommand(NewUserCmdGroup())
	rootCmd.AddCommand(NewGroupCmdGroup())
	rootCmd.AddCommand(NewRoleCmdGroup())
	rootCmd.AddCommand(NewAclCmdGroup())

	rootCmd.AddCommand(NewPoolCmdGroup())
	rootCmd.AddCommand(NewSnapshotCmdGroup())

	rootCmd.AddCommand(NewSystemCmdGroup())

	rootCmd.AddCommand(NewRemoteCmdGroup())
}

func Execute() {
	_ = rootCmd.Execute()
}
