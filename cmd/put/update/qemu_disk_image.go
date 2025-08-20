package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewUpdateQemuDiskImageCmd() *cobra.Command {
	var (
		format        string
		size          int
		preallocation string
		clusterSize   int
		refcountBits  int
		lazyRefcounts string
		subformat     string
		static        string
		zeroedGrain   string
		adapterType   string
		useJSON       string
	)

	cmd := &cobra.Command{
		Use:     "qemu-disk-image [project-name/id] [node-name/id] [disk-name]",
		Short:   "Update a new disk for a node",
		Long:    "Update a new disk for a node in a project.",
		Example: "gns3util -s https://controller:3080 update qemu-disk-image [project-id] [node-id] [disk-name] -f qcow2 -z 1024",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			projectID := args[0]
			nodeID := args[1]
			diskName := args[2]

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
			if err := validateChoice(format, []string{"qcow2", "qcow", "vpc", "vdi", "vdmk", "raw"}, "--format"); err != nil {
				return err
			}
			if err := validateChoice(preallocation, []string{"off", "metadata", "falloc", "full"}, "--preallocation"); err != nil {
				return err
			}
			if err := validateChoice(lazyRefcounts, []string{"on", "off"}, "--lazy_refcounts"); err != nil {
				return err
			}
			if err := validateChoice(subformat, []string{"dynamic", "fixed", "streamOptimized", "twoGbMaxExtentSparse", "twoGbMaxExtentFlat", "monolithicSparse", "monolithicFlat"}, "--subformat"); err != nil {
				return err
			}
			if err := validateChoice(static, []string{"on", "off"}, "--static"); err != nil {
				return err
			}
			if err := validateChoice(zeroedGrain, []string{"on", "off"}, "--zeroed-grain"); err != nil {
				return err
			}
			if err := validateChoice(adapterType, []string{"idle", "lsilogic", "buslogic", "legacyESX"}, "--adapter-type"); err != nil {
				return err
			}

			var payload map[string]any
			if useJSON == "" {
				if format == "" && size == 0 && preallocation == "" && clusterSize == 0 && refcountBits == 0 && lazyRefcounts == "" && subformat == "" && static == "" && zeroedGrain == "" && adapterType == "" {
					return fmt.Errorf("at least one field is required or provide --use-json")
				}

				payload = map[string]any{}
				if format != "" {
					payload["format"] = format
				}
				if size != 0 {
					payload["size"] = size
				}
				if preallocation != "" {
					payload["preallocation"] = preallocation
				}
				if clusterSize != 0 {
					payload["cluster_size"] = clusterSize
				}
				if refcountBits != 0 {
					payload["refcount_bits"] = refcountBits
				}
				if lazyRefcounts != "" {
					payload["lazy_refcounts"] = lazyRefcounts
				}
				if subformat != "" {
					payload["subformat"] = subformat
				}
				if static != "" {
					payload["static"] = static
				}
				if zeroedGrain != "" {
					payload["zeroed_grain"] = zeroedGrain
				}
				if adapterType != "" {
					payload["adapter_type"] = adapterType
				}
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "updateQemuDiskImage", []string{projectID, nodeID, diskName}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "", "Type of image format")
	cmd.Flags().IntVarP(&size, "size", "", 0, "Size of disk in megabytes")
	cmd.Flags().StringVarP(&preallocation, "preallocation", "p", "", "Desired Qemu disk image pre-allocation option")
	cmd.Flags().IntVarP(&clusterSize, "cluster-size", "c", 0, "Desired cluster size")
	cmd.Flags().IntVarP(&refcountBits, "refcount-bits", "r", 0, "Desired amount of refcount bits")
	cmd.Flags().StringVarP(&lazyRefcounts, "lazy_refcounts", "l", "", "Enabling or disabling lazy refcounts")
	cmd.Flags().StringVarP(&subformat, "subformat", "", "", "Desired image sub-format options")
	cmd.Flags().StringVarP(&static, "static", "", "", "Static option")
	cmd.Flags().StringVarP(&zeroedGrain, "zeroed-grain", "", "", "Zeroed grain option")
	cmd.Flags().StringVarP(&adapterType, "adapter-type", "", "", "Adapter type")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")

	return cmd
}
