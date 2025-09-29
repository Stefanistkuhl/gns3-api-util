package post

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
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
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
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

			if !utils.IsValidUUIDv4(nodeID) {
				fmt.Println("Node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/duplicate", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to duplicate node: %v", err)
				return
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 201 {
				fmt.Printf("%s Node duplicated successfully\n", messageUtils.SuccessMsg("Node duplicated successfully"))
			} else {
				fmt.Printf("Failed to duplicate node with status %d", resp.StatusCode)
			}
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
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
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

			if !utils.IsValidUUIDv4(nodeID) {
				fmt.Println("Node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/console/reset", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to reset node console: %v", err)
				return
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Node console reset successfully\n", messageUtils.SuccessMsg("Node console reset successfully"))
			} else {
				fmt.Printf("Failed to reset node console with status %d", resp.StatusCode)
			}
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
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
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

			if !utils.IsValidUUIDv4(nodeID) {
				fmt.Println("Node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/isolate", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to isolate node: %v", err)
				return
			}
			defer func() {
				if resp != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Node isolated successfully\n", messageUtils.SuccessMsg("Node isolated successfully"))
			} else {
				fmt.Printf("Failed to isolate node with status %d", resp.StatusCode)
			}
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
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
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

			if !utils.IsValidUUIDv4(nodeID) {
				fmt.Println("Node ID must be a valid UUID")
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/%s/unisolate", projectID, nodeID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to unisolate node: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Node unisolated successfully\n", messageUtils.SuccessMsg("Node unisolated successfully"))
			} else {
				fmt.Printf("Failed to unisolate node with status %d", resp.StatusCode)
			}
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/reload", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to reload nodes: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes reloaded successfully\n", messageUtils.SuccessMsg("Nodes reloaded successfully"))
			} else {
				fmt.Printf("Failed to reload nodes with status %d", resp.StatusCode)
			}
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/start", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to start nodes: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes started successfully\n", messageUtils.SuccessMsg("Nodes started successfully"))
			} else {
				fmt.Printf("Failed to start nodes with status %d", resp.StatusCode)
			}
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/stop", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to stop nodes: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes stopped successfully\n", messageUtils.SuccessMsg("Nodes stopped successfully"))
			} else {
				fmt.Printf("Failed to stop nodes with status %d", resp.StatusCode)
			}
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
				WithURL(fmt.Sprintf("/projects/%s/nodes/suspend", projectID)).
				WithMethod(api.POST)

			_, resp, err := client.Do(reqOpts)
			if err != nil {
				fmt.Printf("failed to suspend nodes: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode == 204 {
				fmt.Printf("%s Nodes suspended successfully\n", messageUtils.SuccessMsg("Nodes suspended successfully"))
			} else {
				fmt.Printf("Failed to suspend nodes with status %d", resp.StatusCode)
			}
		},
	}

	return cmd
}
