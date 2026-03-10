package clustercmd

import (
	"errors"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/spf13/cobra"
)

func NewSyncClusterConfigCmdGroup() *cobra.Command {
	var noConfirm bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "sync your cluster config file with the local database",
		Long:  `sync your cluster config file with the local database`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfgLoaded cluster.Config
			var missing bool
			store, err := db.Init()
			if err != nil {
				return fmt.Errorf("failed to init db: %w", err)
			}

			if _, err := cluster.LoadClusterConfig(); err != nil {
				if errors.Is(err, cluster.ErrNoConfig) {
					missing = true
				} else {
					return fmt.Errorf("failed to load config: %w", err)
				}
			}

			if missing {
				if !noConfirm {
					confirmed := utils.ConfirmPrompt(
						fmt.Sprintf("%s no cluster config found. Generate one from the database now?",
							messageUtils.WarningMsg("Warning")),
						false,
					)
					if !confirmed {
						fmt.Println("Aborted.")
						return nil
					}
				}

				cfgGen, changed, genErr := cluster.EnsureConfigSyncedFromDB(cmd.Context())
				if genErr != nil {
					return fmt.Errorf("failed to generate config from DB: %w", genErr)
				}
				if changed {
					if err := cluster.WriteClusterConfig(cfgGen); err != nil {
						return fmt.Errorf("failed to write generated config: %w", err)
					}
				}
				fmt.Printf("%s generated cluster config from the database.\n", messageUtils.SuccessMsg("Success"))
				cfgLoaded = cfgGen
			} else {
				cfgEnsured, _, ensureErr := cluster.EnsureConfigSyncedFromDB(cmd.Context())
				if ensureErr != nil {
					return fmt.Errorf("failed ensuring config: %w", ensureErr)
				}
				cfgLoaded = cfgEnsured
			}

			inSync, checkErr := cluster.CheckConfigWithDB(cmd.Context(), store, cfgLoaded, verbose)
			if checkErr != nil {
				return fmt.Errorf("error checking config: %w", checkErr)
			}

			if inSync {
				fmt.Println("Nothing to do, Config already synced.")
				return nil
			}

			if !noConfirm {
				if !utils.ConfirmPrompt(
					fmt.Sprintf("%s out of sync. Sync config with the Database?",
						messageUtils.WarningMsg("Warning")),
					false,
				) {
					return nil
				}
			}

			cfgNew, changed, syncErr := cluster.SyncConfigWithDB(cmd.Context(), cfgLoaded)
			if syncErr != nil {
				return fmt.Errorf("error syncing config: %w", syncErr)
			}
			if !changed {
				fmt.Printf("%s nothing to sync.\n", messageUtils.SuccessMsg("Success"))
				return nil
			}
			if err := cluster.WriteClusterConfig(cfgNew); err != nil {
				return fmt.Errorf("failed to write to the config file: %w", err)
			}
			fmt.Printf("%s synced config with the Database.\n", messageUtils.SuccessMsg("Success"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&noConfirm, "no-confirm", "n", false, "Run the command without asking for confirmations.")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Run the command verbose to show all missmatches if they occur.")

	return cmd
}
