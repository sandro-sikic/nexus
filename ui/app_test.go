package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func appCfg(uiMode config.UIMode, runMode config.RunMode, cmds ...config.Command) *config.Config {
	return &config.Config{
		Title:    "App Test",
		UIMode:   uiMode,
		RunMode:  runMode,
		Commands: cmds,
	}
}

func streamCmd(name, command string) config.Command {
	return config.Command{Name: name, Command: command, RunMode: config.RunModeStream}
}

func handoffCmd(name, command string) config.Command {
	return config.Command{Name: name, Command: command, RunMode: config.RunModeHandoff}
}

// updateApp sends a message and returns the concrete AppModel.
func updateApp(m AppModel, msg tea.Msg) AppModel {
	result, _ := m.Update(msg)
	app, ok := result.(AppModel)
	if !ok {
		panic("Update returned non-AppModel")
	}
	return app
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewApp_ListMode(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	if m.cfg != cfg {
		t.Error("cfg not stored")
	}
	if m.state != stateMenu {
		t.Errorf("initial state: got %d, want stateMenu", m.state)
	}
	if m.quitting {
		t.Error("should not be quitting on init")
	}
}

func TestNewApp_FuzzyMode(t *testing.T) {
	cfg := appCfg(config.UIModeFuzzy, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	if len(m.fuzzy.all) != 1 {
		t.Errorf("fuzzy model not populated: got %d cmds", len(m.fuzzy.all))
	}
}

func TestNewApp_GroupMode(t *testing.T) {
	cfg := appCfg(config.UIModeGroup, config.RunModeStream, gcmd("A", "a", "G"))
	m := NewApp(cfg)
	if len(m.group.entries) == 0 {
		t.Error("group model entries not populated")
	}
}

func TestAppModel_Init_ReturnsNilForList(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream)
	m := NewApp(cfg)
	if m.Init() != nil {
		t.Error("Init() for list mode should return nil")
	}
}

// ── Quit from menu ────────────────────────────────────────────────────────────

func TestAppModel_QuitFromMenu(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m = updateApp(m, keyMsg("q"))
	if !m.quitting {
		t.Error("q should set quitting=true when in menu")
	}
}

func TestAppModel_CtrlCQuits(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.quitting {
		t.Error("ctrl+c should quit from menu")
	}
}

// ── Back from output view ─────────────────────────────────────────────────────

func TestAppModel_QFromOutputGoesBackToMenu(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateOutput
	m = updateApp(m, keyMsg("q"))
	if m.state != stateMenu {
		t.Errorf("q from output: state = %d, want stateMenu", m.state)
	}
	if m.quitting {
		t.Error("q from output should not quit app")
	}
}

func TestAppModel_EscFromOutputGoesBackToMenu(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateOutput
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.state != stateMenu {
		t.Errorf("esc from output: state = %d, want stateMenu", m.state)
	}
}

func TestAppModel_EscFromBGGoesBackToMenu(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateBG
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.state != stateMenu {
		t.Errorf("esc from bg: state = %d, want stateMenu", m.state)
	}
}

func TestAppModel_BackFromOutputClearsSelection(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("X", "echo x"))
	m := NewApp(cfg)
	m.state = stateOutput
	// Manually set a selection to confirm it's cleared
	sel := cfg.Commands[0]
	m.list.selected = &sel
	m = updateApp(m, keyMsg("q"))
	if m.list.selected != nil {
		t.Error("selection should be cleared on back")
	}
}

func TestAppModel_EscFromMenuDoesNothing(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	before := m.state
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.state != before {
		t.Errorf("esc from menu changed state: %d -> %d", before, m.state)
	}
	if m.quitting {
		t.Error("esc from menu should not quit")
	}
}

// ── Stream launch ─────────────────────────────────────────────────────────────

func TestAppModel_EnterOnStreamCmdTransitionsToOutput(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("Echo", "echo hello"))
	m := NewApp(cfg)
	// Press enter to select the first (and only) command
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateOutput {
		t.Errorf("after selecting stream cmd: state = %d, want stateOutput", m.state)
	}
}

func TestAppModel_OutputStream_CmdNameStored(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("MyEcho", "echo x"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.output.cmd.Name != "MyEcho" {
		t.Errorf("output.cmd.Name: got %q, want MyEcho", m.output.cmd.Name)
	}
}

// ── Handoff launch ────────────────────────────────────────────────────────────

func TestAppModel_EnterOnHandoffCmdSetsQuitting(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeHandoff, handoffCmd("Dev", "npm run dev"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !m.quitting {
		t.Error("handoff mode should set quitting=true")
	}
}

// ── Fuzzy mode integration ────────────────────────────────────────────────────

func TestAppModel_FuzzyModeSelectTransitionsToOutput(t *testing.T) {
	cfg := appCfg(config.UIModeFuzzy, config.RunModeStream, streamCmd("Build", "make build"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateOutput {
		t.Errorf("fuzzy enter: state = %d, want stateOutput", m.state)
	}
}

func TestAppModel_FuzzyModeTypingFilters(t *testing.T) {
	cfg := appCfg(config.UIModeFuzzy, config.RunModeStream,
		streamCmd("Build", "make build"),
		streamCmd("Test", "go test"),
	)
	m := NewApp(cfg)
	m = updateApp(m, runesMsg("bld"))
	if len(m.fuzzy.filtered) != 1 {
		t.Errorf("fuzzy filtered: got %d, want 1", len(m.fuzzy.filtered))
	}
}

// ── Group mode integration ────────────────────────────────────────────────────

func TestAppModel_GroupModeSelectTransitionsToOutput(t *testing.T) {
	cfg := appCfg(config.UIModeGroup, config.RunModeStream, gcmd("Deploy", "kubectl apply", "Ops"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateOutput {
		t.Errorf("group enter: state = %d, want stateOutput", m.state)
	}
}

// ── WindowSize propagation ────────────────────────────────────────────────────

func TestAppModel_WindowSizeInMenuState(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m = updateApp(m, tea.WindowSizeMsg{Width: 200, Height: 50})
	if m.list.width != 200 || m.list.height != 50 {
		t.Errorf("list not resized: got %dx%d", m.list.width, m.list.height)
	}
}

func TestAppModel_WindowSizeInOutputState(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateOutput
	m = updateApp(m, tea.WindowSizeMsg{Width: 150, Height: 35})
	if m.output.width != 150 || m.output.height != 35 {
		t.Errorf("output not resized: got %dx%d", m.output.width, m.output.height)
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestAppModel_ViewEmptyWhenQuitting(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream)
	m := NewApp(cfg)
	m.quitting = true
	if m.View() != "" {
		t.Errorf("view should be empty when quitting, got:\n%s", m.View())
	}
}

func TestAppModel_ViewShowsErrorWhenErrSet(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream)
	m := NewApp(cfg)
	m.err = "something broke"
	v := m.View()
	if !strings.Contains(v, "something broke") {
		t.Errorf("view should show error:\n%s", v)
	}
}

func TestAppModel_ViewDelegatesListInMenuState(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("ListCmd", "lc"))
	m := NewApp(cfg)
	v := m.View()
	if !strings.Contains(v, "ListCmd") {
		t.Errorf("menu view should contain command name:\n%s", v)
	}
}

func TestAppModel_ViewDelegatesOutputInOutputState(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream)
	m := NewApp(cfg)
	m.state = stateOutput
	m.output = NewOutputModel(config.Command{Name: "OutputCmdX", Command: "x"})
	v := m.View()
	if !strings.Contains(v, "OutputCmdX") {
		t.Errorf("output view should contain cmd name:\n%s", v)
	}
}

func TestAppModel_ViewDelegatesBGInBGState(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream)
	m := NewApp(cfg)
	m.state = stateBG
	m.bg = BackgroundModel{cmd: config.Command{Name: "BGTaskX", Command: "x"}, height: 24}
	v := m.View()
	if !strings.Contains(v, "BGTaskX") {
		t.Errorf("bg view should contain cmd name:\n%s", v)
	}
}

// ── selectedCmd ───────────────────────────────────────────────────────────────

func TestAppModel_SelectedCmdNilByDefault(t *testing.T) {
	for _, uiMode := range []config.UIMode{config.UIModeList, config.UIModeFuzzy, config.UIModeGroup} {
		cfg := appCfg(uiMode, config.RunModeStream, streamCmd("A", "a"))
		m := NewApp(cfg)
		if m.selectedCmd() != nil {
			t.Errorf("selectedCmd() should be nil by default for %s", uiMode)
		}
	}
}
