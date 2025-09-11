package get

import (
	"github.com/spf13/cobra"
)

func NewGetCmdGroup() *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get GNS3 resources (e.g., projects, users)",
		Long:  `Provides commands to retrieve various resources from a GNS3v3 server.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	getCmd.AddCommand(NewGetVersionCmd())
	getCmd.AddCommand(NewGetIouLicenseCmd())
	getCmd.AddCommand(NewGetStatisticsCmd())
	getCmd.AddCommand(NewGetMeCmd())
	getCmd.AddCommand(NewGetGroupMembershipsCmd())
	getCmd.AddCommand(NewGetGroupsCmd())
	getCmd.AddCommand(NewGetGroupCmd())
	getCmd.AddCommand(NewGetGroupMembersCmd())
	getCmd.AddCommand(NewGetRolesCmd())
	getCmd.AddCommand(NewGetRoleCmd())
	getCmd.AddCommand(NewGetRolePrivsCmd())
	getCmd.AddCommand(NewGetPrivilegesCmd())
	getCmd.AddCommand(NewGetProjectsCmd())
	getCmd.AddCommand(NewGetProjectCmd())
	getCmd.AddCommand(NewGetProjectStatsCmd())
	getCmd.AddCommand(NewGetProjectLockedCmd())
	getCmd.AddCommand(NewGetProjectExportCmd())
	getCmd.AddCommand(NewGetProjectFileCmd())
	getCmd.AddCommand(NewGetNodeFileCmd())
	getCmd.AddCommand(NewStreamPcapCmd())
	getCmd.AddCommand(NewGetAclCmd())
	getCmd.AddCommand(NewGetAceCmd())
	getCmd.AddCommand(NewGetAclEndpointsCmd())
	getCmd.AddCommand(NewGetImagesCmd())
	getCmd.AddCommand(NewGetImageCmd())
	getCmd.AddCommand(NewGetUsersCmd())
	getCmd.AddCommand(NewGetUserCmd())
	getCmd.AddCommand(NewGetTemplatesCmd())
	getCmd.AddCommand(NewGetTemplateCmd())
	getCmd.AddCommand(NewGetNodesCmd())
	getCmd.AddCommand(NewGetNodeCmd())
	getCmd.AddCommand(NewGetNodeLinksCmd())
	getCmd.AddCommand(NewGetNodesAutoIdlePCCmd())
	getCmd.AddCommand(NewGetNodesAutoIdlePCProposalsCmd())
	getCmd.AddCommand(NewGetLinksCmd())
	getCmd.AddCommand(NewGetLinkCmd())
	getCmd.AddCommand(NewGetLinkIfaceCmd())
	getCmd.AddCommand(NewGetLinkFiltersCmd())
	getCmd.AddCommand(NewGetDrawingCmd())
	getCmd.AddCommand(NewGetDrawingsCmd())
	getCmd.AddCommand(NewGetSymbolsCmd())
	getCmd.AddCommand(NewGetSymbolCmd())
	getCmd.AddCommand(NewGetSymbolDimensionsCmd())
	getCmd.AddCommand(NewGetDefaultSymbolsCmd())
	getCmd.AddCommand(NewGetSnapshotsCmd())
	getCmd.AddCommand(NewGetComputeCmd())
	getCmd.AddCommand(NewGetComputesCmd())
	getCmd.AddCommand(NewGetComputeDockerImagesCmd())
	getCmd.AddCommand(NewGetComputeVirtualboxVMSCmd())
	getCmd.AddCommand(NewGetComputeVmWareVMSCmd())
	getCmd.AddCommand(NewGetAppliancesCmd())
	getCmd.AddCommand(NewGetApplianceCmd())
	getCmd.AddCommand(NewGetPoolsCmd())
	getCmd.AddCommand(NewGetPoolCmd())
	getCmd.AddCommand(NewGetPoolResourcesCmd())
	getCmd.AddCommand(NewGetNotificationsCmd())
	getCmd.AddCommand(NewGetProjectNotificationCmd())

	return getCmd
}
