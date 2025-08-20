package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modify [project-name/id] [link-name/id] [json-data]",
		Short:   "Update a link",
		Long:    "Update a link with JSON data.",
		Example: "gns3util -s https://controller:3080 update link [project-id] [link-id] '{\"nodes\":[...]}'",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			projectID := args[0]
			linkID := args[1]
			jsonData := args[2]

			if jsonData == "" {
				return fmt.Errorf("JSON data is required as a positional argument")
			}

			var payload map[string]any
			if err := json.Unmarshal([]byte(jsonData), &payload); err != nil {
				return fmt.Errorf("invalid JSON data: %w", err)
			}

			utils.ExecuteAndPrintWithBody(cfg, "updateLink", []string{projectID, linkID}, payload)
			return nil
		},
	}

	return cmd
}
