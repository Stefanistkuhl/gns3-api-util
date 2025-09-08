package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/tidwall/gjson"
)

var (
	passwordFlag string
	useFuzzy     bool
)

func NewChangePasswordCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "change-password [user-name/id]",
		Short: "Change a user's password",
		Long:  `Change a user's password with interactive password input and validation`,
		Example: `gns3util -s https://controller:3080 user change-password my-user
gns3util -s https://controller:3080 user change-password -f
gns3util -s https://controller:3080 user change-password my-user -p "newpassword123"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [user-name/id] when --fuzzy is not set")
			}
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.SetEnvPrefix("GNS3")
			viper.AutomaticEnv()

			viper.BindPFlag("password", cmd.Flags().Lookup("password"))

			if !cmd.Flags().Changed("password") {
				passwordFlag = viper.GetString("password")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v\n", err)
				return
			}

			var userID string
			var username string

			if useFuzzy {
				rawData, _, err := utils.CallClient(cfg, "getUsers", nil, nil)
				if err != nil {
					fmt.Printf("%s %v\n", colorUtils.Error("Error:"), err)
					return
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					fmt.Printf("%s Expected array response, got %s\n", colorUtils.Error("Error:"), result.Type)
					return
				}

				var apiData []gjson.Result
				var usernames []string

				result.ForEach(func(_, value gjson.Result) bool {
					apiData = append(apiData, value)
					if val := value.Get("username"); val.Exists() {
						usernames = append(usernames, val.String())
					}
					return true
				})

				if len(usernames) == 0 {
					fmt.Println("No users found")
					return
				}

				selected := fuzzy.NewFuzzyFinder(usernames, false)
				if len(selected) == 0 {
					fmt.Println("No user selected")
					return
				}

				for _, data := range apiData {
					if usernameField := data.Get("username"); usernameField.Exists() && usernameField.String() == selected[0] {
						userID = data.Get("user_id").String()
						username = selected[0]
						break
					}
				}
			} else {
				userID = args[0]
				username = args[0]

				if !utils.IsValidUUIDv4(args[0]) {
					userID, err = utils.ResolveID(cfg, "user", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}

			var newPassword string

			if passwordFlag != "" {
				if !utils.ValidatePassword(passwordFlag) {
					fmt.Printf("%s Password must be at least 8 characters with at least 1 number and 1 lowercase letter\n", colorUtils.Error("Error:"))
					return
				}
				newPassword = passwordFlag
			} else {
				fmt.Printf("Changing password for user: %s\n", colorUtils.Bold(username))
				newPassword, err = utils.GetPasswordFromInput()
				if err != nil {
					fmt.Printf("%s %v\n", colorUtils.Error("Error:"), err)
					return
				}
			}

			userUpdate := schemas.UserUpdate{
				Password: &newPassword,
			}

			data, err := json.Marshal(userUpdate)
			if err != nil {
				fmt.Printf("%s Failed to marshal user update: %v\n", colorUtils.Error("Error:"), err)
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				fmt.Printf("%s Failed to prepare payload: %v\n", colorUtils.Error("Error:"), err)
				return
			}

			utils.ExecuteAndPrintWithBody(cfg, "updateUser", []string{userID}, payload)
		},
	}

	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "New password (env: GNS3_PASSWORD)")
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to select a user")

	return cmd
}
