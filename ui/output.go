package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nexus/config"
	"nexus/runner"
)

// outputLineMsg carries a single log line into the model.
type outputLineMsg runner.LogLine

// outputDoneMsg signals the process has finished.
type outputDoneMsg struct{ err error }

// OutputModel streams command output inside the TUI.
type OutputModel struct {
	cmd    config.Command
	lines  []runner.LogLine
	done   bool
	err    error
	width  int
	height int
	offset int // scroll offset
}

func NewOutputModel(cmd config.Command) OutputModel {
	return OutputModel{cmd: cmd, width: 80, height: 24}
}

func (m OutputModel) Init() tea.Cmd {
	return m.startStream()
}

func (m OutputModel) startStream() tea.Cmd {
	return func() tea.Msg {
		lines := make(chan runner.LogLine, 64)
		if err := runner.Stream(m.cmd, lines); err != nil {
			return outputDoneMsg{err: err}
		}
		// Drain first line to kick things off; subsequent lines polled via tickCmd.
		// We return the channel via a wrapper so we can read it in Update.
		return streamStartMsg{lines: lines}
	}
}

type streamStartMsg struct{ lines chan runner.LogLine }

// tickCmd reads one line from the channel each tick.
func tickCmd(lines chan runner.LogLine) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-lines
		if !ok {
			return outputDoneMsg{}
		}
		return struct {
			line  runner.LogLine
			lines chan runner.LogLine
		}{line, lines}
	}
}

func (m OutputModel) Update(msg tea.Msg) (OutputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case streamStartMsg:
		return m, tickCmd(msg.lines)

	case struct {
		line  runner.LogLine
		lines chan runner.LogLine
	}:
		m.lines = append(m.lines, msg.line)
		// Auto-scroll to bottom
		visible := m.height - 6
		if visible < 1 {
			visible = 1
		}
		if len(m.lines) > visible {
			m.offset = len(m.lines) - visible
		}
		return m, tickCmd(msg.lines)

	case outputDoneMsg:
		m.done = true
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			visible := m.height - 6
			if visible < 1 {
				visible = 1
			}
			if m.offset < len(m.lines)-visible {
				m.offset++
			}
		}
	}
	return m, nil
}

func (m OutputModel) View() string {
	var b strings.Builder

	header := titleStyle.Render(fmt.Sprintf("Running: %s", m.cmd.Name))
	b.WriteString(header + "\n")
	steps := m.cmd.Steps()
	if len(steps) == 1 {
		b.WriteString(cmdStyle.Render("$ "+steps[0]) + "\n\n")
	} else {
		for i, step := range steps {
			b.WriteString(cmdStyle.Render(fmt.Sprintf("  [%d] $ %s", i+1, step)) + "\n")
		}
		b.WriteString("\n")
	}

	visible := m.height - 6
	if visible < 1 {
		visible = 1
	}

	start := m.offset
	end := start + visible
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for _, l := range m.lines[start:end] {
		if l.IsErr {
			b.WriteString(errorStyle.Render(l.Text) + "\n")
		} else {
			b.WriteString(l.Text + "\n")
		}
	}

	if m.done {
		if m.err != nil {
			b.WriteString("\n" + errorStyle.Render("Error: "+m.err.Error()))
		} else {
			b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("Process finished."))
		}
		b.WriteString(helpStyle.Render("\nPress q or esc to go back"))
	} else {
		b.WriteString(helpStyle.Render("\n↑/↓ to scroll  •  running…"))
	}

	return b.String()
}

// BackgroundModel shows a background process log panel.
type BackgroundModel struct {
	cmd    config.Command
	proc   *runner.BackgroundProc
	lines  []runner.LogLine
	done   bool
	width  int
	height int
	offset int
	ticker *time.Ticker
}

func NewBackgroundModel(cmd config.Command) (BackgroundModel, error) {
	proc, err := runner.RunBackground(cmd)
	if err != nil {
		return BackgroundModel{}, err
	}
	return BackgroundModel{
		cmd:    cmd,
		proc:   proc,
		width:  80,
		height: 24,
	}, nil
}

func (m BackgroundModel) Init() tea.Cmd {
	return bgTickCmd(m.proc.Lines)
}

func bgTickCmd(lines chan runner.LogLine) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-lines
		if !ok {
			return outputDoneMsg{}
		}
		return struct {
			line  runner.LogLine
			lines chan runner.LogLine
		}{line, lines}
	}
}

func (m BackgroundModel) Update(msg tea.Msg) (BackgroundModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case struct {
		line  runner.LogLine
		lines chan runner.LogLine
	}:
		m.lines = append(m.lines, msg.line)
		visible := m.height - 6
		if visible < 1 {
			visible = 1
		}
		if len(m.lines) > visible {
			m.offset = len(m.lines) - visible
		}
		return m, bgTickCmd(msg.lines)
	case outputDoneMsg:
		m.done = true
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			visible := m.height - 6
			if visible < 1 {
				visible = 1
			}
			if m.offset < len(m.lines)-visible {
				m.offset++
			}
		}
	}
	return m, nil
}

func (m BackgroundModel) View() string {
	var b strings.Builder

	status := "running"
	if m.done {
		status = "done"
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("[BG] %s (%s)", m.cmd.Name, status)) + "\n")
	steps := m.cmd.Steps()
	if len(steps) == 1 {
		b.WriteString(cmdStyle.Render("$ "+steps[0]) + "\n\n")
	} else {
		for i, step := range steps {
			b.WriteString(cmdStyle.Render(fmt.Sprintf("  [%d] $ %s", i+1, step)) + "\n")
		}
		b.WriteString("\n")
	}

	visible := m.height - 6
	if visible < 1 {
		visible = 1
	}
	start := m.offset
	end := start + visible
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for _, l := range m.lines[start:end] {
		if l.IsErr {
			b.WriteString(errorStyle.Render(l.Text) + "\n")
		} else {
			b.WriteString(l.Text + "\n")
		}
	}

	b.WriteString(helpStyle.Render("\n↑/↓ to scroll  •  q/esc to go back"))
	return b.String()
}
