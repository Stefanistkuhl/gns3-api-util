package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/cmd/get"
)

func NewApplianceCmdGroup() *cobra.Command {
	applianceCmd := &cobra.Command{
		Use:   "appliance",
		Short: "Appliance operations",
		Long:  `Get and manage GNS3 appliances.`,
	}

	// Get subcommands
	applianceCmd.AddCommand(get.NewGetAppliancesCmd())
	applianceCmd.AddCommand(get.NewGetApplianceCmd())

	return applianceCmd
}
