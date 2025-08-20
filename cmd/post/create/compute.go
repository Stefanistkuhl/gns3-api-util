package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateComputeCmd() *cobra.Command {
	var (
		protocol  string
		host      string
		port      int
		user      string
		password  string
		name      string
		computeID string
		connect   bool
		useJSON   string
	)
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a compute",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if err := validateChoice(protocol, []string{"http", "https"}, "--protocol"); err != nil {
				return err
			}
			var payload map[string]any
			if useJSON == "" {
				if protocol == "" && host == "" && port == 0 && user == "" && password == "" && name == "" && computeID == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				data := schemas.ComputeCreate{}
				if protocol != "" {
					v := protocol
					data.Protocol = &v
				}
				if host != "" {
					v := host
					data.Host = &v
				}
				if port != 0 {
					v := port
					data.Port = &v
				}
				if user != "" {
					v := user
					data.User = &v
				}
				if password != "" {
					v := password
					data.Password = &v
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
			utils.ExecuteAndPrintWithBody(cfg, "createCompute", []string{fmt.Sprintf("%t", connect)}, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&protocol, "protocol", "r", "", "Protocol (http/https)")
	cmd.Flags().StringVarP(&host, "host", "o", "", "Remote host")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "Port")
	cmd.Flags().StringVarP(&user, "user", "u", "", "Username")
	cmd.Flags().StringVarP(&password, "password", "w", "", "Password")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the compute")
	cmd.Flags().StringVarP(&computeID, "compute-id", "d", "", "Compute ID (generated if empty)")
	cmd.Flags().BoolVarP(&connect, "connect", "c", false, "Attempt connection after creation")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
