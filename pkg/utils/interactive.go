package utils

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginModel struct {
	username string
	password string
	step     int
	done     bool
	err      error
}

func (m LoginModel) Init() tea.Cmd {
	return nil
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.step == 0 {
				if m.username == "" {
					m.err = fmt.Errorf("username cannot be empty")
					return m, nil
				}
				m.step = 1
				return m, nil
			} else {
				if m.password == "" {
					m.err = fmt.Errorf("password cannot be empty")
					return m, nil
				}
				if !ValidatePassword(m.password) {
					m.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
					return m, nil
				}
				m.done = true
				return m, tea.Quit
			}
		case "backspace":
			if m.step == 0 {
				if len(m.username) > 0 {
					m.username = m.username[:len(m.username)-1]
				}
				m.err = nil
			} else {
				if len(m.password) > 0 {
					m.password = m.password[:len(m.password)-1]
				}
				m.err = nil
			}
		default:
			if m.step == 0 {
				m.username += msg.String()
				m.err = nil
			} else {
				m.password += msg.String()
				m.err = nil
			}
		}
	}
	return m, nil
}

func (m LoginModel) View() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Margin(1, 0).
		Render("GNS3 Login")

	s.WriteString(title)
	s.WriteString("\n")

	if m.step == 0 {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Render("Enter username:")
		s.WriteString(prompt)
		s.WriteString("\n")

		input := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("141")).
			Padding(0, 1).
			Width(50).
			Render(m.username)
		s.WriteString(input)
	} else {
		prompt := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Render("Enter password:")
		s.WriteString(prompt)
		s.WriteString("\n")

		input := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("141")).
			Padding(0, 1).
			Width(50).
			Render(strings.Repeat("*", len(m.password)))
		s.WriteString(input)
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Margin(1, 0).
			Render("Error: " + m.err.Error())
		s.WriteString("\n\n")
		s.WriteString(errorStyle)
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\n\nPress Enter to continue, Ctrl+C to cancel")
	s.WriteString(help)

	return s.String()
}

type PasswordModel struct {
	password string
	done     bool
	err      error
}

func (m PasswordModel) Init() tea.Cmd {
	return nil
}

func (m PasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.password == "" {
				m.err = fmt.Errorf("password cannot be empty")
				return m, nil
			}
			if !ValidatePassword(m.password) {
				m.err = fmt.Errorf("password must be at least 8 characters with at least 1 number and 1 lowercase letter")
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "backspace":
			if len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			}
			m.err = nil
		default:
			m.password += msg.String()
			m.err = nil
		}
	}
	return m, nil
}

func (m PasswordModel) View() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Margin(1, 0).
		Render("Change User Password")

	s.WriteString(title)
	s.WriteString("\n\n")

	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Render("Enter new password (min 8 chars, 1 number, 1 lowercase):")
	s.WriteString(prompt)
	s.WriteString("\n")

	input := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(0, 1).
		Width(50).
		Render(strings.Repeat("*", len(m.password)))
	s.WriteString(input)

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Margin(1, 0).
			Render("Error: " + m.err.Error())
		s.WriteString("\n\n")
		s.WriteString(errorStyle)
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\n\nPress Enter to continue, Ctrl+C to cancel")
	s.WriteString(help)

	return s.String()
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)

	return hasNumber && hasLower
}

func GetLoginCredentials() (string, string, error) {
	p := tea.NewProgram(LoginModel{})

	m, err := p.Run()
	if err != nil {
		return "", "", fmt.Errorf("login cancelled")
	}

	model := m.(LoginModel)

	if model.err != nil {
		return "", "", model.err
	}

	if !model.done {
		return "", "", fmt.Errorf("login cancelled")
	}

	return model.username, model.password, nil
}

func GetPasswordFromInput() (string, error) {
	p := tea.NewProgram(PasswordModel{})

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("password input cancelled")
	}

	model := m.(PasswordModel)

	if model.err != nil {
		return "", model.err
	}

	if !model.done {
		return "", fmt.Errorf("password input cancelled")
	}

	return model.password, nil
}
