package ui

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

// FuzzyModel provides a type-to-filter command selector.
type FuzzyModel struct {
	title    string
	all      []config.Command
	filtered []config.Command
	query    string
	cursor   int
	selected *config.Command
	width    int
	height   int
}

func NewFuzzyModel(cfg *config.Config) FuzzyModel {
	m := FuzzyModel{
		title:  cfg.Title,
		all:    cfg.Commands,
		width:  80,
		height: 24,
	}
	m.filtered = m.all
	return m
}

func (m FuzzyModel) Init() tea.Cmd { return nil }

func (m FuzzyModel) Selected() *config.Command { return m.selected }

func fuzzyMatch(query, target string) bool {
	query = strings.ToLower(query)
	target = strings.ToLower(target)
	qi := 0
	for _, r := range target {
		if qi < len(query) && unicode.ToLower(r) == rune(query[qi]) {
			qi++
		}
	}
	return qi == len(query)
}

func (m *FuzzyModel) applyFilter() {
	if m.query == "" {
		m.filtered = m.all
		return
	}
	var out []config.Command
	for _, c := range m.all {
		matched := fuzzyMatch(m.query, c.Name) || fuzzyMatch(m.query, c.Description)
		if !matched {
			for _, step := range c.Steps() {
				if fuzzyMatch(m.query, step) {
					matched = true
					break
				}
			}
		}
		if matched {
			out = append(out, c)
		}
	}
	m.filtered = out
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m FuzzyModel) Update(msg tea.Msg) (FuzzyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.applyFilter()
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				cmd := m.filtered[m.cursor]
				m.selected = &cmd
			}
		case tea.KeyRunes:
			m.query += msg.String()
			m.applyFilter()
		}
	}
	return m, nil
}

func (m FuzzyModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title) + "\n")
	b.WriteString("> " + m.query + "█\n\n")

	if len(m.filtered) == 0 {
		b.WriteString(descStyle.Render("  no matches") + "\n")
	}

	for i, cmd := range m.filtered {
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
		steps := cmd.Steps()
		if len(steps) == 1 {
			line += "\n    " + cmdStyle.Render("$ "+steps[0])
		} else if len(steps) > 1 {
			line += "\n    " + cmdStyle.Render(fmt.Sprintf("$ %s  (+%d more steps)", steps[0], len(steps)-1))
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(helpStyle.Render("\ntype to filter  •  ↑/↓ navigate  •  enter select  •  q quit"))
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
