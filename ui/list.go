package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

// ListModel is a simple arrow-key navigable list.
type ListModel struct {
	title    string
	commands []config.Command
	cursor   int
	selected *config.Command
	width    int
	height   int
}

func NewListModel(cfg *config.Config) ListModel {
	return ListModel{
		title:    cfg.Title,
		commands: cfg.Commands,
		width:    80,
		height:   24,
	}
}

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) Selected() *config.Command { return m.selected }

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.commands)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.commands) > 0 {
				cmd := m.commands[m.cursor]
				m.selected = &cmd
			}
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title) + "\n\n")

	for i, cmd := range m.commands {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		name := normalStyle.Render(cmd.Name)
		if i == m.cursor {
			name = selectedStyle.Render(cmd.Name)
		}

		line := fmt.Sprintf("%s%s", cursor, name)
		if cmd.Description != "" {
			line += "  " + descStyle.Render(cmd.Description)
		}
		if cmd.Command != "" {
			line += "\n    " + cmdStyle.Render("$ "+cmd.Command)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(helpStyle.Render("\n↑/↓ navigate  •  enter select  •  q quit"))
	return b.String()
}
