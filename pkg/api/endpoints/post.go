package endpoints

import "fmt"

type PostEndpoints struct{}

func (PostEndpoints) Authenticate() string {
	return "/access/users/authenticate"
}

func (PostEndpoints) CreateUser() string {
	return "/access/users"
}

func (PostEndpoints) CreateGroup() string {
	return "/access/groups"
}

func (PostEndpoints) CreateRole() string {
	return "/access/roles"
}

func (PostEndpoints) CreateACL() string {
	return "/access/acl"
}

func (PostEndpoints) CreateQemuImage(imagePath string) string {
	return fmt.Sprintf("/images/qemu/%s", imagePath)
}

func (PostEndpoints) CreateTemplate() string {
	return "/templates"
}

func (PostEndpoints) DuplicateTemplate(templateID string) string {
	return fmt.Sprintf("/templates/%s/duplicate", templateID)
}

func (PostEndpoints) DuplicateProject(projectID string) string {
	return fmt.Sprintf("/projects/%s/duplicate", projectID)
}

func (PostEndpoints) DuplicateNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/duplicate", projectID, nodeID)
}

func (PostEndpoints) StartNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/start", projectID, nodeID)
}

func (PostEndpoints) StopNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/stop", projectID, nodeID)
}

func (PostEndpoints) SuspendNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/suspend", projectID, nodeID)
}

func (PostEndpoints) ReloadNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/reload", projectID, nodeID)
}

func (PostEndpoints) IsolateNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/isolate", projectID, nodeID)
}

func (PostEndpoints) UnisolateNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/unisolate", projectID, nodeID)
}

func (PostEndpoints) NodesConsoleReset(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes/console/reset", projectID)
}

func (PostEndpoints) NodeConsoleReset(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/console/reset", projectID, nodeID)
}

func (PostEndpoints) ResetLink(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s/reset", projectID, linkID)
}

func (PostEndpoints) StartCapture(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s/capture/start", projectID, linkID)
}

func (PostEndpoints) StopCapture(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s/capture/stop", projectID, linkID)
}

func (PostEndpoints) UploadImage(imagePath string) string {
	return fmt.Sprintf("/images/upload/%s", imagePath)
}

func (PostEndpoints) InstallImage() string {
	return "/images/install"
}

func (PostEndpoints) StartNodes(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes/start", projectID)
}

func (PostEndpoints) StopNodes(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes/stop", projectID)
}

func (PostEndpoints) SuspendNodes(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes/suspend", projectID)
}

func (PostEndpoints) ReloadNodes(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes/reload", projectID)
}

func (PostEndpoints) CreateProject() string {
	return "/projects"
}

func (PostEndpoints) CreateProjectNodeFromTemplate(projectID, templateID string) string {
	return fmt.Sprintf("/projects/%s/templates/%s", projectID, templateID)
}

func (PostEndpoints) CreateNode(projectID string) string {
	return fmt.Sprintf("/projects/%s/nodes", projectID)
}

func (PostEndpoints) CreateDiskImage(projectID, nodeID, diskName string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/qemu/disk_image/%s", projectID, nodeID, diskName)
}

func (PostEndpoints) CreateLink(projectID string) string {
	return fmt.Sprintf("/projects/%s/links", projectID)
}

func (PostEndpoints) CreateDrawing(projectID string) string {
	return fmt.Sprintf("/projects/%s/drawings", projectID)
}

func (PostEndpoints) CreateSnapshot(projectID string) string {
	return fmt.Sprintf("/projects/%s/snapshots", projectID)
}

func (PostEndpoints) CreateCompute(connect bool) string {
	return fmt.Sprintf("/computes?connect=%t", connect)
}

func (PostEndpoints) OpenProject(projectID string) string {
	return fmt.Sprintf("/projects/%s/open", projectID)
}

func (PostEndpoints) LockProject(projectID string) string {
	return fmt.Sprintf("/projects/%s/lock", projectID)
}

func (PostEndpoints) UnlockProject(projectID string) string {
	return fmt.Sprintf("/projects/%s/unlock", projectID)
}

func (PostEndpoints) LoadProject() string {
	return "/projects/load"
}

func (PostEndpoints) CloseProject(projectID string) string {
	return fmt.Sprintf("/projects/%s/close", projectID)
}

func (PostEndpoints) CreatePool() string {
	return "/pools"
}

func (PostEndpoints) ProjectImport(projectID string) string {
	return fmt.Sprintf("/projects/%s/import", projectID)
}

func (PostEndpoints) Reload() string {
	return "/reload"
}

func (PostEndpoints) UserAuthenticate() string {
	return "/access/users/authenticate"
}

func (PostEndpoints) Shutdown() string {
	return "/shutdown"
}

func (PostEndpoints) Version() string {
	return "/version"
}
