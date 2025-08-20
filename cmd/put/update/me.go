package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateMeCmd() *cobra.Command {
	var (
		password string
		email    string
		fullName string
		useJSON  string
	)

	cmd := &cobra.Command{
		Use:     "me",
		Short:   "Update the logged in user",
		Long:    "Update the current user's password, email, or full name.",
		Example: "gns3util -s https://controller:3080 update me -p newpassword -e newemail@example.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var payload map[string]any
			if useJSON == "" {
				if password == "" && email == "" && fullName == "" {
					return fmt.Errorf("for this command at least one of -p/--password, -e/--email, or -f/--full-name is required or provide --use-json")
				}
				data := schemas.LoggedInUserUpdate{}
				if password != "" {
					data.Password = &password
				}
				if email != "" {
					data.Email = &email
				}
				if fullName != "" {
					data.FullName = &fullName
				}
				b, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to encode request: %w", err)
				}
				if err := json.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("failed to prepare payload: %w", err)
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "updateMe", nil, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&password, "password", "p", "", "Password to set for the current User")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email to set for the current User")
	cmd.Flags().StringVarP(&fullName, "full-name", "f", "", "Full name to set for the current User")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
