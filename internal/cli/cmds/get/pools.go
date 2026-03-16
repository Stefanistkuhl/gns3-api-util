package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get available resource-pools",
		Long:    `Get available resource-pools`,
		Example: "gns3util -s https://controller:3080 pool ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getPools", nil)
			return nil
		},
	}
	return cmd
}

func NewGetPoolCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [pool-name/id]",
		Aliases: []string{"get", "i"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParams(cfg, "getPools", "name", multi)
				fuzzyErr := fuzzy.FuzzyInfo(params)
				if fuzzyErr != nil {
					return fuzzyErr
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "pool", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getPool", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a pool")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple pools")
	return cmd
}

func NewGetPoolResourcesCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	cmd := &cobra.Command{
		Use:     "resources [pool-name/id]",
		Aliases: []string{"res", "assets"},
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				params := fuzzy.NewFuzzyInfoParamsWithContext(cfg, "getPools", "name", multi, "pool", "Pool:")
				ids, fuzzyErr := fuzzy.FuzzyInfoIDs(params)
				if fuzzyErr != nil {
					return fuzzyErr
				}

				poolResources, poolErr := utils.GetResourceWithContext(cfg, "getPoolResources", ids, "pool", "Pool:")
				if poolErr != nil {
					return fmt.Errorf("error getting pool resources: %w", poolErr)
				}

				utils.PrintResourceWithContext(poolResources, "Pool:")
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "pool", args[0], nil)
					if err != nil {
						return err
					}
				}
				utils.ExecuteAndPrint(cfg, "getPoolResources", []string{id})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a pool")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple pools")
	return cmd
}
