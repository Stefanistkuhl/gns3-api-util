package jobscmd

import "github.com/spf13/cobra"

func NewJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job"},
		Short:   "Commands to manage and run background jobs on cluster nodes",
		Long:    "Commands to manage and run background jobs on cluster nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	listGroup := &cobra.Group{ID: "query", Title: "Query Commands:"}
	actionGroup := &cobra.Group{ID: "action", Title: "Action Commands:"}

	cmd.AddGroup(listGroup, actionGroup)

	listJobsCmd := NewListJobsCmd()
	listJobsCmd.GroupID = "query"

	listRunsCmd := NewListJobRunsCmd()
	listRunsCmd.GroupID = "query"

	getRunCmd := NewGetJobRunCmd()
	getRunCmd.GroupID = "query"

	runJobCmd := NewRunJobCmd()
	runJobCmd.GroupID = "action"

	cmd.AddCommand(listJobsCmd, listRunsCmd, getRunCmd, runJobCmd)

	return cmd
}
