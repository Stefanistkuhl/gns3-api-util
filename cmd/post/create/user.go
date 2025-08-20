package create

import (
	"encoding/json"
	"fmt"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateUserCmd() *cobra.Command {
	var (
		flagUsername string
		flagIsActive bool
		flagEmail    string
		flagFullName string
		flagPassword string
		flagUseJSON  string
	)

	var cmd = &cobra.Command{
		Use:     "new",
		Short:   "Create a user account",
		Long:    "Create a new user account on the GNS3v3 controller. Either provide -u and -p (and optional fields) or pass a full JSON payload using --use-json.",
		Example: "gns3util -s https://controller:3080 create user -u alice -p secret --email alice@example.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var payload map[string]any

			if flagUseJSON == "" {
				if flagUsername == "" || flagPassword == "" {
					return fmt.Errorf("for this command -u/--username and -p/--password are required, or provide --use-json")
				}
				if err := validatePassword(flagPassword); err != nil {
					return err
				}
				data := schemas.UserCreate{
					Username: &flagUsername,
					Password: &flagPassword,
					Email:    nil,
					FullName: nil,
					IsActive: flagIsActive,
				}
				if flagEmail != "" {
					data.Email = &flagEmail
				}
				if flagFullName != "" {
					data.FullName = &flagFullName
				}
				b, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to encode request: %w", err)
				}
				if err := json.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("failed to prepare payload: %w", err)
				}
			} else {
				if err := json.Unmarshal([]byte(flagUseJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createUser", nil, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&flagUsername, "username", "u", "", "Desired username for the user")
	cmd.Flags().BoolVarP(&flagIsActive, "is-active", "a", false, "Mark the user as active")
	cmd.Flags().StringVarP(&flagEmail, "email", "e", "", "Email address for the user")
	cmd.Flags().StringVarP(&flagFullName, "full-name", "f", "", "Full name for the user")
	cmd.Flags().StringVarP(&flagPassword, "password", "p", "", "Password for the user")
	cmd.Flags().StringVarP(&flagUseJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}

func validatePassword(pw string) error {
	if len(pw) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(pw) > 100 {
		return fmt.Errorf("password must be at most 100 characters long")
	}
	hasDigit := false
	for _, r := range pw {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number")
	}
	return nil
}
