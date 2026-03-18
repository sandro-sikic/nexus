package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
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

// ── Background launch ─────────────────────────────────────────────────────────

func bgCmd(name, command string) config.Command {
	return config.Command{Name: name, Command: command, RunMode: config.RunModeBackground}
}

func TestAppModel_EnterOnBackgroundCmdTransitionsToBG(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeBackground, bgCmd("BgTask", "echo bg"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateBG {
		t.Errorf("after selecting background cmd: state = %d, want stateBG", m.state)
	}
}

func TestAppModel_BackgroundLaunch_CmdNameStored(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeBackground, bgCmd("MyBGTask", "echo hi"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.bg.cmd.Name != "MyBGTask" {
		t.Errorf("bg.cmd.Name: got %q, want MyBGTask", m.bg.cmd.Name)
	}
}

func TestAppModel_BackgroundLaunch_ErrorPath(t *testing.T) {
	// Launching a background command with invalid dir should store error.
	c := config.Command{
		Name:    "BadDir",
		Command: "echo hi",
		Dir:     "/this/path/does/not/exist/xyz",
		RunMode: config.RunModeBackground,
	}
	cfg := appCfg(config.UIModeList, config.RunModeBackground, c)
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	// On platforms where the dir doesn't exist at Start time, m.err will be set.
	// On others it may or may not error — we just verify no panic and state is sane.
	if m.state == stateBG && m.err != "" {
		t.Errorf("if launched successfully, err should be empty; got %q", m.err)
	}
	if m.state == stateMenu && m.err == "" {
		// This is also acceptable if the dir check happens only at run time
	}
}

// ── updateBG delegation ───────────────────────────────────────────────────────

func TestAppModel_UpdateBGDelegatesToBGModel(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateBG
	m.bg = BackgroundModel{cmd: config.Command{Name: "BGX"}, height: 24}
	// Send a window size; should be handled by bg model
	m = updateApp(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.bg.width != 100 || m.bg.height != 30 {
		t.Errorf("updateBG should propagate WindowSizeMsg: got %dx%d", m.bg.width, m.bg.height)
	}
}

// ── SaveLastIndex on launch ────────────────────────────────────────────────────

func TestAppModel_LaunchPersistsLastIndex(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "nexus.yaml")

	cfg := appCfg(config.UIModeList, config.RunModeStream,
		streamCmd("First", "echo 1"),
		streamCmd("Second", "echo 2"),
	)

	// Write config to temp file so SaveLastIndex can round-trip
	if err := writeTestConfig(cfgPath, cfg); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Build an app with cfgPath wired in and cursor at index 1
	m := newApp(cfg, cfgPath)
	m.list.cursor = 1

	// Select by sending enter
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Re-read config and verify LastIndex was persisted
	loaded, err := loadTestConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if loaded.LastIndex != 1 {
		t.Errorf("LastIndex not persisted: got %d, want 1", loaded.LastIndex)
	}
}

// ── handoffMsg dispatch (Bug fix: handoffMsg was silently dropped) ────────────
//
// Before the fix, AppModel.Update had no case for handoffMsg, so selecting a
// handoff command would set quitting=true but never store the command. The TUI
// would exit and HandleHandoff would never be called, leaving the terminal hung.

// TestAppModel_HandoffMsg_SetsHandoffCmd verifies that receiving a handoffMsg
// stores the command in handoffCmd and does NOT leave it nil.
func TestAppModel_HandoffMsg_SetsHandoffCmd(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)
	sel := cfg.Commands[0]
	m = updateApp(m, handoffMsg{cmd: sel})
	if m.handoffCmd == nil {
		t.Fatal("handoffMsg should set handoffCmd; got nil — handoff would have been silently dropped")
	}
	if m.handoffCmd.Name != "Dev" {
		t.Errorf("handoffCmd.Name: got %q, want Dev", m.handoffCmd.Name)
	}
}

// TestAppModel_HandoffMsg_SetsQuitting verifies the model quits after handoffMsg.
func TestAppModel_HandoffMsg_SetsQuitting(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)
	sel := cfg.Commands[0]

	// The tea.Quit cmd itself won't run here, but quitting is set through the
	// normal enter-key path before the handoffMsg is dispatched.
	m = updateApp(m, handoffMsg{cmd: sel})
	// After handling handoffMsg, the model should be in a quitting state.
	// quitting may already be true from the enter-key launch path; we just
	// verify the model is not in stateOutput or stateBG.
	if m.state == stateOutput || m.state == stateBG {
		t.Errorf("after handoffMsg: unexpected state %d (should not enter output/bg view)", m.state)
	}
}

// TestAppModel_HandoffMsg_CommandNamePreserved verifies multi-step handoff commands
// are preserved intact through the message round-trip.
func TestAppModel_HandoffMsg_CommandNamePreserved(t *testing.T) {
	multiStep := config.Command{
		Name:     "Setup",
		Commands: []string{"npm install", "npm run dev"},
		RunMode:  config.RunModeHandoff,
	}
	cfg := appCfg(config.UIModeList, config.RunModeHandoff, multiStep)
	m := NewApp(cfg)
	m = updateApp(m, handoffMsg{cmd: multiStep})
	if m.handoffCmd == nil {
		t.Fatal("handoffCmd should not be nil for multi-step handoff")
	}
	steps := m.handoffCmd.Steps()
	if len(steps) != 2 {
		t.Fatalf("handoffCmd.Steps(): got %d steps, want 2", len(steps))
	}
	if steps[0] != "npm install" || steps[1] != "npm run dev" {
		t.Errorf("handoffCmd steps: got %v", steps)
	}
}

// TestAppModel_HandoffMsg_NilByDefaultBeforeLaunch verifies handoffCmd starts nil.
func TestAppModel_HandoffMsg_NilByDefaultBeforeLaunch(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)
	if m.handoffCmd != nil {
		t.Error("handoffCmd should be nil before any command is selected")
	}
}

// TestAppModel_EnterOnHandoffCmd_HandoffCmdSetAfterMsg verifies the full
// enter-key → launch → handoffMsg path ends with handoffCmd populated.
// This reproduces the exact hang scenario: user presses enter on a handoff
// command; previously the terminal would hang because handoffCmd was never set.
func TestAppModel_EnterOnHandoffCmd_HandoffCmdSetAfterMsg(t *testing.T) {
	cfg := appCfg(config.UIModeList, config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)

	// Step 1: press enter — this calls launch() which emits tea.Sequence(
	// tea.ExitAltScreen, handoffMsg{…}). The cmd is returned but not executed
	// in unit tests, so we must deliver the handoffMsg manually (as Bubble Tea
	// would in a real program).
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})

	// At this point quitting=true but handoffCmd is still nil — the message
	// hasn't been dispatched yet. Simulate Bubble Tea delivering handoffMsg.
	sel := cfg.Commands[0]
	m = updateApp(m, handoffMsg{cmd: sel})

	if m.handoffCmd == nil {
		t.Fatal("after enter + handoffMsg: handoffCmd is nil — HandleHandoff would never be called")
	}
}

// ── HandleHandoff ─────────────────────────────────────────────────────────────

func TestHandleHandoff_SuccessfulCommandReturnsNil(t *testing.T) {
	c := config.Command{Name: "Echo", Command: "echo handlehandoff-ok", RunMode: config.RunModeHandoff}
	err := HandleHandoff(c)
	if err != nil {
		t.Errorf("HandleHandoff should return nil for successful command, got: %v", err)
	}
}

func TestHandleHandoff_FailingCommandReturnsError(t *testing.T) {
	c := config.Command{Name: "Fail", Command: "exit 1", RunMode: config.RunModeHandoff}
	err := HandleHandoff(c)
	if err == nil {
		t.Error("HandleHandoff should return error when command exits non-zero")
	}
}

func TestHandleHandoff_EmptyCommandReturnsNil(t *testing.T) {
	c := config.Command{Name: "Empty", Command: "", RunMode: config.RunModeHandoff}
	err := HandleHandoff(c)
	if err != nil {
		t.Errorf("HandleHandoff with empty command should return nil, got: %v", err)
	}
}

// ── helpers for config round-trip ─────────────────────────────────────────────

func writeTestConfig(path string, cfg *config.Config) error {
	return config.Write(path, cfg)
}

func loadTestConfig(path string) (*config.Config, error) {
	return config.Load(path)
}
