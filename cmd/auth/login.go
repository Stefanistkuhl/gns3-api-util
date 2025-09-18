package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
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

			_ = viper.BindPFlag("user", cmd.Flags().Lookup("user"))
			_ = viper.BindPFlag("password", cmd.Flags().Lookup("password"))

			if !cmd.Flags().Changed("user") {
				username = viper.GetString("user")
			}
			if !cmd.Flags().Changed("password") {
				password = viper.GetString("password")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v\n", err)
				return
			}

			if username == "" || password == "" {
				interactiveUsername, interactivePassword, err := utils.GetLoginCredentials()
				if err != nil {
					fmt.Printf("%s %v\n", colorUtils.Error("Error:"), err)
					return
				}

				if username == "" {
					username = interactiveUsername
				}
				if password == "" {
					password = interactivePassword
				}
			}

			if username == "" || password == "" {
				fmt.Printf("%s Username and password are required\n", colorUtils.Error("Error:"))
				return
			}

			credentials := schemas.Credentials{
				Username: username,
				Password: password,
			}

			data, err := json.Marshal(credentials)
			if err != nil {
				fmt.Printf("%s Failed to marshal credentials: %v\n", colorUtils.Error("Error:"), err)
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				fmt.Printf("%s Failed to prepare payload: %v\n", colorUtils.Error("Error:"), err)
				return
			}
			body, status, err := utils.CallClient(cfg, "userAuthenticate", []string{}, payload)
			if err != nil {
				if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Authentication was unsuccessful") {
					fmt.Printf("%v Authentication failed. Please check your username and password.\n", colorUtils.Error("Error:"))
					return
				}
				fmt.Printf("%v %v\n", colorUtils.Error("Error:"), err)
				return
			}

			if status == 200 {
				fmt.Printf("%v Successfully logged in as %s\n", colorUtils.Success("Success:"), colorUtils.Bold(username))
				var token schemas.Token
				marshallErr := json.Unmarshal(body, &token)
				if marshallErr != nil {
					fmt.Printf("%v failed to unmarshall response: %s", colorUtils.Error("Error:"), marshallErr)
					return
				}
				writeErr := authentication.SaveAuthData(cfg, token, credentials.Username)
				if writeErr != nil {
					fmt.Printf("%v failed to write authentication data to the keyfile: %s", colorUtils.Error("Error:"), writeErr)
					return
				}
			} else {
				fmt.Printf("%v Authentication failed (status: %d)\n", colorUtils.Error("Error:"), status)
			}

		},
	}
	cmd.Flags().StringVarP(&username, "user", "u", "", "User to log in as (env: GNS3_USER)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password to use (env: GNS3_PASSWORD)")

	return cmd
}
