package class

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func LoadClassFromFile(filePath string) (schemas.Class, error) {
	var classData schemas.Class

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return classData, fmt.Errorf("file does not exist: %s", colorUtils.Bold(filePath))
	}

	file, err := os.Open(filePath)
	if err != nil {
		return classData, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return classData, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, &classData); err != nil {
		return classData, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if classData.Name == "" {
		return classData, fmt.Errorf("class name is required")
	}

	if len(classData.Groups) == 0 {
		return classData, fmt.Errorf("at least one group is required")
	}

	fmt.Printf("%v Loaded class %v with %d groups\n",
		colorUtils.Info("Info:"),
		colorUtils.Bold(classData.Name),
		len(classData.Groups))

	return classData, nil
}

func CreateClass(cfg config.GlobalOptions, classData schemas.Class) (bool, error) {
	classGroupData := schemas.UserGroupCreate{
		Name: &classData.Name,
	}

	classGroupBody, status, err := utils.CallClient(cfg, "createGroup", []string{}, classGroupData)
	if err != nil {
		return false, fmt.Errorf("failed to create class group: %w", err)
	}

	if status != 201 {
		return false, fmt.Errorf("failed to create class group: status %d", status)
	}

	var classGroupResponse schemas.UserGroupResponse
	if err := json.Unmarshal(classGroupBody, &classGroupResponse); err != nil {
		return false, fmt.Errorf("failed to parse class group response: %w", err)
	}

	classGroupName := classGroupResponse.Name

	fmt.Printf("%v Created class group %v\n",
		colorUtils.Success("Success:"),
		colorUtils.Bold(classGroupName))

	for _, group := range classData.Groups {
		studentGroupData := schemas.UserGroupCreate{
			Name: &group.Name,
		}

		studentGroupBody, status, err := utils.CallClient(cfg, "createGroup", []string{}, studentGroupData)
		if err != nil {
			return false, fmt.Errorf("failed to create student group %s: %w", group.Name, err)
		}

		if status != 201 {
			return false, fmt.Errorf("failed to create student group %s: status %d", group.Name, status)
		}

		var studentGroupResponse schemas.UserGroupResponse
		if err := json.Unmarshal(studentGroupBody, &studentGroupResponse); err != nil {
			return false, fmt.Errorf("failed to parse student group response: %w", err)
		}

		studentGroupName := studentGroupResponse.Name

		fmt.Printf("%v Created student group %v\n",
			colorUtils.Success("Success:"),
			colorUtils.Bold(studentGroupName))

		for _, student := range group.Students {

			userData := schemas.UserCreate{
				Username: &student.UserName,
				Password: &student.Password,
				Email:    student.Email,
				FullName: student.FullName,
				IsActive: true,
			}

			userBody, status, err := utils.CallClient(cfg, "createUser", []string{}, userData)
			if err != nil {
				return false, fmt.Errorf("failed to create user %s: %w", student.UserName, err)
			}

			if status != 201 {
				return false, fmt.Errorf("failed to create user %s: status %d", student.UserName, status)
			}

			var userResponse schemas.UserResponse
			if err := json.Unmarshal(userBody, &userResponse); err != nil {
				return false, fmt.Errorf("failed to parse user response: %w", err)
			}

			userID := userResponse.UserID.String()
			username := userResponse.Username

			fmt.Printf("%v Created user %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(username))

			classGroupID := classGroupResponse.UserGroupID.String()
			studentGroupID := studentGroupResponse.UserGroupID.String()

			if err := addUserToGroup(cfg, userID, classGroupID); err != nil {
				return false, fmt.Errorf("failed to add user %s to class group: %w", username, err)
			}

			if err := addUserToGroup(cfg, userID, studentGroupID); err != nil {
				return false, fmt.Errorf("failed to add user %s to student group: %w", username, err)
			}
		}
	}

	return true, nil
}

func addUserToGroup(cfg config.GlobalOptions, userID, groupID string) error {
	_, status, err := utils.CallClient(cfg, "addGroupMember", []string{groupID, userID}, nil)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	if status != 200 && status != 204 {
		return fmt.Errorf("failed to add user to group: status %d", status)
	}

	return nil
}

func DeleteClass(cfg config.GlobalOptions, className string) error {
	groupsBody, status, err := utils.CallClient(cfg, "getGroups", []string{}, nil)
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to get groups: status %d", status)
	}

	var groups []schemas.UserGroupResponse
	if err := json.Unmarshal(groupsBody, &groups); err != nil {
		return fmt.Errorf("failed to parse groups response: %w", err)
	}

	classGroups, studentGroups := findClassAndStudentGroups(groups, className)

	if len(classGroups) == 0 && len(studentGroups) == 0 {
		return fmt.Errorf("no groups found for class %s", colorUtils.Bold(className))
	}

	fmt.Printf("%v Found %d class groups and %d student groups for class %v\n",
		colorUtils.Info("Info:"),
		len(classGroups),
		len(studentGroups),
		colorUtils.Bold(className))

	allGroupsToDelete := append(studentGroups, classGroups...)

	allUsersToDelete := make(map[string]string)

	for _, group := range allGroupsToDelete {
		groupID := group.UserGroupID.String()
		groupName := group.Name

		members, err := getGroupMembers(cfg, groupID)
		if err != nil {
			fmt.Printf("%v Warning: failed to get members for group %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(groupName),
				err)
		} else {
			for _, member := range members {
				userID := member.UserID.String()
				username := member.Username
				allUsersToDelete[userID] = username
			}
		}
	}

	for userID, username := range allUsersToDelete {
		if err := deleteUser(cfg, userID); err != nil {
			fmt.Printf("%v Warning: failed to delete user %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(username),
				err)
		} else {
			fmt.Printf("%v Deleted user %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(username))
		}
	}

	for _, group := range studentGroups {
		groupID := group.UserGroupID.String()
		groupName := group.Name

		if err := deleteGroup(cfg, groupID); err != nil {
			fmt.Printf("%v Warning: failed to delete student group %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(groupName),
				err)
		} else {
			fmt.Printf("%v Deleted student group %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(groupName))
		}
	}

	for _, group := range classGroups {
		groupID := group.UserGroupID.String()
		groupName := group.Name

		if err := deleteGroup(cfg, groupID); err != nil {
			fmt.Printf("%v Warning: failed to delete class group %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(groupName),
				err)
		} else {
			fmt.Printf("%v Deleted class group %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(groupName))
		}
	}

	return nil
}

func findClassAndStudentGroups(groups []schemas.UserGroupResponse, className string) ([]schemas.UserGroupResponse, []schemas.UserGroupResponse) {
	var classGroups []schemas.UserGroupResponse
	var studentGroups []schemas.UserGroupResponse

	for _, group := range groups {
		if group.Name == className {
			classGroups = append(classGroups, group)
		} else if strings.HasPrefix(group.Name, className+"-") {
			studentGroups = append(studentGroups, group)
		}
	}

	return classGroups, studentGroups
}

func getGroupMembers(cfg config.GlobalOptions, groupID string) ([]schemas.UserResponse, error) {
	membersBody, status, err := utils.CallClient(cfg, "getGroupMembers", []string{groupID}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("failed to get group members: status %d", status)
	}

	var members []schemas.UserResponse
	if err := json.Unmarshal(membersBody, &members); err != nil {
		return nil, fmt.Errorf("failed to parse group members response: %w", err)
	}

	return members, nil
}

func deleteUser(cfg config.GlobalOptions, userID string) error {
	_, status, err := utils.CallClient(cfg, "deleteUser", []string{userID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to delete user: status %d", status)
	}
	return nil
}

func deleteGroup(cfg config.GlobalOptions, groupID string) error {
	_, status, err := utils.CallClient(cfg, "deleteGroup", []string{groupID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to delete group: status %d", status)
	}
	return nil
}

func DeleteExercise(cfg config.GlobalOptions, exerciseName, className, groupName string) error {
	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to get projects: status %d", status)
	}

	var projects []schemas.ProjectResponse
	if err := json.Unmarshal(projectsBody, &projects); err != nil {
		return fmt.Errorf("failed to parse projects response: %w", err)
	}

	var exerciseProjects []schemas.ProjectResponse
	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 && parts[1] == exerciseName {
			exerciseProjects = append(exerciseProjects, project)
		}
	}

	if len(exerciseProjects) == 0 {
		return fmt.Errorf("exercise %s not found", colorUtils.Bold(exerciseName))
	}

	fmt.Printf("%v Found %d projects for exercise %v\n",
		colorUtils.Info("Info:"),
		len(exerciseProjects),
		colorUtils.Bold(exerciseName))

	for _, project := range exerciseProjects {
		projectID := project.ProjectID
		projectName := project.Name

		if err := closeProject(cfg, projectID); err != nil {
			fmt.Printf("%v Warning: failed to close project %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(projectName),
				err)
		}

		pools, err := getPoolsForExercise(cfg, exerciseName, className, groupName)
		if err != nil {
			fmt.Printf("%v Warning: failed to get pools for exercise %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(exerciseName),
				err)
		} else {
			for _, pool := range pools {
				poolID := pool.ResourcePoolID
				if err := deleteACLsForPool(cfg, poolID); err != nil {
					fmt.Printf("%v Warning: failed to delete ACLs for pool %v: %v\n",
						colorUtils.Warning("Warning:"),
						colorUtils.Bold(poolID),
						err)
				}
			}

			for _, pool := range pools {
				poolID := pool.ResourcePoolID
				poolName := pool.Name
				if err := deletePool(cfg, poolID); err != nil {
					fmt.Printf("%v Warning: failed to delete pool %v: %v\n",
						colorUtils.Warning("Warning:"),
						colorUtils.Bold(poolName),
						err)
				} else {
					fmt.Printf("%v Deleted pool %v\n",
						colorUtils.Success("Success:"),
						colorUtils.Bold(poolName))
				}
			}
		}

		if err := deleteProject(cfg, projectID); err != nil {
			fmt.Printf("%v Warning: failed to delete project %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(projectName),
				err)
		} else {
			fmt.Printf("%v Deleted project %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(projectName))
		}
	}

	fmt.Printf("%v Deleted exercise %v\n",
		colorUtils.Success("Success:"),
		colorUtils.Bold(exerciseName))

	return nil
}

func DeleteAllExercisesForClass(cfg config.GlobalOptions, className string) error {
	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to get projects: status %d", status)
	}

	var projects []schemas.ProjectResponse
	if err := json.Unmarshal(projectsBody, &projects); err != nil {
		return fmt.Errorf("failed to parse projects response: %w", err)
	}

	var classExercises []string
	seenExercises := make(map[string]bool)

	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 && parts[0] == className {
			exerciseName := parts[1]
			if !seenExercises[exerciseName] {
				classExercises = append(classExercises, exerciseName)
				seenExercises[exerciseName] = true
			}
		}
	}

	if len(classExercises) == 0 {
		fmt.Printf("%v No exercises found for class %v\n",
			colorUtils.Info("Info:"),
			colorUtils.Bold(className))
		return nil
	}

	fmt.Printf("%v Found %d exercises for class %v\n",
		colorUtils.Info("Info:"),
		len(classExercises),
		colorUtils.Bold(className))

	for _, exerciseName := range classExercises {
		if err := DeleteExercise(cfg, exerciseName, className, ""); err != nil {
			fmt.Printf("%v Warning: failed to delete exercise %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(exerciseName),
				err)
		}
	}

	return nil
}

func closeProject(cfg config.GlobalOptions, projectID string) error {
	_, status, err := utils.CallClient(cfg, "closeProject", []string{projectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to close project: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to close project: status %d", status)
	}
	return nil
}

func getPoolsForExercise(cfg config.GlobalOptions, exerciseName, className, groupName string) ([]schemas.ResourcePoolResponse, error) {
	poolsBody, status, err := utils.CallClient(cfg, "getPools", []string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("failed to get pools: status %d", status)
	}

	var allPools []schemas.ResourcePoolResponse
	if err := json.Unmarshal(poolsBody, &allPools); err != nil {
		return nil, fmt.Errorf("failed to parse pools response: %w", err)
	}

	var matchingPools []schemas.ResourcePoolResponse
	for _, pool := range allPools {
		if strings.HasPrefix(pool.Name, className+"-"+exerciseName+"-") && strings.HasSuffix(pool.Name, "-pool") {
			matchingPools = append(matchingPools, pool)
		}
	}

	return matchingPools, nil
}

func deleteACLsForPool(cfg config.GlobalOptions, poolID string) error {
	aclsBody, status, err := utils.CallClient(cfg, "getACL", []string{}, nil)
	if err != nil {
		return fmt.Errorf("failed to get ACLs: %w", err)
	}
	if status != 200 {
		return fmt.Errorf("failed to get ACLs: status %d", status)
	}

	var acls []schemas.ACLResponse
	if err := json.Unmarshal(aclsBody, &acls); err != nil {
		return fmt.Errorf("failed to parse ACLs response: %w", err)
	}

	poolPath := fmt.Sprintf("/pools/%s", poolID)
	for _, acl := range acls {
		if acl.Path == poolPath {
			aclID := acl.ACLID
			if err := deleteACL(cfg, aclID); err != nil {
				return fmt.Errorf("failed to delete ACL: %w", err)
			}
		}
	}

	return nil
}

func deletePool(cfg config.GlobalOptions, poolID string) error {
	_, status, err := utils.CallClient(cfg, "deletePool", []string{poolID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete pool: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to delete pool: status %d", status)
	}
	return nil
}

func deleteProject(cfg config.GlobalOptions, projectID string) error {
	_, status, err := utils.CallClient(cfg, "deleteProject", []string{projectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to delete project: status %d", status)
	}
	return nil
}

func deleteACL(cfg config.GlobalOptions, aclID string) error {
	_, status, err := utils.CallClient(cfg, "deleteACL", []string{aclID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete ACL: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to delete ACL: status %d", status)
	}
	return nil
}
