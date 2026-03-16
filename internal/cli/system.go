package cli

import (
	"github.com/0xveya/gns3util/internal/cli/cmds/get"
	"github.com/0xveya/gns3util/internal/cli/cmds/post"
	"github.com/0xveya/gns3util/internal/cli/cmds/put/update"
	"github.com/spf13/cobra"
)

func NewSystemCmdGroup() *cobra.Command {
	systemCmd := &cobra.Command{
		Use:     "system",
		Aliases: []string{"sys"},
		Short:   "System operations",
		Long:    `Manage GNS3 system operations and settings.`,
	}

	systemCmd.AddCommand(get.NewGetVersionCmd())
	systemCmd.AddCommand(get.NewGetStatisticsCmd())
	systemCmd.AddCommand(get.NewGetNotificationsCmd())
	systemCmd.AddCommand(get.NewGetIouLicenseCmd())

	systemCmd.AddCommand(post.NewCheckVersionCmd())
	systemCmd.AddCommand(post.NewControllerCmdGroup())

	systemCmd.AddCommand(update.NewUpdateIOULicenseCmd())

	return systemCmd
}
