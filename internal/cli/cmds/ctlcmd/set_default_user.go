package ctlcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
)

func NewSetDefaultClusterUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default-user [username]",
		Short: "Mark a stored user as the default for a cluster",
		Long: `Mark one of the saved users as the default for a cluster.
When no --user flag is given on subsequent 'ctl' commands, the default user for
that cluster is used automatically.

Requires --cluster (-c) to identify which cluster to update.
If [username] is omitted a fuzzy picker opens so you can choose interactively.`,
		Args: cobra.MaximumNArgs(1),
		Example: `  gns3util ctl auth set-default-user alice -c mycluster
  gns3util ctl auth set-default-user -c mycluster   # opens picker`,
		Annotations: map[string]string{"auth-mode": "cluster-only"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			clusterName := cfg.Cluster

			keyFilePath, err := pathutils.ResolveKeyFilePath(cfg.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to resolve key file path: %w", err)
			}

			kf, err := pathutils.LoadGNS3KeysFile(keyFilePath)
			if err != nil {
				return fmt.Errorf("failed to load key file: %w", err)
			}

			users := kf.ListClusterUsers(clusterName)
			if len(users) == 0 {
				return fmt.Errorf("no users found for cluster %q in the key file", clusterName)
			}

			var targetUser string
			if len(args) == 1 {
				targetUser = args[0]
			} else {
				choices := make([]string, 0, len(users))
				for _, u := range users {
					choices = append(choices, u.User)
				}
				picked := fuzzy.NewFuzzyFinderWithTitle(choices, false, "Select default user")
				if len(picked) == 0 {
					return fmt.Errorf("no user selected")
				}
				targetUser = picked[0]
			}

			if err := kf.SetDefaultClusterUser(clusterName, targetUser); err != nil {
				return err
			}

			if err := pathutils.SaveKeysFile(keyFilePath, kf); err != nil {
				return fmt.Errorf("failed to save key file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Default cluster user for %q set to %q.\n", clusterName, targetUser)
			return nil
		},
	}
}
