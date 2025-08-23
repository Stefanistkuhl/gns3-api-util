package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateSnapshotCmd() *cobra.Command {
	var (
		name    string
		useJSON string
	)
	cmd := &cobra.Command{
		Use:     utils.CreateSingleElementCmdName + " [project-name/id]",
		Short:   "Create a snapshot of a project",
		Long:    "Create a snapshot of a project with specified name",
		Example: "gns3util -s https://controller:3080 project create my-project --name backup-2024",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			projectID := args[0]
			var payload map[string]any
			if useJSON == "" {
				if name == "" {
					return fmt.Errorf("for this command -n/--name is required or provide --use-json")
				}
				data := schemas.SnapshotCreate{Name: &name}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createSnapshot", []string{projectID}, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the snapshot")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
