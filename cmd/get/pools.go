package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetPoolsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get available resource-pools",
		Long:    `Get available resource-pools`,
		Example: "gns3util -s https://controller:3080 pool ls",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getPools", nil)
		},
	}
	return cmd
}

func NewGetPoolCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [pool-name/id]",
		Short:   "Get a resource-pool by name or id",
		Long:    `Get a resource-pool by name or id`,
		Example: "gns3util -s https://controller:3080 pool info my-pool",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "pool", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getPool", []string{id})
		},
	}
	return cmd
}

func NewGetPoolResourcesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "resources [pool-name/id]",
		Short:   "Get resources of a pool by name or id",
		Long:    `Get resources of a pool by name or id`,
		Example: "gns3util -s https://controller:3080 pool resources my-pool",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "pool", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getPoolResources", []string{id})
		},
	}
	return cmd
}
