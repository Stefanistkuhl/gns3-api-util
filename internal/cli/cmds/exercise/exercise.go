package exercise

import (
	"github.com/spf13/cobra"
)

func NewExerciseCmdGroup() *cobra.Command {
	exerciseCmd := &cobra.Command{
		Use:     "exercise",
		Aliases: []string{"ex"},
		Short:   "Exercise operations",
		Long:    `Create, manage, and manipulate GNS3 exercises.`,
	}

	exerciseCmd.AddCommand(
		NewExerciseCreateCmd(),
		NewExerciseDeleteCmd(),
		NewExerciseLsCmd(),
		NewExerciseInfoCmd(),
	)

	return exerciseCmd
}
