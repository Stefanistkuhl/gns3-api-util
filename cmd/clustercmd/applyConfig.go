package clustercmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewApplyConfigCmd() *cobra.Command {
	var noConfirm bool
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply your config file to the local database",
		Long:  `apply your config file to the local database`,
		Run: func(cmd *cobra.Command, args []string) {
			cfgLoaded, err := cluster.LoadClusterConfig()
			if err != nil {
				fmt.Printf("%s failed to load config: %v\n", messageUtils.ErrorMsg("Error"), err)
				return
			}

			if !noConfirm {
				if !utils.ConfirmPrompt(fmt.Sprintf("%s do you want to apply this config to the Database?", messageUtils.WarningMsg("Warning")), false) {
					return
				}
			}

			applyErr := cluster.ApplyConfig(cfgLoaded)
			if applyErr != nil {
				fmt.Printf("%s %v\n", messageUtils.ErrorMsg("Error applying config"), applyErr)
				return
			}
			fmt.Printf("%s applied config to the Database.\n", messageUtils.SuccessMsg("Success"))
		},
	}
	cmd.Flags().BoolVarP(&noConfirm, "no-confirm", "n", false, "Run the command without asking for confirmations.")

	return cmd
}
