package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [project-name/id]",
		Short: "Delete a project",
		Long:  `Delete a project from the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 project remove my-project
gns3util -s https://controller:3080 project remove 123e4567-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(1),
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

			utils.ExecuteAndPrint(cfg, "deleteProject", []string{projectID})
		},
	}

	return cmd
}
