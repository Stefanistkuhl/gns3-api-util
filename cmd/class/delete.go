package class

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
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

func NewClassDeleteCmd() *cobra.Command {
	var deleteClassCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a class and its students",
		Long: `Delete a class and all its associated groups and users. This command can:
- Delete a specific class by name
- Use fuzzy finder to select classes interactively
- Delete all classes at once
- Delete exercises associated with classes`,
		Example: `
  # Delete a class interactively using fuzzy finder
  gns3util -s http://server:3080 class delete

  # Delete a specific class by name
  gns3util -s http://server:3080 class delete --name "MyClass"

  # Delete a class non-interactively (no fuzzy finder)
  gns3util -s http://server:3080 class delete --non-interactive "MyClass"

  # Delete all classes
  gns3util -s http://server:3080 class delete --all

  # Delete multiple classes using fuzzy finder
  gns3util -s http://server:3080 class delete --multi

  # Delete without confirmation
  gns3util -s http://server:3080 class delete --name "MyClass" --no-confirm

  # Delete all classes without confirmation
  gns3util -s http://server:3080 class delete --all --no-confirm

  # Delete class and its exercises
  gns3util -s http://server:3080 class delete --name "MyClass" --delete-exercises
		`,
		RunE: runDeleteClass,
	}

	deleteClassCmd.Flags().String("name", "", "Name of the class to delete")
	deleteClassCmd.Flags().String("non-interactive", "", "Run the command non-interactively with specified class name")
	deleteClassCmd.Flags().Bool("all", false, "Delete all classes")
	deleteClassCmd.Flags().Bool("multi", false, "Enable multi-select mode for fuzzy finder")
	deleteClassCmd.Flags().Bool("confirm", true, "Require confirmation before deletion")
	deleteClassCmd.Flags().Bool("no-confirm", false, "Skip confirmation prompt")
	deleteClassCmd.Flags().Bool("delete-exercises", false, "Delete all exercises of the class")
	deleteClassCmd.Flags().Bool("db-first", true, "Check database first for classes (default: true)")
	deleteClassCmd.Flags().StringP("cluster", "c", "", "Cluster name")

	return deleteClassCmd
}

func runDeleteClass(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}

	className, _ := cmd.Flags().GetString("name")
	nonInteractive, _ := cmd.Flags().GetString("non-interactive")
	deleteAll, _ := cmd.Flags().GetBool("all")
	multi, _ := cmd.Flags().GetBool("multi")
	confirm, _ := cmd.Flags().GetBool("confirm")
	noConfirm, _ := cmd.Flags().GetBool("no-confirm")
	deleteExercises, _ := cmd.Flags().GetBool("delete-exercises")
	dbFirst, _ := cmd.Flags().GetBool("db-first")
	clusterName, _ := cmd.Flags().GetString("cluster")

	if noConfirm {
		confirm = false
	}

	if className != "" && nonInteractive != "" {
		return fmt.Errorf("cannot specify both --name and --non-interactive")
	}
	if className != "" && deleteAll {
		return fmt.Errorf("cannot specify both --name and --all")
	}
	if nonInteractive != "" && deleteAll {
		return fmt.Errorf("cannot specify both --non-interactive and --all")
	}

	var targetClassName string
	if nonInteractive != "" {
		targetClassName = nonInteractive
	} else if className != "" {
		targetClassName = className
	}

	if targetClassName != "" {
		if clusterName != "" {
			if err := deleteClassInCluster(cfg, clusterName, targetClassName, confirm, deleteExercises); err != nil {
				return fmt.Errorf("failed to delete class: %w", err)
			}
			return nil
		}
		if err := deleteClassWithConfirmation(cfg, targetClassName, confirm, deleteExercises); err != nil {
			return fmt.Errorf("failed to delete class: %w", err)
		}
	} else if deleteAll {
		var classNames []string
		var err error
		if clusterName != "" {
			classNames, err = getAllClassNamesFromDBForClusterName(clusterName)
		} else {
			classNames, err = getAllClassNames(cfg, dbFirst)
		}
		if err != nil {
			return fmt.Errorf("failed to get class names: %w", err)
		}

		if len(classNames) == 0 {
			fmt.Printf("%v No classes found to delete\n", messageUtils.InfoMsg("No classes found to delete"))
			return nil
		}

		if confirm {
			fmt.Printf("%v Found %d classes to delete:\n", messageUtils.WarningMsgf("Found %d classes to delete", len(classNames)), len(classNames))
			for _, name := range classNames {
				fmt.Printf("  - %v\n", messageUtils.Bold(name))
			}

			if !confirmAction("Are you sure you want to delete ALL classes?") {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		for _, name := range classNames {
			var derr error
			if clusterName != "" {
				derr = deleteClassInCluster(cfg, clusterName, name, false, deleteExercises)
			} else {
				derr = deleteClassWithConfirmation(cfg, name, false, deleteExercises)
			}
			if derr != nil {
				fmt.Printf("%v Failed to delete class %v: %v\n",
					messageUtils.ErrorMsg("Failed to delete class"),
					messageUtils.Bold(name),
					derr)
			}
		}
	} else {
		var classNames []string
		var err error
		if clusterName != "" {
			classNames, err = selectClassesWithFuzzyForClusterName(clusterName, multi)
		} else {
			classNames, err = selectClassesWithFuzzy(cfg, multi, dbFirst)
		}
		if err != nil {
			return fmt.Errorf("failed to select classes: %w", err)
		}

		if len(classNames) == 0 {
			fmt.Printf("%v No classes selected for deletion\n", messageUtils.InfoMsg("No classes selected for deletion"))
			return nil
		}

		for _, name := range classNames {
			var derr error
			if clusterName != "" {
				derr = deleteClassInCluster(cfg, clusterName, name, confirm, deleteExercises)
			} else {
				derr = deleteClassWithConfirmation(cfg, name, confirm, deleteExercises)
			}
			if derr != nil {
				fmt.Printf("%v Failed to delete class %v: %v\n",
					messageUtils.ErrorMsg("Failed to delete class"),
					messageUtils.Bold(name),
					derr)
			}
		}
	}

	return nil
}

func deleteClassInCluster(cfg config.GlobalOptions, clusterName, className string, confirm bool, deleteExercises bool) error {
	if confirm {
		message := fmt.Sprintf("Delete class '%s' from cluster '%s'?", className, clusterName)
		if deleteExercises {
			message = fmt.Sprintf("Delete class '%s' and all its exercises from cluster '%s'?", className, clusterName)
		}
		if !confirmAction(message) {
			fmt.Printf("Deletion of class %v cancelled\n", messageUtils.Bold(className))
			return nil
		}
	}

	conn, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}()

	clusters, err := db.GetClusters(conn)
	if err != nil {
		return fmt.Errorf("failed to get clusters: %w", err)
	}
	clusterID := 0
	for _, c := range clusters {
		if c.Name == clusterName {
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

	if err := db.DeleteClassFromDB(conn, clusterID, className); err != nil {
		fmt.Printf("%v Warning: failed to delete class %v from database: %v\n",
			messageUtils.WarningMsgf("Warning: failed to delete class %s from database", className),
			messageUtils.Bold(className),
			err)
	}

	for _, n := range nodes {
		if n.ClusterID != clusterID {
			continue
		}
		nodeCfg := cfg
		nodeCfg.Server = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)

		if deleteExercises {
			if err := class.DeleteAllExercisesForClass(nodeCfg, className); err != nil {
				fmt.Printf("%v failed to delete exercises for class %v on %s: %v\n",
					messageUtils.WarningMsg("Warning"), messageUtils.Bold(className), nodeCfg.Server, err)
			}
		}

		if err := class.DeleteClass(nodeCfg, className); err != nil {
			fmt.Printf("%v Failed to delete class %v on %s: %v\n",
				messageUtils.ErrorMsg("Failed to delete class"), messageUtils.Bold(className), nodeCfg.Server, err)
		} else {
			fmt.Printf("%v Deleted class %v on %s\n",
				messageUtils.SuccessMsg("Deleted class"), messageUtils.Bold(className), nodeCfg.Server)
		}
	}

	return nil
}

func getAllClassNames(cfg config.GlobalOptions, dbFirst bool) ([]string, error) {
	if dbFirst {
		classNames, err := getAllClassNamesFromDB(cfg)
		if err == nil && len(classNames) > 0 {
			return classNames, nil
		}
	}
	return getAllClassNamesFromAPI(cfg)
}

func getAllClassNamesFromDB(cfg config.GlobalOptions) ([]string, error) {
	clusterID, err := getClusterIDForServer(cfg)
	if err != nil {
		return nil, err
	}

	conn, err := db.InitIfNeeded()
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}()

	classes, err := db.GetClassesFromDB(conn, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes from db: %w", err)
	}

	var classNames []string
	for _, class := range classes {
		classNames = append(classNames, class.Name)
	}

	return classNames, nil
}

func getAllClassNamesFromAPI(cfg config.GlobalOptions) ([]string, error) {
	groupsBody, status, err := utils.CallClient(cfg, "getGroups", []string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("failed to get groups: status %d", status)
	}

	var groups []schemas.UserGroupResponse
	if err := json.Unmarshal(groupsBody, &groups); err != nil {
		return nil, fmt.Errorf("failed to parse groups response: %w", err)
	}

	return getClassNamesFromGroups(groups)
}

func selectClassesWithFuzzy(cfg config.GlobalOptions, multi bool, dbFirst bool) ([]string, error) {
	if dbFirst {
		classNames, err := getAllClassNamesFromDB(cfg)
		if err == nil && len(classNames) > 0 {
			finder := fuzzy.NewFuzzyFinder(classNames, multi)
			return finder, nil
		}
	}

	return selectClassesWithFuzzyFromAPI(cfg, multi)
}

func selectClassesWithFuzzyFromAPI(cfg config.GlobalOptions, multi bool) ([]string, error) {
	groupsBody, status, err := utils.CallClient(cfg, "getGroups", []string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	if status != 200 {
		return nil, fmt.Errorf("failed to get groups: status %d", status)
	}

	var groups []schemas.UserGroupResponse
	if err := json.Unmarshal(groupsBody, &groups); err != nil {
		return nil, fmt.Errorf("failed to parse groups response: %w", err)
	}

	classNames, err := getClassNamesFromGroups(groups)
	if err != nil {
		return nil, fmt.Errorf("failed to extract class names: %w", err)
	}

	if len(classNames) == 0 {
		return nil, fmt.Errorf("no classes found")
	}

	finder := fuzzy.NewFuzzyFinder(classNames, multi)
	return finder, nil
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

func getClassNamesFromGroups(groups []schemas.UserGroupResponse) ([]string, error) {
	var classNames []string
	seenClasses := make(map[string]bool)

	for _, group := range groups {
		parts := strings.Split(group.Name, "-")
		if len(parts) == 3 {
			className := parts[0]
			if !seenClasses[className] {
				for _, classGroup := range groups {
					if classGroup.Name == className {
						classNames = append(classNames, className)
						seenClasses[className] = true
						break
					}
				}
			}
		}
	}

	return classNames, nil
}

func deleteClassWithConfirmation(cfg config.GlobalOptions, className string, confirm bool, deleteExercises bool) error {
	if confirm {
		message := fmt.Sprintf("Delete class '%s'?", className)
		if deleteExercises {
			message = fmt.Sprintf("Delete class '%s' and all its exercises?", className)
		}

		if !confirmAction(message) {
			fmt.Printf("Deletion of class %v cancelled\n", messageUtils.Bold(className))
			return nil
		}
	}

	if deleteExercises {
		fmt.Printf("%v Deleting exercises for class %v...\n",
			messageUtils.InfoMsg("Deleting exercises for class"),
			messageUtils.Bold(className))

		if err := class.DeleteAllExercisesForClass(cfg, className); err != nil {
			fmt.Printf("%v failed to delete exercises for class %v: %v\n",
				messageUtils.WarningMsg("Warning: failed to delete exercises for class"),
				messageUtils.Bold(className),
				err)
		} else {
			fmt.Printf("%v Successfully deleted exercises for class %v\n",
				messageUtils.SuccessMsg("Successfully deleted exercises for class"),
				messageUtils.Bold(className))
		}
	}

	return class.DeleteClass(cfg, className)
}

func confirmAction(message string) bool {
	fmt.Printf("%v %s (y/N): ", messageUtils.WarningMsg("Warning:"), message)
	var response string
	_, _ = fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}

func getAllClassNamesFromDBForClusterName(clusterName string) ([]string, error) {
	conn, err := db.InitIfNeeded()
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close database connection: %v", err)
		}
	}()

	clusters, err := db.GetClusters(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters: %w", err)
	}
	clusterID := 0
	for _, c := range clusters {
		if c.Name == clusterName {
			clusterID = c.Id
			break
		}
	}
	if clusterID == 0 {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}

	classes, err := db.GetClassesFromDB(conn, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes from db: %w", err)
	}

	var classNames []string
	for _, class := range classes {
		classNames = append(classNames, class.Name)
	}
	return classNames, nil
}

func selectClassesWithFuzzyForClusterName(clusterName string, multi bool) ([]string, error) {
	classNames, err := getAllClassNamesFromDBForClusterName(clusterName)
	if err != nil {
		return nil, err
	}
	if len(classNames) == 0 {
		return nil, fmt.Errorf("no classes found in cluster %s", clusterName)
	}
	finder := fuzzy.NewFuzzyFinder(classNames, multi)
	return finder, nil
}
