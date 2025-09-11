package class

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
		if err := deleteClassWithConfirmation(cfg, targetClassName, confirm, deleteExercises); err != nil {
			return fmt.Errorf("failed to delete class: %w", err)
		}
	} else if deleteAll {
		classNames, err := getAllClassNames(cfg)
		if err != nil {
			return fmt.Errorf("failed to get class names: %w", err)
		}

		if len(classNames) == 0 {
			fmt.Printf("%v No classes found to delete\n", colorUtils.Info("Info:"))
			return nil
		}

		if confirm {
			fmt.Printf("%v Found %d classes to delete:\n", colorUtils.Warning("Warning:"), len(classNames))
			for _, name := range classNames {
				fmt.Printf("  - %v\n", colorUtils.Bold(name))
			}

			if !confirmAction("Are you sure you want to delete ALL classes?") {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		for _, name := range classNames {
			if err := deleteClassWithConfirmation(cfg, name, false, deleteExercises); err != nil {
				fmt.Printf("%v Failed to delete class %v: %v\n",
					colorUtils.Error("Error:"),
					colorUtils.Bold(name),
					err)
			}
		}
	} else {
		classNames, err := selectClassesWithFuzzy(cfg, multi)
		if err != nil {
			return fmt.Errorf("failed to select classes: %w", err)
		}

		if len(classNames) == 0 {
			fmt.Printf("%v No classes selected for deletion\n", colorUtils.Info("Info:"))
			return nil
		}

		for _, name := range classNames {
			if err := deleteClassWithConfirmation(cfg, name, confirm, deleteExercises); err != nil {
				fmt.Printf("%v Failed to delete class %v: %v\n",
					colorUtils.Error("Error:"),
					colorUtils.Bold(name),
					err)
			}
		}
	}

	return nil
}

func selectClassesWithFuzzy(cfg config.GlobalOptions, multi bool) ([]string, error) {
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

func getAllClassNames(cfg config.GlobalOptions) ([]string, error) {
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
			fmt.Printf("Deletion of class %v cancelled\n", colorUtils.Bold(className))
			return nil
		}
	}

	if deleteExercises {
		fmt.Printf("%v Deleting exercises for class %v...\n",
			colorUtils.Info("Info:"),
			colorUtils.Bold(className))

		if err := class.DeleteAllExercisesForClass(cfg, className); err != nil {
			fmt.Printf("%v Warning: failed to delete exercises for class %v: %v\n",
				colorUtils.Warning("Warning:"),
				colorUtils.Bold(className),
				err)
		} else {
			fmt.Printf("%v Successfully deleted exercises for class %v\n",
				colorUtils.Success("Success:"),
				colorUtils.Bold(className))
		}
	}

	return class.DeleteClass(cfg, className)
}

func confirmAction(message string) bool {
	fmt.Printf("%v %s (y/N): ", colorUtils.Warning("Warning:"), message)
	var response string
	_, _ = fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
