package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/authentication"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/spf13/cobra"
)

func NewGetProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get the projects of the GNS3 Server",
		Long:    `Get the projects of the GNS3 Server`,
		Example: "gns3util -s https://controller:3080 project ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getProjects", nil)
			return nil
		},
	}
	return cmd
}

func NewGetProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info [project-name/id]",
		Aliases: []string{"get", "i"},
		Short:   "Get a project by id or name",
		Long:    `Get a project by id or name`,
		Example: "gns3util -s https://controller:3080 project info my-project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getProject", []string{id})
			return nil
		},
	}
	return cmd
}

func NewGetProjectStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats [project-name/id]",
		Aliases: []string{"st"},
		Short:   "Get project-stats by id or name",
		Long:    `Get project-stats by id or name`,
		Example: "gns3util -s https://controller:3080 project stats my-project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getProjectStats", []string{id})
			return nil
		},
	}
	return cmd
}

func NewGetProjectLockedCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "locked [project-name/id]",
		Short:   "Get if a project is locked by id or name",
		Long:    `Get if a project is locked by id or name`,
		Example: "gns3util -s https://controller:3080 project locked my-project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getProjectLocked", []string{id})
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a project")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple projects")
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

	cmd := &cobra.Command{
		Use:     "export [project-name/id]",
		Short:   "Export a project from GNS3",
		Long:    `Export a project from GNS3 with specified options`,
		Example: "gns3util -s https://controller:3080 project export my-project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}

			projectName := args[0]
			if utils.IsValidUUIDv4(args[0]) {
				projectName, err = getProjectNameFromID(cfg, id)
				if err != nil {
					return fmt.Errorf("failed to get project name: %w", err)
				}
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("%s.gns3project", projectName)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(&settings)

			reqOpts := api.NewRequestOptions(&settings).
				WithURL(fmt.Sprintf("/projects/%s/export", id)).
				WithMethod(api.GET)

			exportData, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to export project: %w", err)
			}
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					// Keep silent on body close errors
					_ = closeErr
				}
			}()

			if resp.StatusCode != 200 {
				return fmt.Errorf("export failed with status %d: %s", resp.StatusCode, string(exportData))
			}

			err = os.WriteFile(outputFile, exportData, 0o600)
			if err != nil {
				return fmt.Errorf("failed to save export file: %w", err)
			}

			fmt.Printf("%s Project exported successfully to %s", messageUtils.SuccessMsg("Project exported successfully"), messageUtils.Bold(outputFile))
			return nil
		},
	}

	cmd.Flags().BoolVar(&includeSnapshots, "include-snapshots", false, "Include snapshots in the export")
	cmd.Flags().BoolVar(&includeImages, "include-images", false, "Include images in the export")
	cmd.Flags().BoolVar(&resetMacAddresses, "reset-mac-addresses", false, "Reset MAC addresses in the export")
	cmd.Flags().BoolVar(&keepComputeIds, "keep-compute-ids", false, "Keep compute IDs in the export")
	cmd.Flags().StringVar(&compression, "compression", "zstd", "Compression type for the export (deflate, bz2, xz, zstd, none)")
	cmd.Flags().IntVar(&compressionLevel, "compression-level", 3, "Compression level for the export (0-9)")
	cmd.Flags().StringVarP(&outputFile, "output", "", "", "Output filename (default: project-name.gns3project)")

	return cmd
}

func getProjectNameFromID(cfg *config.GlobalOptions, projectID string) (string, error) {
	token, err := authentication.GetKeyForServer(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(cfg.Insecure),
		api.WithToken(token),
	)
	client := api.NewGNS3Client(&settings)

	reqOpts := api.NewRequestOptions(&settings).
		WithURL(fmt.Sprintf("/projects/%s", projectID)).
		WithMethod(api.GET)

	body, resp, err := client.Do(reqOpts)
	if err != nil {
		return "", fmt.Errorf("API error: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Keep silent on body close errors
			_ = err
		}
	}()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get project with status %d", resp.StatusCode)
	}

	var project map[string]any
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

	cmd := &cobra.Command{
		Use:     "file [project-name/id] [file-path]",
		Short:   "Get a file from a project",
		Long:    `Get a file from a project by project ID/name and file path`,
		Example: "gns3util -s https://controller:3080 project file my-project /path/to/file",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			filePath := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}

			if outputFile == "" {
				outputFile = filepath.Base(filePath)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(&settings)

			reqOpts := api.NewRequestOptions(&settings).
				WithURL(fmt.Sprintf("/projects/%s/files/%s", projectID, filePath)).
				WithMethod(api.GET)

			fileData, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to get project file: %w", err)
			}
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					// Keep silent on body close errors
					_ = closeErr
				}
			}()

			if resp.StatusCode != 200 {
				return fmt.Errorf("failed to get project file with status %d: %s", resp.StatusCode, string(fileData))
			}

			err = os.WriteFile(outputFile, fileData, 0o600)
			if err != nil {
				return fmt.Errorf("failed to save project file: %w", err)
			}

			fmt.Printf("%s Project file downloaded successfully to %s", messageUtils.SuccessMsg("Project file downloaded successfully"), messageUtils.Bold(outputFile))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "", "", "Output filename (default: original filename)")

	return cmd
}

func NewGetNodeFileCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:     "node-file [project-name/id] [node-name/id] [file-path]",
		Short:   "Get a file from a node",
		Long:    `Get a file from a node by project ID/name, node ID/name, and file path`,
		Example: "gns3util -s https://controller:3080 project node-file my-project my-node /path/to/file",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			nodeID := args[1]
			filePath := args[2]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}

			if !utils.IsValidUUIDv4(args[1]) {
				nodeID, err = utils.ResolveID(cfg, "node", args[1], nil)
				if err != nil {
					return err
				}
			}

			if outputFile == "" {
				outputFile = filepath.Base(filePath)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(&settings)

			reqOpts := api.NewRequestOptions(&settings).
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/files/%s", projectID, nodeID, filePath)).
				WithMethod(api.GET)

			fileData, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to get node file: %w", err)
			}
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					// Keep silent on body close errors
					_ = closeErr
				}
			}()

			if resp.StatusCode != 200 {
				return fmt.Errorf("failed to get node file with status %d: %s", resp.StatusCode, string(fileData))
			}

			err = os.WriteFile(outputFile, fileData, 0o600)
			if err != nil {
				return fmt.Errorf("failed to save node file: %w", err)
			}

			fmt.Printf("%s Node file downloaded successfully to %s", messageUtils.SuccessMsg("Node file downloaded successfully"), messageUtils.Bold(outputFile))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "", "", "Output filename (default: original filename)")

	return cmd
}

func NewStreamPcapCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:     "stream-pcap [project-name/id] [link-name/id]",
		Short:   "Stream PCAP capture file from compute",
		Long:    `Stream the PCAP capture file from compute by project ID/name and link ID/name`,
		Example: "gns3util -s https://controller:3080 project stream-pcap my-project my-link",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			linkID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					return err
				}
			}

			if !utils.IsValidUUIDv4(args[1]) {
				linkID, err = utils.ResolveID(cfg, "link", args[1], nil)
				if err != nil {
					return err
				}
			}

			if outputFile == "" {
				outputFile = fmt.Sprintf("capture_%s.pcap", linkID)
			}

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(&settings)

			reqOpts := api.NewRequestOptions(&settings).
				WithURL(fmt.Sprintf("/projects/%s/links/%s/capture/stream", projectID, linkID)).
				WithMethod(api.GET)

			fileData, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to stream PCAP: %w", err)
			}
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					// Keep silent on body close errors
					_ = closeErr
				}
			}()

			if resp.StatusCode != 200 {
				return fmt.Errorf("failed to stream PCAP with status %d: %s", resp.StatusCode, string(fileData))
			}

			err = os.WriteFile(outputFile, fileData, 0o600)
			if err != nil {
				return fmt.Errorf("failed to save PCAP file: %w", err)
			}

			fmt.Printf("%s PCAP file streamed successfully to %s", messageUtils.SuccessMsg("PCAP file streamed successfully"), messageUtils.Bold(outputFile))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "", "", "Output filename (default: capture_{link-id}.pcap)")

	return cmd
}
