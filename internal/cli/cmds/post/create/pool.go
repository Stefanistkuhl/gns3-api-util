package create

import (
	"encoding/json"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/api/schemas"
	"github.com/spf13/cobra"
)

func NewCreatePoolCmd() *cobra.Command {
	var name string
	var useJSON string
	cmd := &cobra.Command{
		Use:     "new",
		Aliases: []string{"create", "c", "add"},
		Short:   "Create a resource pool",
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
				data := schemas.ResourcePoolCreate{Name: &name}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createPool", nil, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the pool")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
