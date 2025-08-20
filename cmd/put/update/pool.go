package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdatePoolCmd() *cobra.Command {
	var (
		name    string
		useJSON string
	)

	cmd := &cobra.Command{
		Use:     "modify [pool-name/id]",
		Short:   "Update a resource pool",
		Long:    "Update a resource pool with a new name.",
		Example: "gns3util -s https://controller:3080 update pool [pool-id] -n new-name",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			poolID := args[0]
			if !utils.IsValidUUIDv4(args[0]) {
				poolID, err = utils.ResolveID(cfg, "pool", args[0], nil)
				if err != nil {
					return err
				}
			}

			var payload map[string]any
			if useJSON == "" {
				if name == "" {
					return fmt.Errorf("for this command -n/--name is required or provide --use-json")
				}
				data := schemas.ResourcePoolUpdate{Name: &name}
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
			utils.ExecuteAndPrintWithBody(cfg, "updatePool", []string{poolID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the pool")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
