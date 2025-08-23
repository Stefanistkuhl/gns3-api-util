package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateNodeCmd() *cobra.Command {
	var (
		computeID        string
		name             string
		nodeType         string
		console          int
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
		Use:     utils.UpdateSingleElementCmdName + " [project-name/id] [node-name/id]",
		Short:   "Update a Node in a Project",
		Long:    "Update a Node in a Project. To use custom adapters the --use-json option has to be used.",
		Example: "gns3util -s https://controller:3080 update [project-name/id] [node-name/id] --name new-name",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			projectID := args[0]
			nodeID := args[1]

			if !utils.IsValidUUIDv4(projectID) {
				resolvedID, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve project ID: %w", err)
				}
				projectID = resolvedID
			}

			if !utils.IsValidUUIDv4(nodeID) {
				resolvedID, err := utils.ResolveID(cfg, "node", args[1], []string{projectID})
				if err != nil {
					return fmt.Errorf("failed to resolve node ID: %w", err)
				}
				nodeID = resolvedID
			}

			// validate choice-like flags
			if err := validateChoice(nodeType, []string{"cloud", "nat", "ethernet_hub", "ethernet_switch", "frame_relay_switch", "atm_switch", "docker", "dynamips", "vpcs", "virtualbox", "vmware", "iou", "qemu"}, "--node-type"); err != nil {
				return err
			}
			if err := validateChoice(consoleType, []string{"vnc", "telnet", "http", "https", "spice", "spice+agent", "none"}, "--console-type"); err != nil {
				return err
			}
			if err := validateChoice(auxType, []string{"vnc", "telnet", "http", "https", "spice", "spice+agent", "none"}, "--aux-type"); err != nil {
				return err
			}

			var payload map[string]any
			if useJSON == "" {
				if computeID == "" && name == "" && nodeType == "" && console == 0 && consoleType == "" && !consoleAutoStart && aux == 0 && auxType == "" && properties == "" && labelText == "" && labelStyle == "" && labelX == 0 && labelY == 0 && labelRotation == 0 && symbol == "" && x == 0 && y == 0 && z == 0 && !locked && portNameFormat == "" && portSegmentSize == 0 && firstPortName == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				data := schemas.NodeUpdate{}
				if computeID != "" {
					data.ComputeID = &computeID
				}
				if name != "" {
					data.Name = &name
				}
				if nodeType != "" {
					data.NodeType = &nodeType
				}
				if console != 0 {
					data.Console = &console
				}
				if consoleType != "" {
					data.ConsoleType = &consoleType
				}
				if consoleAutoStart {
					data.ConsoleAutoStart = &consoleAutoStart
				}
				if aux != 0 {
					data.Aux = &aux
				}
				if auxType != "" {
					data.AuxType = &auxType
				}
				if properties != "" {
					var props map[string]any
					if err := json.Unmarshal([]byte(properties), &props); err != nil {
						return fmt.Errorf("invalid JSON for --properties: %w", err)
					}
					data.Properties = props
				}
				if labelText != "" || labelStyle != "" || labelX != 0 || labelY != 0 || labelRotation != 0 {
					label := schemas.Label{}
					if labelText != "" {
						label.Text = &labelText
					}
					if labelStyle != "" {
						label.Style = &labelStyle
					}
					if labelX != 0 {
						label.X = &labelX
					}
					if labelY != 0 {
						label.Y = &labelY
					}
					if labelRotation != 0 {
						label.Rotation = &labelRotation
					}
					data.Label = &label
				}
				if symbol != "" {
					data.Symbol = &symbol
				}
				if x != 0 {
					data.X = &x
				}
				if y != 0 {
					data.Y = &y
				}
				if z != 0 {
					data.Z = &z
				}
				if locked {
					data.Locked = &locked
				}
				if portNameFormat != "" {
					data.PortNameFormat = &portNameFormat
				}
				if portSegmentSize != 0 {
					data.PortSegmentSize = &portSegmentSize
				}
				if firstPortName != "" {
					data.FirstPortName = &firstPortName
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
			utils.ExecuteAndPrintWithBody(cfg, "updateNode", []string{projectID, nodeID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&computeID, "compute-id", "c", "local", "Compute on that the Node gets created")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Desired name for the Node")
	cmd.Flags().StringVarP(&nodeType, "node-type", "t", "", "Type of Node")
	cmd.Flags().IntVarP(&console, "console", "p", 0, "TCP port of the console")
	cmd.Flags().StringVarP(&consoleType, "console-type", "", "", "Type of the console interface")
	cmd.Flags().BoolVarP(&consoleAutoStart, "console-auto-start", "", false, "Automatically start the console when the node has started")
	cmd.Flags().IntVarP(&aux, "aux", "a", 0, "Auxiliary console TCP port")
	cmd.Flags().StringVarP(&auxType, "aux-type", "", "", "Type of the aux console")
	cmd.Flags().StringVarP(&properties, "properties", "", "", "Custom properties JSON")
	cmd.Flags().StringVarP(&labelText, "label-text", "", "", "Text of the label of the Node")
	cmd.Flags().StringVarP(&labelStyle, "label-style-attribute", "", "", "SVG style attribute")
	cmd.Flags().IntVarP(&labelX, "label-x-position", "", 0, "X-Position of the label")
	cmd.Flags().IntVarP(&labelY, "label-y-position", "", 0, "Y-Position of the label")
	cmd.Flags().IntVarP(&labelRotation, "label-rotation", "", 0, "Rotation of the label")
	cmd.Flags().StringVar(&symbol, "symbol", "", "Name of the desired symbol")
	cmd.Flags().IntVarP(&x, "x", "x", 0, "X-Position of the node")
	cmd.Flags().IntVarP(&y, "y", "y", 0, "Y-Position of the node")
	cmd.Flags().IntVarP(&z, "z", "z", 1, "Z-Position (layer) of the node")
	cmd.Flags().BoolVarP(&locked, "locked", "l", false, "Whether the node is locked or not")
	cmd.Flags().StringVarP(&portNameFormat, "port-name-format", "", "", "Name format for the port for example: Ethernet{0}")
	cmd.Flags().IntVarP(&portSegmentSize, "port-segment-size", "", 0, "Port segment size")
	cmd.Flags().StringVarP(&firstPortName, "first-port-name", "", "", "Name of the first port")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
