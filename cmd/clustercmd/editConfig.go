package clustercmd

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func NewEditConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edit your configuration with your $EDITOR",
		Long:  `edit your configuration with your $EDITOR`,
		Run: func(cmd *cobra.Command, args []string) {
			cfgLoaded, err := cluster.LoadClusterConfig()
			if err != nil {
				fmt.Printf("%s failed to load config: %v\n", colorUtils.Error("Error:"), err)
				return
			}
			res, marshallErr := toml.Marshal(&cfgLoaded)
			if marshallErr != nil {
				fmt.Printf("%s failed to marshall config %s", colorUtils.Error("Error:"), marshallErr)
				return

			}
			str, editErr := utils.EditTextWithEditor(string(res), "toml")
			if editErr != nil {
				fmt.Printf("%s failed to edit config %s", colorUtils.Error("Error:"), editErr)
				return

			}
			var cfgNew cluster.Config
			unMarshallErr := toml.Unmarshal([]byte(str), &cfgNew)
			if unMarshallErr != nil {
				fmt.Printf("%s failed to unmarshall config %s", colorUtils.Error("Error:"), unMarshallErr)
				return
			}
			writeErr := cluster.WriteClusterConfig(cfgNew)
			if writeErr != nil {
				fmt.Printf("%s failed to write edtied config to the config file %s", colorUtils.Error("Error:"), writeErr)
				return
			}
			fmt.Printf("%s wrote new config to the config file.", colorUtils.Success("Success:"))

		},
	}

	return cmd
}
