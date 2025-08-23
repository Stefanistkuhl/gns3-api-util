package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateLinkCmd() *cobra.Command {
	var useJSON string
	cmd := &cobra.Command{
		Use:     utils.CreateSingleElementCmdName + " [project-name/id] [json-data]",
		Short:   "Create a link between two nodes in a project",
		Long:    "Create a link between two nodes in a project using JSON data",
		Example: "gns3util -s https://controller:3080 link create my-project '{\"nodes\": [{\"node_id\": \"node1\", \"port_number\": 0}, {\"node_id\": \"node2\", \"port_number\": 0}]}'",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			projectID := args[0]
			jsonData := args[1]
			var payload map[string]any
			if useJSON == "" {
				if jsonData == "" {
					return fmt.Errorf("JSON data is required as a positional argument or provide --use-json")
				}
				if err := json.Unmarshal([]byte(jsonData), &payload); err != nil {
					return fmt.Errorf("invalid JSON: %w", err)
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createLink", []string{projectID}, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of positional JSON argument")
	return cmd
}
