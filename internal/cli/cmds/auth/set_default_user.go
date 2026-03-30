package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

func NewSetDefaultUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default-user [username]",
		Short: "Mark a stored user as the default for a standalone GNS3 server",
		Long: `Mark one of the saved users as the default for a standalone GNS3 server.
When no --user flag is given on subsequent commands, the default user for that
server URL is used automatically.

Requires --server (or the global -s flag) to identify which server to update.
If [username] is omitted a fuzzy picker opens so you can choose interactively.`,
		Args: cobra.MaximumNArgs(1),
		Example: `  gns3util auth set-default-user alice -s http://gns3.example.com
  gns3util auth set-default-user -s http://gns3.example.com   # opens picker`,
		Annotations: map[string]string{"auth-mode": "none"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			srv := cfg.Server
			if srv == "" {
				return fmt.Errorf("--server (-s) is required for set-default-user")
			}

			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load key file: %w", err)
			}

			entries := kf.ListUsersForServer(srv)
			if len(entries) == 0 {
				return fmt.Errorf("no users found for %q in the key file — login first with 'auth login'", srv)
			}

			var targetUser string
			if len(args) == 1 {
				targetUser = args[0]
			} else {
				choices := make([]string, 0, len(entries))
				for _, e := range entries {
					choices = append(choices, e.User)
				}
				picked := fuzzy.NewFuzzyFinderWithTitle(choices, false, "Select default user")
				if len(picked) == 0 {
					return fmt.Errorf("no user selected")
				}
				targetUser = picked[0]
			}

			if err := kf.SetDefaultUserForServer(srv, targetUser); err != nil {
				return err
			}

			if err := pathutils.SaveKeysFile(keyFilePath, kf); err != nil {
				return fmt.Errorf("failed to save key file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Default user for %q set to %q.\n", srv, targetUser)
			return nil
		},
	}

	return cmd
}
