package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateNodeCmd() *cobra.Command {
	var (
		computeID        string
		name             string
		nodeType         string
		consolePort      int
		consoleType      string
		consoleAutoStart bool
		aux              int
		auxType          string
		properties       string
		labelText        string
		labelStyle       string
		labelX           int
		labelY           int
		labelRotation    int
		symbol           string
		x                int
		y                int
		z                int
		locked           bool
		portNameFormat   string
		portSegmentSize  int
		firstPortName    string
		useJSON          string
	)
	cmd := &cobra.Command{
		Use:   "new [project-id]",
		Short: "Create a node in a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			// emulate click.Choice validation for selected flags where applicable
			if err := validateChoice(consoleType, []string{"vnc", "telnet", "http", "https", "spice", "spice+agent", "none"}, "--console-type"); err != nil {
				return err
			}
			if err := validateChoice(auxType, []string{"vnc", "telnet", "http", "https", "spice", "spice+agent", "none"}, "--aux-type"); err != nil {
				return err
			}
			projectID := args[0]
			var payload map[string]any
			if useJSON == "" {
				if name == "" || nodeType == "" {
					return fmt.Errorf("for this command -n/--name and --node-type are required or provide --use-json")
				}
				data := schemas.NodeCreate{}
				if computeID != "" {
					v := computeID
					data.ComputeID = &v
				}
				if name != "" {
					v := name
					data.Name = &v
				}
				if nodeType != "" {
					v := nodeType
					data.NodeType = &v
				}
				if consolePort != 0 {
					v := consolePort
					data.Console = &v
				}
				if consoleType != "" {
					v := consoleType
					data.ConsoleType = &v
				}
				if consoleAutoStart {
					v := true
					data.ConsoleAutoStart = &v
				}
				if aux != 0 {
					v := aux
					data.Aux = &v
				}
				if auxType != "" {
					v := auxType
					data.AuxType = &v
				}
				if labelText != "" {
					lbl := schemas.Label{}
					if labelText != "" {
						v := labelText
						lbl.Text = &v
					}
					if labelStyle != "" {
						v := labelStyle
						lbl.Style = &v
					}
					if labelX != 0 {
						v := labelX
						lbl.X = &v
					}
					if labelY != 0 {
						v := labelY
						lbl.Y = &v
					}
					if labelRotation != 0 {
						v := labelRotation
						lbl.Rotation = &v
					}
					data.Label = &lbl
				}
				if symbol != "" {
					v := symbol
					data.Symbol = &v
				}
				if x != 0 {
					v := x
					data.X = &v
				}
				if y != 0 {
					v := y
					data.Y = &v
				}
				if z != 0 {
					v := z
					data.Z = &v
				}
				if locked {
					v := true
					data.Locked = &v
				}
				if portNameFormat != "" {
					v := portNameFormat
					data.PortNameFormat = &v
				}
				if portSegmentSize != 0 {
					v := portSegmentSize
					data.PortSegmentSize = &v
				}
				if firstPortName != "" {
					v := firstPortName
					data.FirstPortName = &v
				}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
				if properties != "" {
					var props map[string]any
					if err := json.Unmarshal([]byte(properties), &props); err == nil {
						payload["properties"] = props
					}
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createNode", []string{projectID}, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&computeID, "compute-id", "c", "", "Compute ID (default 'local')")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Node name")
	cmd.Flags().StringVar(&nodeType, "node-type", "", "Node type")
	cmd.Flags().IntVar(&consolePort, "console-port", 0, "Console TCP port")
	cmd.Flags().StringVar(&consoleType, "console-type", "", "Console type (e.g., telnet)")
	cmd.Flags().BoolVar(&consoleAutoStart, "console-auto-start", false, "Automatically start console")
	cmd.Flags().IntVar(&aux, "aux", 0, "Aux console port")
	cmd.Flags().StringVar(&auxType, "aux-type", "", "Aux console type")
	cmd.Flags().StringVar(&properties, "properties", "", "Custom properties JSON")
	cmd.Flags().StringVar(&labelText, "label-text", "", "Label text")
	cmd.Flags().StringVarP(&labelStyle, "label-style-attribute", "a", "", "SVG style attribute for label")
	cmd.Flags().IntVarP(&labelX, "label-x-position", "x", 0, "Label X position")
	cmd.Flags().IntVarP(&labelY, "label-y-position", "y", 0, "Label Y position")
	cmd.Flags().IntVarP(&labelRotation, "label-rotation", "r", 0, "Label rotation")
	cmd.Flags().StringVar(&symbol, "symbol", "", "Symbol name")
	cmd.Flags().IntVar(&x, "x", 0, "X coordinate")
	cmd.Flags().IntVar(&y, "y", 0, "Y coordinate")
	cmd.Flags().IntVar(&z, "z", 1, "Z layer")
	cmd.Flags().BoolVar(&locked, "locked", false, "Lock node")
	cmd.Flags().StringVarP(&portNameFormat, "port-name-format", "f", "", "Port name format (e.g., Ethernet{0})")
	cmd.Flags().IntVarP(&portSegmentSize, "port-segment-size", "m", 0, "Port segment size")
	cmd.Flags().StringVarP(&firstPortName, "first-port-name", "o", "", "First port name")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
