package clustercmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

var name string
var desc string

func NewCreateClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a cluster",
		Long:  `create a cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			err := db.CheckIfCluterExists(name)
			switch err {
			case db.ErrClusterExists:
				fmt.Printf("A cluster with the name %s already exists.\n", name)
				return

			case nil, db.ErrNoDb:
				data, openErr := db.InitIfNeeded()
				if openErr != nil {
					fmt.Printf("%s failed to open db: %v\n", messageUtils.ErrorMsg("Error"), openErr)
					return
				}
				defer func() {
					_ = data.Close()
				}()

				if _, insertErr := db.UpdateRows(
					data,
					"INSERT INTO clusters (name, description) VALUES (?, ?)",
					name, desc,
				); insertErr != nil {
					fmt.Printf("%s inserting data into the db: %v\n", messageUtils.ErrorMsg("Error"), insertErr)
					return
				}

				cfg, cfgErr := cluster.LoadClusterConfig()
				if cfgErr != nil {
					if cfgErr == cluster.ErrNoConfig {
						cfg = cluster.NewConfig()
					} else {
						fmt.Printf("%s failed to load config: %v\n", messageUtils.ErrorMsg("Error"), cfgErr)
						return
					}
				}

				cfg.Clusters = append(cfg.Clusters, cluster.Cluster{
					Name:        name,
					Description: desc,
				})
				if writeErr := cluster.WriteClusterConfig(cfg); writeErr != nil {
					fmt.Printf("%s failed to write to config file: %v\n", messageUtils.ErrorMsg("Error"), writeErr)
					return
				}

				fmt.Printf("%s created new empty cluster %s\n", messageUtils.SuccessMsg("Success"), name)
				return

			default:
				fmt.Printf("%s failed to check cluster existence: %v\n", messageUtils.ErrorMsg("Error"), err)
				return
			}
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "name for the cluster")
	cmd.Flags().StringVarP(&desc, "description", "d", "", "description for the cluster")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return cmd
}
