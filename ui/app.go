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

	quitting   bool
	handoffCmd *config.Command // set when a handoff is pending
	err        string
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

	// Handle handoff message: store the command and quit the TUI.
	if hm, ok := msg.(handoffMsg); ok {
		m.handoffCmd = &hm.cmd
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

func (m AppModel) launch(sel config.Command) (tea.Model, tea.Cmd) {
	switch sel.RunMode {
	case config.RunModeHandoff:
		// For handoff we need to quit the TUI first, then exec
		m.quitting = true
		return m, tea.Sequence(
			tea.ExitAltScreen,
			func() tea.Msg { return handoffMsg{cmd: sel} },
		)

	case config.RunModeStream:
		m.output = NewOutputModel(sel)
		m.state = stateOutput
		return m, m.output.Init()

	case config.RunModeBackground:
		bg, err := NewBackgroundModel(sel)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.bg = bg
		m.state = stateBG
		return m, m.bg.Init()
	}
	return m, nil
}

type handoffMsg struct{ cmd config.Command }

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

// RunFirstMatch fuzzy-matches query against all commands and immediately
// executes the first hit, bypassing the TUI entirely.
func RunFirstMatch(cfg *config.Config, query string) error {
	for _, cmd := range cfg.Commands {
		matched := fuzzyMatch(query, cmd.Name) || fuzzyMatch(query, cmd.Description)
		if !matched {
			for _, step := range cmd.AllSteps() {
				if fuzzyMatch(query, step) {
					matched = true
					break
				}
			}
		}
		if matched {
			return HandleHandoff(cmd)
		}
	}
	return fmt.Errorf("no command matched %q", query)
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
	if final, ok := finalModel.(AppModel); ok && final.handoffCmd != nil {
		return HandleHandoff(*final.handoffCmd)
	}

	return nil
}

// handleHandoff is called outside the TUI after it exits for handoff mode.
func HandleHandoff(cmd config.Command) error {
	fmt.Fprintf(os.Stderr, "Running: %s\n$ %s\n\n", cmd.Name, cmd.Command)
	return runner.Handoff(cmd)
}
