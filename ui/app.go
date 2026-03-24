package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
	"nexus/runner"
)

type appState int

const (
	stateMenu   appState = iota
	stateOutput          // streaming output view
	stateBG              // background log view
)

// AppModel is the root bubbletea model.
type AppModel struct {
	cfg     *config.Config
	cfgPath string // path to the config file, used to persist state
	state   appState

	// Menu — always the combined fuzzy+group model
	fuzzy FuzzyModel

	// Output sub-models
	output OutputModel
	bg     BackgroundModel

	quitting    bool
	handoffTask *config.Task // set when a handoff is pending
	err         string
}

func NewApp(cfg *config.Config) AppModel {
	return newApp(cfg, "")
}

func newApp(cfg *config.Config, cfgPath string) AppModel {
	return AppModel{
		cfg:     cfg,
		cfgPath: cfgPath,
		fuzzy:   NewFuzzyModel(cfg),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.fuzzy.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global quit
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "ctrl+c":
			if m.state == stateMenu {
				m.quitting = true
				return m, tea.Quit
			}
			// In output/bg state, go back to menu
			m.state = stateMenu
			m.fuzzy.selected = nil
			return m, nil
		case "esc":
			if m.state != stateMenu {
				m.state = stateMenu
				m.fuzzy.selected = nil
				return m, nil
			}
		}
	}

	// Handle handoff message: store the task and quit the TUI.
	if hm, ok := msg.(handoffMsg); ok {
		m.handoffTask = &hm.task
		return m, tea.Quit
	}

	switch m.state {
	case stateMenu:
		return m.updateMenu(msg)
	case stateOutput:
		return m.updateOutput(msg)
	case stateBG:
		return m.updateBG(msg)
	}
	return m, nil
}

func (m AppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.fuzzy, cmd = m.fuzzy.Update(msg)

	if sel := m.fuzzy.Selected(); sel != nil {
		return m.launch(*sel)
	}

	return m, cmd
}

func (m AppModel) launch(sel config.Task) (tea.Model, tea.Cmd) {
	// Check if the task has handoff enabled
	if sel.Handoff {
		// For handoff we need to quit the TUI first, then exec the last action
		m.quitting = true
		return m, tea.Sequence(
			tea.ExitAltScreen,
			func() tea.Msg { return handoffMsg{task: sel} },
		)
	}

	// Check if any action has background enabled
	if sel.HasBackgroundActions() {
		bg, err := NewBackgroundModel(sel)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.bg = bg
		m.state = stateBG
		return m, m.bg.Init()
	}

	// Default: stream output inside TUI
	m.output = NewOutputModel(sel)
	m.state = stateOutput
	return m, m.output.Init()
}

type handoffMsg struct{ task config.Task }

func (m AppModel) updateOutput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.output, cmd = m.output.Update(msg)
	return m, cmd
}

func (m AppModel) updateBG(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.bg, cmd = m.bg.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	if m.quitting {
		return ""
	}
	if m.err != "" {
		return errorStyle.Render("Error: "+m.err) + "\n" + helpStyle.Render("Press q to quit")
	}

	switch m.state {
	case stateOutput:
		return m.output.View()
	case stateBG:
		return m.bg.View()
	}

	return m.fuzzy.View()
}

// RunFirstMatch fuzzy-matches query against all tasks and immediately
// executes the first hit, bypassing the TUI entirely.
func RunFirstMatch(cfg *config.Config, query string) error {
	for _, task := range cfg.Tasks {
		matched := fuzzyMatch(query, task.Name) || fuzzyMatch(query, task.Description)
		if !matched {
			for _, cmd := range task.AllCommands() {
				if fuzzyMatch(query, cmd) {
					matched = true
					break
				}
			}
		}
		if matched {
			return HandleHandoff(task)
		}
	}
	return fmt.Errorf("no task matched %q", query)
}

// Run starts the TUI and handles any post-TUI handoff.
func Run(cfg *config.Config, cfgPath string) error {
	app := newApp(cfg, cfgPath)
	p := tea.NewProgram(app, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	// If a handoff was requested, execute it now that the TUI has exited.
	if final, ok := finalModel.(AppModel); ok && final.handoffTask != nil {
		return HandleHandoff(*final.handoffTask)
	}

	return nil
}

// HandleHandoff is called outside the TUI after it exits for handoff mode.
func HandleHandoff(task config.Task) error {
	fmt.Fprintf(os.Stderr, "Running: %s\n", task.Name)

	// Stream all actions except the last one (if handoff)
	if len(task.Actions) > 1 && task.Handoff {
		fmt.Fprintln(os.Stderr, "Executing setup actions...")
		lines := make(chan runner.LogLine, 64)
		if err := runner.Stream(task, lines); err != nil {
			return err
		}
		// Drain the lines channel
		for line := range lines {
			if line.IsErr {
				fmt.Fprintln(os.Stderr, line.Text)
			} else {
				fmt.Fprintln(os.Stderr, line.Text)
			}
		}
		fmt.Fprintln(os.Stderr)
	}

	// Handoff the last action to the raw terminal
	if len(task.Actions) > 0 {
		lastAction := task.Actions[len(task.Actions)-1]
		fmt.Fprintf(os.Stderr, "$ %s\n\n", lastAction.Command)
	}

	return runner.HandoffLastAction(task)
}
