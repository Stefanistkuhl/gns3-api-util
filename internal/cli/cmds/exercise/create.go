package exercise

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/authentication"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/cluster/db/sqlc"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils/messageUtils"
	"github.com/0xveya/gns3util/pkg/api"
	"github.com/0xveya/gns3util/pkg/api/endpoints"
	"github.com/0xveya/gns3util/pkg/api/schemas"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewExerciseCreateCmd() *cobra.Command {
	createExerciseCmd := &cobra.Command{
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
	createExerciseCmd.Flags().Bool("delete-template-project", false, "Delete the template when using a project as a template")
	createExerciseCmd.Flags().StringP("cluster", "", "", "Cluster name (note: create is server-scoped; use -s)")

	_ = createExerciseCmd.MarkFlagRequired("class")

	return createExerciseCmd
}

func selectAndReplicateTemplateAcrossCluster(cfg *config.GlobalOptions, clusterID int) (map[string]string, error) {
	store, err := db.Init()
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	nodes, err := store.GetNodes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get nodes: %w", err)
	}

	perNode := make(map[string]map[string]string)
	nameSet := make(map[string]struct{})

	for _, node := range nodes {
		if int(node.ClusterID) != clusterID {
			continue
		}

		nodeURL := fmt.Sprintf("%s://%s:%d", node.Protocol, node.Host, node.Port)
		cfgServer := cfg
		cfgServer.Server = nodeURL

		body, status, apiErr := utils.CallClient(cfgServer, "getProjects", []string{}, nil)
		if apiErr != nil {
			return nil, fmt.Errorf("[%s] getProjects: %w", nodeURL, apiErr)
		}
		if status != 200 {
			return nil, fmt.Errorf("[%s] getProjects status: %d", nodeURL, status)
		}
		var projects []schemas.ProjectResponse
		if unmarshalErr := json.Unmarshal(body, &projects); unmarshalErr != nil {
			return nil, fmt.Errorf("[%s] parse projects: %w", nodeURL, unmarshalErr)
		}

		nodeProjects := make(map[string]string)
		for i := range projects {
			p := &projects[i]
			nodeProjects[p.Name] = p.ProjectID
			nameSet[p.Name] = struct{}{}
		}
		perNode[nodeURL] = nodeProjects
	}

	if len(nameSet) == 0 {
		return nil, fmt.Errorf("no projects found on any node in cluster")
	}

	names := make([]string, 0, len(nameSet))
	for k := range nameSet {
		names = append(names, k)
	}
	slices.Sort(names)

	if len(names) == 0 {
		return nil, fmt.Errorf("no template projects available")
	}

	selected := fuzzy.NewFuzzyFinder(names, false)
	if len(selected) == 0 {
		return nil, fmt.Errorf("no items selected")
	}
	selName := selected[0]

	var srcURL, srcProjID string
	for nodeURL, pm := range perNode {
		if id, ok := pm[selName]; ok {
			srcURL = nodeURL
			srcProjID = id
			break
		}
	}

	if srcURL == "" {
		return nil, fmt.Errorf("selected template not found on any node")
	}

	srcCfg := cfg
	srcCfg.Server = srcURL
	exportData, err := exportProjectArchive(srcCfg, srcProjID)
	if err != nil {
		return nil, fmt.Errorf("export from %s: %w", srcURL, err)
	}

	result := make(map[string]string)
	for nodeURL, pm := range perNode {
		tgtCfg := cfg
		tgtCfg.Server = nodeURL
		if id, ok := pm[selName]; ok {
			result[nodeURL] = id
			continue
		}
		newID, err := importProjectArchive(tgtCfg, exportData, selName)
		if err != nil {
			return nil, fmt.Errorf("import to %s failed: %w", nodeURL, err)
		}
		result[nodeURL] = newID
		fmt.Printf("%s Imported template '%s' to %s\n", messageUtils.SuccessMsg("Imported template"), messageUtils.Bold(selName), messageUtils.Highlight(nodeURL))
	}

	return result, nil
}

func exportProjectArchive(cfg *config.GlobalOptions, projectID string) ([]byte, error) {
	body, status, err := utils.CallClient(cfg, "exportProject", []string{projectID}, nil)
	if err != nil {
		return nil, fmt.Errorf("export project: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("export project status %d", status)
	}
	return body, nil
}

func importProjectArchive(cfg *config.GlobalOptions, archive []byte, projectName string) (string, error) {
	token, err := authentication.GetKeyForServer(cfg)
	if err != nil {
		return "", fmt.Errorf("get token: %w", err)
	}
	settings := api.NewSettings(api.WithBaseURL(cfg.Server), api.WithVerify(cfg.Insecure), api.WithToken(token))
	client := api.NewGNS3Client(&settings)

	ep := endpoints.Endpoints{}
	newID := uuid.New().String()
	urlStr := ep.Post.ProjectImport(newID) + fmt.Sprintf("?name=%s", url.QueryEscape(projectName))

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", fmt.Sprintf("%s.gns3project", projectName))
	if err != nil {
		return "", err
	}
	if _, err = fw.Write(archive); err != nil {
		return "", err
	}
	_ = w.Close()

	req := api.NewRequestOptions(&settings).WithURL(urlStr).WithMethod(api.POST).WithData(buf.String())
	_, resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("import status %d", resp.StatusCode)
	}
	return newID, nil
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
	deleteTemplate, _ := cmd.Flags().GetBool("delete-template-project")
	clusterName, _ := cmd.Flags().GetString("cluster")

	if exerciseName == "" && len(args) > 0 {
		exerciseName = args[0]
	}

	ctx := context.Background()

	if clusterName != "" {
		store, err := db.Init()
		if err != nil {
			return fmt.Errorf("failed to init db: %w", err)
		}

		clusters, err := store.GetClusters(ctx)
		if err != nil {
			return fmt.Errorf("failed to get clusters: %w", err)
		}

		clusterID := 0
		for _, c := range clusters {
			if strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(clusterName)) {
				clusterID = int(c.ClusterID)
				break
			}
		}
		if clusterID == 0 {
			return fmt.Errorf("cluster not found: %s", clusterName)
		}

		classesInDb, getClassesErr := store.GetClasses(ctx, int64(clusterID))
		if getClassesErr != nil {
			return fmt.Errorf("failed to get classes from db: %w", getClassesErr)
		}

		if len(classesInDb) == 0 {
			return fmt.Errorf("no classes found in database for cluster %d", clusterID)
		}

		classFound := false
		for _, class := range classesInDb {
			if class.Name == className {
				classFound = true
			}
		}
		if !classFound {
			return fmt.Errorf("class %s not found in database for cluster %d", className, clusterID)
		}

		data, err := store.GetNodeGroupNamesForClass(ctx, sqlc.GetNodeGroupNamesForClassParams{ClusterID: int64(clusterID), Name: className})
		if err != nil {
			return fmt.Errorf("failed to get node groups: %w", err)
		}

		var plans []db.NodeGroupsForClass
		nodeMap := make(map[string]*db.NodeGroupsForClass)
		groupMap := make(map[string]*db.GroupData)

		for _, dat := range data {
			nodeURL := fmt.Sprintf("%v", dat.NodeUrl)

			node, ok := nodeMap[nodeURL]
			if !ok {
				plans = append(plans, db.NodeGroupsForClass{
					NodeURL: nodeURL,
					Groups:  []db.GroupData{},
				})
				node = &plans[len(plans)-1]
				nodeMap[nodeURL] = node
			}

			groupKey := nodeURL + dat.GroupName
			group, ok := groupMap[groupKey]
			if !ok {
				node.Groups = append(node.Groups, db.GroupData{
					Name:     dat.GroupName,
					Students: []db.UserData{},
				})
				group = &node.Groups[len(node.Groups)-1]
				groupMap[groupKey] = group
			}

			fullName := ""
			if dat.FullName.Valid {
				fullName = dat.FullName.String
			}

			group.Students = append(group.Students, db.UserData{
				Username:  dat.Username,
				FullName:  fullName,
				Password:  dat.DefaultPassword,
				GroupName: dat.GroupName,
			})
		}

		var templateIDByNode map[string]string
		if selectTemplate {
			templateIDByNode, err = selectAndReplicateTemplateAcrossCluster(cfg, clusterID)
			if err != nil {
				return fmt.Errorf("template selection/replication failed: %w", err)
			}
			if templateIDByNode == nil {
				return nil
			}
		}

		totalCreated := 0
		for _, plan := range plans {
			cfgServer := cfg
			cfgServer.Server = plan.NodeURL

			groupsBody, status, err := utils.CallClient(cfgServer, "getGroups", []string{}, nil)
			if err != nil {
				return fmt.Errorf("[%s] getGroups: %w", plan.NodeURL, err)
			}
			if status != 200 {
				return fmt.Errorf("[%s] getGroups status: %d", plan.NodeURL, status)
			}
			var groups []schemas.UserGroupResponse
			if unmarshalErr := json.Unmarshal(groupsBody, &groups); unmarshalErr != nil {
				return fmt.Errorf("[%s] parse groups: %w", plan.NodeURL, unmarshalErr)
			}
			want := make(map[string]bool)
			for _, g := range plan.Groups {
				want[g.Name] = true
			}
			var classGroups []schemas.UserGroupResponse
			for _, g := range groups {
				if want[g.Name] {
					classGroups = append(classGroups, g)
				}
			}

			if len(classGroups) == 0 {
				continue
			}

			if confirm && totalCreated == 0 {
				var totalGroups int
				for i := range plans {
					totalGroups += len(plans[i].Groups)
				}
				if !utils.ConfirmPrompt(fmt.Sprintf("Create exercise '%s' for %d groups?", exerciseName, totalGroups), false) {
					fmt.Println("Exercise creation cancelled.")
					return nil
				}
			}

			var preselectedTemplate string
			if selectTemplate {
				preselectedTemplate = templateIDByNode[plan.NodeURL]
			}
			_, created, err := createForGroupsOnServer(
				cfgServer, className, exerciseName, format, templatePath,
				selectTemplate, deleteTemplate, classGroups, preselectedTemplate,
			)
			if err != nil {
				return err
			}
			totalCreated += created
		}

		fmt.Printf("\n%v Created %d projects for exercise '%s' across cluster %s\n",
			messageUtils.SuccessMsg("Created projects for exercise"),
			totalCreated,
			messageUtils.Bold(exerciseName),
			messageUtils.Bold(clusterName),
		)
		return nil
	}

	return nil
}

func createForGroupsOnServer(cfg *config.GlobalOptions, className, exerciseName, format, templatePath string, selectTemplate, deleteTemplate bool, classGroups []schemas.UserGroupResponse, preselectedTemplateID string) (templateID string, createdCount int, err error) {
	fmt.Printf("%v Found %d groups for class %v on %s\n",
		messageUtils.InfoMsg("Found groups for class"),
		len(classGroups),
		messageUtils.Bold(className),
		messageUtils.Highlight(cfg.Server),
	)
	for _, group := range classGroups {
		fmt.Printf("  - %v\n", messageUtils.Highlight(group.Name))
	}

	roleID, err := getUserRoleID(cfg)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get User role ID: %w", err)
	}

	var templateProjectID string
	switch {
	case preselectedTemplateID != "":
		templateProjectID = preselectedTemplateID
	case selectTemplate:
		templateProjectID, err = selectTemplateWithFuzzy(cfg)
		if err != nil {
			return "", 0, fmt.Errorf("failed to select template project: %w", err)
		}
		fmt.Printf("%v Selected template project: %v\n",
			messageUtils.SuccessMsg("Selected template project"),
			messageUtils.Bold(templateProjectID))
	case templatePath != "":
		if _, statErr := os.Stat(templatePath); statErr == nil {
			templateProjectID, err = importTemplateProject(cfg, templatePath, className, exerciseName)
			if err != nil {
				return "", 0, fmt.Errorf("failed to import template project: %w", err)
			}
			fmt.Printf("%v Imported template project: %v\n",
				messageUtils.SuccessMsg("Imported template project"),
				messageUtils.Bold(templateProjectID))
		} else {
			templateProjectID, err = resolveTemplateProject(cfg, templatePath)
			if err != nil {
				return "", 0, fmt.Errorf("failed to resolve template project: %w", err)
			}
			fmt.Printf("%v Using existing template project: %v\n",
				messageUtils.SuccessMsg("Using existing template project"),
				messageUtils.Bold(templateProjectID))
		}
	}

	var exportData []byte
	if templateProjectID != "" {
		_, _, _ = utils.CallClient(cfg, "closeProject", []string{templateProjectID}, nil)
		exportData, err = exportProjectArchive(cfg, templateProjectID)
		if err != nil {
			return "", 0, fmt.Errorf("export template: %w", err)
		}
	}

	existingExercises, err := checkExistingExercises(cfg, className, classGroups)
	if err != nil {
		return templateProjectID, 0, fmt.Errorf("failed to check existing exercises: %w", err)
	}
	if len(existingExercises) > 0 {
		fmt.Printf("%v Some groups already have exercises (server %s):\n", messageUtils.WarningMsg("Some groups already have exercises"), messageUtils.Highlight(cfg.Server))
		for _, groupName := range existingExercises {
			fmt.Printf("  - %v\n", messageUtils.Bold(groupName))
		}
		fmt.Printf("%v Skipping groups that already have exercises\n", messageUtils.InfoMsg("Skipping groups that already have exercises"))
	}

	store, err := db.Init()
	if err != nil {
		return templateProjectID, 0, fmt.Errorf("failed to init db: %w", err)
	}
	ctx := context.Background()
	tx, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return templateProjectID, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				fmt.Printf("failed to rollback transaction: %v", rollbackErr)
			}
		}
		_ = tx.Commit()
	}()

	qtx := store.WithTx(tx)

	successCount := 0
	for _, group := range classGroups {
		groupName := group.Name
		groupID := group.UserGroupID.String()
		if slices.Contains(existingExercises, groupName) {
			continue
		}

		groupNumber := extractGroupNumber(groupName, className)
		projectName := generateProjectName(format, className, exerciseName, groupNumber)

		var projectID string
		if len(exportData) > 0 {
			projectID, err = importProjectArchive(cfg, exportData, projectName)
			if err != nil {
				fmt.Printf("%v Failed to import template for %s on %s: %v\n",
					messageUtils.ErrorMsg("Failed to import template"),
					messageUtils.Bold(groupName), messageUtils.Highlight(cfg.Server), err)
				continue
			}
		} else {
			projectData := schemas.ProjectCreate{Name: &projectName}
			body, status, err := utils.CallClient(cfg, "createProject", []string{}, projectData)
			if err != nil || status != 201 {
				if err == nil {
					err = fmt.Errorf("status %d", status)
				}
				fmt.Printf("%v Failed to create project for %s on %s: %v\n",
					messageUtils.ErrorMsg("Failed to create project"),
					messageUtils.Bold(groupName), messageUtils.Highlight(cfg.Server), err)
				continue
			}
			var pr schemas.ProjectResponse
			if err := json.Unmarshal(body, &pr); err != nil {
				fmt.Printf("%v Failed to parse project response for %s: %v\n",
					messageUtils.ErrorMsg("Failed to parse project response"), messageUtils.Bold(groupName), err)
				continue
			}
			projectID = pr.ProjectID
		}
		if _, status, err := utils.CallClient(cfg, "closeProject", []string{projectID}, nil); err != nil || (status != 200 && status != 204) {
			if err == nil {
				err = fmt.Errorf("status %d", status)
			}
			fmt.Printf("%v Failed to close project %s on %s: %v\n",
				messageUtils.WarningMsg("Failed to close project"), messageUtils.Bold(projectName), messageUtils.Highlight(cfg.Server), err)
		}

		poolName := fmt.Sprintf("%s-pool", projectName)
		poolData := schemas.ResourcePoolCreate{Name: &poolName}
		poolBody, status, err := utils.CallClient(cfg, "createPool", []string{}, poolData)
		if err != nil || status != 201 {
			if err == nil {
				err = fmt.Errorf("status %d", status)
			}
			fmt.Printf("%v Failed to create pool for %s: %v\n",
				messageUtils.ErrorMsg("Failed to create resource pool"), messageUtils.Bold(projectName), err)
			continue
		}
		var poolResp schemas.ResourcePoolResponse
		if unmarshalErr := json.Unmarshal(poolBody, &poolResp); unmarshalErr != nil {
			fmt.Printf("%v Failed to parse pool response for %s: %v\n",
				messageUtils.ErrorMsg("Failed to parse pool response"), messageUtils.Bold(projectName), unmarshalErr)
			continue
		}
		poolID := poolResp.ResourcePoolID

		if _, status, err = utils.CallClient(cfg, "addToPool", []string{poolID, projectID}, nil); err != nil || (status != 201 && status != 204) {
			if err == nil {
				err = fmt.Errorf("status %d", status)
			}
			fmt.Printf("%v Failed to add project to pool for %s: %v\n",
				messageUtils.ErrorMsg("Failed to add project to pool"), messageUtils.Bold(projectName), err)
			continue
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
		if _, status, err = utils.CallClient(cfg, "createACL", []string{}, aclData); err != nil || (status != 201 && status != 204) {
			if err == nil {
				err = fmt.Errorf("status %d", status)
			}
			fmt.Printf("%v Failed to create ACL for %s: %v\n",
				messageUtils.ErrorMsg("Failed to create ACL"), messageUtils.Bold(projectName), err)
			continue
		}

		if store != nil {
			if nodes, nerr := qtx.GetNodes(ctx); nerr == nil {
				var clusterID int
				for _, n := range nodes {
					nodeURL := fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
					if nodeURL == cfg.Server {
						clusterID = int(n.ClusterID)
						break
					}
				}
				if clusterID != 0 {
					groupIDint, convertErr := strconv.Atoi(groupName)
					if convertErr != nil {
						return templateProjectID, 0, fmt.Errorf("failed to convert group name to int: %w", convertErr)
					}
					insertErr := qtx.InsertExerciseRecord(ctx, sqlc.InsertExerciseRecordParams{ProjectUuid: projectID, GroupID: int64(groupIDint), Name: exerciseName, State: sql.NullString{String: "created", Valid: true}})
					if insertErr != nil {
						rollbackErr := tx.Rollback()
						if rollbackErr != nil {
							fmt.Printf("failed to rollback transaction: %v", rollbackErr)
						}
						return templateProjectID, 0, fmt.Errorf("failed to insert exercise record: %w", insertErr)
					}
				}
			}
		}

		successCount++
		commitErr := tx.Commit()
		if commitErr != nil {
			return templateProjectID, 0, fmt.Errorf("failed to commit transaction: %w", commitErr)
		}
		fmt.Printf("%v Created project %s for group %s on %s\n",
			messageUtils.SuccessMsg("Created project"),
			messageUtils.Bold(projectName),
			messageUtils.Highlight(groupName),
			messageUtils.Highlight(cfg.Server),
		)
	}

	if templateProjectID != "" && deleteTemplate {
		if err := cleanupTemplateProject(cfg, templateProjectID, deleteTemplate); err != nil {
			fmt.Printf("%v Failed to clean up template project on %s: %v\n",
				messageUtils.WarningMsg("Failed to clean up template project"), messageUtils.Highlight(cfg.Server), err)
		} else {
			fmt.Printf("%v Cleaned up template project on %s\n",
				messageUtils.InfoMsg("Cleaned up template project"), messageUtils.Highlight(cfg.Server))
		}
	}

	return templateProjectID, successCount, nil
}

func extractGroupNumber(groupName, className string) string {
	prefix := className + "-"
	if after, ok := strings.CutPrefix(groupName, prefix); ok {
		return after
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

func getUserRoleID(cfg *config.GlobalOptions) (string, error) {
	rolesBody, status, err := utils.CallClient(cfg, "getRoles", []string{}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get roles: %w", err)
	}

	if status != 200 {
		return "", fmt.Errorf("failed to get roles: status %d", status)
	}

	var roles []schemas.RoleResponse
	if err := json.Unmarshal(rolesBody, &roles); err != nil {
		return "", fmt.Errorf("failed to parse roles: %w", err)
	}

	for _, role := range roles {
		if role.Name == "User" {
			return role.RoleID, nil
		}
	}

	return "", fmt.Errorf("user role not found")
}

func checkExistingExercises(cfg *config.GlobalOptions, className string, classGroups []schemas.UserGroupResponse) ([]string, error) {
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
	for i := range projects {
		project := &projects[i]
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

func selectTemplateWithFuzzy(cfg *config.GlobalOptions) (string, error) {
	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	if status != 200 {
		return "", fmt.Errorf("failed to get projects: status %d", status)
	}

	var projects []schemas.ProjectResponse
	if unmarshalErr := json.Unmarshal(projectsBody, &projects); unmarshalErr != nil {
		return "", fmt.Errorf("failed to parse projects: %w", unmarshalErr)
	}

	var projectNames []string
	projectMap := make(map[string]string)
	for i := range projects {
		project := &projects[i]
		projectNames = append(projectNames, project.Name)
		projectMap[project.Name] = project.ProjectID
	}

	if len(projectNames) == 0 {
		return "", fmt.Errorf("no projects found to select from")
	}

	fmt.Println("Select template project:")
	for i, name := range projectNames {
		fmt.Printf("%d. %s\n", i+1, name)
	}

	var selectedIndex int
	fmt.Print("Enter number: ")
	_, err = fmt.Scanln(&selectedIndex)
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	if selectedIndex < 1 || selectedIndex > len(projectNames) {
		return "", fmt.Errorf("invalid selection: %d", selectedIndex)
	}

	selected := projectNames[selectedIndex-1]

	return projectMap[selected], nil
}

func importTemplateProject(cfg *config.GlobalOptions, filePath, className, exerciseName string) (string, error) {
	file, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	projectName := fmt.Sprintf("%s-%s-template", className, exerciseName)
	_ = writer.WriteField("name", projectName)

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	resp, status, err := utils.CallClient(cfg, "postProject", []string{}, body)
	if err != nil {
		return "", fmt.Errorf("failed to import project: %w", err)
	}

	if status != 201 {
		return "", fmt.Errorf("failed to import project: status %d", status)
	}

	respBody := resp

	var project schemas.ProjectResponse
	if err := json.Unmarshal(respBody, &project); err != nil {
		return "", fmt.Errorf("failed to parse project response: %w", err)
	}

	return project.ProjectID, nil
}

func resolveTemplateProject(cfg *config.GlobalOptions, projectRef string) (string, error) {
	_, status, err := utils.CallClient(cfg, "getProject", []string{projectRef}, nil)
	if err == nil && status == 200 {
		return projectRef, nil
	}

	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	if status != 200 {
		return "", fmt.Errorf("failed to get projects: status %d", status)
	}

	var projects []schemas.ProjectResponse
	if unmarshalErr := json.Unmarshal(projectsBody, &projects); unmarshalErr != nil {
		return "", fmt.Errorf("failed to parse projects: %w", unmarshalErr)
	}

	for i := range projects {
		project := &projects[i]
		if project.Name == projectRef {
			return project.ProjectID, nil
		}
	}

	return "", fmt.Errorf("template project not found: %s", projectRef)
}

func cleanupTemplateProject(cfg *config.GlobalOptions, projectID string, deleteTemplate bool) error {
	_, status, err := utils.CallClient(cfg, "closeProject", []string{projectID}, nil)
	if err != nil {
		return fmt.Errorf("failed to close project: %w", err)
	}

	if status != 204 && status != 404 {
		return fmt.Errorf("failed to close project: status %d", status)
	}

	if deleteTemplate {
		_, status, err = utils.CallClient(cfg, "deleteProject", []string{projectID}, nil)
		if err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}

		if status != 204 && status != 404 {
			return fmt.Errorf("failed to delete project: status %d", status)
		}
	}

	return nil
}
