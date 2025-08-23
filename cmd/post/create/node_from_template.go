package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateNodeFromTemplateCmd() *cobra.Command {
	var (
		x         int
		y         int
		name      string
		computeID string
		useJSON   string
	)
	cmd := &cobra.Command{
		Use:     "from-template [project-name/id] [template-name/id]",
		Short:   "Create a node from a template",
		Long:    "Create a node from a template in a project",
		Example: "gns3util -s https://controller:3080 node from-template my-project my-template --x 100 --y 200",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			projectID := args[0]
			templateID := args[1]
			var payload map[string]any
			if useJSON == "" {
				if x == 0 || y == 0 {
					return fmt.Errorf("for this command --x and --y are required or provide --use-json")
				}
				data := schemas.TemplateUsage{}
				{
					v := x
					data.X = &v
				}
				{
					v := y
					data.Y = &v
				}
				if name != "" {
					v := name
					data.Name = &v
				}
				if computeID != "" {
					v := computeID
					data.ComputeID = &v
				}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createProjectNodeFromTemplate", []string{projectID, templateID}, payload)
			return nil
		},
	}
	cmd.Flags().IntVar(&x, "x", 0, "X coordinate")
	cmd.Flags().IntVar(&y, "y", 0, "Y coordinate")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Node name")
	cmd.Flags().StringVarP(&computeID, "compute-id", "c", "local", "Compute ID")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
