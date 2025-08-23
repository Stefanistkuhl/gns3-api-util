package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetNodesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Short:   "Get the nodes within a project by name or id",
		Long:    `Get the nodes within a project by name or id`,
		Example: "gns3util -s https://controller:3080 node ls my-project",
		Args:    cobra.ExactArgs(1),
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
			utils.ExecuteAndPrint(cfg, "getNodes", []string{id})
		},
	}
	return cmd
}

func NewGetNodeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [node-name/id]",
		Short:   "Get a node in a project by name or id",
		Long:    `Get a node in a project by name or id`,
		Example: "gns3util -s https://controller:3080 node info my-project my-node",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if !utils.IsValidUUIDv4(args[1]) {
				nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getNode", []string{projectID, nodeID})
		},
	}
	return cmd
}

func NewGetNodeLinksCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "links [project-name/id] [node-name/id]",
		Short:   "Get links of a given node in a project by id or name",
		Long:    `Get links of a given node in a project by id or name`,
		Example: "gns3util -s https://controller:3080 node links my-project my-node",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if !utils.IsValidUUIDv4(args[1]) {
				nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getNodeLinks", []string{projectID, nodeID})
		},
	}
	return cmd
}

func NewGetNodesAutoIdlePCCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "auto-idle-pc [project-name/id] [node-name/id]",
		Short:   "Get the auto-idle-pc of a node in a project by id or name",
		Long:    `Get the auto-idle-pc of a node in a project by id or name`,
		Example: "gns3util -s https://controller:3080 node auto-idle-pc my-project my-node",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if !utils.IsValidUUIDv4(args[1]) {
				nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePc", []string{projectID, nodeID})
		},
	}
	return cmd
}

func NewGetNodesAutoIdlePCProposalsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "auto-idle-pc-proposals [project-name/id] [node-name/id]",
		Short:   "Get the auto-idle-pc-proposals of a node in a project by id or name",
		Long:    `Get the auto-idle-pc-proposals of a node in a project by id or name`,
		Example: "gns3util -s https://controller:3080 node auto-idle-pc-proposals my-project my-node",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			nodeID := args[1]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				projectID, err = utils.ResolveID(cfg, "project", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if !utils.IsValidUUIDv4(args[1]) {
				nodeID, err = utils.ResolveID(cfg, "node", args[1], []string{projectID})
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getNodeAutoIdlePcProposals", []string{projectID, nodeID})
		},
	}
	return cmd
}
