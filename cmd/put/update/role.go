package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateRoleCmd() *cobra.Command {
	var (
		name        string
		description string
		useJSON     string
	)

	cmd := &cobra.Command{
		Use:     "role [role-id]",
		Short:   "Update a role",
		Long:    "Update an RBAC role.",
		Example: "gns3util -s https://controller:3080 update role [role-name/id] --name new-name --description new-description",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			roleID := args[0]

			if !utils.IsValidUUIDv4(roleID) {
				resolvedID, err := utils.ResolveID(cfg, "role", roleID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve role ID: %w", err)
				}
				roleID = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				if name == "" && description == "" {
					return fmt.Errorf("for this command at least one of -n/--name or -d/--description is required or provide --use-json")
				}
				data := schemas.RoleUpdate{}
				if name != "" {
					data.Name = &name
				}
				if description != "" {
					data.Description = &description
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
			utils.ExecuteAndPrintWithBody(cfg, "updateRole", []string{roleID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the role")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description for the role")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
