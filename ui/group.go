package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
)

type groupEntry struct {
	isHeader bool
	group    string
	task     *config.Task
}

// GroupModel presents tasks organized by their group field.
type GroupModel struct {
	title    string
	entries  []groupEntry
	cursor   int
	selected *config.Task
	width    int
	height   int
}

func NewGroupModel(cfg *config.Config) GroupModel {
	// Build entries with group headers
	seen := map[string]bool{}
	var entries []groupEntry

	// Collect group order
	var groupOrder []string
	groupTasks := map[string][]config.Task{}
	for _, task := range cfg.Tasks {
		g := task.Group
		if g == "" {
			g = "General"
		}
		if !seen[g] {
			seen[g] = true
			groupOrder = append(groupOrder, g)
		}
		groupTasks[g] = append(groupTasks[g], task)
	}

	for _, g := range groupOrder {
		entries = append(entries, groupEntry{isHeader: true, group: g})
		for i := range groupTasks[g] {
			t := groupTasks[g][i]
			entries = append(entries, groupEntry{task: &t})
		}
	}

	// Set initial cursor to first non-header
	cursor := 0
	for i, e := range entries {
		if !e.isHeader {
			cursor = i
			break
		}
	}

	return GroupModel{
		title:   cfg.Title,
		entries: entries,
		cursor:  cursor,
		width:   80,
		height:  24,
	}
}

func (m GroupModel) Init() tea.Cmd { return nil }

func (m GroupModel) Selected() *config.Task { return m.selected }

func (m *GroupModel) moveCursor(delta int) {
	next := m.cursor + delta
	for next >= 0 && next < len(m.entries) {
		if !m.entries[next].isHeader {
			m.cursor = next
			return
		}
		next += delta
	}
}

func (m GroupModel) Update(msg tea.Msg) (GroupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "enter", " ":
			if m.cursor < len(m.entries) && !m.entries[m.cursor].isHeader {
				m.selected = m.entries[m.cursor].task
			}
		}
	}
	return m, nil
}

func (m GroupModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title) + "\n\n")

	for i, e := range m.entries {
		if e.isHeader {
			b.WriteString(groupHeaderStyle.Render("── "+e.group+" ──") + "\n")
			continue
		}

		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		name := normalStyle.Render(e.task.Name)
		if i == m.cursor {
			name = selectedStyle.Render(e.task.Name)
		}

		line := fmt.Sprintf("  %s%s", cursor, name)
		if e.task.Description != "" {
			line += "  " + descStyle.Render(e.task.Description)
		}
		cmds := e.task.AllCommands()
		if len(cmds) == 1 {
			line += "\n      " + cmdStyle.Render("$ "+cmds[0])
		} else if len(cmds) > 1 {
			line += "\n      " + cmdStyle.Render(fmt.Sprintf("$ %s  (+%d more)", cmds[0], len(cmds)-1))
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(helpStyle.Render("\n↑/↓ navigate  •  enter select  •  q quit"))
	return b.String()
}
