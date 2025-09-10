package exercise

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/class"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
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
  gns3util -s http://server:3080 exercise delete

  # Delete a specific exercise by name
  gns3util -s http://server:3080 exercise delete --name "MyExercise"

  # Delete an exercise non-interactively
  gns3util -s http://server:3080 exercise delete --non-interactive "MyExercise"

  # Delete all exercises for a class
  gns3util -s http://server:3080 exercise delete --set-class "MyClass"

  # Delete exercises for a specific class and group
  gns3util -s http://server:3080 exercise delete --set-class "MyClass" --set-group "Group1"

  # Delete all exercises
  gns3util -s http://server:3080 exercise delete --all

  # Delete without confirmation
  gns3util -s http://server:3080 exercise delete --name "MyExercise" --no-confirm
		`,
		RunE: runDeleteExercise,
	}

	deleteExerciseCmd.Flags().String("name", "", "Name of the exercise to delete")
	deleteExerciseCmd.Flags().String("non-interactive", "", "Run the command non-interactively with specified exercise name")
	deleteExerciseCmd.Flags().String("set-class", "", "Set the class from which to delete the exercise")
	deleteExerciseCmd.Flags().String("set-group", "", "Set the group from which to delete the exercise")
	deleteExerciseCmd.Flags().Bool("select-class", false, "Select class interactively")
	deleteExerciseCmd.Flags().Bool("select-group", false, "Select group interactively")
	deleteExerciseCmd.Flags().Bool("all", false, "Delete all exercises")
	deleteExerciseCmd.Flags().Bool("multi", false, "Enable multi-select mode for fuzzy finder")
	deleteExerciseCmd.Flags().Bool("confirm", true, "Require confirmation before deletion")
	deleteExerciseCmd.Flags().Bool("no-confirm", false, "Skip confirmation prompt")

	return deleteExerciseCmd
}

func runDeleteExercise(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}

	exerciseName, _ := cmd.Flags().GetString("name")
	nonInteractive, _ := cmd.Flags().GetString("non-interactive")
	setClass, _ := cmd.Flags().GetString("set-class")
	setGroup, _ := cmd.Flags().GetString("set-group")
	selectClass, _ := cmd.Flags().GetBool("select-class")
	selectGroup, _ := cmd.Flags().GetBool("select-group")
	deleteAll, _ := cmd.Flags().GetBool("all")
	multi, _ := cmd.Flags().GetBool("multi")
	confirm, _ := cmd.Flags().GetBool("confirm")
	noConfirm, _ := cmd.Flags().GetBool("no-confirm")

	if noConfirm {
		confirm = false
	}

	if exerciseName != "" && nonInteractive != "" {
		return fmt.Errorf("cannot specify both --name and --non-interactive")
	}
	if exerciseName != "" && deleteAll {
		return fmt.Errorf("cannot specify both --name and --all")
	}
	if nonInteractive != "" && deleteAll {
		return fmt.Errorf("cannot specify both --non-interactive and --all")
	}
	if nonInteractive != "" && setClass != "" {
		return fmt.Errorf("cannot specify both --non-interactive and --set-class")
	}
	if nonInteractive != "" && setGroup != "" {
		return fmt.Errorf("cannot specify both --non-interactive and --set-group")
	}

	if nonInteractive != "" {
		if setClass == "" && setGroup != "" {
			return fmt.Errorf("in non-interactive mode, --set-class and --set-group must both be provided with string values")
		}
	}

	if nonInteractive == "" && !deleteAll {
		if selectClass == false && selectGroup == true {
			return fmt.Errorf("in interactive mode, --select-class and --select-group must both be set")
		}
		if (selectClass || selectGroup) && multi {
			return fmt.Errorf("in interactive mode when either --select-class or --select-group are set, multi mode is not supported")
		}
	}

	var targetExerciseName string
	if nonInteractive != "" {
		targetExerciseName = nonInteractive
	} else if exerciseName != "" {
		targetExerciseName = exerciseName
	}

	if targetExerciseName != "" {
		if err := deleteExerciseWithConfirmation(cfg, targetExerciseName, setClass, setGroup, confirm); err != nil {
			return fmt.Errorf("failed to delete exercise: %w", err)
		}
	} else if deleteAll {
		exerciseNames, err := getAllExerciseNames(cfg)
		if err != nil {
			return fmt.Errorf("failed to get exercise names: %w", err)
		}

		if len(exerciseNames) == 0 {
			fmt.Printf("%v No exercises found to delete\n", colorUtils.Info("Info:"))
			return nil
		}

		if confirm {
			fmt.Printf("%v Found %d exercises to delete:\n", colorUtils.Warning("Warning:"), len(exerciseNames))
			for _, name := range exerciseNames {
				fmt.Printf("  - %v\n", colorUtils.Bold(name))
			}

			if !confirmAction("Are you sure you want to delete ALL exercises?") {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		for _, name := range exerciseNames {
			if err := deleteExerciseWithConfirmation(cfg, name, "", "", false); err != nil {
				fmt.Printf("%v Failed to delete exercise %v: %v\n",
					colorUtils.Error("Error:"),
					colorUtils.Bold(name),
					err)
			}
		}
	} else if setClass != "" {
		if err := deleteAllExercisesForClassWithConfirmation(cfg, setClass, confirm); err != nil {
			return fmt.Errorf("failed to delete exercises for class: %w", err)
		}
	} else {
		exerciseNames, err := selectExercisesWithFuzzy(cfg, multi)
		if err != nil {
			return fmt.Errorf("failed to select exercises: %w", err)
		}

		if len(exerciseNames) == 0 {
			fmt.Printf("%v No exercises selected for deletion\n", colorUtils.Info("Info:"))
			return nil
		}

		for _, name := range exerciseNames {
			if err := deleteExerciseWithConfirmation(cfg, name, "", "", confirm); err != nil {
				fmt.Printf("%v Failed to delete exercise %v: %v\n",
					colorUtils.Error("Error:"),
					colorUtils.Bold(name),
					err)
			}
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

	finder := fuzzy.NewFuzzyFinder(exerciseNames, multi)
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

		if !confirmAction(message) {
			fmt.Printf("Deletion of exercise %v cancelled\n", colorUtils.Bold(exerciseName))
			return nil
		}
	}

	return class.DeleteExercise(cfg, exerciseName, className, groupName)
}

func deleteAllExercisesForClassWithConfirmation(cfg config.GlobalOptions, className string, confirm bool) error {
	if confirm {
		if !confirmAction(fmt.Sprintf("Delete all exercises for class '%s'?", className)) {
			fmt.Printf("Deletion of exercises for class %v cancelled\n", colorUtils.Bold(className))
			return nil
		}
	}

	return class.DeleteAllExercisesForClass(cfg, className)
}
