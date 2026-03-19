package utils

import (
	"fmt"
	"os"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/globals"
	"github.com/charmbracelet/lipgloss"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func PrintError(err error, cfg *config.GlobalOptions) {
	if err == nil {
		return
	}

	formatStr := cfg.OutputFormat.String()

	if formatStr == globals.OutputJSON.String() ||
		formatStr == globals.OutputJSONColorless.String() ||
		formatStr == globals.OutputYAML.String() ||
		formatStr == globals.OutputTOML.String() {

		printer, getErr := GetPrinter(formatStr)
		if getErr == nil {
			_ = printer.PrintObj(ErrorResponse{Error: err.Error()}, os.Stderr)
			return
		}
	}

	errorPrefix := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true).
		Render("Error:")

	fmt.Fprintf(os.Stderr, "%s %v\n", errorPrefix, err)
}
