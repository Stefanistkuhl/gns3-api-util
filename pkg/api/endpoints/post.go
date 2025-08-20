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

func (PostEndpoints) LockProject(projectID string) string {
	return fmt.Sprintf("/projects/%s/lock", projectID)
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
