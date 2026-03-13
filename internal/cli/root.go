package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/internal/cli/cmds/auth"
	"github.com/0xveya/gns3util/internal/cli/cmds/class"
	"github.com/0xveya/gns3util/internal/cli/cmds/exercise"
)

var (
	server       string
	keyFile      string
	insecure     bool
	version      bool
	outputFormat string
	cluster      string
)

var Version = "1.3.1"

var rootCmd = &cobra.Command{
	Use:           "gns3util",
	Short:         "A cute little utility for GNS3v3",
	Long:          `A cute little utility for GNS3v3 for managing GNS3v3 projects and devices.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "completion" || cmd.Name() == "_carapace" ||
			cmd.Name() == "help" || version {
			return nil
		}
		if len(args) > 0 && args[0] == "_carapace" {
			return nil
		}

		serverFlagSet := cmd.Flags().Changed("server")
		clusterFlagSet := cmd.Flags().Changed("cluster")

		if hasAuthModeClusterOnly(cmd) {
			if serverFlagSet {
				return fmt.Errorf("command %q does not support --server; you must use --cluster", cmd.Name())
			}
			if !clusterFlagSet && viper.GetString("cluster") == "" {
				return fmt.Errorf("command %q requires the --cluster flag", cmd.Name())
			}
		}
		if !serverFlagSet {
			server = viper.GetString("server")
		}
		if !clusterFlagSet {
			cluster = viper.GetString("cluster")
		}
		if keyFile == "" {
			keyFile = viper.GetString("key-file")
		}
		if outputFormat == "" || outputFormat == "kv" {
			if v := viper.GetString("output"); v != "" {
				outputFormat = v
			}
		}
		keyFile, _ = pathutils.ExpandPath(keyFile)

		if err := validateGlobalFlags(); err != nil {
			return err
		}

		if serverFlagSet && clusterFlagSet {
			return fmt.Errorf("--server and --cluster are mutually exclusive")
		}

		requiresServer := !isCtlCommand(cmd) && !hasAuthModeFlexible(cmd) && !hasAuthModeNone(cmd)
		if requiresServer && server == "" && cluster == "" {
			return fmt.Errorf("required flag(s) \"--server\" not set")
		}

		var clusterEntry *pathutils.ClusterEntry
		if cluster != "" {
			keyFilePath, err := pathutils.ResolveKeyFilePath(keyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file: %w", err)
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load key file: %w", err)
			}

			entry, ok := findClusterByName(kf, cluster)
			if !ok {
				return fmt.Errorf("cluster %q not found in key file", cluster)
			}
			clusterEntry = entry
		}

		cmdPath := cmd.CommandPath()
		fmtType := globals.ParseOutputFormat(outputFormat)

		opts := &config.GlobalOptions{
			Server:       server,
			Insecure:     insecure,
			KeyFile:      keyFile,
			OutputFormat: fmtType,
			CommandPath:  cmdPath,
			Cluster:      cluster,
			ClusterEntry: clusterEntry,
		}
		ctx := config.WithGlobalOptions(cmd.Context(), opts)
		cmd.SetContext(ctx)

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if version {
			fmt.Printf("gns3util version %s\n", Version)
			return nil
		}
		return cmd.Help()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("GNS3_STORE_LAST_SETTINGS") == "true" {
			gns3Dir, err := pathutils.GetGNS3Dir()
			if err != nil {
				return nil
			}

			viper.Set("server", server)
			viper.Set("output", outputFormat)
			viper.Set("key-file", keyFile)
			viper.Set("insecure", insecure)

			configPath := filepath.Join(gns3Dir, "config.toml")
			_ = viper.WriteConfigAs(configPath)
		}
		return nil
	},
}

func init() {
	carapace.Gen(rootCmd)
	cobra.OnFinalize()
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "",
		"GNS3v3 Server URL. Can be set via GNS3_SERVER or config.toml")

	rootCmd.PersistentFlags().StringVarP(&keyFile, "key-file", "k", "",
		"Path to authentication keyfile. Can be set via GNS3_KEY_FILE or config.toml")

	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "i", false,
		"Ignore unsigned SSL-Certificates. Can be set via GNS3_INSECURE")

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "kv",
		"Output format: [kv, json, json-colorless, collapsed, yaml, toml]. Can be set via GNS3_OUTPUT")
	rootCmd.PersistentFlags().StringVarP(&cluster, "cluster", "c", "",
		"Cluster name from keyfile. Mutually exclusive with --server")

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
	rootCmd.AddCommand(NewACLCmdGroup())

	rootCmd.AddCommand(NewPoolCmdGroup())
	rootCmd.AddCommand(NewSnapshotCmdGroup())

	rootCmd.AddCommand(NewSystemCmdGroup())

	rootCmd.AddCommand(NewRemoteCmdGroup())

	rootCmd.AddCommand(NewClusterCmdGroup())
	rootCmd.AddCommand(NewShareCmdGroup())
	rootCmd.AddCommand(NewCtlCmdGroup())
	carapace.Gen(rootCmd).FlagCompletion(carapace.ActionMap{
		"key-file": carapace.ActionFiles(),
		"server":   carapace.ActionValues("http://localhost:3080", "https://gns3.example.com"),
	})
	carapace.Gen(rootCmd).FlagCompletion(carapace.ActionMap{
		"output": carapace.ActionValuesDescribed(
			"kv", "Classic Key-Value pairs (default)",
			"json", "Pretty-printed JSON with colors",
			"json-colorless", "Pretty-printed JSON without colors",
			"collapsed", "Minified JSON (really ugly)",
			"yaml", "YAML format",
			"toml", "TOML format",
		),
	})
	gns3Dir, err := pathutils.GetGNS3Dir()
	if err == nil {
		viper.AddConfigPath(gns3Dir)
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	viper.SetEnvPrefix("GNS3")
	viper.AutomaticEnv()

	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("key-file", rootCmd.PersistentFlags().Lookup("key-file"))
	_ = viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
	_ = viper.BindPFlag("cluster", rootCmd.PersistentFlags().Lookup("cluster"))

	_ = viper.ReadInConfig()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v\n", messageUtils.ErrorMsg(err.Error()))
	}
}

func validateGlobalFlags() error {
	validFormats := map[string]bool{
		"kv": true, "json": true, "json-colorless": true,
		"collapsed": true, "yaml": true, "toml": true, "table": true,
	}

	if !validFormats[outputFormat] {
		return fmt.Errorf("invalid output format %q: choose from kv, json, json-colorless, collapsed, yaml, toml, table", outputFormat)
	}
	return nil
}

func isCtlCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "ctl" {
			return true
		}
	}
	return false
}

func hasAuthModeFlexible(cmd *cobra.Command) bool {
	return cmd.Annotations != nil && cmd.Annotations["auth-mode"] == "flexible"
}

func hasAuthModeNone(cmd *cobra.Command) bool {
	return cmd.Annotations != nil && cmd.Annotations["auth-mode"] == "none"
}

func hasAuthModeClusterOnly(cmd *cobra.Command) bool {
	return cmd.Annotations != nil && cmd.Annotations["auth-mode"] == "cluster-only"
}

func findClusterByName(kf *pathutils.KeyFileV2, name string) (*pathutils.ClusterEntry, bool) {
	for i := range kf.Clusters {
		if kf.Clusters[i].Name == name {
			return &kf.Clusters[i], true
		}
	}
	return nil, false
}
