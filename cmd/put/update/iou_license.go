package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateIOULicenseCmd() *cobra.Command {
	var (
		iourcContent string
		licenseCheck bool
		useJSON      string
	)

	cmd := &cobra.Command{
		Use:     "iou-license",
		Short:   "Update the IOULicense",
		Long:    "Update the IOULicense with new content and license check settings.",
		Example: "gns3util -s https://controller:3080 update iou-license -iourc-content 'some license content' -license-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var payload map[string]any
			if useJSON == "" {
				if iourcContent == "" && !licenseCheck {
					return fmt.Errorf("for this command either --iourc-content and -l/--license-check are required or provide --use-json")
				}
				data := schemas.IOULicense{}
				if iourcContent != "" {
					data.IOURCContent = &iourcContent
				}
				data.LicenseCheck = &licenseCheck
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
			utils.ExecuteAndPrintWithBody(cfg, "updateIOULicense", nil, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&iourcContent, "iourc-content", "", "", "Contents of the license")
	cmd.Flags().BoolVarP(&licenseCheck, "license-check", "l", false, "Enable license checking")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
