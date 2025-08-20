package post

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewLinkCmdGroup() *cobra.Command {
	linkCmd := &cobra.Command{
		Use:   "link",
		Short: "Link operations",
		Long:  `Link operations for managing GNS3 links.`,
	}

	linkCmd.AddCommand(
		NewResetLinkCmd(),
		NewStartCaptureCmd(),
		NewStopCaptureCmd(),
	)

	return linkCmd
}

func NewResetLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset [project-name/id] [link-name/id]",
		Short: "Reset a link",
		Long:  `Reset a link in a project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post link reset my-project my-link
gns3util -s https://controller:3080 post link reset 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			linkID := args[1]
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

			if !utils.IsValidUUIDv4(linkID) {
				id, err := utils.ResolveID(cfg, "link", linkID, []string{projectID})
				if err != nil {
					return fmt.Errorf("failed to resolve link ID: %w", err)
				}
				linkID = id
			}

			utils.ExecuteAndPrint(cfg, "resetLink", []string{projectID, linkID})
			return nil
		},
	}

	return cmd
}

func NewStartCaptureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-capture [project-name/id] [link-name/id]",
		Short: "Start packet capture on a link",
		Long:  `Start packet capture on a link in a project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post link start-capture my-project my-link
gns3util -s https://controller:3080 post link start-capture 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			linkID := args[1]
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

			if !utils.IsValidUUIDv4(linkID) {
				id, err := utils.ResolveID(cfg, "link", linkID, []string{projectID})
				if err != nil {
					return fmt.Errorf("failed to resolve link ID: %w", err)
				}
				linkID = id
			}

			utils.ExecuteAndPrint(cfg, "startCapture", []string{projectID, linkID})
			return nil
		},
	}

	return cmd
}

func NewStopCaptureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-capture [project-name/id] [link-name/id]",
		Short: "Stop packet capture on a link",
		Long:  `Stop packet capture on a link in a project on the GNS3 server.`,
		Example: `gns3util -s https://controller:3080 post link stop-capture my-project my-link
gns3util -s https://controller:3080 post link stop-capture 123e4567-e89b-12d3-a456-426614174000 456e7890-e89b-12d3-a456-426614174000`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			linkID := args[1]
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

			if !utils.IsValidUUIDv4(linkID) {
				id, err := utils.ResolveID(cfg, "link", linkID, []string{projectID})
				if err != nil {
					return fmt.Errorf("failed to resolve link ID: %w", err)
				}
				linkID = id
			}

			utils.ExecuteAndPrint(cfg, "stopCapture", []string{projectID, linkID})
			return nil
		},
	}

	return cmd
}
