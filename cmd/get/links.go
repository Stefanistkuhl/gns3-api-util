package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetLinksCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName + " [project-name/id]",
		Short:   "Get the links within a project by name or id",
		Long:    `Get the links within a project by name or id`,
		Example: "gns3util -s https://controller:3080 link ls my-project",
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
			utils.ExecuteAndPrint(cfg, "getLinks", []string{id})
		},
	}
	return cmd
}

func NewGetLinkCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [project-name/id] [link-name/id]",
		Short:   "Get a link within a project by name or id",
		Long:    `Get a link within a project by name or id`,
		Example: "gns3util -s https://controller:3080 link info my-project my-link",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			linkID := args[1]
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
			utils.ExecuteAndPrint(cfg, "getLink", []string{projectID, linkID})
		},
	}
	return cmd
}

func NewGetLinkIfaceCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "iface [project-name/id] [link-name/id]",
		Short:   "Get the interface of a link within a project by name or id for Cloud or NAT devices.",
		Long:    `Return iface info for links to Cloud or NAT devices for a link in a project by id or name.`,
		Example: "gns3util -s https://controller:3080 link iface my-project my-link",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			linkID := args[1]
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
			utils.ExecuteAndPrint(cfg, "getLinkIface", []string{projectID, linkID})
		},
	}
	return cmd
}

func NewGetLinkFiltersCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "filters [project-name/id] [link-name/id]",
		Short:   "Get the filters for a link within a project by name or id",
		Long:    `Get the filters for a link within a project by name or id`,
		Example: "gns3util -s https://controller:3080 link filters my-project my-link",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			projectID := args[0]
			linkID := args[1]
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
			utils.ExecuteAndPrint(cfg, "getLinkFilters", []string{projectID, linkID})
		},
	}
	return cmd
}
