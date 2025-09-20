package create

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewCreateExerciseCmd() *cobra.Command {
	var createExerciseCmd = &cobra.Command{
		Use:   "exercise",
		Short: "Create an exercise (project) for every group in a class with ACLs",
		Long: `Create an exercise (project) for every group in a class with ACLs to lock down access.

This command will:
- Create a project for each group in the specified class
- Create resource pools for each project
- Create ACLs to restrict access to each group's project
- Assign the "User" role to each group for their respective projects`,
		Example: `
  # Create exercise for a class
  gns3util -s https://controller:3080 create exercise --class "CS101" --exercise "Lab1"

  # Create exercise with custom project name format
  gns3util -s https://controller:3080 create exercise --class "CS101" --exercise "Lab1" --format "{{class}}-{{exercise}}-{{group}}"
		`,
		RunE: runCreateExercise,
	}

	createExerciseCmd.Flags().String("class", "", "Class name to create exercise for")
	createExerciseCmd.Flags().String("exercise", "", "Exercise name")
	createExerciseCmd.Flags().String("format", "{{class}}-{{exercise}}-{{group}}-{{uuid}}", "Project name format (supports {{class}}, {{exercise}}, {{group}}, {{uuid}})")
	createExerciseCmd.Flags().Bool("confirm", true, "Confirm before creating projects")

	_ = createExerciseCmd.MarkFlagRequired("class")
	_ = createExerciseCmd.MarkFlagRequired("exercise")

	return createExerciseCmd
}

func runCreateExercise(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}

	className, _ := cmd.Flags().GetString("class")
	exerciseName, _ := cmd.Flags().GetString("exercise")
	format, _ := cmd.Flags().GetString("format")
	confirm, _ := cmd.Flags().GetBool("confirm")

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

	classGroups := findClassGroups(groups, className)
	if len(classGroups) == 0 {
		return fmt.Errorf("no groups found for class %s", messageUtils.Bold(className))
	}

	fmt.Printf("%v Found %d groups for class %v:\n",
		messageUtils.InfoMsg("Found groups for class"),
		len(classGroups),
		messageUtils.Bold(className))
	for _, group := range classGroups {
		fmt.Printf("  - %v\n", messageUtils.Highlight(group.Name))
	}

	if confirm {
		if !confirmAction(fmt.Sprintf("Create exercise '%s' for %d groups?", exerciseName, len(classGroups))) {
			fmt.Println("Exercise creation cancelled.")
			return nil
		}
	}

	// Get the "User" role ID
	roleID, err := getUserRoleID(cfg)
	if err != nil {
		return fmt.Errorf("failed to get User role ID: %w", err)
	}

	// Check if groups already have exercises
	existingExercises, err := checkExistingExercises(cfg, className, classGroups)
	if err != nil {
		return fmt.Errorf("failed to check existing exercises: %w", err)
	}

	if len(existingExercises) > 0 {
		fmt.Printf("%v Warning: Some groups already have exercises:\n", messageUtils.WarningMsg("Some groups already have exercises"))
		for _, groupName := range existingExercises {
			fmt.Printf("  - %v\n", messageUtils.Bold(groupName))
		}
		fmt.Printf("%v Skipping groups that already have exercises\n", messageUtils.InfoMsg("Skipping groups that already have exercises"))
	}

	successCount := 0
	for _, group := range classGroups {
		groupName := group.Name
		groupID := group.UserGroupID.String()

		hasExercise := false
		for _, existingGroup := range existingExercises {
			if existingGroup == groupName {
				hasExercise = true
				break
			}
		}
		if hasExercise {
			continue
		}

		groupNumber := extractGroupNumber(groupName, className)

		projectName := generateProjectName(format, className, exerciseName, groupNumber)

		if err := createProjectForGroup(cfg, projectName, groupID, roleID); err != nil {
			fmt.Printf("%v Failed to create project for group %s: %v\n",
				messageUtils.ErrorMsg("Failed to create project for group"),
				messageUtils.Bold(groupName),
				err)
			continue
		}

		successCount++
		fmt.Printf("%v Created project %s for group %s\n",
			messageUtils.SuccessMsg("Created project"),
			messageUtils.Bold(projectName),
			messageUtils.Highlight(groupName))
	}

	fmt.Printf("\n%v Created %d projects for exercise '%s'\n",
		messageUtils.SuccessMsg("Created projects for exercise"),
		successCount,
		messageUtils.Bold(exerciseName))

	return nil
}

func findClassGroups(groups []schemas.UserGroupResponse, className string) []schemas.UserGroupResponse {
	var classGroups []schemas.UserGroupResponse

	for _, group := range groups {
		if strings.HasPrefix(group.Name, className+"-") && group.Name != className {
			classGroups = append(classGroups, group)
		}
	}

	return classGroups
}

func extractGroupNumber(groupName, className string) string {
	prefix := className + "-"
	if strings.HasPrefix(groupName, prefix) {
		return strings.TrimPrefix(groupName, prefix)
	}
	return groupName
}

func generateProjectName(format, className, exerciseName, groupNumber string) string {
	uuidStr := uuid.New().String()[:8]

	projectName := format
	projectName = strings.ReplaceAll(projectName, "{{class}}", className)
	projectName = strings.ReplaceAll(projectName, "{{exercise}}", exerciseName)
	projectName = strings.ReplaceAll(projectName, "{{group}}", groupNumber)
	projectName = strings.ReplaceAll(projectName, "{{uuid}}", uuidStr)

	return projectName
}

func getUserRoleID(cfg config.GlobalOptions) (string, error) {
	rolesBody, status, err := utils.CallClient(cfg, "getRoles", []string{}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get roles: %w", err)
	}

	if status != 200 {
		return "", fmt.Errorf("failed to get roles: status %d", status)
	}

	var roles []schemas.RoleResponse
	if err := json.Unmarshal(rolesBody, &roles); err != nil {
		return "", fmt.Errorf("failed to parse roles response: %w", err)
	}

	for _, role := range roles {
		if role.Name == "User" {
			return role.RoleID, nil
		}
	}

	return "", fmt.Errorf("user role not found")
}

func createProjectForGroup(cfg config.GlobalOptions, projectName, groupID, roleID string) error {
	projectData := schemas.ProjectCreate{
		Name: &projectName,
	}

	projectBody, status, err := utils.CallClient(cfg, "createProject", []string{}, projectData)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	if status != 201 {
		return fmt.Errorf("failed to create project: status %d", status)
	}

	var projectResponse schemas.ProjectResponse
	if err := json.Unmarshal(projectBody, &projectResponse); err != nil {
		return fmt.Errorf("failed to parse project response: %w", err)
	}

	projectID := projectResponse.ProjectID

	_, status, err = utils.CallClient(cfg, "closeProject", []string{projectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to close project: %w", err)
	}

	if status != 200 && status != 204 {
		return fmt.Errorf("failed to close project: status %d", status)
	}

	poolName := fmt.Sprintf("%s-pool", projectName)
	poolData := schemas.ResourcePoolCreate{
		Name: &poolName,
	}

	poolBody, status, err := utils.CallClient(cfg, "createPool", []string{}, poolData)
	if err != nil {
		return fmt.Errorf("failed to create resource pool: %w", err)
	}

	if status != 201 {
		return fmt.Errorf("failed to create resource pool: status %d", status)
	}

	var poolResponse schemas.ResourcePoolResponse
	if err := json.Unmarshal(poolBody, &poolResponse); err != nil {
		return fmt.Errorf("failed to parse pool response: %w", err)
	}

	poolID := poolResponse.ResourcePoolID

	_, status, err = utils.CallClient(cfg, "addToPool", []string{poolID, projectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to add project to pool: %w", err)
	}

	if status != 201 && status != 204 {
		return fmt.Errorf("failed to add project to pool: status %d", status)
	}

	aceType := "group"
	path := fmt.Sprintf("/pools/%s", poolID)
	propagate := true
	allowed := true

	aclData := schemas.ACECreate{
		ACEType:   &aceType,
		Path:      &path,
		Propagate: &propagate,
		Allowed:   &allowed,
		GroupID:   &groupID,
		RoleID:    &roleID,
	}

	_, status, err = utils.CallClient(cfg, "createACL", []string{}, aclData)
	if err != nil {
		return fmt.Errorf("failed to create ACL: %w", err)
	}

	if status != 201 && status != 204 {
		return fmt.Errorf("failed to create ACL: status %d", status)
	}

	return nil
}

func checkExistingExercises(cfg config.GlobalOptions, className string, classGroups []schemas.UserGroupResponse) ([]string, error) {
	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("failed to get projects: status %d", status)
	}

	var projects []schemas.ProjectResponse
	if err := json.Unmarshal(projectsBody, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	groupMap := make(map[string]bool)
	for _, group := range classGroups {
		groupMap[group.Name] = true
	}

	var existingExercises []string
	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 && parts[0] == className {
			groupName := strings.Join(parts[:3], "-")
			if groupMap[groupName] {
				existingExercises = append(existingExercises, groupName)
			}
		}
	}

	return existingExercises, nil
}

func confirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	var response string
	_, _ = fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}
