package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetSymbolsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get the available symbols",
		Long:    `Get the available symbols`,
		Example: "gns3util -s https://controller:3080 symbol ls",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getSymbols", nil)
		},
	}
	return cmd
}

func NewGetSymbolCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [symbol-id]",
		Short:   "Get a symbol by id",
		Long:    `Get a symbol by id`,
		Example: "gns3util -s https://controller:3080 symbol info symbol-id",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [symbol-id] when --fuzzy is not set")
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
				params := fuzzy.NewFuzzyInfoParams(cfg, "getSymbols", "symbol_id", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "symbol", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getSymbol", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a symbol")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple symbols")
	return cmd
}

func NewGetSymbolDimensionsCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var cmd = &cobra.Command{
		Use:     "dimensions [symbol-id]",
		Short:   "Get the dimensions of a symbol by id",
		Long:    `Get the dimensions of a symbol by id`,
		Example: "gns3util -s https://controller:3080 symbol dimensions symbol-id",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [symbol-id] when --fuzzy is not set")
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
				params := fuzzy.NewFuzzyInfoParams(cfg, "getSymbols", "symbol_id", multi)
				err := fuzzy.FuzzyInfo(params)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				id := args[0]
				if !utils.IsValidUUIDv4(args[0]) {
					id, err = utils.ResolveID(cfg, "symbol", args[0], nil)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				utils.ExecuteAndPrint(cfg, "getSymbolDimensions", []string{id})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find a symbol")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get dimensions for multiple symbols")
	return cmd
}

func NewGetDefaultSymbolsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "defaults",
		Short:   "Get the available default-symbols",
		Long:    `Get the available symbols`,
		Example: "gns3util -s https://controller:3080 symbol defaults",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getDefaultSymbols", nil)
		},
	}
	return cmd
}
