package clustercmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathUtils"
	"github.com/spf13/cobra"
)

func NewPurgeClusterConfigCMD() *cobra.Command {
	var noConfirm bool
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "purge the config file and the local database",
		Long:  `purge the config file and the local database`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgLoaded, err := cluster.LoadClusterConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if !noConfirm {
				if !utils.ConfirmPrompt(fmt.Sprintf("%s do you want to purge this config?", messageUtils.WarningMsg("Warning")), false) {
					return nil
				}
			}
			dir, getDirErr := pathUtils.GetGNS3Dir()
			if getDirErr != nil {
				return getDirErr
			}
			path := filepath.Join(dir, "cluster_config.toml")
			applyErr := cluster.PurgeConfig(cfgLoaded, cmd.Context())
			if applyErr != nil {
				return fmt.Errorf("error purging config: %w", applyErr)
			}
			removeErr := os.Remove(path)
			if removeErr != nil {
				return fmt.Errorf("failed to remove config file: %w", removeErr)
			}
			fmt.Printf("%s purged config.\n", messageUtils.SuccessMsg("Success"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&noConfirm, "no-confirm", "n", false, "Run the command without asking for confirmations.")

	return cmd
}
