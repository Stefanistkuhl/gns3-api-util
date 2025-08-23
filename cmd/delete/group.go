package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewDeleteGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.DeleteSingleElementCmdName + " [group-name/id]",
		Short:   "Delete a group",
		Long:    `Delete a group from the GNS3 server.`,
		Example: "gns3util -s https://controller:3080 group delete my-group",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			groupID := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
				return
			}

			if !utils.IsValidUUIDv4(groupID) {
				id, err := utils.ResolveID(cfg, "group", groupID, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				groupID = id
			}

			utils.ExecuteAndPrint(cfg, "deleteGroup", []string{groupID})
		},
	}

	return cmd
}
