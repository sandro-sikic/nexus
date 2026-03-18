package ui

// Wizard is a multi-step TUI that walks the user through creating or editing
// a runner.yaml.
//
// CREATE flow (no existing config):
//  1. Welcome notice
//  2. Project title
//  3. UI mode  (list | fuzzy | group)
//  4. Default run mode  (stream | handoff | background)
//  5. Add first command  → sub-steps: name, description, command(s), dir,
//     group, per-command run_mode override
//  6. "Add another command?" → yes loops back to 5, no proceeds
//  7. Summary + confirm save
//
// EDIT flow (existing config):
//  1. Hub screen — shows config summary + command list with delete checkboxes,
//     plus an action menu:
//       [d] delete marked commands
//       [e] edit general settings (title / ui_mode / run_mode)
//       [a] add a new command
//       [s] save & exit

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

	wizActionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33"))

	wizActionKeyStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))
)

// ── Step enum ─────────────────────────────────────────────────────────────────

type wizStep int

const (
	wizStepWelcome         wizStep = iota // 0  – info screen, press enter       (create only)
	wizStepEditHub                        // 1  – overview + action menu          (edit only)
	wizStepTitle                          // 2  – text input
	wizStepUIMode                         // 3  – option picker
	wizStepRunMode                        // 4  – option picker
	wizStepCmdName                        // 5  – text input
	wizStepCmdDesc                        // 6  – text input (optional)
	wizStepCmdCommand                     // 7  – text input (first shell command)
	wizStepCmdMoreCommands                // 8  – repeat: add more shell commands (leave blank to stop)
	wizStepCmdDir                         // 9  – text input (optional)
	wizStepCmdGroup                       // 10 – text input (optional)
	wizStepCmdRunMode                     // 11 – option picker (inherit | stream | handoff | background)
	wizStepAddAnother                     // 12 – yes/no picker                  (create only)
	wizStepDeleteCmds                     // 13 – multi-select delete             (create only, after add-another)
	wizStepSummary                        // 14 – review + confirm
	wizStepDone                           // 15 – finished
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
	optCursor int    // cursor in an option list or command list
	validErr  string // inline validation message
	aborted   bool   // user pressed ctrl+c
	saved     bool   // file was written successfully
	saveErr   string // error from write
	savePath  string // destination path
	editing   bool   // true when editing an existing config (vs. creating new)

	// Where to return after a sub-flow triggered from the hub.
	returnToHub bool

	// collected global config values
	cfgTitle   string
	cfgUIMode  config.UIMode
	cfgRunMode config.RunMode

	// command being built
	cmdName          string
	cmdDesc          string
	cmdCommand       string   // first shell command
	cmdExtraCommands []string // additional shell commands (multi-step)
	cmdDir           string
	cmdGroup         string
	cmdRunMode       config.RunMode // "" means "inherit"

	// accumulated commands
	commands []config.Command

	// hub / delete step: parallel bool slice — true means "mark for deletion"
	deleteMarks []bool

	width  int
	height int
}

// NewWizard creates a fresh WizardModel for creating a new config.
// savePath is where the file will be written.
func NewWizard(savePath string) WizardModel {
	return WizardModel{
		step:     wizStepWelcome,
		savePath: savePath,
		width:    80,
		height:   24,
	}
}

// NewWizardFromConfig creates a WizardModel pre-populated from an existing
// config and opens the edit hub directly.
func NewWizardFromConfig(savePath string, cfg *config.Config) WizardModel {
	m := NewWizard(savePath)
	m.editing = true
	m.step = wizStepEditHub

	// Pre-fill global fields.
	m.cfgTitle = cfg.Title
	m.cfgUIMode = cfg.UIMode
	m.cfgRunMode = cfg.RunMode

	// Deep-copy commands.
	m.commands = make([]config.Command, len(cfg.Commands))
	copy(m.commands, cfg.Commands)

	// Initialise delete marks (all false).
	m.deleteMarks = make([]bool, len(m.commands))

	return m
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

// optionIndex returns the index of the option whose value matches v,
// or 0 if not found (safe fallback for the first option).
func optionIndex(opts []wizOption, v string) int {
	for i, o := range opts {
		if o.value == v {
			return i
		}
	}
	return 0
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

	// ── Edit hub has its own rich navigation ──────────────────────────────
	if m.step == wizStepEditHub {
		return m.handleHubKey(msg)
	}

	// ── Delete-commands step (create flow) ────────────────────────────────
	if m.step == wizStepDeleteCmds {
		switch msg.Type {
		case tea.KeyUp:
			if m.optCursor > 0 {
				m.optCursor--
				m.validErr = ""
			}
		case tea.KeyDown:
			if m.optCursor < len(m.commands)-1 {
				m.optCursor++
				m.validErr = ""
			}
		case tea.KeyRunes:
			if msg.String() == " " && len(m.deleteMarks) > 0 {
				m.deleteMarks[m.optCursor] = !m.deleteMarks[m.optCursor]
				m.validErr = ""
			}
		case tea.KeyEnter:
			return m.advance()
		}
		return m, nil
	}

	opts := m.currentOptions()
	isOptionStep := opts != nil
	isTextStep := !isOptionStep &&
		m.step != wizStepWelcome &&
		m.step != wizStepSummary &&
		m.step != wizStepDone

	switch msg.Type {
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

	case tea.KeyBackspace:
		if isTextStep && len(m.inputBuf) > 0 {
			runes := []rune(m.inputBuf)
			m.inputBuf = string(runes[:len(runes)-1])
			m.validErr = ""
		}
		return m, nil

	case tea.KeyEnter:
		return m.advance()

	case tea.KeyRunes:
		if isTextStep {
			m.inputBuf += msg.String()
			m.validErr = ""
		}
		return m, nil
	}

	return m, nil
}

// handleHubKey processes key input on the edit hub screen.
func (m WizardModel) handleHubKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.optCursor > 0 {
			m.optCursor--
			m.validErr = ""
		}
		return m, nil

	case tea.KeyDown:
		if m.optCursor < len(m.commands)-1 {
			m.optCursor++
			m.validErr = ""
		}
		return m, nil

	case tea.KeyRunes:
		switch msg.String() {
		case " ":
			// Toggle delete mark on the currently focused command.
			if len(m.deleteMarks) > m.optCursor {
				m.deleteMarks[m.optCursor] = !m.deleteMarks[m.optCursor]
				m.validErr = ""
			}

		case "d", "D":
			// Delete all marked commands.
			markedCount := 0
			for _, marked := range m.deleteMarks {
				if marked {
					markedCount++
				}
			}
			if markedCount == 0 {
				m.validErr = "No commands marked for deletion"
				return m, nil
			}
			if markedCount == len(m.commands) {
				m.validErr = "At least one command must remain"
				return m, nil
			}
			kept := m.commands[:0:0]
			for i, cmd := range m.commands {
				if !m.deleteMarks[i] {
					kept = append(kept, cmd)
				}
			}
			m.commands = kept
			m.deleteMarks = make([]bool, len(m.commands))
			// Keep cursor in bounds.
			if m.optCursor >= len(m.commands) {
				m.optCursor = max(0, len(m.commands)-1)
			}
			m.validErr = ""

		case "e", "E":
			// Edit general settings: go to Title step, return to hub after RunMode.
			m.returnToHub = true
			m.inputBuf = m.cfgTitle
			m.optCursor = optionIndex(uiModeOptions, string(m.cfgUIMode))
			m.step = wizStepTitle

		case "a", "A":
			// Add a new command: go to CmdName step, return to hub after RunMode.
			m.returnToHub = true
			m.resetCmdFields()
			m.optCursor = 0
			m.step = wizStepCmdName

		case "s", "S":
			// Save immediately — go to summary.
			m.step = wizStepSummary
		}
		return m, nil
	}

	return m, nil
}

// advance validates the current step, commits its value, and moves forward.
func (m WizardModel) advance() (tea.Model, tea.Cmd) {
	m.validErr = ""

	switch m.step {

	// ── Welcome (create flow) ─────────────────────────────────────────────
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
		m.optCursor = optionIndex(uiModeOptions, string(m.cfgUIMode))
		m.step = wizStepUIMode

	// ── UI mode ───────────────────────────────────────────────────────────
	case wizStepUIMode:
		m.cfgUIMode = config.UIMode(uiModeOptions[m.optCursor].value)
		m.optCursor = optionIndex(runModeOptions, string(m.cfgRunMode))
		m.step = wizStepRunMode

	// ── Default run mode ──────────────────────────────────────────────────
	case wizStepRunMode:
		m.cfgRunMode = config.RunMode(runModeOptions[m.optCursor].value)
		m.inputBuf = ""
		m.optCursor = 0
		if m.returnToHub {
			// Back to hub after editing general settings.
			m.returnToHub = false
			m.step = wizStepEditHub
		} else {
			m.step = wizStepCmdName
		}

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
		m.cmdExtraCommands = nil
		m.inputBuf = ""
		m.step = wizStepCmdMoreCommands

	// ── Command: additional shell commands (optional, repeating) ──────────
	case wizStepCmdMoreCommands:
		extra := strings.TrimSpace(m.inputBuf)
		m.inputBuf = ""
		if extra == "" {
			m.step = wizStepCmdDir
		} else {
			m.cmdExtraCommands = append(m.cmdExtraCommands, extra)
			// Stay on this step so the user can add another.
		}

	// ── Command: working directory (optional) ────────────────────────────
	case wizStepCmdDir:
		m.cmdDir = strings.TrimSpace(m.inputBuf)
		m.inputBuf = ""
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
		if m.returnToHub {
			// Back to hub after adding a new command.
			m.returnToHub = false
			// Grow deleteMarks to cover the newly added command.
			m.deleteMarks = make([]bool, len(m.commands))
			m.step = wizStepEditHub
		} else {
			m.step = wizStepAddAnother
		}

	// ── Add another? (create flow only) ───────────────────────────────────
	case wizStepAddAnother:
		if addAnotherOptions[m.optCursor].value == "yes" {
			m.resetCmdFields()
			m.optCursor = 0
			m.step = wizStepCmdName
		} else {
			m.deleteMarks = make([]bool, len(m.commands))
			m.optCursor = 0
			m.step = wizStepDeleteCmds
		}

	// ── Delete commands (create flow only) ────────────────────────────────
	case wizStepDeleteCmds:
		markedCount := 0
		for _, marked := range m.deleteMarks {
			if marked {
				markedCount++
			}
		}
		if markedCount == len(m.commands) {
			m.validErr = "At least one command must remain"
			return m, nil
		}
		if markedCount > 0 {
			kept := m.commands[:0:0]
			for i, cmd := range m.commands {
				if !m.deleteMarks[i] {
					kept = append(kept, cmd)
				}
			}
			m.commands = kept
		}
		m.deleteMarks = nil
		m.validErr = ""
		m.step = wizStepSummary

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
	cmd := config.Command{
		Name:        m.cmdName,
		Description: m.cmdDesc,
		Dir:         m.cmdDir,
		Group:       m.cmdGroup,
		RunMode:     rm,
	}
	if len(m.cmdExtraCommands) > 0 {
		cmd.Commands = append([]string{m.cmdCommand}, m.cmdExtraCommands...)
	} else {
		cmd.Command = m.cmdCommand
	}
	m.commands = append(m.commands, cmd)
}

// resetCmdFields clears the per-command scratch space.
func (m *WizardModel) resetCmdFields() {
	m.cmdName = ""
	m.cmdDesc = ""
	m.cmdCommand = ""
	m.cmdExtraCommands = nil
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
	case wizStepEditHub:
		return m.viewEditHub()
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
			"The first (or only) shell command to execute.",
			"e.g. npm run dev",
			false,
		)
	case wizStepCmdMoreCommands:
		return m.viewMoreCommands()
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
	case wizStepDeleteCmds:
		return m.viewDeleteCmds()
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

// viewEditHub renders the edit-mode hub: config summary + command list + actions.
func (m WizardModel) viewEditHub() string {
	var b strings.Builder

	b.WriteString(wizTitleStyle.Render("Runner — Edit Configuration") + "\n")
	b.WriteString(wizSeparator + "\n\n")

	// ── Global settings summary ───────────────────────────────────────────
	b.WriteString(wizLabelStyle.Render("  Settings") + "\n")
	b.WriteString(wizDimStyle.Render("    title     ") + wizValueStyle.Render(m.cfgTitle) + "\n")
	b.WriteString(wizDimStyle.Render("    ui_mode   ") + wizValueStyle.Render(string(m.cfgUIMode)) + "\n")
	b.WriteString(wizDimStyle.Render("    run_mode  ") + wizValueStyle.Render(string(m.cfgRunMode)) + "\n")
	b.WriteString("\n")

	// ── Command list with delete checkboxes ───────────────────────────────
	b.WriteString(wizLabelStyle.Render("  Commands") + "\n")
	if len(m.commands) == 0 {
		b.WriteString("  " + wizDimStyle.Render("(no commands yet)") + "\n")
	} else {
		for i, cmd := range m.commands {
			cursor := "   "
			if i == m.optCursor {
				cursor = "  " + wizCursorStyle.Render("▶")
			}

			var checkbox string
			if len(m.deleteMarks) > i && m.deleteMarks[i] {
				checkbox = wizErrorStyle.Render("[✗]")
			} else {
				checkbox = wizDimStyle.Render("[ ]")
			}

			name := cmd.Name
			if i == m.optCursor {
				name = wizSelectedOptStyle.Render(name)
			} else {
				name = wizOptStyle.Render(name)
			}

			steps := cmd.Steps()
			var cmdPreview string
			if len(steps) == 1 {
				cmdPreview = wizDimStyle.Render("  $ " + steps[0])
			} else if len(steps) > 1 {
				cmdPreview = wizDimStyle.Render(fmt.Sprintf("  $ %s  (+%d more)", steps[0], len(steps)-1))
			}

			b.WriteString(fmt.Sprintf("%s %s %s%s\n", cursor, checkbox, name, cmdPreview))
		}
	}

	if m.validErr != "" {
		b.WriteString("\n  " + wizErrorStyle.Render("✗ "+m.validErr) + "\n")
	}

	// ── Action menu ───────────────────────────────────────────────────────
	b.WriteString("\n" + wizSeparator + "\n")
	b.WriteString(wizLabelStyle.Render("  Actions") + "\n")
	b.WriteString("  " + wizActionKeyStyle.Render("[space]") + "  " + wizActionStyle.Render("toggle delete mark on selected command") + "\n")
	b.WriteString("  " + wizActionKeyStyle.Render("[d]    ") + "  " + wizActionStyle.Render("delete all marked commands") + "\n")
	b.WriteString("  " + wizActionKeyStyle.Render("[e]    ") + "  " + wizActionStyle.Render("edit general settings  (title / ui_mode / run_mode)") + "\n")
	b.WriteString("  " + wizActionKeyStyle.Render("[a]    ") + "  " + wizActionStyle.Render("add a new command") + "\n")
	b.WriteString("  " + wizActionKeyStyle.Render("[s]    ") + "  " + wizActionStyle.Render("save & exit") + "\n")
	b.WriteString(wizHelpStyle.Render("\n  ↑/↓ navigate commands  •  ctrl+c abort"))

	return b.String()
}

func (m WizardModel) viewTextInput(title, hint, placeholder string, optional bool) string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — "+title) + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString("  " + wizSubtitleStyle.Render(hint) + "\n\n")

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

func (m WizardModel) viewMoreCommands() string {
	var b strings.Builder
	idx := len(m.commands) + 1
	stepNum := len(m.cmdExtraCommands) + 2 // first command is step 1
	b.WriteString(wizTitleStyle.Render(fmt.Sprintf("Runner — Command #%d — Step %d", idx, stepNum)) + "\n")
	b.WriteString(wizSeparator + "\n\n")

	b.WriteString("  " + wizSubtitleStyle.Render("Commands added so far:") + "\n")
	b.WriteString("  " + wizDimStyle.Render(fmt.Sprintf("  1. %s", m.cmdCommand)) + "\n")
	for i, c := range m.cmdExtraCommands {
		b.WriteString("  " + wizDimStyle.Render(fmt.Sprintf("  %d. %s", i+2, c)) + "\n")
	}
	b.WriteString("\n")

	b.WriteString("  " + wizSubtitleStyle.Render("Add another shell command, or leave blank to finish adding steps.") + "\n\n")

	display := m.inputBuf
	if display == "" {
		display = wizDimStyle.Render("e.g. npm run build  (leave blank to stop)")
	} else {
		display = wizValueStyle.Render(display)
	}
	b.WriteString("  " + wizLabelStyle.Render(">") + " " + display + wizCursorStyle.Render("█") + "\n")

	if m.validErr != "" {
		b.WriteString("\n  " + wizErrorStyle.Render("✗ "+m.validErr) + "\n")
	}

	b.WriteString(wizHelpStyle.Render("\n  enter to add step / leave blank to finish  •  ctrl+c to abort"))
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

func (m WizardModel) viewDeleteCmds() string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — Delete Commands") + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString("  " + wizSubtitleStyle.Render("Mark commands to delete, then press enter to confirm.\n  At least one command must remain.") + "\n\n")

	for i, cmd := range m.commands {
		cursor := "  "
		if i == m.optCursor {
			cursor = wizCursorStyle.Render("▶ ")
		}

		var checkbox string
		if len(m.deleteMarks) > i && m.deleteMarks[i] {
			checkbox = wizErrorStyle.Render("[✗]")
		} else {
			checkbox = wizDimStyle.Render("[ ]")
		}

		name := cmd.Name
		if i == m.optCursor {
			name = wizSelectedOptStyle.Render(name)
		} else {
			name = wizOptStyle.Render(name)
		}

		desc := ""
		if cmd.Description != "" {
			desc = "  " + wizDimStyle.Render(cmd.Description)
		}

		b.WriteString(fmt.Sprintf("  %s%s %s%s\n", cursor, checkbox, name, desc))
	}

	if m.validErr != "" {
		b.WriteString("\n  " + wizErrorStyle.Render("✗ "+m.validErr) + "\n")
	}

	b.WriteString(wizHelpStyle.Render("\n  ↑/↓ navigate  •  space toggle  •  enter confirm  •  ctrl+c abort"))
	return b.String()
}

func (m WizardModel) viewSummary() string {
	var b strings.Builder
	b.WriteString(wizTitleStyle.Render("Runner — Summary") + "\n")
	b.WriteString(wizSeparator + "\n\n")
	b.WriteString(wizSubtitleStyle.Render("  Review your configuration before saving.") + "\n")

	b.WriteString(wizLabelStyle.Render("  Title      ") + wizValueStyle.Render(m.cfgTitle) + "\n")
	b.WriteString(wizLabelStyle.Render("  UI Mode    ") + wizValueStyle.Render(string(m.cfgUIMode)) + "\n")
	b.WriteString(wizLabelStyle.Render("  Run Mode   ") + wizValueStyle.Render(string(m.cfgRunMode)) + "\n")
	b.WriteString("\n")

	for i, c := range m.commands {
		b.WriteString(wizLabelStyle.Render(fmt.Sprintf("  Command #%d", i+1)) + "\n")
		b.WriteString(wizDimStyle.Render("    name     ") + wizValueStyle.Render(c.Name) + "\n")
		if c.Description != "" {
			b.WriteString(wizDimStyle.Render("    desc     ") + wizValueStyle.Render(c.Description) + "\n")
		}
		steps := c.Steps()
		if len(steps) == 1 {
			b.WriteString(wizDimStyle.Render("    command  ") + wizValueStyle.Render(steps[0]) + "\n")
		} else {
			b.WriteString(wizDimStyle.Render("    commands ") + "\n")
			for j, step := range steps {
				b.WriteString(wizDimStyle.Render(fmt.Sprintf("      %d. ", j+1)) + wizValueStyle.Render(step) + "\n")
			}
		}
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
	} else if m.editing {
		b.WriteString(wizSuccessStyle.Render("  ✓ "+m.savePath+" updated successfully!") + "\n\n")
		b.WriteString(wizSubtitleStyle.Render("  Runner will now start with your updated configuration.\n"))
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

// runWizardModel is the shared runner used by both RunWizard and RunWizardEdit.
func runWizardModel(m WizardModel) (WizardResult, error) {
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

// RunWizardEdit runs the wizard pre-populated with the values from an existing
// config. On completion the updated config is written to savePath and returned.
func RunWizardEdit(savePath string, cfg *config.Config) (WizardResult, error) {
	return runWizardModel(NewWizardFromConfig(savePath, cfg))
}

// RunWizard runs the TUI wizard and returns its result.
func RunWizard(savePath string) (WizardResult, error) {
	return runWizardModel(NewWizard(savePath))
}
