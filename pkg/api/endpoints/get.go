package endpoints

import "fmt"

type GetEndpoints struct{}

func (GetEndpoints) Version() string {
	return "/version"
}

func (GetEndpoints) IouLicense() string {
	return "/iou_license"
}

func (GetEndpoints) Statistics() string {
	return "/statistics"
}

func (GetEndpoints) Me() string {
	return "/access/users/me"
}

func (GetEndpoints) GroupMemberships(userID string) string {
	return fmt.Sprintf("/access/users/%s/groups", userID)
}

func (GetEndpoints) Projects() string {
	return "/projects"
}

func (GetEndpoints) Project(projectID string) string {
	return fmt.Sprintf("/projects/%s", projectID)
}

func (GetEndpoints) ProjectStats(projectID string) string {
	return fmt.Sprintf("/projects/%s/stats", projectID)
}

func (GetEndpoints) ProjectLocked(projectID string) string {
	return fmt.Sprintf("/projects/%s/locked", projectID)
}

func (GetEndpoints) Users() string {
	return "/access/users"
}

func (GetEndpoints) User(userID string) string {
	return fmt.Sprintf("/access/users/%s", userID)
}

func (GetEndpoints) Groups() string {
	return "/access/groups"
}

func (GetEndpoints) Group(groupID string) string {
	return fmt.Sprintf("/access/groups/%s", groupID)
}

func (GetEndpoints) GroupMembers(groupID string) string {
	return fmt.Sprintf("/access/groups/%s/members", groupID)
}

func (GetEndpoints) Roles() string {
	return "/access/roles"
}

func (GetEndpoints) Role(roleID string) string {
	return fmt.Sprintf("/access/roles/%s", roleID)
}

func (GetEndpoints) RolePrivs(roleID string) string {
	return fmt.Sprintf("/access/roles/%s/privileges", roleID)
}

func (GetEndpoints) ACL() string {
	return "/access/acl"
}

func (GetEndpoints) ACLEndpoints() string {
	return "/access/acl/endpoints"
}

func (GetEndpoints) ACE(aceID string) string {
	return fmt.Sprintf("/access/acl/%s", aceID)
}

func (GetEndpoints) Images(imageType string) string {
	if imageType != "" {
		return fmt.Sprintf("/images?image_type=%s", imageType)
	}
	return "/images"
}

func (GetEndpoints) Image(imagePath string) string {
	return fmt.Sprintf("/images/%s", imagePath)
}

func (GetEndpoints) Templates() string {
	return "/templates"
}

func (GetEndpoints) Template(templateID string) string {
	return fmt.Sprintf("/templates/%s", templateID)
}

func (GetEndpoints) Computes() string {
	return "/computes"
}

func (GetEndpoints) Compute(computeID string) string {
	return fmt.Sprintf("/computes/%s", computeID)
}

func (GetEndpoints) ComputeDocker(computeID string) string {
	return fmt.Sprintf("/computes/%s/docker/images", computeID)
}

func (GetEndpoints) ComputeVirtualbox(computeID string) string {
	return fmt.Sprintf("/computes/%s/virtualbox/vms", computeID)
}

func (GetEndpoints) ComputeVmware(computeID string) string {
	return fmt.Sprintf("/computes/%s/vmware/vms", computeID)
}

func (GetEndpoints) Pools() string {
	return "/pools"
}

func (GetEndpoints) Pool(poolID string) string {
	return fmt.Sprintf("/pools/%s", poolID)
}

func (GetEndpoints) PoolResources(poolID string) string {
	return fmt.Sprintf("/pools/%s/resources", poolID)
}

func (GetEndpoints) Nodes(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes", projectID)
}

func (GetEndpoints) Node(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s", projectID, nodeID)
}

func (GetEndpoints) NodeLinks(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/links", projectID, nodeID)
}

func (GetEndpoints) NodeAutoIdlePc(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/dynamips/auto_idlepc", projectID, nodeID)
}

func (GetEndpoints) NodeAutoIdlePcProposals(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/dynamips/idlepc_proposals", projectID, nodeID)
}

func (GetEndpoints) Links(projectID string) string {
	return fmt.Sprintf("/projects/%s/links", projectID)
}

func (GetEndpoints) Link(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s", projectID, linkID)
}

func (GetEndpoints) LinkIface(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s/iface", projectID, linkID)
}

func (GetEndpoints) LinkFilters(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s/available_filters", projectID, linkID)
}

func (GetEndpoints) Drawings(projectID string) string {
	return fmt.Sprintf("/projects/%s/drawings", projectID)
}

func (GetEndpoints) Drawing(projectID, drawingID string) string {
	return fmt.Sprintf("/projects/%s/drawings/%s", projectID, drawingID)
}

func (GetEndpoints) Symbols() string {
	return "/symbols"
}

func (GetEndpoints) DefaultSymbols() string {
	return "/symbols/default_symbols"
}

func (GetEndpoints) Symbol(symbolID string) string {
	return fmt.Sprintf("/symbols/%s/raw", symbolID)
}

func (GetEndpoints) SymbolDimensions(symbolID string) string {
	return fmt.Sprintf("/symbols/%s/dimensions", symbolID)
}

func (GetEndpoints) Snapshots(projectID string) string {
	return fmt.Sprintf("/projects/%s/snapshots", projectID)
}

func (GetEndpoints) Appliances() string {
	return "/appliances"
}

func (GetEndpoints) Appliance(applianceID string) string {
	return fmt.Sprintf("/appliances/%s/", applianceID)
}

func (GetEndpoints) Notifications() string {
	return "/notifications"
}

func (GetEndpoints) ProjectNotifications(projectID string) string {
	return fmt.Sprintf("/projects/%s/notifications", projectID)
}

func (GetEndpoints) ProjectExport(projectID string) string {
	return fmt.Sprintf("/projects/%s/export", projectID)
}

func (GetEndpoints) ProjectFile(projectID, filePath string) string {
	return fmt.Sprintf("/projects/%s/files/%s", projectID, filePath)
}

func (GetEndpoints) NodeFile(projectID, nodeID, filePath string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/files/%s", projectID, nodeID, filePath)
}
func (GetEndpoints) StreamPcap(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s/capture/stream", projectID, linkID)
}

func (GetEndpoints) GetPrivileges() string {
	return "/access/privileges"
}
