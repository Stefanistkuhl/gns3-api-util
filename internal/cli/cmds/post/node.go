package post

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/authentication"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/spf13/cobra"
)

func NewNodeCmdGroup() *cobra.Command {
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Node operations",
		Long:  `Node operations for managing GNS3 nodes.`,
	}

	nodeCmd.AddCommand(
		NewNodeDuplicateCmd(),
		NewNodeConsoleResetCmd(),
		NewNodeIsolateCmd(),
		NewNodeUnisolateCmd(),
		NewReloadNodesCmd(),
		NewStartNodesCmd(),
		NewStopNodesCmd(),
		NewSuspendNodesCmd(),
	)

	return nodeCmd
}

func NewNodeDuplicateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "duplicate [project-name/id] [node-name/id]",
		Short:   "Duplicate a Node in a Project",
		Long:    `Duplicate a node in a project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node duplicate [project-name/id] [node-name/id]`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(nodeID) {
				return fmt.Errorf("node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/duplicate", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to duplicate node: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 201 {
				fmt.Printf("%s Node duplicated successfully\n", messageUtils.SuccessMsg("Node duplicated successfully"))
			} else {
				return fmt.Errorf("failed to duplicate node with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewNodeConsoleResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "console-reset [project-name/id] [node-name/id]",
		Short:   "Reset a console for a given node",
		Long:    `Reset a console for a given node on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node node-console-reset [project-name/id] [node-name/id]`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(nodeID) {
				return fmt.Errorf("node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/console/reset", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to reset console: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Console reset successfully\n", messageUtils.SuccessMsg("Console reset successfully"))
			} else {
				return fmt.Errorf("failed to reset console with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewNodeIsolateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node-isolate [project-name/id] [node-name/id]",
		Short:   "Isolate a node (suspend all attached links)",
		Long:    `Isolate a node by suspending all attached links on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node node-isolate [project-name/id] [node-name/id]`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(nodeID) {
				return fmt.Errorf("node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/isolate", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to isolate node: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Node isolated successfully\n", messageUtils.SuccessMsg("Node isolated successfully"))
			} else {
				return fmt.Errorf("failed to isolate node with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewNodeUnisolateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node-unisolate [project-name/id] [node-name/id]",
		Short:   "Un-isolate a node (resume all attached suspended links)",
		Long:    `Un-isolate a node by resuming all attached suspended links on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node node-unisolate [project-name/id] [node-name/id]`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
			}

			if !utils.IsValidUUIDv4(nodeID) {
				return fmt.Errorf("node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/unisolate", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to un-isolate node: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Node un-isolated successfully\n", messageUtils.SuccessMsg("Node un-isolated successfully"))
			} else {
				return fmt.Errorf("failed to un-isolate node with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewReloadNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reload-all [project-name/id]",
		Short:   "Reload all nodes belonging to a project",
		Long:    `Reload all nodes belonging to a given project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node reload-all [project-name/id]`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/reload", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to reload nodes: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes reloaded successfully\n", messageUtils.SuccessMsg("Nodes reloaded successfully"))
			} else {
				return fmt.Errorf("failed to reload nodes with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewStartNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start-all [project-name/id]",
		Short:   "Start all nodes belonging to a project",
		Long:    `Start all nodes belonging to a given project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node start-all [project-name/id]`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/start", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to start nodes: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes started successfully\n", messageUtils.SuccessMsg("Nodes started successfully"))
			} else {
				return fmt.Errorf("failed to start nodes with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewStopNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stop-all [project-name/id]",
		Short:   "Stop all nodes belonging to a project",
		Long:    `Stop all nodes belonging to a given project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node stop-all [project-name/id]`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/stop", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to stop nodes: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes stopped successfully\n", messageUtils.SuccessMsg("Nodes stopped successfully"))
			} else {
				return fmt.Errorf("failed to stop nodes with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}

func NewSuspendNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "suspend-all [project-name/id]",
		Short:   "Suspend all nodes belonging to a project",
		Long:    `Suspend all nodes belonging to a given project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post node suspend-all [project-name/id]`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if !utils.IsValidUUIDv4(projectID) {
				id, resolveErr := utils.ResolveID(cfg, "project", projectID, nil)
				if resolveErr != nil {
					return resolveErr
				}
				projectID = id
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/suspend", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				return fmt.Errorf("failed to suspend nodes: %w", err)
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes suspended successfully\n", messageUtils.SuccessMsg("Nodes suspended successfully"))
			} else {
				return fmt.Errorf("failed to suspend nodes with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	return cmd
}
