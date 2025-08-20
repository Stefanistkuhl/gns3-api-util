package post

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCheckVersionCmd() *cobra.Command {
	var (
		flagVersion        string
		flagUseJSON        string
		flagControllerHost string
		flagLocal          bool
	)

	cmd := &cobra.Command{
		Use:     "check-version",
		Short:   "Check server version against provided data",
		Long:    `Check server version against provided data on the GNS3 server. Either provide --version or pass a full JSON payload using --use-json.`,
		Example: `gns3util -s https://controller:3080 post check-version --version "3.0.5"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			var payload map[string]any

			if flagUseJSON == "" {
				if flagVersion == "" {
					return fmt.Errorf("for this command --version is required, or provide --use-json")
				}
				data := schemas.Version{
					ControllerHost: flagControllerHost,
					Version:        flagVersion,
					Local:          flagLocal,
				}
				b, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to encode request: %w", err)
				}
				if err := json.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("failed to prepare payload: %w", err)
				}
			} else {
				if err := json.Unmarshal([]byte(flagUseJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}

			utils.ExecuteAndPrintWithBody(cfg, "checkVersion", nil, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&flagVersion, "version", "v", "", "Version to check against")
	cmd.Flags().StringVarP(&flagUseJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	cmd.Flags().StringVarP(&flagControllerHost, "controller-host", "c", "", "Controller host to use")
	cmd.Flags().BoolVarP(&flagLocal, "local", "l", false, "Use local mode")

	return cmd
}
