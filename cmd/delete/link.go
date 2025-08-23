package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [project-name/id] [link-name/id]",
		Short:   "Delete a link from a project",
		Long:    `Delete a link from a project on the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 link delete my-project my-link",
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

			utils.ExecuteAndPrint(cfg, "deleteLink", []string{projectID, linkID})
		},
	}

	return cmd
}
