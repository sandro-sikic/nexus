package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nexus/config"
	"nexus/runner"
)

// outputLineMsg carries a single log line into the model.
type outputLineMsg runner.LogLine

// outputDoneMsg signals the process has finished.
type outputDoneMsg struct{ err error }

// OutputModel streams task output inside the TUI.
type OutputModel struct {
	task    config.Task
	lines   []runner.LogLine
	done    bool
	err     error
	width   int
	height  int
	offset  int // scroll offset
	spinner spinner.Model
}

func NewOutputModel(task config.Task) OutputModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	return OutputModel{task: task, width: 80, height: 24, spinner: s}
}

func (m OutputModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startStream())
}

func (m OutputModel) startStream() tea.Cmd {
	return func() tea.Msg {
		lines := make(chan runner.LogLine, 64)
		// Stream handles both foreground and background actions
		if err := runner.Stream(m.task, lines); err != nil {
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

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

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

	header := titleStyle.Render(fmt.Sprintf("Running: %s", m.task.Name))
	b.WriteString(header + "\n")

	// Show actions list
	if len(m.task.Actions) == 1 {
		// Single action: show simple format without numbering
		action := m.task.Actions[0]
		b.WriteString(cmdStyle.Render(fmt.Sprintf("  $ %s", action.Command)) + "\n")
	} else {
		// Multiple actions: show numbered steps
		for i, action := range m.task.Actions {
			prefix := "  "
			if action.Background {
				prefix = "[BG]"
			}
			b.WriteString(cmdStyle.Render(fmt.Sprintf("%s [%d] $ %s", prefix, i+1, action.Command)) + "\n")
		}
	}
	b.WriteString("\n")

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
		b.WriteString(helpStyle.Render(fmt.Sprintf("\n↑/↓ to scroll  •  %s running…", m.spinner.View())))
	}

	return b.String()
}

// BackgroundModel shows a background process log panel.
type BackgroundModel struct {
	task    config.Task
	proc    *runner.BackgroundProc
	lines   []runner.LogLine
	done    bool
	width   int
	height  int
	offset  int
	ticker  *time.Ticker
	spinner spinner.Model
}

func NewBackgroundModel(task config.Task) (BackgroundModel, error) {
	proc, err := runner.RunBackground(task)
	if err != nil {
		return BackgroundModel{}, err
	}
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	return BackgroundModel{
		task:    task,
		proc:    proc,
		width:   80,
		height:  24,
		spinner: s,
	}, nil
}

func (m BackgroundModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, bgTickCmd(m.proc.Lines))
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
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
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

	b.WriteString(titleStyle.Render(fmt.Sprintf("[BG] %s (%s)", m.task.Name, status)) + "\n")
	cmds := m.task.AllCommands()
	if len(cmds) == 1 {
		b.WriteString(cmdStyle.Render("$ "+cmds[0]) + "\n\n")
	} else {
		for i, cmd := range cmds {
			b.WriteString(cmdStyle.Render(fmt.Sprintf("  [%d] $ %s", i+1, cmd)) + "\n")
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

	if !m.done {
		b.WriteString(helpStyle.Render(fmt.Sprintf("\n↑/↓ to scroll  •  %s  •  q/esc to go back", m.spinner.View())))
	} else {
		b.WriteString(helpStyle.Render("\n↑/↓ to scroll  •  q/esc to go back"))
	}
	return b.String()
}
