package exercise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api"
	"github.com/stefanistkuhl/gns3util/pkg/api/endpoints"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/authentication"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func NewExerciseCreateCmd() *cobra.Command {
	var createExerciseCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an exercise (project) for every group in a class with ACLs",
		Long: `Create an exercise (project) for every group in a class with ACLs to lock down access.

This command will:
- Use an existing template project from the server (recommended) or create empty projects for each group
- Create resource pools for each project
- Create ACLs to restrict access to each group's project
- Assign the "User" role to each group for their respective projects`,
		Example: `
  # Create exercise for a class
  gns3util -s https://controller:3080 exercise create --class "CS101" --exercise "Lab1"

  # Create exercise with custom project name format
  gns3util -s https://controller:3080 exercise create --class "CS101" --exercise "Lab1" --format "{{class}}-{{exercise}}-{{group}}"

  # Create exercise using interactive template selection (recommended)
  gns3util -s https://controller:3080 exercise create --class "CS101" --exercise "Lab1" --select-template

  # Create exercise using a specific template project by name/ID
  gns3util -s https://controller:3080 exercise create --class "CS101" --exercise "Lab1" --template "MyTemplateProject"

  # Create exercise using a template file (fallback)
  gns3util -s https://controller:3080 exercise create --class "CS101" --exercise "Lab1" --template "/path/to/template.gns3project"
		`,
		RunE: runCreateExercise,
	}

	createExerciseCmd.Flags().String("class", "", "Class name to create exercise for")
	createExerciseCmd.Flags().String("exercise", "", "Exercise name")
	createExerciseCmd.Flags().String("format", "{{class}}-{{exercise}}-{{group}}-{{uuid}}", "Project name format (supports {{class}}, {{exercise}}, {{group}}, {{uuid}})")
	createExerciseCmd.Flags().String("template", "", "Existing project name/ID or path to template file (.gns3project) to use as base for all exercise projects")
	createExerciseCmd.Flags().Bool("select-template", false, "Interactively select a template project from existing projects on the server (recommended)")
	createExerciseCmd.Flags().Bool("confirm", true, "Confirm before creating projects")

	createExerciseCmd.MarkFlagRequired("class")
	createExerciseCmd.MarkFlagRequired("exercise")

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
	templatePath, _ := cmd.Flags().GetString("template")
	selectTemplate, _ := cmd.Flags().GetBool("select-template")
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
		return fmt.Errorf("no groups found for class %s", className)
	}

	fmt.Printf("%v Found %d groups for class %v:\n",
		colorUtils.Info("Info:"),
		len(classGroups),
		colorUtils.Bold(className))
	for _, group := range classGroups {
		fmt.Printf("  - %v\n", colorUtils.Highlight(group.Name))
	}

	if confirm {
		if !confirmAction(fmt.Sprintf("Create exercise '%s' for %d groups?", exerciseName, len(classGroups))) {
			fmt.Println("Exercise creation cancelled.")
			return nil
		}
	}

	roleID, err := getUserRoleID(cfg)
	if err != nil {
		return fmt.Errorf("failed to get User role ID: %w", err)
	}

	var templateProjectID string
	if selectTemplate {
		templateProjectID, err = selectTemplateWithFuzzy(cfg)
		if err != nil {
			return fmt.Errorf("failed to select template project: %w", err)
		}
		fmt.Printf("%v Selected template project: %v\n",
			colorUtils.Success("Success:"),
			colorUtils.Bold(templateProjectID))
	} else if templatePath != "" {
		if _, err := os.Stat(templatePath); err == nil {
			templateProjectID, err = importTemplateProject(cfg, templatePath, className, exerciseName)
			if err != nil {
				return fmt.Errorf("failed to import template project: %w", err)
			}
			fmt.Printf("%v Imported template project: %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(templateProjectID))
		} else {
			templateProjectID, err = resolveTemplateProject(cfg, templatePath)
			if err != nil {
				return fmt.Errorf("failed to resolve template project: %w", err)
			}
			fmt.Printf("%v Using existing template project: %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(templateProjectID))
		}
	}

	existingExercises, err := checkExistingExercises(cfg, className, classGroups)
	if err != nil {
		return fmt.Errorf("failed to check existing exercises: %w", err)
	}

	if len(existingExercises) > 0 {
		fmt.Printf("%v Warning: Some groups already have exercises:\n", colorUtils.Warning("Warning:"))
		for _, groupName := range existingExercises {
			fmt.Printf("  - %v\n", colorUtils.Bold(groupName))
		}
		fmt.Printf("%v Skipping groups that already have exercises\n", colorUtils.Info("Info:"))
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

		if err := createProjectForGroup(cfg, projectName, groupID, roleID, templateProjectID); err != nil {
			fmt.Printf("%v Failed to create project for group %s: %v\n",
				colorUtils.Error("Error:"),
				colorUtils.Bold(groupName),
				err)
			continue
		}

		successCount++
		fmt.Printf("%v Created project %s for group %s\n",
			colorUtils.Success("Success:"),
			colorUtils.Bold(projectName),
			colorUtils.Highlight(groupName))
	}

	if templateProjectID != "" {
		if err := cleanupTemplateProject(cfg, templateProjectID); err != nil {
			fmt.Printf("%v Warning: Failed to clean up template project: %v\n",
				colorUtils.Warning("Warning:"),
				err)
		} else {
			fmt.Printf("%v Cleaned up template project\n", colorUtils.Info("Info:"))
		}
	}

	fmt.Printf("\n%v Created %d projects for exercise '%s'\n",
		colorUtils.Success("Success:"),
		successCount,
		colorUtils.Bold(exerciseName))

	return nil
}

func findClassGroups(groups []schemas.UserGroupResponse, className string) []schemas.UserGroupResponse {
	var classGroups []schemas.UserGroupResponse

	for _, group := range groups {
		if strings.HasPrefix(group.Name, className+"-") {
			classGroups = append(classGroups, group)
		}
	}

	return classGroups
}

func extractGroupNumber(groupName, className string) string {
	withoutClass := strings.TrimPrefix(groupName, className+"-")
	parts := strings.Split(withoutClass, "-")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return "1"
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

func importTemplateProject(cfg config.GlobalOptions, templatePath, className, exerciseName string) (string, error) {
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file does not exist: %s", templatePath)
	}

	file, err := os.Open(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(templatePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	token, err := authentication.GetKeyForServer(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	settings := api.NewSettings(
		api.WithBaseURL(cfg.Server),
		api.WithVerify(cfg.Insecure),
		api.WithToken(token),
	)
	client := api.NewGNS3Client(settings)

	ep := endpoints.Endpoints{}
	templateProjectID := uuid.New().String()
	templateProjectName := fmt.Sprintf("%s-%s-template", className, exerciseName)
	urlStr := ep.Post.ProjectImport(templateProjectID) + fmt.Sprintf("?name=%s", url.QueryEscape(templateProjectName))

	reqOpts := api.NewRequestOptions(settings).
		WithURL(urlStr).
		WithMethod(api.POST).
		WithData(buf.String())

	_, resp, err := client.Do(reqOpts)
	if err != nil {
		return "", fmt.Errorf("failed to import template project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("failed to import template project with status %d", resp.StatusCode)
	}

	return templateProjectID, nil
}

func resolveTemplateProject(cfg config.GlobalOptions, templateIdentifier string) (string, error) {
	if utils.IsValidUUIDv4(templateIdentifier) {
		return templateIdentifier, nil
	}

	projectID, err := utils.ResolveID(cfg, "project", templateIdentifier, nil)
	if err != nil {
		return "", fmt.Errorf("failed to resolve template project '%s': %w", templateIdentifier, err)
	}

	return projectID, nil
}

func selectTemplateWithFuzzy(cfg config.GlobalOptions) (string, error) {
	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}
	if status != 200 {
		return "", fmt.Errorf("failed to get projects: status %d", status)
	}

	var projects []schemas.ProjectResponse
	if err := json.Unmarshal(projectsBody, &projects); err != nil {
		return "", fmt.Errorf("failed to parse projects response: %w", err)
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found on server")
	}

	var projectNames []string
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	selectedNames := fuzzy.NewFuzzyFinder(projectNames, false)
	if len(selectedNames) == 0 {
		return "", fmt.Errorf("no project selected")
	}

	selectedName := selectedNames[0]
	for _, project := range projects {
		if project.Name == selectedName {
			return project.ProjectID, nil
		}
	}

	return "", fmt.Errorf("selected project not found")
}

func cleanupTemplateProject(cfg config.GlobalOptions, templateProjectID string) error {
	_, status, err := utils.CallClient(cfg, "closeProject", []string{templateProjectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to close template project: %w", err)
	}
	if status != 200 && status != 204 {
		return fmt.Errorf("failed to close template project: status %d", status)
	}

	_, status, err = utils.CallClient(cfg, "deleteProject", []string{templateProjectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete template project: %w", err)
	}
	if status != 204 {
		return fmt.Errorf("failed to delete template project: status %d", status)
	}

	return nil
}

func createProjectForGroup(cfg config.GlobalOptions, projectName, groupID, roleID, templateProjectID string) error {
	var projectID string

	if templateProjectID != "" {
		duplicateData := schemas.ProjectDuplicate{
			Name: projectName,
		}

		projectBody, status, err := utils.CallClient(cfg, "duplicateProject", []string{templateProjectID}, duplicateData)
		if err != nil {
			return fmt.Errorf("failed to duplicate template project: %w", err)
		}

		if status != 201 {
			return fmt.Errorf("failed to duplicate template project: status %d", status)
		}

		var projectResponse schemas.ProjectResponse
		if err := json.Unmarshal(projectBody, &projectResponse); err != nil {
			return fmt.Errorf("failed to parse project response: %w", err)
		}

		projectID = projectResponse.ProjectID
	} else {
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

		projectID = projectResponse.ProjectID
	}

	_, status, err := utils.CallClient(cfg, "closeProject", []string{projectID}, nil)
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
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}
