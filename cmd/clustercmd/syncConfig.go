package clustercmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewSyncClusterConfigCmdGroup() *cobra.Command {
	var noConfirm bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "sync your cluster config file with the local database",
		Long:  `sync your cluster config file with the local database`,
		Run: func(cmd *cobra.Command, args []string) {
			cfgLoaded, err := cluster.LoadClusterConfig()
			missing := false
			if err != nil {
				if errors.Is(err, cluster.ErrNoConfig) {
					missing = true
				} else {
					fmt.Printf("%s failed to load config: %v\n", messageUtils.ErrorMsg("Error"), err)
					return
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
						return
					}
				}

				cfgGen, changed, genErr := cluster.EnsureConfigSyncedFromDB()
				if genErr != nil {
					fmt.Printf("%s failed to generate config from DB: %v\n", messageUtils.ErrorMsg("Error"), genErr)
					return
				}
				if changed {
					if err := cluster.WriteClusterConfig(cfgGen); err != nil {
						fmt.Printf("%s failed to write generated config: %v\n", messageUtils.ErrorMsg("Error"), err)
						return
					}
				}
				fmt.Printf("%s generated cluster config from the database.\n", messageUtils.SuccessMsg("Success"))
				cfgLoaded = cfgGen
			} else {
				cfgEnsured, _, ensureErr := cluster.EnsureConfigSyncedFromDB()
				if ensureErr != nil {
					fmt.Printf("%s failed ensuring config: %v\n", messageUtils.ErrorMsg("Error"), ensureErr)
					return
				}
				cfgLoaded = cfgEnsured
			}

			inSync, checkErr := cluster.CheckConfigWithDb(cfgLoaded, verbose)
			if checkErr != nil {
				fmt.Printf("%s %v\n", messageUtils.ErrorMsg("Error checking config"), checkErr)
				return
			}

			if inSync {
				fmt.Println("Nothing to do, Config already synced.")
				return
			}

			if !noConfirm {
				if !utils.ConfirmPrompt(
					fmt.Sprintf("%s out of sync. Sync config with the Database?",
						messageUtils.WarningMsg("Warning")),
					false,
				) {
					return
				}
			}

			cfgNew, changed, syncErr := cluster.SyncConfigWithDb(cfgLoaded)
			if syncErr != nil {
				fmt.Printf("%s %v\n", messageUtils.ErrorMsg("Error syncing config"), syncErr)
				return
			}
			if !changed {
				fmt.Printf("%s nothing to sync.\n", messageUtils.SuccessMsg("Success"))
				return
			}
			if err := cluster.WriteClusterConfig(cfgNew); err != nil {
				fmt.Printf("%s failed to write to the config file: %v\n", messageUtils.ErrorMsg("Error"), err)
				return
			}
			fmt.Printf("%s synced config with the Database.\n", messageUtils.SuccessMsg("Success"))
		},
	}
	cmd.Flags().BoolVarP(&noConfirm, "no-confirm", "n", false, "Run the command without asking for confirmations.")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Run the command verbose to show all missmatches if they occur.")

	return cmd
}
