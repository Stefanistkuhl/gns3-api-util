package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateUserCmd() *cobra.Command {
	var (
		username string
		isActive bool
		email    string
		fullName string
		password string
		useJSON  string
	)

	cmd := &cobra.Command{
		Use:     "modify [user-name/id]",
		Short:   "Update a user",
		Long:    "Update a given User with a given ID or name which will be resolved to a ID if a User with a matching name exists.",
		Example: "gns3util -s https://controller:3080 update user [user-name/id] --username newname --password newpassword",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			userID := args[0]

			if !utils.IsValidUUIDv4(userID) {
				resolvedID, err := utils.ResolveID(cfg, "user", userID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve user ID: %w", err)
				}
				userID = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				if username == "" && password == "" && email == "" && fullName == "" && !isActive {
					return fmt.Errorf("for this command at least one field is required or provide --use-json")
				}
				data := schemas.UserUpdate{}
				if username != "" {
					data.Username = &username
				}
				if password != "" {
					data.Password = &password
				}
				if email != "" {
					data.Email = &email
				}
				if fullName != "" {
					data.FullName = &fullName
				}
				if isActive {
					data.IsActive = &isActive
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
			utils.ExecuteAndPrintWithBody(cfg, "updateUser", []string{userID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Desired username for the User")
	cmd.Flags().BoolVarP(&isActive, "is-active", "a", false, "Marking the user as currently active")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Desired email for the user")
	cmd.Flags().StringVarP(&fullName, "full-name", "f", "", "Desired full name for the user")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Desired password for the user")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
