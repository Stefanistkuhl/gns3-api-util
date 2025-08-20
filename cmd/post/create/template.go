package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateTemplateCmd() *cobra.Command {
	var (
		templateID   string
		name         string
		version      string
		category     string
		defaultFmt   string
		symbol       string
		templateType string
		computeID    string
		usage        string
		useJSON      string
	)

	cmd := &cobra.Command{
		Use:     "new",
		Short:   "Create a template",
		Example: "gns3util -s https://controller:3080 create template -n some_name -t vpcs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			// validate choice flags similar to click.Choice
			if err := validateChoice(category, []string{"router", "switch", "guest", "firewall"}, "--category"); err != nil {
				return err
			}
			if err := validateChoice(templateType, []string{"cloud", "nat", "ethernet_hub", "frame_relay_switch", "atm_switch", "docker", "dynamips", "vpcs", "virtualbox", "vmware", "iou", "qemu"}, "--template-type"); err != nil {
				return err
			}
			var payload map[string]any
			if useJSON == "" {
				if name == "" || templateType == "" {
					return fmt.Errorf("for this command -n/--name and -t/--template-type are required or provide --use-json")
				}
				data := schemas.TemplateCreate{}
				if templateID != "" {
					v := templateID
					data.TemplateID = &v
				}
				if name != "" {
					v := name
					data.Name = &v
				}
				if version != "" {
					v := version
					data.Version = &v
				}
				if category != "" {
					v := category
					data.Category = &v
				}
				if defaultFmt != "" {
					v := defaultFmt
					data.DefaultNameFormat = &v
				}
				if symbol != "" {
					v := symbol
					data.Symbol = &v
				}
				if templateType != "" {
					v := templateType
					data.TemplateType = &v
				}
				if computeID != "" {
					v := computeID
					data.ComputeID = &v
				}
				if usage != "" {
					v := usage
					data.Usage = &v
				}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createTemplate", nil, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&templateID, "template-id", "d", "", "Desired ID for template, leave empty to use a generated one")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name of the template")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version of the template")
	cmd.Flags().StringVarP(&category, "category", "c", "", "Category")
	cmd.Flags().StringVarP(&defaultFmt, "default-name-format", "f", "", "Default name format")
	cmd.Flags().StringVar(&symbol, "symbol", "", "Symbol name")
	cmd.Flags().StringVarP(&templateType, "template-type", "t", "", "Template type")
	cmd.Flags().StringVarP(&computeID, "compute-id", "o", "", "Compute ID")
	cmd.Flags().StringVar(&usage, "usage", "", "Usage")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}

func validateChoice(val string, allowed []string, flag string) error {
	if val == "" {
		return nil
	}
	for _, v := range allowed {
		if v == val {
			return nil
		}
	}
	return fmt.Errorf("invalid value %q for %s (allowed: %v)", val, flag, allowed)
}
