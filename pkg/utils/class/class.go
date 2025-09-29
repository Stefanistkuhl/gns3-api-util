package class

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

type NodeAndGroups struct {
	NodeID    int
	NumGroups int
}

var (
	ErrInsufficientCapacity = errors.New("not enough capacity for all groups")
	ErrExerciseNotFound     = errors.New("exercise not found")
	ErrClassNotFound        = errors.New("class not found")
)

func LoadClassFromFile(filePath string) (schemas.Class, error) {
	var classData schemas.Class

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return classData, fmt.Errorf("file does not exist: %s", messageUtils.Bold(filePath))
	}

	file, err := os.Open(filePath)
	if err != nil {
		return classData, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

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
		messageUtils.InfoMsgf("Loaded class %s with %d groups", classData.Name, len(classData.Groups)),
		messageUtils.Bold(classData.Name),
		len(classData.Groups))

	return classData, nil
}

func CreateClass(cfg config.GlobalOptions, clusterID int, classData schemas.Class, insertedNodes []db.NodeDataAll) (bool, error) {

	conn, err := db.InitIfNeeded()
	if err != nil {
		return false, fmt.Errorf("failed to init db: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	assignedPerNode := make(map[int]int)
	rows, qerr := conn.Query(`
SELECT ga.node_id, COUNT(*)
FROM group_assignments ga
JOIN nodes n ON n.node_id = ga.node_id
WHERE n.cluster_id = ?
GROUP BY ga.node_id;
`, clusterID)
	if qerr == nil {
		defer func() {
			if err := rows.Err(); err != nil {
				fmt.Printf("failed to iterate over rows: %v", err)
			}
		}()
		for rows.Next() {
			var nodeID, cnt int
			if err := rows.Scan(&nodeID, &cnt); err == nil {
				assignedPerNode[nodeID] = cnt
			}
		}
		_ = rows.Err()
	}

	nodesAdjusted := make([]db.NodeDataAll, 0, len(insertedNodes))
	availableGroups := 0
	for _, node := range insertedNodes {
		used := assignedPerNode[node.ID]
		remaining := max(0, node.MaxGroups-used)
		n := node
		n.MaxGroups = remaining
		nodesAdjusted = append(nodesAdjusted, n)
		availableGroups += remaining
	}
	totalGroups := len(classData.Groups)

	overCapacity := availableGroups < totalGroups
	allowOverAssign := false
	if overCapacity {
		warning := fmt.Sprintf(
			"not enough groups for class %s: available %d, needed %d. Allow assigning ABOVE MaxGroups based on weights?",
			messageUtils.Bold(classData.Name),
			availableGroups,
			totalGroups,
		)
		allowOverAssign = utils.ConfirmPrompt(warning, false)
		if !allowOverAssign {
			return false, fmt.Errorf("not enough groups available for class %s", messageUtils.Bold(classData.Name))
		}
	}

	getClassErr := db.CheckIfClassExists(conn, clusterID, classData.Name)
	if getClassErr != nil {
		if getClassErr == db.ErrClassExists {
			return false, fmt.Errorf("the class %s already exists", messageUtils.Bold(classData.Name))
		} else {
			return false, fmt.Errorf("failed to check if the class %s exists %w", messageUtils.Bold(classData.Name), getClassErr)
		}
	}

	insertedData, addDbErr := db.InsertClassIntoDB(conn, clusterID, classData)
	if addDbErr != nil {
		return false, fmt.Errorf("failed to add class to db: %w", addDbErr)
	}

	dist, err := distributeGroupsWithMode(
		func() []db.NodeDataAll {
			if allowOverAssign {
				return insertedNodes
			}
			return nodesAdjusted
		}(),
		classData,
		!allowOverAssign,
	)
	if err != nil {
		return false, fmt.Errorf("distribution failed: %w", err)
	}

	groupIDVals := make([]int, 0, len(insertedData.GroupIDs))
	for _, v := range insertedData.GroupIDs {
		groupIDVals = append(groupIDVals, v)
	}
	sort.Ints(groupIDVals)

	var assignments []db.AssignmentIds
	offset := 0
	for _, g := range dist {
		a := db.AssignmentIds{NodeID: g.NodeID}
		end := offset + g.NumGroups
		end = min(end, len(groupIDVals))
		a.GroupIDs = append(a.GroupIDs, groupIDVals[offset:end]...)
		assignments = append(assignments, a)
		offset = end
	}

	if err := db.AssignGroupsToNode(conn, db.NodeAndGroupIds{Assignments: assignments}); err != nil {
		return false, fmt.Errorf("failed to assign groups to nodes: %w", err)
	}

	plans, err := db.GetNodeGroupNamesForClass(conn, clusterID, classData.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get node group names: %w", err)
	}

	emailByGroupUser := make(map[string]map[string]string)
	for _, g := range classData.Groups {
		if _, ok := emailByGroupUser[g.Name]; !ok {
			emailByGroupUser[g.Name] = make(map[string]string)
		}
		for _, s := range g.Students {
			if s.Email != nil && *s.Email != "" {
				emailByGroupUser[g.Name][s.UserName] = *s.Email
			}
		}
	}
	for pi := range plans {
		for gi := range plans[pi].Groups {
			grpName := plans[pi].Groups[gi].Name
			for si := range plans[pi].Groups[gi].Students {
				u := &plans[pi].Groups[gi].Students[si]
				if u.Email == "" {
					if em, ok := emailByGroupUser[grpName][u.Username]; ok {
						u.Email = em
					}
				}
			}
		}
	}

	if err := runPlans(cfg, classData, plans); err != nil {
		return false, fmt.Errorf("failed to run plans: %w", err)
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
	clusterID, err := getClusterIDForServer(cfg)

	var (
		dbDeleted   bool
		nodeServers []string
	)

	if err == nil && clusterID != 0 {
		dbConn, dbErr := db.InitIfNeeded()
		if dbErr == nil {
			defer func() {
				if err := dbConn.Close(); err != nil {
					fmt.Printf("failed to close database connection: %v", err)
				}
			}()

			if dbErr = deleteClassFromDB(clusterID, className); dbErr == nil {
				dbDeleted = true
				fmt.Printf("%v Deleted class %v from database\n",
					messageUtils.SuccessMsg("Success"),
					messageUtils.Bold(className))
			} else {
				fmt.Printf("%v Failed to delete class %v from database: %v\n",
					messageUtils.WarningMsg("Warning"),
					messageUtils.Bold(className),
					dbErr)
			}

			nodes, nerr := db.GetNodes(dbConn)
			if nerr != nil {
				fmt.Printf("%v Failed to get nodes for class deletion: %v\n",
					messageUtils.WarningMsg("Warning"),
					nerr)
			} else {
				nodeServers = make([]string, 0, len(nodes))
				for _, n := range nodes {
					if n.ClusterID == clusterID {
						nodeServers = append(nodeServers, fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port))
					}
				}
			}
		}
	}

	if len(nodeServers) > 0 {
		var wg sync.WaitGroup
		errCh := make(chan error, len(nodeServers))

		for _, srv := range nodeServers {
			wg.Add(1)
			go func(server string) {
				defer wg.Done()
				nodeCfg := cfg
				nodeCfg.Server = server
				if err := deleteClassFromAPI(nodeCfg, className); err != nil {
					if errors.Is(err, ErrClassNotFound) {
						fmt.Printf("%v Class %v not present on %s; skipping.\n",
							messageUtils.WarningMsg("Warning"),
							messageUtils.Bold(className),
							server,
						)
						return
					}
					errCh <- fmt.Errorf("%s: %w", server, err)
				}
			}(srv)
		}

		wg.Wait()
		close(errCh)

		var apiErrors []error
		for e := range errCh {
			if e != nil {
				apiErrors = append(apiErrors, e)
			}
		}

		if dbDeleted && len(apiErrors) == 0 {
			return nil
		}

		if len(apiErrors) > 0 {
			return fmt.Errorf("failed to delete from some nodes: %v", apiErrors)
		}

		return nil
	}

	if !dbDeleted {
		fmt.Printf("%v Falling back to direct API deletion for class %v\n",
			messageUtils.InfoMsg("Info"),
			messageUtils.Bold(className))
	}

	if err := deleteClassFromAPI(cfg, className); err != nil {
		if errors.Is(err, ErrClassNotFound) {
			fmt.Printf("%v Class %v not present on %s; skipping.\n",
				messageUtils.WarningMsg("Warning"),
				messageUtils.Bold(className),
				cfg.Server,
			)
			return nil
		}
		return err
	}

	return nil
}

func deleteClassFromAPI(cfg config.GlobalOptions, className string) error {
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
		return fmt.Errorf("%w: %s", ErrClassNotFound, className)
	}

	fmt.Printf("%v Found %d class groups and %d student groups for class %v\n",
		messageUtils.InfoMsg("Found groups for class"),
		len(classGroups),
		len(studentGroups),
		messageUtils.Bold(className))

	allGroupsToDelete := append(studentGroups, classGroups...)

	allUsersToDelete := make(map[string]string)

	for _, group := range allGroupsToDelete {
		groupID := group.UserGroupID.String()
		groupName := group.Name

		members, err := getGroupMembers(cfg, groupID)
		if err != nil {
			fmt.Printf("%v failed to get members for group %v: %v\n",
				messageUtils.WarningMsgf("failed to get members for group %s", groupName),
				messageUtils.Bold(groupName),
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
			fmt.Printf("%v failed to delete user %v: %v\n",
				messageUtils.WarningMsgf("failed to delete user %s", username),
				messageUtils.Bold(username),
				err)
		} else {
			fmt.Printf("%v Deleted user %v\n",
				messageUtils.SuccessMsg("Deleted user"),
				messageUtils.Bold(username))
		}
	}

	for _, group := range studentGroups {
		groupID := group.UserGroupID.String()
		groupName := group.Name

		if err := deleteGroup(cfg, groupID); err != nil {
			fmt.Printf("%v failed to delete student group %v: %v\n",
				messageUtils.WarningMsgf("failed to delete student group %s", groupName),
				messageUtils.Bold(groupName),
				err)
		} else {
			fmt.Printf("%v Deleted student group %v\n",
				messageUtils.SuccessMsg("Deleted student group"),
				messageUtils.Bold(groupName))
		}
	}

	for _, group := range classGroups {
		groupID := group.UserGroupID.String()
		groupName := group.Name

		if err := deleteGroup(cfg, groupID); err != nil {
			fmt.Printf("%v failed to delete class group %v: %v\n",
				messageUtils.WarningMsgf("failed to delete class group %s", groupName),
				messageUtils.Bold(groupName),
				err)
		} else {
			fmt.Printf("%v Deleted class group %v\n",
				messageUtils.SuccessMsg("Deleted class group"),
				messageUtils.Bold(groupName))
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

func DeleteExercise(cfg config.GlobalOptions, exerciseName, className, groupName string) error {
	dbConn, err := db.InitIfNeeded()
	if err != nil {
		fmt.Printf("%v Failed to initialize database: %v\n",
			messageUtils.WarningMsg("Warning"),
			err)
	} else {
		defer func() {
			if err := dbConn.Close(); err != nil {
				fmt.Printf("failed to close database connection: %v", err)
			}
		}()
	}

	projects, err := getProjectsForExercise(cfg, exerciseName, className, groupName)
	if err != nil {
		return fmt.Errorf("failed to get projects for exercise: %w", err)
	}

	if len(projects) == 0 {
		if dbConn != nil {
			_, err := dbConn.Exec(`
				DELETE FROM exercises 
				WHERE name = ?`,
				exerciseName)
			if err != nil {
				fmt.Printf("%v Failed to clean up database entries: %v\n",
					messageUtils.WarningMsg("Warning"),
					err)
			}
		}
		return fmt.Errorf("%w: %s", ErrExerciseNotFound, exerciseName)
	}

	fmt.Printf("%v Found %d projects for exercise %v\n",
		messageUtils.InfoMsg("Found projects for exercise"),
		len(projects),
		messageUtils.Bold(exerciseName))

	for _, project := range projects {
		projectID := project.ProjectID
		projectName := project.Name

		parts := strings.Split(projectName, "-")
		resolvedClass := className
		resolvedExercise := exerciseName
		if resolvedClass == "" && len(parts) > 0 {
			resolvedClass = parts[0]
		}
		if resolvedExercise == "" && len(parts) > 1 {
			resolvedExercise = parts[1]
		}

		if err := closeProject(cfg, projectID); err != nil {
			fmt.Printf("%v Failed to close project %s: %v\n",
				messageUtils.WarningMsg("Warning"),
				projectName,
				err)
		}

		pools, err := getPoolsForProject(cfg, projectID, projectName, resolvedClass, resolvedExercise)
		if err != nil {
			fmt.Printf("%v Failed to get pools for exercise %s: %v\n",
				messageUtils.WarningMsg("Warning"),
				exerciseName,
				err)
		} else {
			for _, pool := range pools {
				aclsToDelete, collectErr := listACLsForPool(cfg, pool.ResourcePoolID, pool.Name)
				if collectErr != nil {
					fmt.Printf("%v Failed to enumerate ACLs for pool %s: %v\n",
						messageUtils.WarningMsg("Warning"),
						pool.Name,
						collectErr)
					continue
				}

				if len(aclsToDelete) == 0 {
					fmt.Printf("%v No ACL entries matched pool %s; pool deletion skipped.\n",
						messageUtils.InfoMsg("Info"),
						messageUtils.Bold(pool.Name))
					continue
				}

				fmt.Printf("%v Found %d ACL entries for pool %s; deleting...\n",
					messageUtils.InfoMsg("Info"),
					len(aclsToDelete),
					messageUtils.Bold(pool.Name))

				for _, aclID := range aclsToDelete {
					if err := deleteACL(cfg, aclID); err != nil {
						fmt.Printf("%v Failed to delete ACL %s for pool %s: %v\n",
							messageUtils.WarningMsg("Warning"),
							aclID,
							pool.Name,
							err)
					}
				}

				if len(aclsToDelete) > 0 {
					fmt.Printf("%v Deleted %d ACL entry(ies) for pool %s\n",
						messageUtils.SuccessMsg("Deleted ACLs"),
						len(aclsToDelete),
						messageUtils.Bold(pool.Name))
				}

				if err := deletePool(cfg, pool.ResourcePoolID); err != nil {
					fmt.Printf("%v Failed to delete pool %s: %v\n",
						messageUtils.WarningMsg("Warning"),
						pool.Name,
						err)
				} else {
					fmt.Printf("%v Deleted pool %s\n",
						messageUtils.SuccessMsg("Success"),
						messageUtils.Bold(pool.Name))
				}
			}
		}

		if err := deleteProject(cfg, projectID); err != nil {
			fmt.Printf("%v Failed to delete project %s: %v\n",
				messageUtils.WarningMsg("Warning"),
				projectName,
				err)
		} else {
			fmt.Printf("%v Deleted project %s\n",
				messageUtils.SuccessMsg("Success"),
				messageUtils.Bold(projectName))
		}

		if dbConn != nil {
			_, err := dbConn.Exec(`
				DELETE FROM exercises 
				WHERE project_uuid = ?`,
				projectID)
			if err != nil {
				fmt.Printf("%v Failed to delete database entry for project %s: %v\n",
					messageUtils.WarningMsg("Warning"),
					projectID,
					err)
			}
		}
	}

	if dbConn != nil {
		_, err := dbConn.Exec(`
			DELETE FROM exercises 
			WHERE name = ?`,
			exerciseName)
		if err != nil {
			fmt.Printf("%v Failed to clean up exercise database entries: %v\n",
				messageUtils.WarningMsg("Warning"),
				err)
		}
	}

	fmt.Printf("%v Successfully deleted exercise %s\n",
		messageUtils.SuccessMsg("Success"),
		messageUtils.Bold(exerciseName))

	return nil
}

func getProjectsForExercise(cfg config.GlobalOptions, exerciseName, className, groupName string) ([]schemas.ProjectResponse, error) {
	if exerciseName == "" {
		return nil, fmt.Errorf("exercise name cannot be empty")
	}

	dbConn, err := db.InitIfNeeded()
	if err != nil {
		fmt.Printf("%v Failed to initialize database (will try API): %v\n",
			messageUtils.WarningMsg("Warning"), err)
	} else {
		defer func() {
			if err := dbConn.Close(); err != nil {
				fmt.Printf("failed to close database connection: %v", err)
			}
		}()

		query := `
			SELECT e.project_uuid, e.name, c.name as class_name, g.name as group_name
			FROM exercises e
			JOIN groups g ON e.group_id = g.group_id
			JOIN classes c ON g.class_id = c.class_id
			WHERE e.name = ?
		`
		args := []any{exerciseName}

		if className != "" {
			query += " AND c.name = ?"
			args = append(args, className)
		}

		if groupName != "" {
			query += " AND g.name = ?"
			args = append(args, groupName)
		}

		rows, err := dbConn.Query(query, args...)
		if err != nil {
			fmt.Printf("%v Database query failed (will try API): %v\n",
				messageUtils.WarningMsg("Warning"), err)
		} else {
			defer func() {
				if err := rows.Err(); err != nil {
					fmt.Printf("failed to iterate over rows: %v", err)
				}
			}()

			validProjects := make(map[string]struct{})
			var scanErrors []error

			for rows.Next() {
				var projectUUID, name, dbClassName, dbGroupName string
				if err := rows.Scan(&projectUUID, &name, &dbClassName, &dbGroupName); err != nil {
					scanErrors = append(scanErrors, fmt.Errorf("failed to scan project: %w", err))
					continue
				}

				validProjects[projectUUID] = struct{}{}
			}

			for _, scanErr := range scanErrors {
				fmt.Printf("%v %v\n", messageUtils.WarningMsg("Warning"), scanErr)
			}

			if len(validProjects) > 0 {
				projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to get projects from API: %w", err)
				}
				if status != 200 {
					return nil, fmt.Errorf("unexpected status code %d when getting projects", status)
				}

				var allProjects []schemas.ProjectResponse
				if err := json.Unmarshal(projectsBody, &allProjects); err != nil {
					return nil, fmt.Errorf("failed to parse projects response: %w", err)
				}

				var matchingProjects []schemas.ProjectResponse
				for _, project := range allProjects {
					for shortUUID := range validProjects {
						if strings.Contains(project.Name, shortUUID) {
							matchingProjects = append(matchingProjects, project)
							break
						}
					}
				}

				if len(matchingProjects) > 0 {
					fmt.Printf("%v Found %d projects in database for exercise %s\n",
						messageUtils.InfoMsg("Info"),
						len(matchingProjects), messageUtils.Bold(exerciseName))
					return matchingProjects, nil
				}
			}
		}
	}

	fmt.Printf("%v Querying API for projects matching %s-%s-*\n",
		messageUtils.InfoMsg("Info"),
		messageUtils.Bold(className),
		messageUtils.Bold(exerciseName))

	projectsBody, status, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects from API: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("unexpected status code %d when getting projects", status)
	}

	var allProjects []schemas.ProjectResponse
	if err := json.Unmarshal(projectsBody, &allProjects); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	var matchingProjects []schemas.ProjectResponse

	for _, project := range allProjects {
		parts := strings.Split(project.Name, "-")
		if len(parts) < 2 {
			continue
		}

		if className != "" && (len(parts) < 2 || parts[0] != className) {
			continue
		}

		if parts[1] != exerciseName {
			continue
		}

		if groupName != "" && (len(parts) < 3 || parts[2] != groupName) {
			continue
		}

		matchingProjects = append(matchingProjects, project)
	}

	fmt.Printf("%v Found %d matching projects in API\n",
		messageUtils.InfoMsg("Info"),
		len(matchingProjects))

	return matchingProjects, nil
}

func DeleteAllExercisesForClass(cfg config.GlobalOptions, className string) error {
	dbConn, err := db.InitIfNeeded()
	if err != nil {
		fmt.Printf("%v Failed to initialize database: %v\n",
			messageUtils.WarningMsg("Warning"),
			err)
	} else {
		defer func() {
			if err := dbConn.Close(); err != nil {
				fmt.Printf("failed to close database connection: %v", err)
			}
		}()
	}

	clusterID, err := getClusterIDForServer(cfg)
	var nodeServers []string
	if err == nil && clusterID != 0 {
		nodes, nerr := db.GetNodes(dbConn)
		if nerr == nil {
			for _, n := range nodes {
				if n.ClusterID == clusterID {
					nodeServers = append(nodeServers, fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port))
				}
			}
		}
	}

	if len(nodeServers) == 0 {
		return deleteAllExercisesForClassOnNode(cfg, className)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(nodeServers))
	for _, srv := range nodeServers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			nodeCfg := cfg
			nodeCfg.Server = server
			if e := deleteAllExercisesForClassOnNode(nodeCfg, className); e != nil {
				errCh <- fmt.Errorf("%s: %w", server, e)
			}
		}(srv)
	}

	if dbConn != nil {
		if _, err := dbConn.Exec(`
			DELETE FROM exercises 
			WHERE class = ?`,
			className); err != nil {
			errCh <- fmt.Errorf("failed to delete exercises for class %s from database: %w", className, err)
		} else {
			fmt.Printf("%v Deleted database records for exercises in class %v\n",
				messageUtils.SuccessMsg("Deleted database records"),
				messageUtils.Bold(className))
		}
	}

	wg.Wait()
	close(errCh)
	for e := range errCh {
		if e != nil {
			return e
		}
	}
	return nil
}

func deleteAllExercisesForClassOnNode(cfg config.GlobalOptions, className string) error {
	dbConn, err := db.InitIfNeeded()
	if err == nil {
		defer func() {
			if err := dbConn.Close(); err != nil {
				fmt.Printf("failed to close database connection: %v", err)
			}
		}()
		_, err = dbConn.Exec(`
			DELETE FROM exercises 
			WHERE class = ?`,
			className)
		if err != nil {
			fmt.Printf("%v Failed to delete exercises for class %s from database: %v\n",
				messageUtils.WarningMsg("Warning"),
				className, err)
		} else {
			fmt.Printf("%v Deleted database records for exercises in class %v\n",
				messageUtils.SuccessMsg("Success"),
				messageUtils.Bold(className))
		}
	} else {
		fmt.Printf("%v Failed to initialize database: %v\n",
			messageUtils.WarningMsg("Warning"),
			err)
	}

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
		if len(parts) >= 2 && parts[0] == className {
			exerciseName := parts[1]
			if !seenExercises[exerciseName] {
				classExercises = append(classExercises, exerciseName)
				seenExercises[exerciseName] = true
			}
		}
	}

	if len(classExercises) == 0 {
		fmt.Printf("%v No exercises found for class %v on %s\n",
			messageUtils.InfoMsg("No exercises found for class"),
			messageUtils.Bold(className), cfg.Server)
		return nil
	}

	fmt.Printf("%v Found %d exercises for class %v on %s\n",
		messageUtils.InfoMsgf("Found %d exercises for class", len(classExercises)),
		len(classExercises),
		messageUtils.Bold(className), cfg.Server)

	deleted := 0
	errorCount := 0
	for _, exerciseName := range classExercises {
		err := DeleteExercise(cfg, exerciseName, className, "")
		if err != nil {
			if errors.Is(err, ErrExerciseNotFound) {
				fmt.Printf("%v Exercise %v not present on %s; skipping.\n",
					messageUtils.WarningMsg("Warning"),
					messageUtils.Bold(exerciseName),
					cfg.Server,
				)
				continue
			}
			fmt.Printf("%v Failed to delete exercise %v on %s: %v\n",
				messageUtils.ErrorMsg("Error"),
				messageUtils.Bold(exerciseName),
				cfg.Server,
				err)
			errorCount++
		} else {
			deleted++
		}
	}

	if deleted > 0 {
		fmt.Printf("%v Successfully deleted %d/%d exercises for class %s\n",
			messageUtils.SuccessMsg("Success"),
			deleted,
			len(classExercises),
			messageUtils.Bold(className))
	}

	if errorCount > 0 {
		return fmt.Errorf("encountered %d errors while deleting exercises", errorCount)
	}

	return nil
}

func closeProject(cfg config.GlobalOptions, projectID string) error {
	_, status, err := utils.CallClient(cfg, "closeProject", []string{projectID}, nil)
	if err != nil {
		if strings.Contains(err.Error(), "UUID") || strings.Contains(err.Error(), "uuid_parsing") {
			projectsBody, _, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
			if err != nil {
				return fmt.Errorf("failed to get projects: %w", err)
			}

			var projects []schemas.ProjectResponse
			if err := json.Unmarshal(projectsBody, &projects); err != nil {
				return fmt.Errorf("failed to parse projects response: %w", err)
			}

			for _, p := range projects {
				if p.Name == projectID {
					_, status, err = utils.CallClient(cfg, "closeProject", []string{p.ProjectID}, nil)
					if err != nil && status != 404 {
						return fmt.Errorf("failed to close project %s: %w", p.ProjectID, err)
					}
					return nil
				}
			}
			return fmt.Errorf("project with name '%s' not found", projectID)
		}
		return fmt.Errorf("failed to close project: %w", err)
	}

	if status != 200 && status != 204 {
		return fmt.Errorf("failed to close project: status %d", status)
	}
	return nil
}

func getPoolsForProject(cfg config.GlobalOptions, projectID, projectName, className, exerciseName string) ([]schemas.ResourcePoolResponse, error) {
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
		name := strings.TrimSpace(pool.Name)
		if name == "" {
			continue
		}

		if projectName != "" && strings.HasPrefix(name, projectName) && strings.HasSuffix(name, "-pool") {
			matchingPools = append(matchingPools, pool)
			continue
		}

		contains, cerr := poolContainsProject(cfg, pool.ResourcePoolID, projectID)
		if cerr != nil {
			fmt.Printf("%v Failed to inspect pool %s for project membership: %v\n",
				messageUtils.WarningMsg("Warning"),
				messageUtils.Bold(pool.Name),
				cerr)
		} else if contains {
			matchingPools = append(matchingPools, pool)
			continue
		}

		if className != "" && exerciseName != "" && strings.HasPrefix(name, fmt.Sprintf("%s-%s-", className, exerciseName)) && strings.HasSuffix(name, "-pool") {
			matchingPools = append(matchingPools, pool)
			continue
		}
		if exerciseName != "" && strings.Contains(name, "-"+exerciseName+"-") && strings.HasSuffix(name, "-pool") {
			matchingPools = append(matchingPools, pool)
		}
	}

	return matchingPools, nil
}

func poolContainsProject(cfg config.GlobalOptions, poolID, projectID string) (bool, error) {
	body, status, err := utils.CallClient(cfg, "getPoolResources", []string{poolID}, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get pool resources: %w", err)
	}
	if status != 200 {
		return false, fmt.Errorf("failed to get pool resources: status %d", status)
	}

	var resources []struct {
		ResourceID string `json:"resource_id"`
	}
	if err := json.Unmarshal(body, &resources); err != nil {
		return false, fmt.Errorf("failed to parse pool resources response: %w", err)
	}

	for _, res := range resources {
		if res.ResourceID == projectID {
			return true, nil
		}
	}

	return false, nil
}

func listACLsForPool(cfg config.GlobalOptions, poolID, poolName string) ([]string, error) {
	aclsBody, status, err := utils.CallClient(cfg, "getAcl", []string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get ACLs: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("failed to get ACLs: status %d", status)
	}

	var acls []schemas.ACLResponse
	if err := json.Unmarshal(aclsBody, &acls); err != nil {
		return nil, fmt.Errorf("failed to parse ACLs response: %w", err)
	}

	poolPath := fmt.Sprintf("/pools/%s", poolID)
	poolPathLower := strings.ToLower(poolPath)
	trimmedName := strings.TrimSpace(poolName)
	trimmedNameLower := strings.ToLower(trimmedName)
	poolIDLower := strings.ToLower(poolID)

	var matched []string

	for _, acl := range acls {
		path := strings.TrimSpace(acl.Path)
		if path == "" {
			continue
		}
		pathLower := strings.ToLower(path)

		match := path == poolPath || path == strings.TrimPrefix(poolPath, "/")
		if !match {
			if pathLower == poolPathLower || pathLower == strings.TrimPrefix(poolPathLower, "/") {
				match = true
			}
		}
		if !match && trimmedName != "" {
			if strings.Contains(pathLower, poolIDLower) || strings.Contains(pathLower, trimmedNameLower) || strings.Contains(pathLower, fmt.Sprintf("\"%s\"", trimmedNameLower)) {
				match = true
			}
		}
		if !match && trimmedName != "" {
			resourceLabel := fmt.Sprintf("resource pool \"%s\"", trimmedNameLower)
			if strings.Contains(pathLower, resourceLabel) {
				match = true
			}
		}
		fmt.Printf("%v Inspecting ACL %s with path %q for pool %s (%s); match=%v\n",
			messageUtils.InfoMsg("ACL"),
			messageUtils.Bold(acl.ACLID),
			path,
			messageUtils.Bold(poolName),
			poolID,
			match)
		if match {
			id := strings.TrimSpace(acl.ACLID)
			if id != "" {
				matched = append(matched, id)
			}
		}
	}

	if len(matched) == 0 && trimmedName != "" {
		for _, acl := range acls {
			id := strings.TrimSpace(acl.ACLID)
			if id == "" {
				continue
			}
			if strings.Contains(strings.ToLower(id), trimmedNameLower) {
				matched = append(matched, id)
			}
		}
	}

	if len(matched) == 0 {
		fmt.Printf("%v No ACL entries matched pool %s (%s) after inspecting %d entries.\n",
			messageUtils.WarningMsg("Warning"),
			messageUtils.Bold(poolName),
			poolID,
			len(acls))
	}

	return matched, nil
}

func deletePool(cfg config.GlobalOptions, poolID string) error {
	_, status, err := utils.CallClient(cfg, "deletePool", []string{poolID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete pool: %w", err)
	}
	if status != 204 && status != 404 {
		return fmt.Errorf("unexpected status code: %d", status)
	}
	return nil
}

func deleteGroup(cfg config.GlobalOptions, groupID string) error {
	_, status, err := utils.CallClient(cfg, "deleteGroup", []string{groupID}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	if status != 204 && status != 404 {
		return fmt.Errorf("unexpected status code: %d", status)
	}
	return nil
}

func deleteProject(cfg config.GlobalOptions, projectID string) error {
	_, status, err := utils.CallClient(cfg, "closeProject", []string{projectID}, nil)
	if err != nil && status != 404 {
		return fmt.Errorf("failed to close project: %w", err)
	}

	_, status, err = utils.CallClient(cfg, "deleteProject", []string{projectID}, nil)
	if err != nil {
		if strings.Contains(err.Error(), "UUID") || strings.Contains(err.Error(), "uuid_parsing") {
			projectsBody, _, err := utils.CallClient(cfg, "getProjects", []string{}, nil)
			if err != nil {
				return fmt.Errorf("failed to get projects: %w", err)
			}

			var projects []schemas.ProjectResponse
			if err := json.Unmarshal(projectsBody, &projects); err != nil {
				return fmt.Errorf("failed to parse projects response: %w", err)
			}

			for _, p := range projects {
				if p.Name == projectID {
					_, status, err = utils.CallClient(cfg, "closeProject", []string{p.ProjectID}, nil)
					if err != nil && status != 404 {
						return fmt.Errorf("failed to close project %s: %w", p.ProjectID, err)
					}

					_, _, err = utils.CallClient(cfg, "deleteProject", []string{p.ProjectID}, nil)
					if err != nil {
						return fmt.Errorf("failed to delete project %s: %w", p.ProjectID, err)
					}
					return nil
				}
			}
			return fmt.Errorf("project with name '%s' not found", projectID)
		}
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if status != 200 && status != 204 {
		return fmt.Errorf("failed to delete project: status %d", status)
	}
	return nil
}

func deleteACL(cfg config.GlobalOptions, aclID string) error {
	cleanID := strings.TrimSpace(aclID)
	if cleanID == "" {
		return fmt.Errorf("empty ACL id")
	}

	variants := []string{cleanID, url.PathEscape(cleanID)}

	for _, id := range variants {
		for _, suffix := range []string{"", "/"} {
			candidate := strings.TrimSuffix(id, "/") + suffix

			_, status, err := utils.CallClient(cfg, "deleteACE", []string{candidate}, nil)
			if err != nil {
				lastErr := fmt.Errorf("failed to delete ACL %s: %w", candidate, err)
				if suffix == "/" || id == variants[len(variants)-1] {
					return lastErr
				}
				continue
			}

			switch status {
			case 200, 204, 404:
				return nil
			case 405:
				if suffix == "/" || id == variants[len(variants)-1] {
					return fmt.Errorf("failed to delete ACL %s: status %d", candidate, status)
				}
			default:
				return fmt.Errorf("failed to delete ACL %s: status %d", candidate, status)
			}
		}
	}

	return fmt.Errorf("failed to delete ACL %s", cleanID)
}

func getClusterIDForServer(cfg config.GlobalOptions) (int, error) {
	if cfg.Server == "" {
		return 0, fmt.Errorf("no server configured")
	}

	urlObj := utils.ValidateUrlWithReturn(cfg.Server)
	clusterName := fmt.Sprintf("%s%s", urlObj.Hostname(), "_single_node_cluster")

	conn, err := db.InitIfNeeded()
	if err != nil {
		return 0, fmt.Errorf("failed to init db: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}()

	clusters, err := db.GetClusters(conn)
	if err != nil {
		return 0, fmt.Errorf("failed to get clusters: %w", err)
	}

	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			return cluster.Id, nil
		}
	}

	return 0, fmt.Errorf("cluster not found")
}

func deleteClassFromDB(clusterID int, className string) error {
	conn, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}()

	return db.DeleteClassFromDB(conn, clusterID, className)
}

func distributeGroupsWithMode(nodes []db.NodeDataAll, classData schemas.Class, respectCaps bool) ([]NodeAndGroups, error) {
	totalGroups := len(classData.Groups)

	totalWeight := 0
	for _, n := range nodes {
		w := max(n.Weight, 0)
		totalWeight += w
	}

	type share struct {
		nodeID    int
		max       int
		weight    int
		assigned  int
		remainder int
	}

	shares := make([]share, len(nodes))

	if respectCaps {
		totalCap := 0
		for _, n := range nodes {
			totalCap += n.MaxGroups
		}
		if totalGroups > totalCap {
			return nil, fmt.Errorf("not enough capacity for all groups")
		}
	}

	if totalWeight == 0 {
		per := 0
		if len(nodes) > 0 {
			per = totalGroups / len(nodes)
		}
		assignedTotal := 0
		for i, n := range nodes {
			assign := per
			if respectCaps {
				assign = min(assign, n.MaxGroups)
			}
			shares[i] = share{
				nodeID:    n.ID,
				max:       n.MaxGroups,
				weight:    0,
				assigned:  assign,
				remainder: 0,
			}
			assignedTotal += assign
		}
		remaining := totalGroups - assignedTotal
		for remaining > 0 {
			progress := false
			for i := range shares {
				if remaining == 0 {
					break
				}
				if !respectCaps || shares[i].assigned < shares[i].max {
					shares[i].assigned++
					remaining--
					progress = true
				}
			}
			if !progress {
				break
			}
		}
	} else {
		assignedTotal := 0
		for i, n := range nodes {
			w := n.Weight
			w = max(w, 0)
			num := w * totalGroups
			base := num / totalWeight
			rem := num % totalWeight
			if respectCaps && base > n.MaxGroups {
				base = n.MaxGroups
			}
			shares[i] = share{
				nodeID:    n.ID,
				max:       n.MaxGroups,
				weight:    w,
				assigned:  base,
				remainder: rem,
			}
			assignedTotal += base
		}

		remaining := totalGroups - assignedTotal
		for remaining > 0 {
			sort.SliceStable(shares, func(i, j int) bool {
				if shares[i].remainder != shares[j].remainder {
					return shares[i].remainder > shares[j].remainder
				}
				freeI := shares[i].max - shares[i].assigned
				freeJ := shares[j].max - shares[j].assigned
				if !respectCaps {
					freeI, freeJ = 1<<30, 1<<30
				}
				if freeI != freeJ {
					return freeI > freeJ
				}
				return shares[i].nodeID < shares[j].nodeID
			})
			progress := false
			for k := range shares {
				if remaining == 0 {
					break
				}
				if !respectCaps || shares[k].assigned < shares[k].max {
					shares[k].assigned++
					remaining--
					progress = true
					break
				}
			}
			if !progress {
				break
			}
		}
	}

	result := make([]NodeAndGroups, 0, len(shares))
	for _, s := range shares {
		result = append(result, NodeAndGroups{
			NodeID:    s.nodeID,
			NumGroups: s.assigned,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].NodeID < result[j].NodeID })
	return result, nil
}

func runPlans(cfg config.GlobalOptions, classData schemas.Class, plans []db.NodeGroupsForClass) error {

	// 1. create class group on all nodes

	if len(plans) == 0 {
		return nil
	}

	classGroupData := schemas.UserGroupCreate{
		Name: &classData.Name,
	}

	var wg sync.WaitGroup

	errChan := make(chan error, len(plans))
	for _, plan := range plans {
		wg.Add(1)
		go func(plan db.NodeGroupsForClass) {
			defer wg.Done()
			nodeCfg := cfg
			nodeCfg.Server = plan.NodeURL

			classGroupBody, status, err := utils.CallClient(nodeCfg, "createGroup", []string{}, classGroupData)
			if err != nil {
				errChan <- fmt.Errorf("failed to create class group: %w", err)
				return
			}

			if status != 201 {
				errChan <- fmt.Errorf("failed to create class group: status %d", status)
				return
			}

			var classGroupResponse schemas.UserGroupResponse
			if err := json.Unmarshal(classGroupBody, &classGroupResponse); err != nil {
				errChan <- fmt.Errorf("failed to parse class group response: %w", err)
				return
			}

			classGroupName := classGroupResponse.Name

			fmt.Printf("%v Created class group %v\n",
				messageUtils.SuccessMsg("Created class group"),
				messageUtils.Bold(classGroupName))

		}(plan)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// 2. create student groups on the nodes

	var wg2 sync.WaitGroup

	errchan2 := make(chan error, len(plans))
	for _, plan := range plans {
		wg2.Add(1)
		go func(plan db.NodeGroupsForClass) {
			defer wg2.Done()
			nodeCfg := cfg
			nodeCfg.Server = plan.NodeURL
			for _, group := range plan.Groups {
				if group.Name == classData.Name {
					continue
				}
				sgData := schemas.UserGroupCreate{
					Name: &group.Name,
				}
				studentGroupBody, status, err := utils.CallClient(nodeCfg, "createGroup", []string{}, sgData)
				if err != nil {
					errchan2 <- fmt.Errorf("failed to create student group %s: %w", group.Name, err)
					return
				}

				if status != 201 {
					errchan2 <- fmt.Errorf("failed to create student group %s: status %d", group.Name, status)
					return
				}

				var studentGroupResponse schemas.UserGroupResponse
				if err := json.Unmarshal(studentGroupBody, &studentGroupResponse); err != nil {
					errchan2 <- fmt.Errorf("failed to parse student group response: %w", err)
					return
				}

				studentGroupName := studentGroupResponse.Name

				fmt.Printf("%v Created student group %v\n",
					messageUtils.SuccessMsg("Created student group"),
					messageUtils.Bold(studentGroupName))

			}
		}(plan)

	}
	wg2.Wait()
	close(errchan2)
	for err := range errchan2 {
		if err != nil {
			return err
		}
	}

	// 3. create users and add them to groups concurrently

	var wg3 sync.WaitGroup

	errchan3 := make(chan error, len(plans))
	for _, plan := range plans {
		wg3.Add(1)
		go func(plan db.NodeGroupsForClass) {
			defer wg3.Done()
			nodeCfg := cfg
			nodeCfg.Server = plan.NodeURL

			// Create users and add them to groups
			for _, group := range plan.Groups {
				for _, user := range group.Students {
					userData := schemas.UserCreate{
						Username: &user.Username,
						Password: &user.Password,
						Email:    &user.Email,
						FullName: &user.FullName,
					}

					userBody, status, err := utils.CallClient(nodeCfg, "createUser", []string{}, userData)
					if err != nil {
						errchan3 <- fmt.Errorf("failed to create user %s: %w", user.Username, err)
						return
					}

					if status != 201 {
						errchan3 <- fmt.Errorf("failed to create user %s: status %d", user.Username, status)
						return
					}

					var userResponse schemas.UserResponse
					if err := json.Unmarshal(userBody, &userResponse); err != nil {
						errchan3 <- fmt.Errorf("failed to parse user response: %w", err)
						return
					}

					userID := userResponse.UserID.String()
					username := userResponse.Username

					fmt.Printf("%v Created user %v\n",
						messageUtils.SuccessMsg("Created user"),
						messageUtils.Bold(username))

					// Get group IDs for class and student groups
					groupsBody, status, err := utils.CallClient(nodeCfg, "getGroups", []string{}, nil)
					if err != nil {
						errchan3 <- fmt.Errorf("failed to get groups: %w", err)
						return
					}
					if status != 200 {
						errchan3 <- fmt.Errorf("failed to get groups: status %d", status)
						return
					}

					var groups []schemas.UserGroupResponse
					if err := json.Unmarshal(groupsBody, &groups); err != nil {
						errchan3 <- fmt.Errorf("failed to parse groups response: %w", err)
						return
					}

					var classGroupID, studentGroupID string
					for _, group := range groups {
						switch group.Name {
						case classData.Name:
							classGroupID = group.UserGroupID.String()
						case user.GroupName:
							studentGroupID = group.UserGroupID.String()
						}
					}

					if err := addUserToGroup(nodeCfg, userID, classGroupID); err != nil {
						errchan3 <- fmt.Errorf("failed to add user %s to class group: %w", username, err)
						return
					}

					if err := addUserToGroup(nodeCfg, userID, studentGroupID); err != nil {
						errchan3 <- fmt.Errorf("failed to add user %s to student group: %w", username, err)
						return
					}
				}
			}
		}(plan)
	}
	wg3.Wait()
	close(errchan3)
	for err := range errchan3 {
		if err != nil {
			return err
		}
	}

	return nil
}
