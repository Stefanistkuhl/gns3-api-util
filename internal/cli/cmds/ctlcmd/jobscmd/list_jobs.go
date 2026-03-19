package jobscmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
)

type JobRow struct {
	Name         string `json:"name"`
	NodeID       string `json:"node_id"`
	Interval     string `json:"interval"`
	Description  string `json:"description"`
	RegisteredAt string `json:"registered_at"`
}

func (j *JobRow) GetHeaders() []string {
	return []string{"NAME", "NODE ID", "INTERVAL", "DESCRIPTION", "REGISTERED AT"}
}

func (j *JobRow) GetRow() []string {
	return []string{
		j.Name,
		j.NodeID,
		j.Interval,
		j.Description,
		j.RegisteredAt,
	}
}

func NewListJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered background jobs",
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := newMasterClient(cfg)
			if err != nil {
				return err
			}

			resp, err := client.ListJobs(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list jobs: %w", err)
			}

			var records []utils.TableRecord
			for _, j := range resp.Jobs {
				records = append(records, &JobRow{
					Name:         j.Name,
					NodeID:       j.NodeID,
					Interval:     j.Interval,
					Description:  j.Description,
					RegisteredAt: j.RegisteredAt,
				})
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(records, cmd.OutOrStdout())
		},
	}

	return cmd
}
