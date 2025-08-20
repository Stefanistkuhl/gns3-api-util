package endpoints

import "fmt"

type PutEndpoints struct{}

func (PutEndpoints) UpdateIOULicense() string {
	return "/iou_license"
}

func (PutEndpoints) UpdateMe() string {
	return "/access/users/me"
}

func (PutEndpoints) UpdateUser(userID string) string {
	return fmt.Sprintf("/access/users/%s", userID)
}

func (PutEndpoints) UpdateGroup(groupID string) string {
	return fmt.Sprintf("/access/groups/%s", groupID)
}

func (PutEndpoints) UpdateRole(roleID string) string {
	return fmt.Sprintf("/access/roles/%s", roleID)
}

func (PutEndpoints) UpdateACE(aceID string) string {
	return fmt.Sprintf("/access/acl/%s", aceID)
}

func (PutEndpoints) UpdateTemplate(templateID string) string {
	return fmt.Sprintf("/templates/%s", templateID)
}

func (PutEndpoints) UpdateProject(projectID string) string {
	return fmt.Sprintf("/projects/%s", projectID)
}

func (PutEndpoints) UpdateNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s", projectID, nodeID)
}

func (PutEndpoints) UpdateQemuDiskImage(projectID, nodeID, diskName string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s/qemu/disk_image/%s", projectID, nodeID, diskName)
}

func (PutEndpoints) UpdateLink(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s", projectID, linkID)
}

func (PutEndpoints) UpdateDrawing(projectID, drawingID string) string {
	return fmt.Sprintf("/projects/%s/drawings/%s", projectID, drawingID)
}

func (PutEndpoints) UpdateCompute(computeID string) string {
	return fmt.Sprintf("/v3/gns3/computes/%s", computeID)
}

func (PutEndpoints) UpdatePool(poolID string) string {
	return fmt.Sprintf("/pools/%s", poolID)
}

func (PutEndpoints) AddGroupMember(groupID, userID string) string {
	return fmt.Sprintf("/access/groups/%s/members/%s", groupID, userID)
}

func (PutEndpoints) AddPrivilege(roleID, privilegeID string) string {
	return fmt.Sprintf("/access/roles/%s/privileges/%s", roleID, privilegeID)
}

func (PutEndpoints) AddToPool(poolID, resourceID string) string {
	return fmt.Sprintf("/pools/%s/resources/%s", poolID, resourceID)
}
