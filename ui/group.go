package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

type groupEntry struct {
	isHeader bool
	group    string
	cmd      *config.Command
}

// GroupModel presents commands organized by their group field.
type GroupModel struct {
	title    string
	entries  []groupEntry
	cursor   int
	selected *config.Command
	width    int
	height   int
}

func NewGroupModel(cfg *config.Config) GroupModel {
	// Build entries with group headers
	seen := map[string]bool{}
	var entries []groupEntry

	// Collect group order
	var groupOrder []string
	groupCmds := map[string][]config.Command{}
	for _, cmd := range cfg.Commands {
		g := cmd.Group
		if g == "" {
			g = "General"
		}
		if !seen[g] {
			seen[g] = true
			groupOrder = append(groupOrder, g)
		}
		groupCmds[g] = append(groupCmds[g], cmd)
	}

	for _, g := range groupOrder {
		entries = append(entries, groupEntry{isHeader: true, group: g})
		for i := range groupCmds[g] {
			c := groupCmds[g][i]
			entries = append(entries, groupEntry{cmd: &c})
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

func (m GroupModel) Selected() *config.Command { return m.selected }

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
				m.selected = m.entries[m.cursor].cmd
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

		name := normalStyle.Render(e.cmd.Name)
		if i == m.cursor {
			name = selectedStyle.Render(e.cmd.Name)
		}

		line := fmt.Sprintf("  %s%s", cursor, name)
		if e.cmd.Description != "" {
			line += "  " + descStyle.Render(e.cmd.Description)
		}
		steps := e.cmd.Steps()
		if len(steps) == 1 {
			line += "\n      " + cmdStyle.Render("$ "+steps[0])
		} else if len(steps) > 1 {
			line += "\n      " + cmdStyle.Render(fmt.Sprintf("$ %s  (+%d more steps)", steps[0], len(steps)-1))
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(helpStyle.Render("\n↑/↓ navigate  •  enter select  •  q quit"))
	return b.String()
}
