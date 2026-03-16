package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/spf13/cobra"
)

func NewSymbolCmdGroup() *cobra.Command {
	symbolCmd := &cobra.Command{
		Use:     "symbol",
		Aliases: []string{"sym"},
		Short:   "Symbol operations",
		Long:    `Get and manage GNS3 symbols.`,
	}

	symbolCmd.AddCommand(get.NewGetSymbolsCmd())
	symbolCmd.AddCommand(get.NewGetSymbolCmd())
	symbolCmd.AddCommand(get.NewGetSymbolDimensionsCmd())
	symbolCmd.AddCommand(get.NewGetDefaultSymbolsCmd())

	return symbolCmd
}
