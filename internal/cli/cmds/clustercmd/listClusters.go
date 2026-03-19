package clustercmd

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
)

type ClusterRow struct {
	sqlc.Cluster
}

func (c ClusterRow) GetHeaders() []string {
	return []string{"ID", "NAME", "DESCRIPTION"}
}

func (c ClusterRow) GetRow() []string {
	desc := "N/A"
	if c.Description.Valid {
		desc = c.Description.String
	}
	return []string{
		fmt.Sprintf("%d", c.ClusterID),
		c.Name,
		desc,
	}
}

func NewLsClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "list all clusters",
		Long:  `list all clusters`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}

			store, openErr := db.Init()
			if openErr != nil {
				return fmt.Errorf("failed to initialize database: %w", openErr)
			}
			clusters, fetchErr := store.GetClusters(cmd.Context())
			if fetchErr != nil {
				if errors.Is(fetchErr, sql.ErrNoRows) {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No clusters found")
					return nil
				}
				return fmt.Errorf("failed to get clusters: %w", fetchErr)
			}

			formatStr := cfg.OutputFormat.String()
			raw, _ := cmd.InheritedFlags().GetBool("raw")
			noColor, _ := cmd.InheritedFlags().GetBool("no-color")

			if raw {
				if noColor {
					formatStr = globals.OutputJSONColorless.String()
				} else {
					formatStr = globals.OutputJSON.String()
				}
			}

			printer, err := utils.GetPrinter(formatStr)
			if err != nil {
				return err
			}

			if formatStr == globals.OutputJSON.String() || formatStr == globals.OutputJSONColorless.String() {
				return printer.PrintObj(clusters, cmd.OutOrStdout())
			}

			var records []utils.TableRecord
			for _, c := range clusters {
				records = append(records, ClusterRow{Cluster: c})
			}

			return printer.PrintObj(records, cmd.OutOrStdout())
		},
	}

	return cmd
}
