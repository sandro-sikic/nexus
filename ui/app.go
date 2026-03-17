package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
	"runner/runner"
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

	// Menu sub-models (only one active per ui_mode)
	list  ListModel
	fuzzy FuzzyModel
	group GroupModel

	// Output sub-models
	output OutputModel
	bg     BackgroundModel

	quitting bool
	err      string
}

func NewApp(cfg *config.Config) AppModel {
	return newApp(cfg, "")
}

func newApp(cfg *config.Config, cfgPath string) AppModel {
	m := AppModel{cfg: cfg, cfgPath: cfgPath}
	switch cfg.UIMode {
	case config.UIModeList:
		m.list = NewListModel(cfg)
	case config.UIModeFuzzy:
		m.fuzzy = NewFuzzyModel(cfg)
	case config.UIModeGroup:
		m.group = NewGroupModel(cfg)
	}
	return m
}

func (m AppModel) Init() tea.Cmd {
	switch m.cfg.UIMode {
	case config.UIModeList:
		return m.list.Init()
	case config.UIModeFuzzy:
		return m.fuzzy.Init()
	case config.UIModeGroup:
		return m.group.Init()
	}
	return nil
}

// selectedCmd returns whichever menu model has a selection.
func (m *AppModel) selectedCmd() *config.Command {
	switch m.cfg.UIMode {
	case config.UIModeList:
		return m.list.Selected()
	case config.UIModeFuzzy:
		return m.fuzzy.Selected()
	case config.UIModeGroup:
		return m.group.Selected()
	}
	return nil
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
			// Reset selections
			m.list.selected = nil
			m.fuzzy.selected = nil
			m.group.selected = nil
			return m, nil
		case "esc":
			if m.state != stateMenu {
				m.state = stateMenu
				m.list.selected = nil
				m.fuzzy.selected = nil
				m.group.selected = nil
				return m, nil
			}
		}
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
	switch m.cfg.UIMode {
	case config.UIModeList:
		m.list, cmd = m.list.Update(msg)
	case config.UIModeFuzzy:
		m.fuzzy, cmd = m.fuzzy.Update(msg)
	case config.UIModeGroup:
		m.group, cmd = m.group.Update(msg)
	}

	// Check if a command was selected
	if sel := m.selectedCmd(); sel != nil {
		return m.launch(*sel)
	}

	return m, cmd
}

func (m AppModel) launch(sel config.Command) (tea.Model, tea.Cmd) {
	// Persist the selected index for list mode so the next run pre-selects it.
	if m.cfg.UIMode == config.UIModeList && m.cfgPath != "" {
		// Best-effort: ignore save errors so a read-only config doesn't block execution.
		_ = config.SaveLastIndex(m.cfgPath, m.list.cursor)
	}

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

	// Menu
	switch m.cfg.UIMode {
	case config.UIModeList:
		return m.list.View()
	case config.UIModeFuzzy:
		return m.fuzzy.View()
	case config.UIModeGroup:
		return m.group.View()
	}
	return ""
}

// Run starts the TUI and handles any post-TUI handoff.
func Run(cfg *config.Config, cfgPath string) error {
	app := newApp(cfg, cfgPath)
	p := tea.NewProgram(app, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	// Check if we need to do a handoff
	if final, ok := finalModel.(AppModel); ok {
		if final.selectedCmd() != nil && final.cfg != nil {
			// Should not happen – handled via handoffMsg
		}
		_ = final
	}

	return nil
}

// RunWithHandoff starts the program and performs handoff after the TUI exits.
func RunWithHandoff(cfg *config.Config, cfgPath string) error {
	app := newApp(cfg, cfgPath)
	p := tea.NewProgram(app, tea.WithAltScreen())

	rawModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	// If a handoff was requested, the model carries it
	if final, ok := rawModel.(AppModel); ok && final.quitting {
		// Nothing more to do; handoff already happened or user quit
		_ = final
	}
	return nil
}

// handleHandoff is called outside the TUI after it exits for handoff mode.
func HandleHandoff(cmd config.Command) error {
	fmt.Fprintf(os.Stderr, "Running: %s\n$ %s\n\n", cmd.Name, cmd.Command)
	return runner.Handoff(cmd)
}
