package objstorecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/pkg/models"
)

// Scaffolded by tools/scaffoldctl from ".codegen/scaffoldctl/packages/vm-images.yaml"; edit command bodies as needed.
func NewVMImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vm-images",
		Aliases: []string{"vms", "vm"},
		Short:   "Manage VM image objects in the filestore",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	queryGroup := &cobra.Group{ID: "query", Title: "Query commands:"}
	actionGroup := &cobra.Group{ID: "action", Title: "Action commands:"}
	cmd.AddGroup(
		queryGroup,
		actionGroup,
	)
	deleteVMImageCmd := newDeleteVMImageCmd()
	deleteVMImageCmd.GroupID = "action"
	getVMImageCmd := newGetVMImageCmd()
	getVMImageCmd.GroupID = "query"
	listVMImagesCmd := newListVMImagesCmd()
	listVMImagesCmd.GroupID = "query"
	updateVMImageCmd := newUpdateVMImageCmd()
	updateVMImageCmd.GroupID = "action"
	uploadVMImageCmd := newUploadVMImageCmd()
	uploadVMImageCmd.GroupID = "action"

	cmd.AddCommand(
		deleteVMImageCmd,
		getVMImageCmd,
		listVMImagesCmd,
		updateVMImageCmd,
		uploadVMImageCmd,
	)

	return cmd
}

type vmImageRow struct {
	FileUUID  string    `json:"file_uuid"`
	Filename  string    `json:"filename"`
	Status    string    `json:"status"`
	VirtType  string    `json:"virt_type"`
	Format    string    `json:"format"`
	VCPUs     int64     `json:"vcpus"`
	RAMMB     int64     `json:"ram_mb"`
	CreatedAt time.Time `json:"created_at"`
}

func (v *vmImageRow) GetHeaders() []string {
	return []string{
		"FILE UUID",
		"FILENAME",
		"STATUS",
		"VIRT TYPE",
		"FORMAT",
		"VCPUS",
		"RAM MB",
		"CREATED AT",
	}
}

func (v *vmImageRow) GetRow() []string {
	return []string{
		v.FileUUID,
		v.Filename,
		v.Status,
		v.VirtType,
		v.Format,
		fmt.Sprint(v.VCPUs),
		fmt.Sprint(v.RAMMB),
		v.CreatedAt.Format(time.RFC3339),
	}
}

type vmImageUploadResult struct {
	FileUUID string `json:"file_uuid"`
	Status   string `json:"status"`
}

func (v *vmImageUploadResult) GetHeaders() []string {
	return []string{
		"FILE UUID",
		"STATUS",
	}
}

func (v *vmImageUploadResult) GetRow() []string {
	return []string{
		v.FileUUID,
		v.Status,
	}
}

type vmImageDeleteResult struct {
	FileUUID string `json:"file_uuid"`
	Status   string `json:"status"`
}

func (v *vmImageDeleteResult) GetHeaders() []string {
	return []string{
		"FILE UUID",
		"STATUS",
	}
}

func (v *vmImageDeleteResult) GetRow() []string {
	return []string{
		v.FileUUID,
		v.Status,
	}
}

func newDeleteVMImageCmd() *cobra.Command {
	var fileStoreName string
	cmd := &cobra.Command{
		Use:     "delete <file-uuid>",
		Aliases: []string{"rm"},
		Short:   "Delete a VM image object",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			fileUUID := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			resp, err := client.DeleteVMImage(cmd.Context(), fileUUID)
			if err != nil {
				return err
			}
			return printRowOrObject[vmImageDeleteResult](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newGetVMImageCmd() *cobra.Command {
	var fileStoreName string
	cmd := &cobra.Command{
		Use:   "get <file-uuid>",
		Short: "Get VM image details",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			fileUUID := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			resp, err := client.GetVMImage(cmd.Context(), fileUUID)
			if err != nil {
				return err
			}
			return printRowOrObject[vmImageRow](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newListVMImagesCmd() *cobra.Command {
	var fileStoreName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List VM image objects",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			resp, err := client.ListVMImages(cmd.Context())
			if err != nil {
				return err
			}
			return printRowsOrObject[vmImageRow](cmd, cfg, resp.VMs, "No VM images found.")
		},
	}
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newUpdateVMImageCmd() *cobra.Command {
	var (
		fileStoreName       string
		virtType            string
		format              string
		vcpus               int64
		rammb               int64
		extraAttributesJSON string
	)
	cmd := &cobra.Command{
		Use:   "update <file-uuid>",
		Short: "Update VM image metadata",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			fileUUID := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			req := &models.UpdateVMImageRequest{}
			if !cmd.Flags().Changed("virt-type") && !cmd.Flags().Changed("format") && !cmd.Flags().Changed("vcpus") && !cmd.Flags().Changed("ram-mb") && !cmd.Flags().Changed("extra-attributes-json") {
				return fmt.Errorf("no update flags were provided")
			}
			if cmd.Flags().Changed("vcpus") {
				req.VCPUs = &vcpus
			}
			if cmd.Flags().Changed("ram-mb") {
				req.RAMMB = &rammb
			}
			if cmd.Flags().Changed("virt-type") {
				req.VirtType = virtType
			}
			if cmd.Flags().Changed("format") {
				req.Format = format
			}
			if cmd.Flags().Changed("extra-attributes-json") {
				req.ExtraAttributesJSON = extraAttributesJSON
			}
			resp, err := client.UpdateVMImage(cmd.Context(), fileUUID, req)
			if err != nil {
				return err
			}
			return printRowOrObject[vmImageRow](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVar(&virtType, "virt-type", "", "Virtualization type: qemu, iou, docker, dynamips, vmware, or virtualbox")
	cmd.Flags().StringVar(&format, "format", "", "VM disk/image format")
	cmd.Flags().Int64Var(&vcpus, "vcpus", 0, "Virtual CPU count")
	cmd.Flags().Int64Var(&rammb, "ram-mb", 0, "RAM size in MiB")
	cmd.Flags().StringVar(&extraAttributesJSON, "extra-attributes-json", "", "Extra metadata JSON for the VM image")
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")

	return cmd
}

func newUploadVMImageCmd() *cobra.Command {
	var (
		fileStoreName       string
		contentType         string
		retentionPeriod     int64
		bucketID            string
		virtType            string
		format              string
		vcpus               int64
		rammb               int64
		extraAttributesJSON string
	)
	cmd := &cobra.Command{
		Use:     "upload <file-path>",
		Aliases: []string{"add"},
		Short:   "Upload a VM image object",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			filePath := args[0]
			client, _, err := newFilestoreClient(cfg, fileStoreName)
			if err != nil {
				return err
			}
			file, err := os.Open(filePath) // #nosec G304
			if err != nil {
				return err
			}
			defer file.Close()
			stat, err := file.Stat()
			if err != nil {
				return err
			}
			req := &models.InitVMUploadRequest{
				Filename:            filepath.Base(filePath),
				SizeBytes:           stat.Size(),
				ContentType:         contentType,
				BucketID:            bucketID,
				VirtType:            virtType,
				Format:              format,
				ExtraAttributesJSON: extraAttributesJSON,
			}
			if req.BucketID == "" {
				req.BucketID = models.GlobalBucketID
			}
			if cmd.Flags().Changed("retention") {
				req.RetentionPeriod = &retentionPeriod
			}
			if cmd.Flags().Changed("vcpus") {
				req.VCPUs = &vcpus
			}
			if cmd.Flags().Changed("ram-mb") {
				req.RAMMB = &rammb
			}
			initResp, err := client.InitVMUpload(cmd.Context(), req)
			if err != nil {
				return err
			}
			status, err := client.GetUploadStatus(cmd.Context(), initResp.FileUUID)
			if err != nil {
				return err
			}
			if status.Offset >= stat.Size() {
				return printRowOrObject[vmImageUploadResult](cmd, cfg, &models.FinalizeUploadResponse{
					FileUUID: initResp.FileUUID,
					Status:   models.FileStatusAvailable,
				})
			}
			if status.Offset > 0 {
				if _, seekErr := file.Seek(status.Offset, 0); seekErr != nil {
					return fmt.Errorf("failed to seek local file: %w", seekErr)
				}
			}
			resp, err := client.StreamUpload(cmd.Context(), initResp.FileUUID, file, status.Offset)
			if err != nil {
				return err
			}
			return printRowOrObject[vmImageUploadResult](cmd, cfg, resp)
		},
	}
	cmd.Flags().StringVarP(&contentType, "type", "t", "application/octet-stream", "Explicit Content-Type for the VM image file")
	cmd.Flags().Int64VarP(&retentionPeriod, "retention", "r", 0, "Retention period in hours")
	cmd.Flags().StringVar(&bucketID, "bucket-id", "", "Target bucket ID")
	cmd.Flags().StringVar(&virtType, "virt-type", "", "Virtualization type: qemu, iou, docker, dynamips, vmware, or virtualbox")
	cmd.Flags().StringVar(&format, "format", "", "VM disk/image format")
	cmd.Flags().Int64Var(&vcpus, "vcpus", 0, "Virtual CPU count")
	cmd.Flags().Int64Var(&rammb, "ram-mb", 0, "RAM size in MiB")
	cmd.Flags().StringVar(&extraAttributesJSON, "extra-attributes-json", "", "Extra metadata JSON for the VM image")
	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	if err := cmd.MarkFlagRequired("virt-type"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("format"); err != nil {
		panic(err)
	}

	return cmd
}
