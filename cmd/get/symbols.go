package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetSymbolsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "symbols",
		Short: "Get the avaliable symbols",
		Long:  `Get the avaliable symbols`,
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
	var cmd = &cobra.Command{
		Use:   "symbol",
		Short: "Get a symbol by id",
		Long:  `Get a symbol by id`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getSymbol", []string{id})
		},
	}
	return cmd
}

func NewGetSymbolDimensionsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "symbol-dimensions",
		Short: "Get the dimensions of a symbol by id",
		Long:  `Get the dimensions of a symbol by id`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getSymbolDimensions", []string{id})
		},
	}
	return cmd
}

func NewGetDefaultSymbolsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "default-symbols",
		Short: "Get the avaliable default-symbols",
		Long:  `Get the avaliable symbols`,
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
