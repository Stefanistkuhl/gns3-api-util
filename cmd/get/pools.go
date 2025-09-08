package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
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
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [pool-name/id]",
		Short:   "Get a resource-pool by name or id",
		Long:    `Get a resource-pool by name or id`,
		Example: "gns3util -s https://controller:3080 pool info my-pool",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [pool-name/id] when --fuzzy is not set")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if multi && !useFuzzy {
				return fmt.Errorf("the --multi (-m) flag can only be used together with --fuzzy (-f)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getPools", "name", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "pool", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getPool", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a pool")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple pools")
	return cmd
}

func NewGetPoolResourcesCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     "resources [pool-name/id]",
		Short:   "Get resources of a pool by name or id",
		Long:    `Get resources of a pool by name or id`,
		Example: "gns3util -s https://controller:3080 pool resources my-pool",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [pool-name/id] when --fuzzy is not set")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if multi && !useFuzzy {
				return fmt.Errorf("the --multi (-m) flag can only be used together with --fuzzy (-f)")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getPools", "name", multi, "pool", "Pool:")
				ids, err := fuzzy.FuzzyInfoIDs(params)
				if err != nil {
					fmt.Println(err)
					return
				}

				poolResources, err := utils.GetResourceWithContext(cfg, "getPoolResources", ids, "pool", "Pool:")
				if err != nil {
					fmt.Printf("Error getting pool resources: %v\n", err)
					return
				}

				utils.PrintResourceWithContext(poolResources, "Pool:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "pool", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getPoolResources", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a pool")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple pools")
	return cmd
}
