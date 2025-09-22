package exercise

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/cluster/db"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/class"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewExerciseDeleteCmd() *cobra.Command {
	var deleteExerciseCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an exercise",
		Long: `Delete an exercise (project) and its associated resources. This command can:
- Delete a specific exercise by name
- Use fuzzy finder to select exercises interactively
- Delete all exercises for a class
- Delete exercises for specific class and group combinations`,
		Example: `
  # Delete an exercise interactively using fuzzy finder
  gns3util -s http://server:3080 exercise delete --select-exercise

  # Delete a specific exercise by name
  gns3util -s http://server:3080 exercise delete --name "MyExercise"

  # Delete an exercise non-interactively
  gns3util -s http://server:3080 exercise delete --non-interactive "MyExercise"

  # Delete all exercises for a class
  gns3util -s http://server:3080 exercise delete --class "MyClass"

  # Delete exercises for a specific class and group
  gns3util -s http://server:3080 exercise delete --class "MyClass" --group "Group1"

  # Delete all exercises (use with caution!)
  gns3util -s http://server:3080 exercise delete --all

  # Delete without confirmation
  gns3util -s http://server:3080 exercise delete --name "MyExercise" --no-confirm

  # Delete multiple exercises using fuzzy finder with multi-select
  gns3util -s http://server:3080 exercise delete --select-exercise --multi

  # Delete exercises for a specific class (interactive selection)
  gns3util -s http://server:3080 exercise delete --select-class

  # Delete exercises for a specific class and group (interactive selection)
  gns3util -s http://server:3080 exercise delete --select-class --select-group
		`,
		RunE: runDeleteExercise,
	}

	deleteExerciseCmd.Flags().String("name", "", "Name of the exercise to delete")
	deleteExerciseCmd.Flags().String("non-interactive", "", "Run the command non-interactively with specified exercise name")
	deleteExerciseCmd.Flags().String("class", "", "Class name for the exercise")
	deleteExerciseCmd.Flags().String("group", "", "Group name for the exercise")
	deleteExerciseCmd.Flags().Bool("select-exercise", false, "Select exercise interactively from a list")
	deleteExerciseCmd.Flags().Bool("select-class", false, "Select class interactively from a list")
	deleteExerciseCmd.Flags().Bool("select-group", false, "Select group interactively from a list")
	deleteExerciseCmd.Flags().Bool("all", false, "Delete all exercises (use with caution!)")
	deleteExerciseCmd.Flags().Bool("multi", false, "Enable multi-select mode for fuzzy finder (only for exercise selection)")
	deleteExerciseCmd.Flags().Bool("confirm", true, "Require confirmation before deletion")
	deleteExerciseCmd.Flags().Bool("no-confirm", false, "Skip confirmation prompt")
	deleteExerciseCmd.Flags().StringP("cluster", "c", "", "Cluster name")

	return deleteExerciseCmd
}

func deleteExerciseInCluster(cfg config.GlobalOptions, clusterName, exerciseName, className, groupName string, confirm bool) error {
	if confirm {
		msg := fmt.Sprintf("Delete exercise '%s' across cluster '%s'?", exerciseName, clusterName)
		if className != "" {
			msg = fmt.Sprintf("Delete exercise '%s' for class '%s' across cluster '%s'?", exerciseName, className, clusterName)
		}
		if groupName != "" {
			msg = fmt.Sprintf("Delete exercise '%s' for class '%s' group '%s' across cluster '%s'?", exerciseName, className, groupName, clusterName)
		}
		if !utils.ConfirmPrompt(msg, false) {
			fmt.Printf("Deletion of exercise %v cancelled\n", messageUtils.Bold(exerciseName))
			return nil
		}
	}

	conn, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer conn.Close()

	clusters, err := db.GetClusters(conn)
	if err != nil {
		return fmt.Errorf("failed to get clusters: %w", err)
	}
	clusterID := 0
	for _, c := range clusters {
		if strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(clusterName)) {
			clusterID = c.Id
			break
		}
	}
	if clusterID == 0 {
		return fmt.Errorf("cluster not found: %s", clusterName)
	}

	nodes, err := db.GetNodes(conn)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	for _, n := range nodes {
		if n.ClusterID != clusterID {
			continue
		}
		nodeCfg := cfg
		nodeCfg.Server = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
		if err := deleteExerciseWithConfirmation(nodeCfg, exerciseName, className, groupName, false); err != nil {
			fmt.Printf("%v Failed to delete exercise %v on %s: %v\n", messageUtils.ErrorMsg("Failed to delete exercise"), messageUtils.Bold(exerciseName), nodeCfg.Server, err)
		} else {
			fmt.Printf("%v Deleted exercise %v on %s\n", messageUtils.SuccessMsg("Deleted exercise"), messageUtils.Bold(exerciseName), nodeCfg.Server)
		}
	}
	return nil
}

func getAllExerciseNamesFromCluster(cfg config.GlobalOptions, clusterName string) ([]string, error) {
	conn, err := db.InitIfNeeded()
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}
	defer conn.Close()
	clusters, err := db.GetClusters(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters: %w", err)
	}
	clusterID := 0
	for _, c := range clusters {
		if strings.EqualFold(strings.TrimSpace(c.Name), strings.TrimSpace(clusterName)) {
			clusterID = c.Id
			break
		}
	}
	if clusterID == 0 {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}
	nodes, err := db.GetNodes(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	seen := make(map[string]bool)
	var out []string
	for _, n := range nodes {
		if n.ClusterID != clusterID {
			continue
		}
		nodeCfg := cfg
		nodeCfg.Server = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
		names, err := getAllExerciseNames(nodeCfg)
		if err != nil {
			continue
		}
		for _, nm := range names {
			if !seen[nm] {
				seen[nm] = true
				out = append(out, nm)
			}
		}
	}
	return out, nil
}

func selectExercisesWithFuzzyFromCluster(cfg config.GlobalOptions, clusterName string, multi bool) ([]string, error) {
	names, err := getAllExerciseNamesFromCluster(cfg, clusterName)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no exercises found")
	}
	title := fmt.Sprintf("Select exercises from cluster %s:", clusterName)
	if multi {
		title = fmt.Sprintf("Select exercises from cluster %s (multi-select):", clusterName)
	}
	finder := fuzzy.NewFuzzyFinderWithTitle(names, multi, title)
	return finder, nil
}

func runDeleteExercise(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}

	exerciseName, _ := cmd.Flags().GetString("name")
	nonInteractive, _ := cmd.Flags().GetString("non-interactive")
	className, _ := cmd.Flags().GetString("class")
	groupName, _ := cmd.Flags().GetString("group")
	selectExercise, _ := cmd.Flags().GetBool("select-exercise")
	selectClass, _ := cmd.Flags().GetBool("select-class")
	selectGroup, _ := cmd.Flags().GetBool("select-group")
	deleteAll, _ := cmd.Flags().GetBool("all")
	multi, _ := cmd.Flags().GetBool("multi")
	confirm, _ := cmd.Flags().GetBool("confirm")
	noConfirm, _ := cmd.Flags().GetBool("no-confirm")
	clusterName, _ := cmd.Flags().GetString("cluster")

	if noConfirm {
		confirm = false
	}

	// Validate flag combinations
	if exerciseName != "" && nonInteractive != "" {
		return fmt.Errorf("cannot specify both --name and --non-interactive")
	}
	if exerciseName != "" && deleteAll {
		return fmt.Errorf("cannot specify both --name and --all")
	}
	if nonInteractive != "" && deleteAll {
		return fmt.Errorf("cannot specify both --non-interactive and --all")
	}
	if nonInteractive != "" && (className != "" || groupName != "") {
		return fmt.Errorf("cannot specify both --non-interactive and class/group flags")
	}

	if selectClass || selectGroup {
		if className != "" || groupName != "" {
			return fmt.Errorf("cannot specify both selection flags and explicit class/group")
		}

		// Interactive class selection
		if selectClass {
			selectedClass, err := selectClassWithFuzzy(cfg)
			if err != nil {
				return fmt.Errorf("failed to select class: %w", err)
			}
			if selectedClass == "" {
				fmt.Printf("%s\n", colorUtils.Warning("No class selected"))
				return nil
			}
			className = selectedClass
		}

		// Interactive group selection (only if class is selected)
		if selectGroup {
			if className == "" {
				return fmt.Errorf("must select a class before selecting a group")
			}
			selectedGroup, err := selectGroupWithFuzzy(cfg, className)
			if err != nil {
				return fmt.Errorf("failed to select group: %w", err)
			}
			if selectedGroup == "" {
				fmt.Printf("%s\n", colorUtils.Warning("No group selected"))
				return nil
			}
			groupName = selectedGroup
		}
	}

	if selectExercise {
		if exerciseName != "" || nonInteractive != "" {
			return fmt.Errorf("cannot specify both selection and explicit exercise name")
		}

	}

	if nonInteractive != "" {
		exerciseName = nonInteractive
	}

	var targetExerciseName string
	if nonInteractive != "" {
		targetExerciseName = nonInteractive
	} else if exerciseName != "" {
		targetExerciseName = exerciseName
	}

	if targetExerciseName != "" {
		if clusterName != "" {
			if err := deleteExerciseInCluster(cfg, clusterName, targetExerciseName, className, groupName, confirm); err != nil {
				return fmt.Errorf("failed to delete exercise: %w", err)
			}
		} else {
			if err := deleteExerciseWithConfirmation(cfg, targetExerciseName, className, groupName, confirm); err != nil {
				return fmt.Errorf("failed to delete exercise: %w", err)
			}
		}
	} else if deleteAll {
		var exerciseNames []string
		var err error
		if clusterName != "" {
			exerciseNames, err = getAllExerciseNamesFromCluster(cfg, clusterName)
		} else {
			exerciseNames, err = getAllExerciseNames(cfg)
		}
		if err != nil {
			return fmt.Errorf("failed to get exercise names: %w", err)
		}

		if len(exerciseNames) == 0 {
			fmt.Printf("%v No exercises found to delete\n", messageUtils.InfoMsg("No exercises found to delete"))
			return nil
		}

		if confirm {
			fmt.Printf("%v Found %d exercises to delete:\n", messageUtils.WarningMsg("Found exercises to delete"), len(exerciseNames))
			for _, name := range exerciseNames {
				fmt.Printf("  - %v\n", messageUtils.Bold(name))
			}

			if !utils.ConfirmPrompt("Are you sure you want to delete ALL exercises?", false) {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		dbConn, err := db.InitIfNeeded()
		if err != nil {
			fmt.Printf("%v Failed to initialize database: %v\n",
				messageUtils.WarningMsg("Warning"),
				err)
		} else {
			defer dbConn.Close()

			_, err = dbConn.Exec(`DELETE FROM exercises`)
			if err != nil {
				fmt.Printf("%v Failed to clean up exercise database entries: %v\n",
					messageUtils.WarningMsg("Warning"),
					err)
			} else {
				fmt.Printf("%v Deleted all exercise database entries\n",
					messageUtils.SuccessMsg("Cleaned up database"))
			}
		}

		for _, name := range exerciseNames {
			var derr error
			if clusterName != "" {
				derr = deleteExerciseInCluster(cfg, clusterName, name, className, groupName, false)
			} else {
				derr = deleteExerciseWithConfirmation(cfg, name, className, groupName, false)
			}
			if derr != nil {
				fmt.Printf("%v Failed to delete exercise %v: %v\n",
					messageUtils.ErrorMsg("Failed to delete exercise"),
					messageUtils.Bold(name),
					derr)
			}
		}
	} else if className != "" {
		if err := deleteAllExercisesForClassWithConfirmation(cfg, className, confirm); err != nil {
			return fmt.Errorf("failed to delete exercises for class: %w", err)
		}
	} else if selectExercise || (!selectClass && !selectGroup) {
		var exerciseNames []string
		var err error
		if clusterName != "" {
			exerciseNames, err = selectExercisesWithFuzzyFromCluster(cfg, clusterName, multi)
		} else {
			exerciseNames, err = selectExercisesWithFuzzy(cfg, multi)
		}
		if err != nil {
			return fmt.Errorf("failed to select exercises: %w", err)
		}

		if len(exerciseNames) == 0 {
			fmt.Printf("%v No exercises selected for deletion\n", messageUtils.InfoMsg("No exercises selected for deletion"))
			return nil
		}

		for _, name := range exerciseNames {
			var derr error
			if clusterName != "" {
				derr = deleteExerciseInCluster(cfg, clusterName, name, className, groupName, confirm)
			} else {
				derr = deleteExerciseWithConfirmation(cfg, name, className, groupName, confirm)
			}
			if derr != nil {
				fmt.Printf("%v Failed to delete exercise %v: %v\n",
					messageUtils.ErrorMsg("Failed to delete exercise"),
					messageUtils.Bold(name),
					derr)
			}
		}
	} else {
		// This handles the case where selectClass or selectGroup was used
		// but no specific action was taken (className should be set by now)
		if className != "" {
			if err := deleteAllExercisesForClassWithConfirmation(cfg, className, confirm); err != nil {
				return fmt.Errorf("failed to delete exercises for class: %w", err)
			}
		} else {
			fmt.Printf("%s\n", colorUtils.Warning("No action taken - no class or exercises selected"))
		}
	}

	return nil
}

func selectExercisesWithFuzzy(cfg config.GlobalOptions, multi bool) ([]string, error) {
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

	var exerciseNames []string
	seenExercises := make(map[string]bool)

	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 {
			exerciseName := parts[1]
			if !seenExercises[exerciseName] {
				exerciseNames = append(exerciseNames, exerciseName)
				seenExercises[exerciseName] = true
			}
		}
	}

	if len(exerciseNames) == 0 {
		return nil, fmt.Errorf("no exercises found")
	}

	title := "Select exercises to delete:"
	if multi {
		title = "Select exercises to delete (multi-select):"
	}
	finder := fuzzy.NewFuzzyFinderWithTitle(exerciseNames, multi, title)
	return finder, nil
}

func getAllExerciseNames(cfg config.GlobalOptions) ([]string, error) {
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

	var exerciseNames []string
	seenExercises := make(map[string]bool)

	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 {
			exerciseName := parts[1]
			if !seenExercises[exerciseName] {
				exerciseNames = append(exerciseNames, exerciseName)
				seenExercises[exerciseName] = true
			}
		}
	}

	return exerciseNames, nil
}

func deleteExerciseWithConfirmation(cfg config.GlobalOptions, exerciseName, className, groupName string, confirm bool) error {
	if confirm {
		message := fmt.Sprintf("Delete exercise '%s'?", exerciseName)
		if className != "" {
			message = fmt.Sprintf("Delete exercise '%s' for class '%s'?", exerciseName, className)
		}
		if groupName != "" {
			message = fmt.Sprintf("Delete exercise '%s' for class '%s' group '%s'?", exerciseName, className, groupName)
		}

		if !utils.ConfirmPrompt(message, false) {
			fmt.Printf("Deletion of exercise %v cancelled\n", messageUtils.Bold(exerciseName))
			return nil
		}
	}

	return class.DeleteExercise(cfg, exerciseName, className, groupName)
}

func deleteAllExercisesForClassWithConfirmation(cfg config.GlobalOptions, className string, confirm bool) error {
	if confirm {
		if !utils.ConfirmPrompt(fmt.Sprintf("Delete all exercises for class '%s'?", className), false) {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	dbConn, err := db.InitIfNeeded()
	if err == nil {
		defer dbConn.Close()
		_, err = dbConn.Exec(`DELETE FROM exercises WHERE class = ?`, className)
		if err != nil {
			fmt.Printf("%v Failed to delete exercises for class %s from database: %v\n",
				messageUtils.WarningMsg("Warning"), className, err)
		}
	}

	return class.DeleteAllExercisesForClass(cfg, className)
}

func selectClassWithFuzzy(cfg config.GlobalOptions) (string, error) {
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

	classSet := make(map[string]bool)
	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 {
			className := parts[2]
			if className != "" {
				classSet[className] = true
			}
		}
	}

	if len(classSet) == 0 {
		return "", fmt.Errorf("no classes found")
	}

	var classes []string
	for className := range classSet {
		classes = append(classes, className)
	}

	selected := fuzzy.NewFuzzyFinderWithTitle(classes, false, "Select a class:")
	if len(selected) == 0 {
		return "", nil
	}
	return selected[0], nil
}

func selectGroupWithFuzzy(cfg config.GlobalOptions, className string) (string, error) {
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

	groupSet := make(map[string]bool)
	for _, project := range projects {
		parts := strings.Split(project.Name, "-")
		if len(parts) >= 4 {
			projectClass := parts[2]
			groupName := parts[3]
			if projectClass == className && groupName != "" {
				groupSet[groupName] = true
			}
		}
	}

	if len(groupSet) == 0 {
		return "", fmt.Errorf("no groups found for class %s", className)
	}

	var groups []string
	for groupName := range groupSet {
		groups = append(groups, groupName)
	}

	title := fmt.Sprintf("Select a group for class %s:", className)
	selected := fuzzy.NewFuzzyFinderWithTitle(groups, false, title)
	if len(selected) == 0 {
		return "", nil
	}
	return selected[0], nil
}
