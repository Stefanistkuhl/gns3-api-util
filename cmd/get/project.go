package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func NewGetProjectsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "projects",
		Short: "Get the projects of the GNS3 Server",
		Long:  `Get the projects of the GNS3 Server`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getProjects", nil)
		},
	}
	return cmd
}

func NewGetProjectCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Get a project by id or name",
		Long:  `Get a project by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getProject", []string{id})
		},
	}
	return cmd
}

func NewGetProjectStatsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "project-stats",
		Short: "Get project-stats by id or name",
		Long:  `Get project-stats by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getProjectStats", []string{id})
		},
	}
	return cmd
}

func NewGetProjectLockedCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "project-locked",
		Short: "Get if a project is locked by id or name",
		Long:  `Get if a project is locked by id or name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getProjectLocked", []string{id})
		},
	}
	return cmd
}

func NewGetProjectExportCmd() *cobra.Command {
	var (
		includeSnapshots  bool
		includeImages     bool
		resetMacAddresses bool
		keepComputeIds    bool
		compression       string
		compressionLevel  int
		outputFile        string
	)

	var cmd = &cobra.Command{
		Use:   "project-export",
		Short: "Export a project from GNS3",
		Long:  `Export a project from GNS3 with specified options`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			// Get project name for default filename if not specified
			projectName := args[0]
			if utils.IsValidUUIDv4(args[0]) {
				projectName, err = getProjectNameFromID(cfg, id)
				if err != nil {
					fmt.Printf("failed to get project name: %v", err)
					return
				}
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("%s.gns3project", projectName)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				fmt.Printf("failed to get token: %v", err)
				return
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(settings)

			reqOpts := api.NewRequestOptions(settings).
				WithURL(fmt.Sprintf("/projects/%s/export", id)).
				WithMethod(api.GET)

			exportData, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to export project: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("export failed with status %d: %s", resp.StatusCode, string(exportData))
				return
			}

			err = os.WriteFile(outputFile, exportData, 0644)
			if err != nil {
				fmt.Printf("failed to save export file: %v", err)
				return
			}

			fmt.Printf("%s Project exported successfully to %s", colorUtils.Success("Success:"), colorUtils.Bold(outputFile))
		},
	}

	cmd.Flags().BoolVar(&includeSnapshots, "include-snapshots", false, "Include snapshots in the export")
	cmd.Flags().BoolVar(&includeImages, "include-images", false, "Include images in the export")
	cmd.Flags().BoolVar(&resetMacAddresses, "reset-mac-addresses", false, "Reset MAC addresses in the export")
	cmd.Flags().BoolVar(&keepComputeIds, "keep-compute-ids", false, "Keep compute IDs in the export")
	cmd.Flags().StringVar(&compression, "compression", "zstd", "Compression type for the export (deflate, bz2, xz, zstd, none)")
	cmd.Flags().IntVar(&compressionLevel, "compression-level", 3, "Compression level for the export (0-9)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (default: project-name.gns3project)")

	return cmd
}

func getProjectNameFromID(cfg config.GlobalOptions, projectID string) (string, error) {
	token, err := authentication.GetKeyForServer(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(cfg.Insecure),
		api.WithToken(token),
	)
	client := api.NewGNS3Client(settings)

	reqOpts := api.NewRequestOptions(settings).
		WithURL(fmt.Sprintf("/projects/%s", projectID)).
		WithMethod(api.GET)

	body, resp, err := client.Do(reqOpts)
	if err != nil {
		return "", fmt.Errorf("API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get project with status %d", resp.StatusCode)
	}

	var project map[string]interface{}
	if err := json.Unmarshal(body, &project); err != nil {
		return "", fmt.Errorf("failed to parse project response: %w", err)
	}

	if name, ok := project["name"].(string); ok {
		return name, nil
	}

	return "", fmt.Errorf("project name not found in response")
}

func NewGetProjectFileCmd() *cobra.Command {
	var outputFile string

	var cmd = &cobra.Command{
		Use:   "project-file",
		Short: "Get a file from a project",
		Long:  `Get a file from a project by project ID/name and file path`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			filePath := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			if outputFile == "" {
				outputFile = filepath.Base(filePath)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				fmt.Printf("failed to get token: %v", err)
				return
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(settings)

			reqOpts := api.NewRequestOptions(settings).
				WithURL(fmt.Sprintf("/projects/%s/files/%s", projectID, filePath)).
				WithMethod(api.GET)

			fileData, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to get project file: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("failed to get project file with status %d: %s", resp.StatusCode, string(fileData))
				return
			}

			err = os.WriteFile(outputFile, fileData, 0644)
			if err != nil {
				fmt.Printf("failed to save project file: %v", err)
				return
			}

			fmt.Printf("%s Project file downloaded successfully to %s", colorUtils.Success("Success:"), colorUtils.Bold(outputFile))
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (default: original filename)")

	return cmd
}

func NewGetNodeFileCmd() *cobra.Command {
	var outputFile string

	var cmd = &cobra.Command{
		Use:   "node-file",
		Short: "Get a file from a node",
		Long:  `Get a file from a node by project ID/name, node ID/name, and file path`,
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
			filePath := args[2]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			if !utils.IsValidUUIDv4(args[1]) {
				nodeID, err = utils.ResolveID(cfg, "node", args[1], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			if outputFile == "" {
				outputFile = filepath.Base(filePath)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				fmt.Printf("failed to get token: %v", err)
				return
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(settings)

			reqOpts := api.NewRequestOptions(settings).
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/files/%s", projectID, nodeID, filePath)).
				WithMethod(api.GET)

			fileData, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to get node file: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("failed to get node file with status %d: %s", resp.StatusCode, string(fileData))
				return
			}

			err = os.WriteFile(outputFile, fileData, 0644)
			if err != nil {
				fmt.Printf("failed to save node file: %v", err)
				return
			}

			fmt.Printf("%s Node file downloaded successfully to %s", colorUtils.Success("Success:"), colorUtils.Bold(outputFile))
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (default: original filename)")

	return cmd
}

func NewStreamPcapCmd() *cobra.Command {
	var outputFile string

	var cmd = &cobra.Command{
		Use:   "stream-pcap",
		Short: "Stream PCAP capture file from compute",
		Long:  `Stream the PCAP capture file from compute by project ID/name and link ID/name`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			linkID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			if !utils.IsValidUUIDv4(args[1]) {
				linkID, err = utils.ResolveID(cfg, "link", args[1], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("capture_%s.pcap", linkID)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				fmt.Printf("failed to get token: %v", err)
				return
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(settings)

			reqOpts := api.NewRequestOptions(settings).
				WithURL(fmt.Sprintf("/projects/%s/links/%s/capture/stream", projectID, linkID)).
				WithMethod(api.GET)

			fileData, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to stream PCAP: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("failed to stream PCAP with status %d: %s", resp.StatusCode, string(fileData))
				return
			}

			err = os.WriteFile(outputFile, fileData, 0644)
			if err != nil {
				fmt.Printf("failed to save PCAP file: %v", err)
				return
			}

			fmt.Printf("%s PCAP file streamed successfully to %s", colorUtils.Success("Success:"), colorUtils.Bold(outputFile))
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (default: capture_{link-id}.pcap)")

	return cmd
}
