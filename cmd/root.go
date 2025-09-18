package cmd

import (
	"fmt"
	"os"

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
	noColor  bool
	version  bool
)

var Version = "1.1.0"

var Foo bool

var rootCmd = &cobra.Command{
	Use:   "gns3util",
	Short: "A utility for GNS3v3",
	Long:  `A utility for GNS3v3 for managing GNS3v3 projects and devices.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "completion" || (cmd.Parent() != nil && cmd.Parent().Name() == "completion") {
			return nil
		}

		if version {
			return nil
		}

		if err := validateGlobalFlags(); err != nil {
			return err
		}

		if err := validateRequiresServer(); err != nil {
			return err
		}

		opts := config.GlobalOptions{
			Server:   server,
			Insecure: insecure,
			KeyFile:  keyFile,
			Raw:      raw,
		}
		ctx := config.WithGlobalOptions(cmd.Context(), opts)
		cmd.SetContext(ctx)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Printf("gns3util version %s\n", Version)
			os.Exit(0)
		}
		_ = cmd.Help()
	},
}

func init() {
	cobra.OnFinalize()
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "GNS3v3 Server URL (required for non cluster commands)")
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key-file", "k", "", "Set a location for a keyfile to use")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "i", false, "Ignore unsigned SSL-Certificates")
	rootCmd.PersistentFlags().BoolVarP(&raw, "raw", "", false, "Output all data in raw json")
	rootCmd.PersistentFlags().BoolVarP(&noColor, "no-color", "", false, "Output all data in raw json and dont use a colored output")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "Print version information")

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

	rootCmd.AddCommand(NewClusterCmdGroup())
}

func Execute() {
	_ = rootCmd.Execute()
}

func validateGlobalFlags() error {
	if noColor && !raw {
		return fmt.Errorf("--no-color can only be used when --raw is also used")
	}
	return nil
}

func validateRequiresServer() error {
	if server == "" {
		return fmt.Errorf("required flag(s) \"server\" not set")
	}
	return nil
}
