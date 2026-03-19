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

	// Input box styles
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")).
			BorderBackground(lipgloss.Color("236")).
			Padding(0, 1).
			Width(40)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)

	inputTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	inputCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)

	inputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)
)
