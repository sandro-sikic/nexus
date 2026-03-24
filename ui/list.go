package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
)

// ListModel is a simple arrow-key navigable list.
type ListModel struct {
	title    string
	tasks    []config.Task
	cursor   int
	selected *config.Task
	width    int
	height   int
}

func NewListModel(cfg *config.Config) ListModel {
	return ListModel{
		title:  cfg.Title,
		tasks:  cfg.Tasks,
		cursor: 0,
		width:  80,
		height: 24,
	}
}

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) Selected() *config.Task { return m.selected }

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
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.tasks) > 0 {
				task := m.tasks[m.cursor]
				m.selected = &task
			}
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title) + "\n\n")

	for i, task := range m.tasks {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		name := normalStyle.Render(task.Name)
		if i == m.cursor {
			name = selectedStyle.Render(task.Name)
		}

		line := fmt.Sprintf("%s%s", cursor, name)
		if task.Description != "" {
			line += "  " + descStyle.Render(task.Description)
		}
		cmds := task.AllCommands()
		if len(cmds) == 1 {
			line += "\n    " + cmdStyle.Render("$ "+cmds[0])
		} else if len(cmds) > 1 {
			line += "\n    " + cmdStyle.Render(fmt.Sprintf("$ %s  (+%d more)", cmds[0], len(cmds)-1))
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(helpStyle.Render("\n↑/↓ navigate  •  enter select  •  q quit"))
	return b.String()
}
