package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewProjectCmdGroup() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Project operations",
		Long:  `Project operations for managing GNS3 projects.`,
	}

	projectCmd.AddCommand(
		NewProjectDuplicateCmd(),
		NewProjectLoadCmd(),
		NewProjectCloseCmd(),
		NewProjectImportCmd(),
		NewProjectLockCmd(),
		NewProjectOpenCmd(),
		NewProjectUnlockCmd(),
		NewProjectWriteFileCmd(),
		NewProjectStartCaptureCmd(),
	)

	return projectCmd
}

func NewProjectDuplicateCmd() *cobra.Command {
	var (
		name        string
		projectID   string
		path        string
		autoClose   bool
		autoOpen    bool
		autoStart   bool
		sceneHeight int
		sceneWidth  int
		zoom        int
		showLayers  bool
		snapToGrid  bool
		showGrid    bool
		useJSON     string
	)

	cmd := &cobra.Command{
		Use:     "duplicate [project-name/id]",
		Short:   "Duplicate a Project",
		Long:    `Duplicate a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project duplicate my-project --name \"duplicated-project\"",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					return fmt.Errorf("failed to resolve project ID: %w", err)
				}
				projectID = id
			}

			var payload map[string]any
			if useJSON == "" {
				if name == "" {
					return fmt.Errorf("for this command --name is required or provide --use-json")
				}
				data := schemas.ProjectDuplicate{
					Name: name,
				}
				if projectID != "" {
					data.ProjectID = &projectID
				}
				if path != "" {
					data.Path = &path
				}
				if autoClose {
					v := true
					data.AutoClose = &v
				}
				if autoOpen {
					v := true
					data.AutoOpen = &v
				}
				if autoStart {
					v := true
					data.AutoStart = &v
				}
				if sceneHeight != 0 {
					data.SceneHeight = &sceneHeight
				}
				if sceneWidth != 0 {
					data.SceneWidth = &sceneWidth
				}
				if zoom != 0 {
					data.Zoom = &zoom
				}
				if showLayers {
					v := true
					data.ShowLayers = &v
				}
				if snapToGrid {
					v := true
					data.SnapToGrid = &v
				}
				if showGrid {
					v := true
					data.ShowGrid = &v
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
					return fmt.Errorf("failed to parse JSON: %w", err)
				}
			}

			utils.ExecuteAndPrintWithBody(cfg, "duplicateProject", []string{projectID}, payload)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the duplicated project (required)")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID for the duplicated project")
	cmd.Flags().StringVar(&path, "path", "", "Path for the duplicated project")
	cmd.Flags().BoolVar(&autoClose, "auto-close", false, "Auto close the project")
	cmd.Flags().BoolVar(&autoOpen, "auto-open", false, "Auto open the project")
	cmd.Flags().BoolVar(&autoStart, "auto-start", false, "Auto start the project")
	cmd.Flags().IntVar(&sceneHeight, "scene-height", 0, "Scene height")
	cmd.Flags().IntVar(&sceneWidth, "scene-width", 0, "Scene width")
	cmd.Flags().IntVar(&zoom, "zoom", 0, "Zoom level")
	cmd.Flags().BoolVar(&showLayers, "show-layers", false, "Show layers")
	cmd.Flags().BoolVar(&snapToGrid, "snap-to-grid", false, "Snap to grid")
	cmd.Flags().BoolVar(&showGrid, "show-grid", false, "Show grid")
	cmd.Flags().StringVar(&useJSON, "use-json", "", "Provide a raw JSON string to send instead of flags")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func NewProjectLoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load [project-path]",
		Short:   "Load a project from a given path",
		Long:    `Load a project from a given path on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post project load /path/to/project.gns3project`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectPath := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
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
				WithURL(fmt.Sprintf("/projects/load?path=%s", projectPath)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to load project: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 201 {
				fmt.Printf("%s Project loaded successfully\n", messageUtils.SuccessMsg("Project loaded successfully"))
			} else {
				fmt.Printf("Failed to load project with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}

func NewProjectCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "close [project-name/id]",
		Short:   "Close a project",
		Long:    `Close a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project close my-project",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/close", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to close project: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Project closed successfully\n", messageUtils.SuccessMsg("Project closed successfully"))
			} else {
				fmt.Printf("Failed to close project with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}

func NewProjectImportCmd() *cobra.Command {
	var projectName string

	cmd := &cobra.Command{
		Use:   "import [archive-file]",
		Short: "Import a project from a portable archive",
		Long:  `Import a project from a portable archive (.gns3project file) on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post project import /path/to/project.gns3project
gns3util -s https://controller:3080 post project import --name "my-project" /path/to/project.gns3project`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			archiveFile := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if _, err := os.Stat(archiveFile); os.IsNotExist(err) {
				return fmt.Errorf("archive file does not exist: %s", archiveFile)
			}

			file, err := os.Open(archiveFile)
			if err != nil {
				return fmt.Errorf("failed to open archive file: %w", err)
			}
			defer func() {
				_ = file.Close()
			}()

			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, err := writer.CreateFormFile("file", filepath.Base(archiveFile))
			if err != nil {
				return fmt.Errorf("failed to create form file: %w", err)
			}

			if _, err := io.Copy(part, file); err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}

			_ = writer.Close()

			token, err := authentication.GetKeyForServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to get token: %w", err)
			}

			settings := api.NewSettings(
				api.WithBaseURL(cfg.Server),
				api.WithVerify(cfg.Insecure),
				api.WithToken(token),
			)
			client := api.NewGNS3Client(settings)

			ep := endpoints.Endpoints{}
			projectID := uuid.New().String()
			urlStr := ep.Post.ProjectImport(projectID)
			if projectName != "" {
				urlStr += fmt.Sprintf("?name=%s", url.QueryEscape(projectName))
			}

			reqOpts := api.NewRequestOptions(settings).
				WithURL(urlStr).
				WithMethod(api.POST).
				WithData(buf.String())

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to import project: %w", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode == 201 {
				fmt.Printf("%s: Project imported successfully\n", messageUtils.SuccessMsg("Success"))
			} else {
				return fmt.Errorf("failed to import project with status %d", resp.StatusCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&projectName, "name", "", "Name for the imported project (optional)")

	return cmd
}

func NewProjectLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lock [project-name/id]",
		Short:   "Lock all drawings and nodes in a project",
		Long:    `Lock all drawings and nodes in a given project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project lock my-project",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/lock", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to lock project: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Project locked successfully\n", messageUtils.SuccessMsg("Project locked successfully"))
			} else {
				fmt.Printf("Failed to lock project with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}

func NewProjectOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open [project-name/id]",
		Short:   "Open a project",
		Long:    `Open a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project open my-project",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/open", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to open project: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Project opened successfully\n", messageUtils.SuccessMsg("Project opened successfully"))
			} else {
				fmt.Printf("Failed to open project with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}

func NewProjectUnlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unlock [project-name/id]",
		Short:   "Unlock all drawings and nodes in a project",
		Long:    `Unlock all drawings and nodes in a given project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project unlock my-project",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/unlock", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to unlock project: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Project unlocked successfully\n", messageUtils.SuccessMsg("Project unlocked successfully"))
			} else {
				fmt.Printf("Failed to unlock project with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}

func NewProjectWriteFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "write-file [project-name/id] [file-path]",
		Short:   "Write a file to a project",
		Long:    `Write a file to a project with the given file path on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project write-file my-project /path/to/file",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			filePath := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/files%s", projectID, filePath)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to write file to project: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 201 {
				fmt.Printf("%s File written to project successfully\n", messageUtils.SuccessMsg("File written to project successfully"))
			} else {
				fmt.Printf("Failed to write file to project with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}

func NewProjectStartCaptureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start-capture [project-name/id] [link-name/id]",
		Short:   "Start a packet capture in a project on a given link",
		Long:    `Start a packet capture in a project on a given link on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 project start-capture my-project my-link",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			linkID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, err := utils.ResolveID(cfg, "project", projectID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(linkID) {
				fmt.Println("Link ID must be a valid UUID")
				return
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
				WithURL(fmt.Sprintf("/projects/%s/links/%s/start_capture", projectID, linkID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to start capture: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 201 {
				fmt.Printf("%s Packet capture started successfully\n", messageUtils.SuccessMsg("Packet capture started successfully"))
			} else {
				fmt.Printf("Failed to start packet capture with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}
