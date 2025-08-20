package create

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewCreateQemuImageCmd() *cobra.Command {
	var (
		imagePath     string
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
		Use:   "qemu-img [image-path]",
		Short: "Create a QEMU disk image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			imagePath = args[0]
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
				if format == "" || size == 0 {
					return fmt.Errorf("for this command --format and --size are required or provide --use-json")
				}
				data := schemas.QemuDiskImageCreate{}
				if format != "" {
					v := format
					data.Format = &v
				}
				if size != 0 {
					v := size
					data.Size = &v
				}
				if preallocation != "" {
					v := preallocation
					data.Preallocation = &v
				}
				if clusterSize != 0 {
					v := clusterSize
					data.ClusterSize = &v
				}
				if refcountBits != 0 {
					v := refcountBits
					data.RefcountBits = &v
				}
				if lazyRefcounts != "" {
					v := lazyRefcounts
					data.LazyRefcounts = &v
				}
				if subformat != "" {
					v := subformat
					data.Subformat = &v
				}
				if static != "" {
					v := static
					data.Static = &v
				}
				if zeroedGrain != "" {
					v := zeroedGrain
					data.ZeroedGrain = &v
				}
				if adapterType != "" {
					v := adapterType
					data.AdapterType = &v
				}
				b, _ := json.Marshal(data)
				_ = json.Unmarshal(b, &payload)
			} else {
				if err := json.Unmarshal([]byte(useJSON), &payload); err != nil {
					return fmt.Errorf("invalid JSON for --use-json: %w", err)
				}
			}
			utils.ExecuteAndPrintWithBody(cfg, "createQemuImage", []string{imagePath}, payload)
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "", "Image format (qcow2, raw, ...)")
	cmd.Flags().IntVarP(&size, "size", "z", 0, "Image size in MB")
	cmd.Flags().StringVarP(&preallocation, "preallocation", "p", "", "Preallocation (off, metadata, falloc, full)")
	cmd.Flags().IntVarP(&clusterSize, "cluster-size", "c", 0, "Cluster size")
	cmd.Flags().IntVarP(&refcountBits, "refcount-bits", "r", 0, "Refcount bits")
	cmd.Flags().StringVarP(&lazyRefcounts, "lazy_refcounts", "l", "", "lazy_refcounts (on/off)")
	cmd.Flags().StringVarP(&subformat, "subformat", "u", "", "Subformat")
	cmd.Flags().StringVarP(&static, "static", "t", "", "static (on/off)")
	cmd.Flags().StringVarP(&zeroedGrain, "zeroed-grain", "g", "", "zeroed-grain (on/off)")
	cmd.Flags().StringVarP(&adapterType, "adapter-type", "a", "", "Adapter type (idle, lsilogic, ...)")
	cmd.Flags().StringVarP(&useJSON, "use-json", "j", "", "Provide a raw JSON string to send instead of flags")
	return cmd
}
