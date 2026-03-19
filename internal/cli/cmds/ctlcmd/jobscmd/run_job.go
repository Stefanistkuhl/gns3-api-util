package jobscmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/pkg/models"
)

type RunJobRow struct {
	RunID   string `json:"run_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (r RunJobRow) GetHeaders() []string {
	return []string{"RUN ID", "STATUS", "MESSAGE"}
}

func (r RunJobRow) GetRow() []string {
	return []string{
		r.RunID,
		r.Status,
		r.Message,
	}
}

func NewRunJobCmd() *cobra.Command {
	var nodeID string

	cmd := &cobra.Command{
		Use:   "run [job_name]",
		Short: "Trigger a background job on a specific node",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"auth-mode": "cluster-only",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			jobName := args[0]

			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			if nodeID == "" {
				return fmt.Errorf("--node-id is required")
			}

			client, err := newMasterClient(cfg)
			if err != nil {
				return err
			}

			resp, err := client.RunJob(cmd.Context(), jobName, &models.RunJobRequest{
				NodeID: nodeID,
			})
			if err != nil {
				return fmt.Errorf("failed to run job: %w", err)
			}

			result := RunJobRow{
				RunID:   resp.RunID,
				Status:  resp.Status,
				Message: resp.Message,
			}

			printer, err := utils.GetPrinter(cfg.OutputFormat.String())
			if err != nil {
				return err
			}

			return printer.PrintObj(result, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&nodeID, "node-id", "", "Node ID to run the job on (required)")
	_ = cmd.MarkFlagRequired("node-id")

	return cmd
}
