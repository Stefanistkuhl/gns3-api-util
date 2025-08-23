package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateTemplateCmd() *cobra.Command {
	var (
		templateID        string
		name              string
		version           string
		category          string
		defaultNameFormat string
		symbol            string
		templateType      string
		computeID         string
		usage             string
		useJSON           string
	)

	cmd := &cobra.Command{
		Use:     utils.UpdateSingleElementCmdName + " [template-name/id]",
		Short:   "Update a template",
		Long:    "Update a template with new settings and properties.",
		Example: "gns3util -s https://controller:3080 template update my-template --name new-name",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			templateIDArg := args[0]

			if !utils.IsValidUUIDv4(templateIDArg) {
				resolvedID, err := utils.ResolveID(cfg, "template", templateIDArg, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve template ID: %w", err)
				}
				templateIDArg = resolvedID
			}

			var payload map[string]any
			if useJSON == "" {
				if templateID == "" && name == "" && version == "" && category == "" && defaultNameFormat == "" && symbol == "" && templateType == "" && computeID == "" && usage == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				data := schemas.TemplateUpdate{}
				if templateID != "" {
					data.TemplateID = &templateID
				}
				if name != "" {
					data.Name = &name
				}
				if version != "" {
					data.Version = &version
				}
				if category != "" {
					data.Category = &category
				}
				if defaultNameFormat != "" {
					data.DefaultNameFormat = &defaultNameFormat
				}
				if symbol != "" {
					data.Symbol = &symbol
				}
				if templateType != "" {
					data.TemplateType = &templateType
				}
				if computeID != "" {
					data.ComputeID = &computeID
				}
				if usage != "" {
					data.Usage = &usage
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
			utils.ExecuteAndPrintWithBody(cfg, "updateTemplate", []string{templateIDArg}, payload)
			return nil
		},
	}

	cmd.Flags().StringVar(&templateID, "template-id", "", "Desired ID for the template")
	cmd.Flags().StringVar(&name, "name", "", "Desired name for the template")
	cmd.Flags().StringVar(&version, "version", "", "Desired version for the template")
	cmd.Flags().StringVar(&category, "category", "", "Desired category for the template")
	cmd.Flags().StringVar(&defaultNameFormat, "default-name-format", "", "Desired default name format for the template")
	cmd.Flags().StringVar(&symbol, "symbol", "", "Desired symbol for the template")
	cmd.Flags().StringVar(&templateType, "template-type", "", "Desired type for the template")
	cmd.Flags().StringVar(&computeID, "compute-id", "", "Desired compute ID for the template")
	cmd.Flags().StringVar(&usage, "usage", "", "Desired usage description for the template")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
