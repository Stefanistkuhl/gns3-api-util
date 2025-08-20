package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateRoleCmd() *cobra.Command {
	var (
		name        string
		description string
		useJSON     string
	)

	cmd := &cobra.Command{
		Use:     "role",
		Short:   "Create a role",
		Long:    "Create an RBAC role.",
		Example: "gns3util -s https://controller:3080 create role -n some-name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var payload map[string]any
			if useJSON == "" {
				if name == "" {
					return fmt.Errorf("for this command -n/--name is required or provide --use-json")
				}
				data := schemas.RoleCreate{Name: &name}
				if description != "" {
					data.Description = &description
				}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createRole", nil, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the role")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description for the role")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
