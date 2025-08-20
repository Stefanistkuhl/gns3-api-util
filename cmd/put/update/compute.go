package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateComputeCmd() *cobra.Command {
	var (
		protocol  string
		host      string
		port      int
		user      string
		password  string
		name      string
		computeID string
		useJSON   string
	)

	cmd := &cobra.Command{
		Use:     "modify [compute-name/id]",
		Short:   "Update a compute",
		Long:    "Update a compute with new settings.",
		Example: "gns3util -s https://controller:3080 update compute [compute-id] -n new-name",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			computeIDArg := args[0]

			if err := validateChoice(protocol, []string{"http", "https"}, "--protocol"); err != nil {
				return err
			}

			var payload map[string]any
			if useJSON == "" {
				if protocol == "" && host == "" && port == 0 && user == "" && password == "" && name == "" && computeID == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				data := schemas.ComputeUpdate{}
				if protocol != "" {
					data.Protocol = &protocol
				}
				if host != "" {
					data.Host = &host
				}
				if port != 0 {
					data.Port = &port
				}
				if user != "" {
					data.User = &user
				}
				if password != "" {
					data.Password = &password
				}
				if name != "" {
					data.Name = &name
				}
				if computeID != "" {
					data.ComputeID = &computeID
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
			utils.ExecuteAndPrintWithBody(cfg, "updateCompute", []string{computeIDArg}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&protocol, "protocol", "", "", "Protocol for connection to the compute")
	cmd.Flags().StringVarP(&host, "host", "", "", "IP or Domain of the remote host")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "TCP port to connect with the remote host")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Username to connect as")
	cmd.Flags().StringVarP(&password, "password", "", "", "Password for the user")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the compute")
	cmd.Flags().StringVarP(&computeID, "compute-id", "", "", "Desired id for the compute")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
