package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateGroupCmd() *cobra.Command {
	var (
		name    string
		useJSON string
	)

	cmd := &cobra.Command{
		Use:     utils.UpdateSingleElementCmdName + " [group-name/id]",
		Short:   "Update a group",
		Long:    "Update a user group.",
		Example: "gns3util -s https://controller:3080 group update my-group --name new-name",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			groupID := args[0]

			if !utils.IsValidUUIDv4(groupID) {
				resolvedID, err := utils.ResolveID(cfg, "group", groupID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve group ID: %w", err)
				}
				groupID = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				if name == "" {
					return fmt.Errorf("for this command -n/--name is required or provide --use-json")
				}
				data := schemas.UserGroupUpdate{Name: &name}
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
			utils.ExecuteAndPrintWithBody(cfg, "updateGroup", []string{groupID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the group")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
