package fuzzy

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

var (
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Bold(true).
			Render
	normalStyle = lipgloss.NewStyle().Render
	cursorGlyph = "› "
	inputBox    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("183"))
	// markedStyle = lipgloss.NewStyle().
	//		Foreground(lipgloss.Color("84")).
	//		Render
)

type selectionMsg []string

type fuzzyFinder struct {
	viewport  viewport.Model
	rows      []string
	cursor    int
	input     textinput.Model
	liveValue string
	termW     int
	boxW      int
	multiMode bool
	selection []string
}

func newFuzzyFinder(data []string, multiMode bool) (*fuzzyFinder, error) {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.Focus()

	vp := viewport.New(0, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("183"))

	f := &fuzzyFinder{
		viewport:  vp,
		rows:      data,
		cursor:    0,
		input:     ti,
		termW:     0,
		boxW:      0,
		multiMode: multiMode,
	}
	return f, nil
}

func (f *fuzzyFinder) refreshContent() {
	contentWidth := max(f.viewport.Width-2, 0)
	pattern := f.liveValue
	matches := fuzzy.Find(pattern, f.rows)
	results := getMatches(matches)

	if pattern == "" {
		lines := make([]string, 0, len(f.rows)+1)
		for i, row := range f.rows {
			prefix := "  "
			if i == f.cursor {
				prefix = cursorGlyph
			}
			marker := " "
			if slices.Contains(f.selection, row) {
				marker = "│"
			}
			line := marker + " " + prefix + row
			padded := lipgloss.NewStyle().Width(contentWidth).Render(line)
			if i == f.cursor {
				lines = append(lines, selectedStyle(padded))
			} else {
				lines = append(lines, normalStyle(padded))
			}
		}
		f.viewport.SetContent(strings.Join(lines, "\n"))
	} else {
		if f.cursor >= len(results) {
			f.cursor = max(0, len(results)-1)
		}
		lines := make([]string, 0, len(results)+1)
		for i, row := range results {
			prefix := "  "
			if i == f.cursor {
				prefix = cursorGlyph
			}
			marker := " "
			if slices.Contains(f.selection, row) {
				marker = "│"
			}
			line := marker + " " + prefix + row
			padded := lipgloss.NewStyle().Width(contentWidth).Render(line)
			if i == f.cursor {
				lines = append(lines, selectedStyle(padded))
			} else {
				lines = append(lines, normalStyle(padded))
			}
		}
		f.viewport.SetContent(strings.Join(lines, "\n"))
	}

	top := f.viewport.YOffset
	bottom := top + f.viewport.Height - 1
	if f.cursor < top {
		f.viewport.YOffset = f.cursor
	} else if f.cursor > bottom {
		f.viewport.YOffset = f.cursor - f.viewport.Height + 1
	}
}

func (f *fuzzyFinder) Init() tea.Cmd {
	return nil
}

func (f *fuzzyFinder) helpView() string {
	if f.multiMode {
		return helpStyle("\n  ↑/↓: Navigate • <tab>: Toggle selection • <enter>: Confirm & quit • Control c/<esc>: Quit\n")
	}
	return helpStyle("\n  ↑/↓: Navigate • <enter>: Select & quit • Control c/<esc>: Quit\n")
}

func (f *fuzzyFinder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.termW = msg.Width
		f.boxW = int(float64(msg.Width) * 0.6)
		promptCells := lipgloss.Width(f.input.Prompt)
		f.input.Width = max(f.boxW-promptCells-3, 0)
		f.viewport.Width = f.boxW
		f.refreshContent()
		return f, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return f, tea.Quit
		case "up":
			if f.cursor > 0 {
				f.cursor--
				f.refreshContent()
			}
			return f, nil
		case "down":
			if f.cursor < len(f.rows)-1 {
				f.cursor++
				f.refreshContent()
			}
			return f, nil
		case "enter":
			var choice string
			if f.liveValue == "" {
				choice = f.rows[f.cursor]
			} else {
				matches := fuzzy.Find(f.liveValue, f.rows)
				results := getMatches(matches)
				if f.cursor < len(results) {
					choice = results[f.cursor]
				}
			}

			if !slices.Contains(f.selection, choice) {
				f.selection = append(f.selection, choice)
			}

			return f, tea.Sequence(
				func() tea.Msg { return selectionMsg(f.selection) },
				tea.Quit,
			)
		case "tab":
			if f.multiMode {
				row := f.rows[f.cursor]
				if slices.Contains(f.selection, row) {
					f.selection = slices.DeleteFunc(f.selection, func(s string) bool {
						return s == row
					})
				} else {
					f.selection = append(f.selection, row)
				}
			}
		}

	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)

	f.liveValue = f.input.Value()

	f.refreshContent()
	return f, cmd
}

func (f *fuzzyFinder) View() string {
	gap := "\n"
	boxedInput := inputBox.Render(f.input.View())
	return boxedInput + gap + f.viewport.View() + f.helpView()
}

func getMatches(matches []fuzzy.Match) []string {
	var results = []string{}
	for _, match := range matches {
		results = append(results, match.Str)
	}
	return results
}

func countLines(s string) int {
	return strings.Count(s, "\n") + 1
}

func clearLines(n int) {
	for range make([]struct{}, n) {
		fmt.Print("\033[1A")
		fmt.Print("\033[2K")
	}
}

// deduplicate removes duplicate strings from a slice while preserving order
func deduplicate(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func NewFuzzyFinder(input []string, multiMode bool) []string {
	var result []string
	model, err := newFuzzyFinder(input, multiMode)

	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}

	a, err := tea.NewProgram(model).Run()
	if err != nil {
		fmt.Println("Bummer, there's been an error:", err)
		os.Exit(1)
	}
	final := a.(*fuzzyFinder)
	lines := countLines(final.View())
	clearLines(lines)

	result = deduplicate(final.selection)
	return result
}
