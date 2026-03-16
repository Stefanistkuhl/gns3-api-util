package objstorecmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/pathutils"
	"github.com/0xveya/gns3util/pkg/models"
	"github.com/0xveya/gns3util/pkg/web/helpers"
)

type UploadResult struct {
	FilestoreID  string                         `json:"filestore_id"`
	FilestoreURL string                         `json:"filestore_url"`
	Filename     string                         `json:"filename"`
	Response     *models.FinalizeUploadResponse `json:"response"`
}

func NewUploadCmd() *cobra.Command {
	var (
		fileStoreName string
		scope         string
		contentType   string
		retention     int64
	)

	cmd := &cobra.Command{
		Use:     "upload [file_path]",
		Aliases: []string{"u", "add"},
		Short:   "Upload a file to the cluster filestore in a bucket",
		Long:    "Upload a file to the cluster filestore in a bucket, if no bucket is set the default bucket is used",
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
			clusterName := cfg.Cluster

			keyPath, err := pathutils.ResolveKeyFilePath("")
			if err != nil {
				return err
			}
			kf, err := pathutils.LoadGNS3KeysFile(keyPath)
			if err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			var targetCluster *pathutils.ClusterEntry
			for i := range kf.Clusters {
				if kf.Clusters[i].Name == clusterName {
					targetCluster = &kf.Clusters[i]
					break
				}
			}

			if targetCluster == nil {
				return fmt.Errorf("cluster %q not found in keys file", clusterName)
			}

			cfg.ClusterEntry.CaCert = targetCluster.CaCert

			var available []pathutils.ServiceEntry
			for _, node := range targetCluster.Nodes {
				if node.Type == pathutils.TypeClusterFileStore {
					available = append(available, node)
				}
			}

			if len(available) == 0 {
				return fmt.Errorf("no filestore found in cluster %q", clusterName)
			}

			var filestore *pathutils.ServiceEntry
			if fileStoreName != "" {
				for _, fs := range available {
					if fs.ID == fileStoreName {
						filestore = &fs
						break
					}
				}
				if filestore == nil {
					return fmt.Errorf("filestore ID %q not found", fileStoreName)
				}
			} else {
				if len(available) > 1 {
					type fsOption struct {
						ID  string `json:"filestore_id"`
						URL string `json:"filestore_url"`
					}
					var options []fsOption
					for _, fs := range available {
						options = append(options, fsOption{ID: fs.ID, URL: fs.URL})
					}

					body, _ := json.Marshal(struct {
						Error   string     `json:"error"`
						Message string     `json:"message"`
						Options []fsOption `json:"options"`
					}{
						Error:   "ambiguous filestore selection",
						Message: "Multiple filestores detected. Specify one with --filestore-id",
						Options: options,
					})

					utils.PrintOutput(body, cfg)

					return fmt.Errorf("ambiguous filestore selection")
				}
				filestore = &available[0]
			}

			uploadReq := models.InitUploadRequest{
				ScopeLabel:      scope,
				ContentType:     contentType,
				RetentionPeriod: retention,
				Filename:        filepath.Base(filePath),
				BucketID:        models.GlobalBucketID,
			}

			resp, err := helpers.RunUpload(
				filestore.URL,
				cfg.ClusterEntry.Master.AccessToken,
				filePath,
				&uploadReq,
				cfg,
			)
			if err != nil {
				return err
			}
			result := UploadResult{
				FilestoreID:  filestore.ID,
				FilestoreURL: filestore.URL,
				Filename:     filepath.Base(filePath),
				Response:     resp,
			}

			body, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed to marshal upload result: %w", err)
			}

			utils.PrintOutput(body, cfg)

			return nil
		},
	}

	cmd.Flags().StringVar(&fileStoreName, "filestore-id", "", "Specific filestore node ID")
	cmd.Flags().StringVarP(&scope, "scope", "", "default", "Storage scope/label (e.g. 'iso', 'images')")
	cmd.Flags().StringVarP(&contentType, "type", "t", "application/octet-stream", "Explicit Content-Type for the file")
	cmd.Flags().Int64VarP(&retention, "retention", "r", 0, "Retention period in hours (0 = permanent)")

	return cmd
}
