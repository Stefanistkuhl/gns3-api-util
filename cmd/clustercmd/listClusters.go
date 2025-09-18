package clustercmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func NewLsClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls [cluster-name]",
		Short: "list all clusters",
		Long:  `list all clusters`,
		Run: func(cmd *cobra.Command, args []string) {
			data, openErr := db.InitIfNeeded()
			if openErr != nil {
				fmt.Printf("%s %s", colorUtils.Error("Error:"), openErr)
				return
			}
			clusters, fetchErr := db.GetClusters(data)
			if fetchErr != nil {
				if fetchErr == sql.ErrNoRows {
					fmt.Printf("No clusters found")
					return
				}
				fmt.Printf("%s %s", colorUtils.Error("Error:"), fetchErr)
				return
			}
			raw, _ := cmd.InheritedFlags().GetBool("raw")
			noColor, _ := cmd.InheritedFlags().GetBool("no-color")
			if raw {
				mar, err := json.Marshal(clusters)
				if err != nil {
					fmt.Printf("%s failed to marshall the results %s", colorUtils.Error("Error:"), err)
				}
				if noColor {
					utils.PrintJsonUgly(mar)
					return
				} else {
					utils.PrintJson(mar)
					return
				}
			}
			utils.PrintTable(clusters, []utils.Column[db.ClusterName]{
				{
					Header: "ID",
					Value: func(c db.ClusterName) string {
						return fmt.Sprintf("%d", c.Id)
					},
				},
				{
					Header: "Name",
					Value: func(c db.ClusterName) string {
						return c.Name
					},
				},
				{
					Header: "Desc",
					Value: func(c db.ClusterName) string {
						if c.Desc.Valid {
							return c.Desc.String
						}
						return "N/A"
					},
				},
			})

		},
	}

	return cmd
}
