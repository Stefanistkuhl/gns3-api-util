package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewCompletionCmd creates the completion command
func NewCompletionCmd() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for gns3util.

To load completions in your current shell session:
  source <(gns3util completion bash)

To load completions for all new sessions, add to your shell profile:
  echo 'source <(gns3util completion bash)' >> ~/.bashrc

Supported shells: bash, zsh, fish, powershell`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shell := args[0]
			switch shell {
			case "bash":
				_ = rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				_ = rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				_ = rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = rootCmd.GenPowerShellCompletion(os.Stdout)
			default:
				fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
				os.Exit(1)
			}
		},
	}
	return completionCmd
}
