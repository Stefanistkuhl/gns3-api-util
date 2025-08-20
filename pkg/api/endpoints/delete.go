package endpoints

import "fmt"

type DeleteEndpoints struct{}

func (DeleteEndpoints) PruneImages() string {
	return "/images/prune"
}

func (DeleteEndpoints) DeleteUser(userID string) string {
	return fmt.Sprintf("/access/users/%s", userID)
}

func (DeleteEndpoints) DeleteCompute(computeID string) string {
	return fmt.Sprintf("/computes/%s", computeID)
}

func (DeleteEndpoints) DeleteProject(projectID string) string {
	return fmt.Sprintf("/projects/%s", projectID)
}

func (DeleteEndpoints) DeleteTemplate(templateID string) string {
	return fmt.Sprintf("/templates/%s", templateID)
}

func (DeleteEndpoints) DeleteImage(imagePath string) string {
	return fmt.Sprintf("/images/%s", imagePath)
}

func (DeleteEndpoints) DeleteACE(aceID string) string {
	return fmt.Sprintf("/access/acl/%s", aceID)
}

func (DeleteEndpoints) DeleteRole(roleID string) string {
	return fmt.Sprintf("/access/roles/%s", roleID)
}

func (DeleteEndpoints) DeleteGroup(groupID string) string {
	return fmt.Sprintf("/access/groups/%s", groupID)
}

func (DeleteEndpoints) DeletePool(poolID string) string {
	return fmt.Sprintf("/pools/%s", poolID)
}

func (DeleteEndpoints) DeletePoolResource(poolID, resourceID string) string {
	return fmt.Sprintf("/pools/%s/resources/%s", poolID, resourceID)
}

func (DeleteEndpoints) DeleteLink(projectID, linkID string) string {
	return fmt.Sprintf("/projects/%s/links/%s", projectID, linkID)
}

func (DeleteEndpoints) DeleteNode(projectID, nodeID string) string {
	return fmt.Sprintf("/projects/%s/nodes/%s", projectID, nodeID)
}

func (DeleteEndpoints) DeleteDrawing(projectID, drawingID string) string {
	return fmt.Sprintf("/projects/%s/drawings/%s", projectID, drawingID)
}

func (DeleteEndpoints) DeleteRolePrivilege(roleID, privilegeID string) string {
	return fmt.Sprintf("/access/roles/%s/privileges/%s", roleID, privilegeID)
}

func (DeleteEndpoints) DeleteUserFromGroup(groupID, userID string) string {
	return fmt.Sprintf("/access/groups/%s/members/%s", groupID, userID)
}

func (DeleteEndpoints) DeleteSnapshot(projectID, snapshotID string) string {
	return fmt.Sprintf("/projects/%s/snapshots/%s", projectID, snapshotID)
}
