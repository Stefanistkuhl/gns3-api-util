package clustercmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewLsClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "list all clusters",
		Long:  `list all clusters`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, openErr := db.Init()
			if openErr != nil {
				return fmt.Errorf("failed to initialize database: %w", openErr)
			}
			clusters, fetchErr := store.GetClusters(cmd.Context())
			if fetchErr != nil {
				if errors.Is(fetchErr, sql.ErrNoRows) {
					fmt.Printf("No clusters found")
					return nil
				}
				return fmt.Errorf("failed to get clusters: %w", fetchErr)
			}
			raw, _ := cmd.InheritedFlags().GetBool("raw")
			noColor, _ := cmd.InheritedFlags().GetBool("no-color")
			if raw {
				mar, err := json.Marshal(clusters)
				if err != nil {
					return fmt.Errorf("failed to marshal results: %w", err)
				}
				if noColor {
					utils.PrintJSONUgly(mar)
					return nil
				} else {
					utils.PrintJSON(mar)
					return nil
				}
			}
			utils.PrintTable(clusters, []utils.Column[sqlc.Cluster]{
				{
					Header: "ID",
					Value: func(c sqlc.Cluster) string {
						return fmt.Sprintf("%d", c.ClusterID)
					},
				},
				{
					Header: "Name",
					Value: func(c sqlc.Cluster) string {
						return c.Name
					},
				},
				{
					Header: "Desc",
					Value: func(c sqlc.Cluster) string {
						if c.Description.Valid {
							return c.Description.String
						}
						return "N/A"
					},
				},
			})
			return nil
		},
	}

	return cmd
}
