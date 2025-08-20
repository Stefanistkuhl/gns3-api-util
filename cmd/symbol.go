package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/get"
)

func NewSymbolCmdGroup() *cobra.Command {
	symbolCmd := &cobra.Command{
		Use:   "symbol",
		Short: "Symbol operations",
		Long:  `Get and manage GNS3 symbols.`,
	}

	// Get subcommands
	symbolCmd.AddCommand(get.NewGetSymbolsCmd())
	symbolCmd.AddCommand(get.NewGetSymbolCmd())
	symbolCmd.AddCommand(get.NewGetSymbolDimensionsCmd())
	symbolCmd.AddCommand(get.NewGetDefaultSymbolsCmd())

	return symbolCmd
}
