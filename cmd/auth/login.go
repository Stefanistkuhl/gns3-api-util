package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/pathUtils"
)

var username string
var password string

func NewAuthLoginCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "login",
		Short: "Log in as user",
		Long:  `Log in as a user`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.SetEnvPrefix("GNS3")
			viper.AutomaticEnv()

			viper.BindPFlag("user", cmd.Flags().Lookup("USER"))
			viper.BindPFlag("password", cmd.Flags().Lookup("PASSWORD"))

			if !cmd.Flags().Changed("user") {
				username = viper.GetString("user")
			}
			if !cmd.Flags().Changed("password") {
				password = viper.GetString("password")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var token schemas.Token
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				log.Fatalf("failed to get global options: %v", err)
			}
			credentials := schemas.Credentials{
				Username: username,
				Password: password,
			}
			data, _ := json.Marshal(credentials)

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
			)

			client := api.NewGNS3Client(settings)

			ep := endpoints.PostEndpoints{}

			reqOpts := api.
				NewRequestOptions(settings).
				WithURL(ep.Authenticate()).
				WithMethod(api.POST).
				WithData(string(data))

			tokenData, res, err := client.Do(reqOpts)
			if err != nil {
				fmt.Println(err)
				return
			}
			if res.StatusCode == 200 {
				err = json.Unmarshal([]byte(tokenData), &token)
				if err != nil {
					log.Fatalf("Error unmarshaling JSON: %v", err)
				}
				err := authentication.SaveAuthData(cfg, token, username)
				if err != nil {
					log.Fatalln(err)
				}
				var keyFilePath string
				if cfg.KeyFile == "" {
					k, err := pathUtils.GetGNS3Dir()
					if err != nil {
						log.Fatalln(err)
					}
					keyFilePath = filepath.Join(k, "gns3key")
				} else {
					keyFilePath = cfg.KeyFile
				}

				fmt.Printf("%s logged in as user %s and saved token to %s", colorUtils.Success("Success:"), colorUtils.Bold(username), colorUtils.Bold(keyFilePath))
			}

		},
	}
	cmd.Flags().StringVarP(&username, "user", "u", "", "User to log in as (env: GNS3_USER)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password to use (env: GNS3_PASSWORD)")
	if os.Getenv("GNS3_USER") == "" {
		cmd.MarkFlagRequired("user")
	}
	if os.Getenv("GNS3_PASSWORD") == "" {
		cmd.MarkFlagRequired("password")
	}

	return cmd
}
