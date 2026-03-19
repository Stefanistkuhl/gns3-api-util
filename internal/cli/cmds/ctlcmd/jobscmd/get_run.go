package jobscmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
)

type JobRunRow struct {
	RunID       string `json:"run_id"`
	JobName     string `json:"job_name"`
	NodeID      string `json:"node_id"`
	Status      string `json:"status"`
	InvokedBy   string `json:"invoked_by"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

func (j *JobRunRow) GetHeaders() []string {
	return []string{"RUN ID", "JOB NAME", "NODE ID", "STATUS", "INVOKED BY", "STARTED AT", "COMPLETED AT"}
}

func (j *JobRunRow) GetRow() []string {
	return []string{
		j.RunID,
		j.JobName,
		j.NodeID,
		j.Status,
		j.InvokedBy,
		j.StartedAt,
		j.CompletedAt,
	}
}

func NewGetJobRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-run [run_id]",
		Short: "Get the status and result of a job run",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			client, err := newMasterClient(cfg)
			if err != nil {
				return err
			}

			resp, err := client.GetJobRun(cmd.Context(), runID)
			if err != nil {
				return fmt.Errorf("failed to get job run: %w", err)
			}

			// For JSON output, return the full response including results
			if cfg.OutputFormat == globals.OutputJSON || cfg.OutputFormat == globals.OutputJSONColorless {
				printer, pErr := utils.GetPrinter(cfg.OutputFormat.String())
				if pErr != nil {
					return pErr
				}
				return printer.PrintObj(resp, cmd.OutOrStdout())
			}

			completedAt := ""
			if resp.CompletedAt != nil {
				completedAt = resp.CompletedAt.Format(time.RFC3339)
			}

			result := JobRunRow{
				RunID:       resp.RunID,
				JobName:     resp.JobName,
				NodeID:      resp.NodeID,
				Status:      resp.Status,
				InvokedBy:   resp.InvokedBy,
				StartedAt:   resp.StartedAt.Format(time.RFC3339),
				CompletedAt: completedAt,
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(&result, cmd.OutOrStdout())
		},
	}

	return cmd
}

func NewListJobRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-runs",
		Short: "List all job runs",
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

			resp, err := client.ListJobRuns(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list job runs: %w", err)
			}

			var records []utils.TableRecord
			for i := range resp.Runs {
				completedAt := ""
				if resp.Runs[i].CompletedAt != nil {
					completedAt = resp.Runs[i].CompletedAt.Format(time.RFC3339)
				}
				records = append(records, &JobRunRow{
					RunID:       resp.Runs[i].RunID,
					JobName:     resp.Runs[i].JobName,
					NodeID:      resp.Runs[i].NodeID,
					Status:      resp.Runs[i].Status,
					InvokedBy:   resp.Runs[i].InvokedBy,
					StartedAt:   resp.Runs[i].StartedAt.Format(time.RFC3339),
					CompletedAt: completedAt,
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
