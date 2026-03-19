package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Italic(true)

	groupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
)
