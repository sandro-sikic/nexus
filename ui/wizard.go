package ui

// Wizard is a multi-step TUI that walks the user through creating a runner.yaml
// from scratch. Steps:
//
//  1. Welcome / missing-file notice
//  2. Project title
//  3. UI mode  (list | fuzzy | group)
//  4. Default run mode  (stream | handoff | background)
//  5. Add first command  → sub-steps: name, description, command, dir, group,
//     per-command run_mode override (or inherit)
//  6. "Add another command?" → yes loops back to 5, no proceeds
//  7. Summary + confirm save

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"runner/config"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	wizTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)

	wizSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				MarginBottom(1)

	wizLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33"))

	wizValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	wizSelectedOptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)

	wizOptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	wizCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	wizHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")).
			MarginTop(1)

	wizErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	wizSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	wizWarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	wizDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	wizSeparator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")).
			Render(strings.Repeat("─", 50))
)

// ── Step enum ─────────────────────────────────────────────────────────────────

type wizStep int

const (
	wizStepWelcome    wizStep = iota // 0 – info screen, press enter
	wizStepTitle                     // 1 – text input
	wizStepUIMode                    // 2 – option picker
	wizStepRunMode                   // 3 – option picker
	wizStepCmdName                   // 4 – text input
	wizStepCmdDesc                   // 5 – text input (optional)
	wizStepCmdCommand                // 6 – text input
	wizStepCmdDir                    // 7 – text input (optional)
	wizStepCmdGroup                  // 8 – text input (optional)
	wizStepCmdRunMode                // 9 – option picker (inherit | stream | handoff | background)
	wizStepAddAnother                // 10 – yes/no picker
	wizStepSummary                   // 11 – review + confirm
	wizStepDone                      // 12 – finished
)

// ── Option picker ─────────────────────────────────────────────────────────────

type wizOption struct {
	value string
	label string
	desc  string
}

// ── WizardModel ───────────────────────────────────────────────────────────────

// WizardModel is the root bubbletea model for the setup wizard.
type WizardModel struct {
	step      wizStep
	inputBuf  string // current text-input buffer
	optCursor int    // cursor in an option list
	validErr  string // inline validation message
	aborted   bool   // user pressed ctrl+c
	saved     bool   // file was written successfully
	saveErr   string // error from write
	savePath  string // destination path

	// collected global config values
	cfgTitle   string
	cfgUIMode  config.UIMode
	cfgRunMode config.RunMode

	// command being built
	cmdName    string
	cmdDesc    string
	cmdCommand string
	cmdDir     string
	cmdGroup   string
	cmdRunMode config.RunMode // "" means "inherit"

	// accumulated commands
	commands []config.Command

	width  int
	height int
}

// NewWizard creates a fresh WizardModel. savePath is where the file will be written.
func NewWizard(savePath string) WizardModel {
	return WizardModel{
		step:     wizStepWelcome,
		savePath: savePath,
		width:    80,
		height:   24,
	}
}

// Result returns the assembled Config after the wizard completes.
// Returns nil if the wizard was aborted.
func (m WizardModel) Result() *config.Config {
	if m.aborted || !m.saved {
		return nil
	}
	return m.buildConfig()
}

// Aborted returns true if the user cancelled.
func (m WizardModel) Aborted() bool { return m.aborted }

// Saved returns true if the file was written without error.
func (m WizardModel) Saved() bool { return m.saved }

func (m WizardModel) buildConfig() *config.Config {
	return &config.Config{
		Title:    m.cfgTitle,
		UIMode:   m.cfgUIMode,
		RunMode:  m.cfgRunMode,
		Commands: m.commands,
	}
}

// ── Options tables ────────────────────────────────────────────────────────────

var uiModeOptions = []wizOption{
	{value: string(config.UIModeList), label: "list", desc: "Arrow-key navigable list"},
	{value: string(config.UIModeFuzzy), label: "fuzzy", desc: "Type to filter commands"},
	{value: string(config.UIModeGroup), label: "group", desc: "Commands grouped by category"},
}

var runModeOptions = []wizOption{
	{value: string(config.RunModeStream), label: "stream", desc: "Stream output inside the TUI"},
	{value: string(config.RunModeHandoff), label: "handoff", desc: "Hand off to the raw terminal"},
	{value: string(config.RunModeBackground), label: "background", desc: "Run in background, tail logs in TUI"},
}

var cmdRunModeOptions = []wizOption{
	{value: "", label: "inherit", desc: "Use the project default"},
	{value: string(config.RunModeStream), label: "stream", desc: "Stream output inside the TUI"},
	{value: string(config.RunModeHandoff), label: "handoff", desc: "Hand off to the raw terminal"},
	{value: string(config.RunModeBackground), label: "background", desc: "Run in background, tail logs in TUI"},
}

var addAnotherOptions = []wizOption{
	{value: "yes", label: "Yes", desc: "Add another command"},
	{value: "no", label: "No", desc: "Finish and save"},
}

// currentOptions returns the option list for the current step (if any).
func (m WizardModel) currentOptions() []wizOption {
	switch m.step {
	case wizStepUIMode:
		return uiModeOptions
	case wizStepRunMode:
		return runModeOptions
	case wizStepCmdRunMode:
		return cmdRunModeOptions
	case wizStepAddAnother:
		return addAnotherOptions
	}
	return nil
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m WizardModel) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m WizardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global abort
	if msg.Type == tea.KeyCtrlC {
		m.aborted = true
		return m, tea.Quit
	}

	opts := m.currentOptions()
	isOptionStep := opts != nil
	isTextStep := !isOptionStep && m.step != wizStepWelcome && m.step != wizStepSummary && m.step != wizStepDone

	switch msg.Type {
	// ── Navigation in option pickers ──────────────────────────────────────
	case tea.KeyUp:
		if isOptionStep && m.optCursor > 0 {
			m.optCursor--
		}
		return m, nil

	case tea.KeyDown:
		if isOptionStep && m.optCursor < len(opts)-1 {
			m.optCursor++
		}
		return m, nil

	// ── Backspace in text inputs ──────────────────────────────────────────
	case tea.KeyBackspace:
		if isTextStep && len(m.inputBuf) > 0 {
			// handle multi-byte runes
			runes := []rune(m.inputBuf)
			m.inputBuf = string(runes[:len(runes)-1])
			m.validErr = ""
		}
		return m, nil

	// ── Enter: advance step ───────────────────────────────────────────────
	case tea.KeyEnter:
		return m.advance()

	// ── Rune input ────────────────────────────────────────────────────────
	case tea.KeyRunes:
		if isTextStep {
			m.inputBuf += msg.String()
			m.validErr = ""
		}
		return m, nil
	}

	return m, nil
}

// advance validates the current step, commits its value, and moves forward.
func (m WizardModel) advance() (tea.Model, tea.Cmd) {
	m.validErr = ""

	switch m.step {

	// ── Welcome ───────────────────────────────────────────────────────────
	case wizStepWelcome:
		m.inputBuf = ""
		m.step = wizStepTitle

	// ── Title ─────────────────────────────────────────────────────────────
	case wizStepTitle:
		title := strings.TrimSpace(m.inputBuf)
		if title == "" {
			title = "Runner"
		}
		m.cfgTitle = title
		m.inputBuf = ""
		m.optCursor = 0
		m.step = wizStepUIMode

	// ── UI mode ───────────────────────────────────────────────────────────
	case wizStepUIMode:
		m.cfgUIMode = config.UIMode(uiModeOptions[m.optCursor].value)
		m.optCursor = 0
		m.step = wizStepRunMode

	// ── Default run mode ──────────────────────────────────────────────────
	case wizStepRunMode:
		m.cfgRunMode = config.RunMode(runModeOptions[m.optCursor].value)
		m.inputBuf = ""
		m.optCursor = 0
		m.step = wizStepCmdName

	// ── Command: name ─────────────────────────────────────────────────────
	case wizStepCmdName:
		name := strings.TrimSpace(m.inputBuf)
		if name == "" {
			m.validErr = "Command name is required"
			return m, nil
		}
		m.cmdName = name
		m.inputBuf = ""
		m.step = wizStepCmdDesc

	// ── Command: description (optional) ──────────────────────────────────
	case wizStepCmdDesc:
		m.cmdDesc = strings.TrimSpace(m.inputBuf)
		m.inputBuf = ""
		m.step = wizStepCmdCommand

	// ── Command: shell command ────────────────────────────────────────────
	case wizStepCmdCommand:
		command := strings.TrimSpace(m.inputBuf)
		if command == "" {
			m.validErr = "Shell command is required"
			return m, nil
		}
		m.cmdCommand = command
		m.inputBuf = ""
		m.step = wizStepCmdDir

	// ── Command: working directory (optional) ────────────────────────────
	case wizStepCmdDir:
		m.cmdDir = strings.TrimSpace(m.inputBuf)
		m.inputBuf = ""
		// Only show group step when ui_mode = group
		if m.cfgUIMode == config.UIModeGroup {
			m.step = wizStepCmdGroup
		} else {
			m.optCursor = 0
			m.step = wizStepCmdRunMode
		}

	// ── Command: group (optional, only in group mode) ─────────────────────
	case wizStepCmdGroup:
		m.cmdGroup = strings.TrimSpace(m.inputBuf)
		m.inputBuf = ""
		m.optCursor = 0
		m.step = wizStepCmdRunMode

	// ── Command: per-command run mode override ────────────────────────────
	case wizStepCmdRunMode:
		m.cmdRunMode = config.RunMode(cmdRunModeOptions[m.optCursor].value)
		m.commitCommand()
		m.optCursor = 0
		m.step = wizStepAddAnother

	// ── Add another? ──────────────────────────────────────────────────────
	case wizStepAddAnother:
		if addAnotherOptions[m.optCursor].value == "yes" {
			m.resetCmdFields()
			m.optCursor = 0
			m.step = wizStepCmdName
		} else {
			m.step = wizStepSummary
		}

	// ── Summary: confirm ──────────────────────────────────────────────────
	case wizStepSummary:
		cfg := m.buildConfig()
		if err := config.Write(m.savePath, cfg); err != nil {
			m.saveErr = err.Error()
		} else {
			m.saved = true
		}
		m.step = wizStepDone

	// ── Done ──────────────────────────────────────────────────────────────
	case wizStepDone:
		return m, tea.Quit
	}

	return m, nil
}

// commitCommand appends the command being built to m.commands.
func (m *WizardModel) commitCommand() {
	rm := m.cmdRunMode
	if rm == "" {
		rm = m.cfgRunMode
	}
	m.commands = append(m.commands, config.Command{
		Name:        m.cmdName,
		Description: m.cmdDesc,
		Command:     m.cmdCommand,
		Dir:         m.cmdDir,
		Group:       m.cmdGroup,
		RunMode:     rm,
	})
}

// resetCmdFields clears the per-command scratch space.
func (m *WizardModel) resetCmdFields() {
	m.cmdName = ""
	m.cmdDesc = ""
	m.cmdCommand = ""
	m.cmdDir = ""
	m.cmdGroup = ""
	m.cmdRunMode = ""
	m.inputBuf = ""
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m WizardModel) View() string {
	switch m.step {
	case wizStepWelcome:
		return m.viewWelcome()
	case wizStepTitle:
		return m.viewTextInput(
			"Project Title",
			"A short name shown at the top of the menu.",
			"e.g. My Project",
			true,
		)
	case wizStepUIMode:
		return m.viewOptionPicker(
			"UI Mode",
			"How should commands be presented?",
			uiModeOptions,
		)
	case wizStepRunMode:
		return m.viewOptionPicker(
			"Default Run Mode",
			"How should commands be executed by default?\n"+
				wizDimStyle.Render("  (individual commands can override this)"),
			runModeOptions,
		)
	case wizStepCmdName:
		return m.viewTextInput(
			fmt.Sprintf("Command #%d — Name", len(m.commands)+1),
			"A short display name for this command.",
			"e.g. Dev Server",
			false,
		)
	case wizStepCmdDesc:
		return m.viewTextInput(
			fmt.Sprintf("Command #%d — Description", len(m.commands)+1),
			"Optional longer description shown next to the name.",
			"e.g. Start the dev server with hot reload  (leave blank to skip)",
			true,
		)
	case wizStepCmdCommand:
		return m.viewTextInput(
			fmt.Sprintf("Command #%d — Shell Command", len(m.commands)+1),
			"The shell command to execute.",
			"e.g. npm run dev",
			false,
		)
	case wizStepCmdDir:
		return m.viewTextInput(
			fmt.Sprintf("Command #%d — Working Directory", len(m.commands)+1),
			"Optional working directory (leave blank for CWD).",
			"e.g. ./frontend",
			true,
		)
	case wizStepCmdGroup:
		return m.viewTextInput(
			fmt.Sprintf("Command #%d — Group", len(m.commands)+1),
			"Group name for the grouped menu (leave blank for \"General\").",
			"e.g. Development",
			true,
		)
	case wizStepCmdRunMode:
		return m.viewOptionPicker(
			fmt.Sprintf("Command #%d — Run Mode", len(m.commands)+1),
			"Override the default run mode for this command, or inherit it.",
			cmdRunModeOptions,
		)
	case wizStepAddAnother:
		return m.viewOptionPicker(
			"Add Another Command?",
			fmt.Sprintf("%d command(s) added so far.", len(m.commands)),
			addAnotherOptions,
		)
	case wizStepSummary:
		return m.viewSummary()
	case wizStepDone:
		return m.viewDone()
	}
	return ""
}

// ── View helpers ──────────────────────────────────────────────────────────────

func (m WizardModel) viewWelcome() string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — Setup Wizard") + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString(wizWarningStyle.Render("  No runner.yaml found at the expected path.") + "\n\n")
	b.WriteString(wizSubtitleStyle.Render(
		"  This wizard will guide you through creating one.\n" +
			"  You'll define global settings and at least one command.\n",
	))
	b.WriteString(wizHelpStyle.Render("\n  Press enter to begin  •  ctrl+c to abort"))
	return b.String()
}

func (m WizardModel) viewTextInput(title, hint, placeholder string, optional bool) string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — "+title) + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString("  " + wizSubtitleStyle.Render(hint) + "\n\n")

	// Input box
	display := m.inputBuf
	if display == "" {
		display = wizDimStyle.Render(placeholder)
	} else {
		display = wizValueStyle.Render(display)
	}
	b.WriteString("  " + wizLabelStyle.Render(">") + " " + display + wizCursorStyle.Render("█") + "\n")

	if m.validErr != "" {
		b.WriteString("\n  " + wizErrorStyle.Render("✗ "+m.validErr) + "\n")
	}

	hint2 := "enter to confirm"
	if optional {
		hint2 = "enter to confirm  •  leave blank to skip"
	}
	b.WriteString(wizHelpStyle.Render("\n  " + hint2 + "  •  ctrl+c to abort"))
	return b.String()
}

func (m WizardModel) viewOptionPicker(title, hint string, opts []wizOption) string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — "+title) + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString("  " + wizSubtitleStyle.Render(hint) + "\n\n")

	for i, o := range opts {
		cursor := "  "
		var line string
		if i == m.optCursor {
			cursor = wizCursorStyle.Render("▶ ")
			label := wizSelectedOptStyle.Render(o.label)
			desc := wizOptStyle.Render("  — " + o.desc)
			line = "  " + cursor + label + desc
		} else {
			label := wizOptStyle.Render(o.label)
			desc := wizDimStyle.Render("  — " + o.desc)
			line = "  " + cursor + label + desc
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(wizHelpStyle.Render("\n  ↑/↓ navigate  •  enter select  •  ctrl+c abort"))
	return b.String()
}

func (m WizardModel) viewSummary() string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — Summary") + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString(wizSubtitleStyle.Render("  Review your configuration before saving.\n\n"))

	// Global
	b.WriteString(wizLabelStyle.Render("  Title      ") + wizValueStyle.Render(m.cfgTitle) + "\n")
	b.WriteString(wizLabelStyle.Render("  UI Mode    ") + wizValueStyle.Render(string(m.cfgUIMode)) + "\n")
	b.WriteString(wizLabelStyle.Render("  Run Mode   ") + wizValueStyle.Render(string(m.cfgRunMode)) + "\n")
	b.WriteString("\n")

	// Commands
	for i, c := range m.commands {
		b.WriteString(wizLabelStyle.Render(fmt.Sprintf("  Command #%d", i+1)) + "\n")
		b.WriteString(wizDimStyle.Render("    name     ") + wizValueStyle.Render(c.Name) + "\n")
		if c.Description != "" {
			b.WriteString(wizDimStyle.Render("    desc     ") + wizValueStyle.Render(c.Description) + "\n")
		}
		b.WriteString(wizDimStyle.Render("    command  ") + wizValueStyle.Render(c.Command) + "\n")
		if c.Dir != "" {
			b.WriteString(wizDimStyle.Render("    dir      ") + wizValueStyle.Render(c.Dir) + "\n")
		}
		if c.Group != "" {
			b.WriteString(wizDimStyle.Render("    group    ") + wizValueStyle.Render(c.Group) + "\n")
		}
		b.WriteString(wizDimStyle.Render("    run_mode ") + wizValueStyle.Render(string(c.RunMode)) + "\n")
		b.WriteString("\n")
	}

	b.WriteString(wizSeparator + "\n")
	b.WriteString(wizLabelStyle.Render("  Save to: ") + wizValueStyle.Render(m.savePath) + "\n")
	b.WriteString(wizHelpStyle.Render("\n  Press enter to save  •  ctrl+c to abort"))
	return b.String()
}

func (m WizardModel) viewDone() string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — Setup Complete") + "\n")
	b.WriteString(wizSeparator + "\n\n")

	if m.saveErr != "" {
		b.WriteString(wizErrorStyle.Render("  ✗ Failed to write "+m.savePath) + "\n")
		b.WriteString(wizErrorStyle.Render("    "+m.saveErr) + "\n")
	} else {
		b.WriteString(wizSuccessStyle.Render("  ✓ "+m.savePath+" created successfully!") + "\n\n")
		b.WriteString(wizSubtitleStyle.Render("  Runner will now start with your new configuration.\n"))
	}

	b.WriteString(wizHelpStyle.Render("\n  Press enter to continue"))
	return b.String()
}

// ── RunWizard is the public entry point ───────────────────────────────────────

// WizardResult holds the outcome of a completed wizard run.
type WizardResult struct {
	Aborted bool
	SaveErr string
	Config  *config.Config
}

// RunWizard runs the TUI wizard and returns its result.
func RunWizard(savePath string) (WizardResult, error) {
	m := NewWizard(savePath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return WizardResult{}, fmt.Errorf("wizard: %w", err)
	}
	wm, ok := final.(WizardModel)
	if !ok {
		return WizardResult{}, fmt.Errorf("wizard: unexpected model type")
	}
	result := WizardResult{
		Aborted: wm.Aborted(),
	}
	if wm.saveErr != "" {
		result.SaveErr = wm.saveErr
	}
	if wm.Saved() {
		result.Config = wm.Result()
	}
	return result, nil
}
