package ui

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
)

// FuzzyModel provides a combined fuzzy-search + group view.
// When no query is active, commands are displayed grouped by their Group field.
// When a query is typed, a flat filtered list is shown instead.
type FuzzyModel struct {
	title    string
	all      []config.Command
	filtered []config.Command
	query    string
	cursor   int
	selected *config.Command
	width    int
	height   int

	// group view state (used when query is empty)
	entries []groupEntry // flat: [header, cmd, cmd, header, cmd, ...]
	gCursor int          // cursor index into entries (non-header only)

	// scrollOffset is the index into renderLines() of the first visible line.
	// It is updated whenever the cursor moves outside the visible window.
	scrollOffset int
}

func NewFuzzyModel(cfg *config.Config) FuzzyModel {
	m := FuzzyModel{
		title:  cfg.Title,
		all:    cfg.Commands,
		width:  80,
		height: 24,
	}
	m.filtered = m.all
	m.entries = buildGroupEntries(cfg.Commands)
	m.gCursor = firstCmdEntry(m.entries)
	return m
}

// buildGroupEntries constructs a flat slice of groupEntry values from commands.
func buildGroupEntries(cmds []config.Command) []groupEntry {
	seen := map[string]bool{}
	var groupOrder []string
	groupCmds := map[string][]config.Command{}

	for _, cmd := range cmds {
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

	var entries []groupEntry
	for _, g := range groupOrder {
		entries = append(entries, groupEntry{isHeader: true, group: g})
		for i := range groupCmds[g] {
			c := groupCmds[g][i]
			entries = append(entries, groupEntry{cmd: &c})
		}
	}
	return entries
}

// firstCmdEntry returns the index of the first non-header entry, or 0.
func firstCmdEntry(entries []groupEntry) int {
	for i, e := range entries {
		if !e.isHeader {
			return i
		}
	}
	return 0
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

// moveGroupCursor moves the group-view cursor by delta, skipping headers.
func (m *FuzzyModel) moveGroupCursor(delta int) {
	next := m.gCursor + delta
	for next >= 0 && next < len(m.entries) {
		if !m.entries[next].isHeader {
			m.gCursor = next
			return
		}
		next += delta
	}
}

// headerText returns the fixed top section as a rendered string.
func (m FuzzyModel) headerText() string {
	return titleStyle.Render(m.title) + "\n\n" + "> " + m.query + "█\n"
}

// footerText returns the fixed bottom section as a rendered string.
// Includes the blank separator line before the help bar.
func footerText() string {
	return "\n" + helpStyle.Render("type to filter  •  ↑/↓ navigate  •  enter select  •  q quit")
}

// countLines counts the number of terminal lines a rendered string occupies.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// listBudget returns how many list lines fit between the header and footer.
func (m FuzzyModel) listBudget() int {
	headerH := countLines(m.headerText())
	footerH := countLines(footerText())
	budget := m.height - headerH - footerH
	if budget < 1 {
		budget = 1
	}
	return budget
}

// adjustScroll keeps scrollOffset in sync with the cursor so that:
//   - the cursor line is always inside [scrollOffset, scrollOffset+budget)
//   - the window never scrolls past the top (scrollOffset >= 0)
func (m *FuzzyModel) adjustScroll() {
	budget := m.listBudget()
	cur := m.cursorLine()

	if cur < m.scrollOffset {
		m.scrollOffset = cur
	} else if cur >= m.scrollOffset+budget {
		m.scrollOffset = cur - budget + 1
	}

	total := len(m.renderLines())
	maxOffset := total - budget
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m FuzzyModel) Update(msg tea.Msg) (FuzzyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustScroll()

	case tea.KeyMsg:
		filtering := m.query != ""

		switch msg.Type {
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.scrollOffset = 0
				m.applyFilter()
			}
		case tea.KeyUp:
			if filtering {
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				m.moveGroupCursor(-1)
			}
			m.adjustScroll()
		case tea.KeyDown:
			if filtering {
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			} else {
				m.moveGroupCursor(1)
			}
			m.adjustScroll()
		case tea.KeyEnter:
			if filtering {
				if len(m.filtered) > 0 {
					cmd := m.filtered[m.cursor]
					m.selected = &cmd
				}
			} else {
				if m.gCursor < len(m.entries) && !m.entries[m.gCursor].isHeader {
					m.selected = m.entries[m.gCursor].cmd
				}
			}
		case tea.KeyRunes:
			m.query += msg.String()
			m.scrollOffset = 0
			m.applyFilter()
		}

		// Space selects in group mode (when not filtering)
		if msg.Type == tea.KeySpace && !filtering {
			if m.gCursor < len(m.entries) && !m.entries[m.gCursor].isHeader {
				m.selected = m.entries[m.gCursor].cmd
			}
		}
	}
	return m, nil
}

// renderLines builds the full list of rendered lines for the current mode,
// returning them as a slice so View can window them to fit the terminal.
func (m FuzzyModel) renderLines() []string {
	var lines []string

	if m.query != "" {
		// Fuzzy filtered list
		if len(m.filtered) == 0 {
			lines = append(lines, descStyle.Render("  no matches"))
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
			lines = append(lines, line)
			steps := cmd.Steps()
			if len(steps) == 1 {
				lines = append(lines, "    "+cmdStyle.Render("$ "+steps[0]))
			} else if len(steps) > 1 {
				lines = append(lines, "    "+cmdStyle.Render(fmt.Sprintf("$ %s  (+%d more steps)", steps[0], len(steps)-1)))
			}
		}
	} else {
		// Grouped view
		for i, e := range m.entries {
			if e.isHeader {
				// blank line before every group header (spacing)
				if len(lines) > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, groupHeaderStyle.Render("── "+e.group+" ──"))
				continue
			}
			cursor := "  "
			if i == m.gCursor {
				cursor = cursorStyle.Render("▶ ")
			}
			name := normalStyle.Render(e.cmd.Name)
			if i == m.gCursor {
				name = selectedStyle.Render(e.cmd.Name)
			}
			line := fmt.Sprintf("  %s%s", cursor, name)
			if e.cmd.Description != "" {
				line += "  " + descStyle.Render(e.cmd.Description)
			}
			lines = append(lines, line)
			steps := e.cmd.Steps()
			if len(steps) == 1 {
				lines = append(lines, "      "+cmdStyle.Render("$ "+steps[0]))
			} else if len(steps) > 1 {
				lines = append(lines, "      "+cmdStyle.Render(fmt.Sprintf("$ %s  (+%d more steps)", steps[0], len(steps)-1)))
			}
		}
	}

	return lines
}

// cursorLine returns the rendered-line index of the active cursor so View
// knows which line must stay visible.
func (m FuzzyModel) cursorLine() int {
	if m.query != "" {
		// Each item takes 1 or 2 rendered lines (name + optional cmd line).
		// Walk the filtered list to find the line index of the selected item.
		line := 0
		for i, cmd := range m.filtered {
			if i == m.cursor {
				return line
			}
			line++ // name line
			if len(cmd.Steps()) > 0 {
				line++ // cmd preview line
			}
		}
		return line
	}

	// Grouped view: walk entries to find the rendered line of gCursor,
	// accounting for blank lines inserted before group headers (except the first).
	line := 0
	headerCount := 0
	for i, e := range m.entries {
		if e.isHeader {
			if headerCount > 0 {
				// blank line before this header (except for the very first header)
				line++
			}
			headerCount++
		}
		// Now line points to the start of this entry's content (header line if header,
		// name line if name). Since gCursor never points to a header, we can safely
		// return line when we hit the cursor.
		if i == m.gCursor {
			return line
		}
		if e.isHeader {
			line++ // consume header line
		} else {
			line++ // consume name line
			if len(e.cmd.Steps()) > 0 {
				line++ // consume preview line
			}
		}
	}
	return line
}

func (m FuzzyModel) View() string {
	allLines := m.renderLines()
	budget := m.listBudget()

	// Slice the visible window using the scroll offset maintained by adjustScroll.
	start := m.scrollOffset
	if start > len(allLines) {
		start = len(allLines)
	}
	end := start + budget
	if end > len(allLines) {
		end = len(allLines)
	}
	visible := allLines[start:end]

	var b strings.Builder
	b.WriteString(m.headerText())
	b.WriteString("\n")
	if len(visible) > 0 {
		b.WriteString(strings.Join(visible, "\n"))
		b.WriteString("\n")
	}
	b.WriteString(footerText())
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
