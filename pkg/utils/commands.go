package utils

import (
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
)

type CommandConfig struct {
	Method   api.HTTPMethod
	Endpoint func(ep endpoints.Endpoints, args []string) string
}

var commandMap = map[string]CommandConfig{
	"getVersion": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Version()
		},
	},
	"getIouLicense": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.IouLicense()
		},
	},
	"getStatistics": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Statistics()
		},
	},
	"getMe": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Me()
		},
	},
	"getUser": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.User(args[0])
		},
	},
	"getUsers": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Users()
		},
	},
	"getGroupMemberships": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.GroupMemberships(args[0])
		},
	},
	"getGroups": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Groups()
		},
	},
	"getGroup": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Group(args[0])
		},
	},
	"getGroupMembers": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.GroupMembers(args[0])
		},
	},
	"getProjects": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Projects()
		},
	},
	"getProject": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Project(args[0])
		},
	},
	"getProjectStats": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ProjectStats(args[0])
		},
	},
	"getProjectLocked": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ProjectLocked(args[0])
		},
	},
	"getRoles": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Roles()
		},
	},
	"getRole": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Role(args[0])
		},
	},
	"getPrivileges": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.GetPrivileges()
		},
	},
	"getRolePrivs": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.RolePrivs(args[0])
		},
	},
	"getAcl": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ACL()
		},
	},
	"getAce": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ACE(args[0])
		},
	},
	"getAclEndpoints": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ACLEndpoints()
		},
	},
	"getImages": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Images(args[0])
		},
	},
	"getImage": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Image(args[0])
		},
	},
	"getTemplates": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Templates()
		},
	},
	"getTemplate": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Template(args[0])
		},
	},
	"getComputes": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Computes()
		},
	},
	"getPools": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Pools()
		},
	},
	"getPool": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Pool(args[0])
		},
	},
	"getPoolResources": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.PoolResources(args[0])
		},
	},
	"getNodes": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Nodes(args[0])
		},
	},
	"getNode": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Node(args[0], args[1])
		},
	},
	"getNodeLinks": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.NodeLinks(args[0], args[1])
		},
	},
	"getNodeAutoIdlePc": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.NodeAutoIdlePc(args[0], args[1])
		},
	},
	"getNodeAutoIdlePcProposals": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.NodeAutoIdlePcProposals(args[0], args[1])
		},
	},
	"getLinks": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Links(args[0])
		},
	},
	"getLink": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Link(args[0], args[1])
		},
	},
	"getLinkIface": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.LinkIface(args[0], args[1])
		},
	},
	"getDrawing": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Drawing(args[0], args[1])
		},
	},
	"getDrawings": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Drawings(args[0])
		},
	},
	"getLinkFilters": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.LinkFilters(args[0], args[1])
		},
	},
	"getSymbols": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Symbols()
		},
	},
	"getSymbol": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Symbol(args[0])
		},
	},
	"getSymbolDimensions": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.SymbolDimensions(args[0])
		},
	},
	"getDefaultSymbols": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.DefaultSymbols()
		},
	},
	"getSnapshots": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Snapshots(args[0])
		},
	},
	"exportProject": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ProjectExport(args[0])
		},
	},
	"getProjectFile": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ProjectFile(args[0], args[1])
		},
	},
	"getNodeFile": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.NodeFile(args[0], args[1], args[2])
		},
	},
	"streamPcap": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.StreamPcap(args[0], args[1])
		},
	},
	"getCompute": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Compute(args[0])
		},
	},
	"getComputeDockerImgs": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ComputeDocker(args[0])
		},
	},
	"getComputeVirtualboxVms": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ComputeVirtualbox(args[0])
		},
	},
	"getAppliances": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Appliances()
		},
	},
	"getAppliance": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.Appliance(args[0])
		},
	},
	"getComputeVmwareVms": {
		Method: api.GET,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Get.ComputeVmware(args[0])
		},
	},
	"lockProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.LockProject(args[0])
		},
	},
	"createUser": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateUser()
		},
	},
	"createGroup": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateGroup()
		},
	},
	"createRole": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateRole()
		},
	},
	"createACL": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateACL()
		},
	},
	"createQemuImage": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateQemuImage(args[0])
		},
	},
	"createTemplate": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateTemplate()
		},
	},
	"createProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateProject()
		},
	},
	"createProjectNodeFromTemplate": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateProjectNodeFromTemplate(args[0], args[1])
		},
	},
	"createNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateNode(args[0])
		},
	},
	"createDiskImage": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateDiskImage(args[0], args[1], args[2])
		},
	},
	"createLink": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateLink(args[0])
		},
	},
	"createDrawing": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateDrawing(args[0])
		},
	},
	"createSnapshot": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreateSnapshot(args[0])
		},
	},
	"createCompute": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			// args[0] should be stringified bool for connect; endpoint formats it properly
			return ep.Post.CreateCompute(args[0] == "true")
		},
	},
	"createPool": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CreatePool()
		},
	},
	"closeProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.CloseProject(args[0])
		},
	},
	"updateIOULicense": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateIOULicense()
		},
	},
	"updateMe": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateMe()
		},
	},
	"updateUser": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateUser(args[0])
		},
	},
	"updateGroup": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateGroup(args[0])
		},
	},
	"updateRole": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateRole(args[0])
		},
	},
	"updateACE": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateACE(args[0])
		},
	},
	"updateTemplate": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateTemplate(args[0])
		},
	},
	"updateProject": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateProject(args[0])
		},
	},
	"updateNode": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateNode(args[0], args[1])
		},
	},
	"updateQemuDiskImage": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateQemuDiskImage(args[0], args[1], args[2])
		},
	},
	"updateLink": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateLink(args[0], args[1])
		},
	},
	"updateDrawing": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateDrawing(args[0], args[1])
		},
	},
	"updateCompute": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdateCompute(args[0])
		},
	},
	"updatePool": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.UpdatePool(args[0])
		},
	},
	// Add commands
	"addGroupMember": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.AddGroupMember(args[0], args[1])
		},
	},
	"addPrivilege": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.AddPrivilege(args[0], args[1])
		},
	},
	"addToPool": {
		Method: api.PUT,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Put.AddToPool(args[0], args[1])
		},
	},
	// Delete commands
	"deleteUser": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteUser(args[0])
		},
	},
	"deleteGroup": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteGroup(args[0])
		},
	},
	"deleteRole": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteRole(args[0])
		},
	},
	"deleteTemplate": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteTemplate(args[0])
		},
	},
	"deleteProject": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteProject(args[0])
		},
	},
	"deleteCompute": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteCompute(args[0])
		},
	},
	"deleteImage": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteImage(args[0])
		},
	},
	"deleteNode": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteNode(args[0], args[1])
		},
	},
	"deleteLink": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteLink(args[0], args[1])
		},
	},
	"deleteDrawing": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteDrawing(args[0], args[1])
		},
	},
	"deletePool": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeletePool(args[0])
		},
	},
	"deletePoolResource": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeletePoolResource(args[0], args[1])
		},
	},
	"deleteACE": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteACE(args[0])
		},
	},
	"deleteRolePrivilege": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteRolePrivilege(args[0], args[1])
		},
	},
	"deleteUserFromGroup": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteUserFromGroup(args[0], args[1])
		},
	},
	"deleteSnapshot": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.DeleteSnapshot(args[0], args[1])
		},
	},
	"deletePruneImages": {
		Method: api.DELETE,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Delete.PruneImages()
		},
	},
	// Post commands
	"userAuthenticate": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.UserAuthenticate()
		},
	},
	"checkVersion": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.Version()
		},
	},
	"reloadController": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.Reload()
		},
	},
	"shutdownController": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.Shutdown()
		},
	},
	"openProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.OpenProject(args[0])
		},
	},
	"loadProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.LoadProject()
		},
	},
	"duplicateProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.DuplicateProject(args[0])
		},
	},
	"projectImport": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.ProjectImport(args[0])
		},
	},
	"duplicateTemplate": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.DuplicateTemplate(args[0])
		},
	},
	"duplicateNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.DuplicateNode(args[0], args[1])
		},
	},
	"startAllNodes": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.StartNodes(args[0])
		},
	},
	"stopAllNodes": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.StopNodes(args[0])
		},
	},
	"suspendAllNodes": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.SuspendNodes(args[0])
		},
	},
	"reloadAllNodes": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.ReloadNodes(args[0])
		},
	},
	"startNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.StartNode(args[0], args[1])
		},
	},
	"stopNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.StopNode(args[0], args[1])
		},
	},
	"suspendNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.SuspendNode(args[0], args[1])
		},
	},
	"reloadNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.ReloadNode(args[0], args[1])
		},
	},
	"isolateNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.IsolateNode(args[0], args[1])
		},
	},
	"unisolateNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.UnisolateNode(args[0], args[1])
		},
	},
	"resetConsoleAllNodes": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.NodesConsoleReset(args[0])
		},
	},
	"resetConsoleNode": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.NodeConsoleReset(args[0], args[1])
		},
	},
	"resetLink": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.ResetLink(args[0], args[1])
		},
	},
	"startCapture": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.StartCapture(args[0], args[1])
		},
	},
	"stopCapture": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.StopCapture(args[0], args[1])
		},
	},
	"uploadImage": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.UploadImage(args[0])
		},
	},
	"installImages": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.InstallImage()
		},
	},
	"unlockProject": {
		Method: api.POST,
		Endpoint: func(ep endpoints.Endpoints, args []string) string {
			return ep.Post.UnlockProject(args[0])
		},
	},
}
