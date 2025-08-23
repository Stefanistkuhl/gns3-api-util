package colorUtils

import "github.com/fatih/color"

var (
	Success = color.New(color.FgGreen).SprintfFunc()

	Error = color.New(color.FgRed, color.Bold).SprintfFunc()

	Warning = color.New(color.FgYellow).SprintfFunc()

	Highlight = color.New(color.FgCyan, color.Bold).SprintfFunc()

	Bold = color.New(color.Bold).SprintfFunc()

	BoldWhite = color.New(color.FgWhite, color.Bold).SprintfFunc()

	Info = color.New(color.FgBlue).SprintfFunc()

	Seperator = color.New(color.FgHiBlack).SprintfFunc()
)

func Emphasize(s string) string {
	return color.New(color.FgMagenta, color.Underline).Sprint(s)
}
