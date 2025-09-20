package loading

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
)

type LoadingModel struct {
	message     string
	spinner     int
	done        bool
	success     bool
	errorMsg    string
	steps       []string
	currentStep int
}

type LoadingMsg struct {
	Type    string
	Message string
	Error   error
}

const (
	spinnerChars = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
)

var (
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	stepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	currentStepStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("33")).
				Bold(true)

	completedStepStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82"))
)

func NewLoadingModel(message string, steps []string) *LoadingModel {
	return &LoadingModel{
		message:     message,
		spinner:     0,
		done:        false,
		success:     false,
		steps:       steps,
		currentStep: 0,
	}
}

func (m LoadingModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return LoadingMsg{Type: "tick"}
	})
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadingMsg:
		switch msg.Type {
		case "tick":
			if !m.done {
				m.spinner = (m.spinner + 1) % len(spinnerChars)
				return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
					return LoadingMsg{Type: "tick"}
				})
			}
		case "step":
			m.currentStep++
			return m, nil
		case "success":
			m.done = true
			m.success = true
			return m, tea.Quit
		case "error":
			m.done = true
			m.success = false
			m.errorMsg = msg.Message
			return m, tea.Quit
		}
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m LoadingModel) View() string {
	if m.done {
		if m.success {
			return m.renderSuccess()
		}
		return m.renderError()
	}
	return m.renderLoading()
}

func (m LoadingModel) renderLoading() string {
	var sb strings.Builder

	sb.WriteString(messageUtils.Bold("🔧 GNS3 SSL Installation\n"))
	sb.WriteString(messageUtils.Seperator(strings.Repeat("─", 50)) + "\n\n")

	spinner := spinnerStyle.Render(string(spinnerChars[m.spinner]))
	sb.WriteString(fmt.Sprintf("%s %s\n\n", spinner, m.message))

	if len(m.steps) > 0 {
		sb.WriteString(messageUtils.InfoMsg("Steps:\n"))
		for i, step := range m.steps {
			var stepText string
			if i < m.currentStep {
				stepText = completedStepStyle.Render("✓ " + step)
			} else if i == m.currentStep {
				stepText = currentStepStyle.Render("→ " + step)
			} else {
				stepText = stepStyle.Render("  " + step)
			}
			sb.WriteString("  " + stepText + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(messageUtils.Seperator("Press Ctrl+C to cancel"))

	return sb.String()
}

func (m LoadingModel) renderSuccess() string {
	var sb strings.Builder

	sb.WriteString(messageUtils.Bold("🎉 GNS3 SSL Installation Complete!\n"))
	sb.WriteString(messageUtils.Seperator(strings.Repeat("─", 50)) + "\n\n")

	sb.WriteString(successStyle.Render("✓ ") + "All steps completed successfully\n\n")

	if len(m.steps) > 0 {
		sb.WriteString(messageUtils.InfoMsg("Completed steps:\n"))
		for _, step := range m.steps {
			sb.WriteString("  " + completedStepStyle.Render("✓ "+step) + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(messageUtils.SuccessMsg("Your GNS3 server is now accessible via HTTPS! 🚀\n"))

	return sb.String()
}

func (m LoadingModel) renderError() string {
	var sb strings.Builder

	sb.WriteString(messageUtils.Bold("❌ GNS3 SSL Installation Failed\n"))
	sb.WriteString(messageUtils.Seperator(strings.Repeat("─", 50)) + "\n\n")

	sb.WriteString(errorStyle.Render("✗ ") + "Installation failed\n\n")

	if m.errorMsg != "" {
		sb.WriteString(messageUtils.ErrorMsg("Error: ") + m.errorMsg + "\n\n")
	}

	if len(m.steps) > 0 && m.currentStep > 0 {
		sb.WriteString(messageUtils.InfoMsg("Completed steps:\n"))
		for i, step := range m.steps {
			if i < m.currentStep {
				sb.WriteString("  " + completedStepStyle.Render("✓ "+step) + "\n")
			} else if i == m.currentStep {
				sb.WriteString("  " + errorStyle.Render("✗ "+step) + "\n")
			} else {
				sb.WriteString("  " + stepStyle.Render("  "+step) + "\n")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(messageUtils.WarningMsg("Please check the error message above and try again.\n"))

	return sb.String()
}

func (m *LoadingModel) NextStep() {
	m.currentStep++
}

func (m *LoadingModel) SetMessage(msg string) {
	m.message = msg
}

func (m *LoadingModel) SetSuccess() {
	m.done = true
	m.success = true
}

func (m *LoadingModel) SetError(err error) {
	m.done = true
	m.success = false
	m.errorMsg = err.Error()
}

func RunLoading(message string, steps []string, fn func(*LoadingModel)) error {
	model := NewLoadingModel(message, steps)

	go func() {
		fn(model)
	}()

	_, err := tea.NewProgram(model).Run()
	return err
}
