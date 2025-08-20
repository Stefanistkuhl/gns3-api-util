package auth

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func NewAuthStatusCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status",
		Short: "Check the current status of your Authentication",
		Long:  `Check the current status of your Authentication`,
		Run: func(cmd *cobra.Command, args []string) {
			var user schemas.User

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				log.Fatalf("failed to get global options: %v", err)
			}

			keys, err := authentication.LoadKeys(cfg.KeyFile)
			if err != nil {
				panic(err)
			}

			userData, err := authentication.TryKeys(keys, cfg)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = json.Unmarshal([]byte(userData), &user)
			if err != nil {
				log.Fatalf("Error unmarshaling JSON: %v", err)
			}
			fmt.Printf("%s logged in as user %s", colorUtils.Success("Success:"), colorUtils.Bold(*user.Username))
		},
	}
	return cmd
}
