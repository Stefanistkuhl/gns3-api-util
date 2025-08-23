package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [project-name/id] [node-name/id]",
		Short:   "Delete a node from a project",
		Long:    `Delete a node from a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 node delete my-project my-node",
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

			utils.ExecuteAndPrint(cfg, "deleteNode", []string{projectID, nodeID})
		},
	}

	return cmd
}
