package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetPoolsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "pools",
		Short: "Get avaliable resource-pools",
		Long:  `Get avaliable resource-pools`,
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
		Use:   "info",
		Short: "Get a resource-pool by name or id",
		Long:  `Get a resource-pool by name or id`,
		Args:  cobra.ExactArgs(1),
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
		Use:   "pool-resources",
		Short: "Get get resources of a pool by name or id",
		Long:  `Get get resources of a pool by name or id`,
		Args:  cobra.ExactArgs(1),
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
