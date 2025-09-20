package clustercmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/cluster"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewAddNodeCmd() *cobra.Command {
	opts := &cluster.AddNodeOptions{}
	cmd := &cobra.Command{
		Use:   "add-node [cluster-name]",
		Short: "Add a single node to a cluster",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			cluster.ValidateClusterAndCreds(args[0], opts, cmd)
			if len(opts.Servers) == 0 {
				fmt.Println(messageUtils.InfoMsg("No servers provided, will enter interactive mode."))
			}
			if len(opts.Servers) > 1 {
				fmt.Printf("%s add-node only supports a single --server. Use add-nodes for multiple.\n",
					messageUtils.ErrorMsg("Error"))
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			nodes, err := cluster.RunAddNodes(opts, cmd)
			if err != nil {
				fmt.Printf("%s %v\n", messageUtils.ErrorMsg("Error"), err)
				os.Exit(1)
			}
			if nodes == nil {
				return
			}

			insertedNodes, err := db.InsertNodes(opts.ClusterID, nodes)
			if err != nil {
				fmt.Printf("%s failed to insert node: %v\n", messageUtils.ErrorMsg("Error"), err)
				os.Exit(1)
			}
			for _, node := range insertedNodes {
				fmt.Printf("Inserted node %s:%d with ID: %d\n", node.Host, node.Port, node.ID)
			}
		},
	}
	addCommonFlags(cmd, opts)
	return cmd
}

func NewAddNodesCmd() *cobra.Command {
	opts := &cluster.AddNodeOptions{}
	cmd := &cobra.Command{
		Use:   "add-nodes [cluster-name]",
		Short: "Add multiple nodes to a cluster",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			cluster.ValidateClusterAndCreds(args[0], opts, cmd)
			if len(opts.Servers) == 0 {
				fmt.Println(messageUtils.InfoMsg("No servers provided, will enter interactive mode."))
			}
			if len(opts.Servers) == 1 {
				fmt.Printf("%s add-nodes requires at least 2 --server entries. Use add-node for one.\n",
					messageUtils.ErrorMsg("Error"))
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			nodes, err := cluster.RunAddNodes(opts, cmd)
			if err != nil {
				fmt.Printf("%s %v\n", messageUtils.ErrorMsg("Error"), err)
				os.Exit(1)
			}
			if nodes == nil {
				return
			}

			insertedNodes, err := db.InsertNodes(opts.ClusterID, nodes)
			if err != nil {
				fmt.Printf("%s DB insert error: %v\n", messageUtils.ErrorMsg("Error"), err)
				os.Exit(1)
			}
			for _, node := range insertedNodes {
				fmt.Printf("Inserted node %s:%d with ID: %d\n", node.Host, node.Port, node.ID)
			}
		},
	}
	addCommonFlags(cmd, opts)
	return cmd
}

func addCommonFlags(cmd *cobra.Command, opts *cluster.AddNodeOptions) {
	cmd.Flags().StringSliceVarP(&opts.Servers, "server", "s", nil, "Server(s) to add")
	cmd.Flags().IntVarP(&opts.Weight, "weight", "w", 5, "Weight to assign to node(s) (0–10, default 5)")
	cmd.Flags().IntVarP(&opts.MaxGroups, "max-groups", "g", 3, "Maximum groups per node (default 3)")
	cmd.Flags().StringVarP(&opts.Username, "user", "u", "", "User to log in as (env: GNS3_USER)")
	cmd.Flags().StringVarP(&opts.Password, "password", "p", "", "Password to use (env: GNS3_PASSWORD)")
}
