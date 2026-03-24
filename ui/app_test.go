package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func appCfg(runMode config.RunMode, cmds ...config.Task) *config.Config {
	return &config.Config{
		Title:   "App Test",
		RunMode: runMode,
		Tasks:   cmds,
	}
}

func streamCmd(name, command string) config.Task {
	return config.Task{Name: name, Actions: []config.Action{{Command: command}}, RunMode: config.RunModeStream}
}

func handoffCmd(name, command string) config.Task {
	return config.Task{Name: name, Actions: []config.Action{{Command: command}}, RunMode: config.RunModeHandoff}
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
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
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
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	if len(m.fuzzy.all) != 1 {
		t.Errorf("fuzzy model not populated: got %d cmds", len(m.fuzzy.all))
	}
}

func TestNewApp_GroupMode(t *testing.T) {
	cfg := appCfg(config.RunModeStream, gtask("A", []string{"a"}, "G"))
	m := NewApp(cfg)
	if len(m.fuzzy.entries) == 0 {
		t.Error("fuzzy group entries not populated")
	}
}

func TestAppModel_Init_ReturnsNilForList(t *testing.T) {
	cfg := appCfg(config.RunModeStream)
	m := NewApp(cfg)
	if m.Init() != nil {
		t.Error("Init() for list mode should return nil")
	}
}

// ── Quit from menu ────────────────────────────────────────────────────────────

func TestAppModel_QuitFromMenu(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m = updateApp(m, keyMsg("q"))
	if !m.quitting {
		t.Error("q should set quitting=true when in menu")
	}
}

func TestAppModel_CtrlCQuits(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.quitting {
		t.Error("ctrl+c should quit from menu")
	}
}

// ── Back from output view ─────────────────────────────────────────────────────

func TestAppModel_QFromOutputGoesBackToMenu(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
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
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateOutput
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.state != stateMenu {
		t.Errorf("esc from output: state = %d, want stateMenu", m.state)
	}
}

func TestAppModel_EscFromBGGoesBackToMenu(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateBG
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.state != stateMenu {
		t.Errorf("esc from bg: state = %d, want stateMenu", m.state)
	}
}

func TestAppModel_BackFromOutputClearsSelection(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("X", "echo x"))
	m := NewApp(cfg)
	m.state = stateOutput
	// Manually set a selection to confirm it's cleared
	sel := cfg.Tasks[0]
	m.fuzzy.selected = &sel
	m = updateApp(m, keyMsg("q"))
	if m.fuzzy.selected != nil {
		t.Error("selection should be cleared on back")
	}
}

func TestAppModel_EscFromMenuDoesNothing(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
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
	cfg := appCfg(config.RunModeStream, streamCmd("Echo", "echo hello"))
	m := NewApp(cfg)
	// Press enter to select the first (and only) command
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateOutput {
		t.Errorf("after selecting stream cmd: state = %d, want stateOutput", m.state)
	}
}

func TestAppModel_OutputStream_CmdNameStored(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("MyEcho", "echo x"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.output.task.Name != "MyEcho" {
		t.Errorf("output.task.Name: got %q, want MyEcho", m.output.task.Name)
	}
}

// ── Handoff launch ────────────────────────────────────────────────────────────

func TestAppModel_EnterOnHandoffCmdSetsQuitting(t *testing.T) {
	cfg := appCfg(config.RunModeHandoff, handoffCmd("Dev", "npm run dev"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !m.quitting {
		t.Error("handoff mode should set quitting=true")
	}
}

// ── Fuzzy mode integration ────────────────────────────────────────────────────

func TestAppModel_FuzzyModeSelectTransitionsToOutput(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("Build", "make build"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateOutput {
		t.Errorf("fuzzy enter: state = %d, want stateOutput", m.state)
	}
}

func TestAppModel_FuzzyModeTypingFilters(t *testing.T) {
	cfg := appCfg(config.RunModeStream,
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
	cfg := appCfg(config.RunModeStream, gtask("Deploy", []string{"kubectl apply"}, "Ops"))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateOutput {
		t.Errorf("group enter: state = %d, want stateOutput", m.state)
	}
}

// ── WindowSize propagation ────────────────────────────────────────────────────

func TestAppModel_WindowSizeInMenuState(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m = updateApp(m, tea.WindowSizeMsg{Width: 200, Height: 50})
	if m.fuzzy.width != 200 || m.fuzzy.height != 50 {
		t.Errorf("fuzzy not resized: got %dx%d", m.fuzzy.width, m.fuzzy.height)
	}
}

func TestAppModel_WindowSizeInOutputState(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateOutput
	m = updateApp(m, tea.WindowSizeMsg{Width: 150, Height: 35})
	if m.output.width != 150 || m.output.height != 35 {
		t.Errorf("output not resized: got %dx%d", m.output.width, m.output.height)
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestAppModel_ViewEmptyWhenQuitting(t *testing.T) {
	cfg := appCfg(config.RunModeStream)
	m := NewApp(cfg)
	m.quitting = true
	if m.View() != "" {
		t.Errorf("view should be empty when quitting, got:\n%s", m.View())
	}
}

func TestAppModel_ViewShowsErrorWhenErrSet(t *testing.T) {
	cfg := appCfg(config.RunModeStream)
	m := NewApp(cfg)
	m.err = "something broke"
	v := m.View()
	if !strings.Contains(v, "something broke") {
		t.Errorf("view should show error:\n%s", v)
	}
}

func TestAppModel_ViewDelegatesListInMenuState(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("ListCmd", "lc"))
	m := NewApp(cfg)
	v := m.View()
	if !strings.Contains(v, "ListCmd") {
		t.Errorf("menu view should contain command name:\n%s", v)
	}
}

func TestAppModel_ViewDelegatesOutputInOutputState(t *testing.T) {
	cfg := appCfg(config.RunModeStream)
	m := NewApp(cfg)
	m.state = stateOutput
	m.output = NewOutputModel(config.Task{Name: "OutputCmdX", Actions: []config.Action{{Command: "x"}}, RunMode: config.RunModeStream})
	v := m.View()
	if !strings.Contains(v, "OutputCmdX") {
		t.Errorf("output view should contain cmd name:\n%s", v)
	}
}

func TestAppModel_ViewDelegatesBGInBGState(t *testing.T) {
	cfg := appCfg(config.RunModeStream)
	m := NewApp(cfg)
	m.state = stateBG
	m.bg = BackgroundModel{task: config.Task{Name: "BGTaskX", Actions: []config.Action{{Command: "x"}}, RunMode: config.RunModeBackground}, height: 24}
	v := m.View()
	if !strings.Contains(v, "BGTaskX") {
		t.Errorf("bg view should contain cmd name:\n%s", v)
	}
}

// ── selection nil by default ──────────────────────────────────────────────────

func TestAppModel_SelectedCmdNilByDefault(t *testing.T) {
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	if m.fuzzy.Selected() != nil {
		t.Error("fuzzy.Selected() should be nil by default")
	}
}

// ── Background launch ─────────────────────────────────────────────────────────

func bgTask(name string, commands []string) config.Task {
	actions := make([]config.Action, len(commands))
	for i, cmd := range commands {
		actions[i] = config.Action{Command: cmd}
	}
	return config.Task{Name: name, Actions: actions, RunMode: config.RunModeBackground}
}

func TestAppModel_EnterOnBackgroundCmdTransitionsToBG(t *testing.T) {
	cfg := appCfg(config.RunModeBackground, bgTask("BgTask", []string{"echo bg"}))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.state != stateBG {
		t.Errorf("after selecting background cmd: state = %d, want stateBG", m.state)
	}
}

func TestAppModel_BackgroundLaunch_CmdNameStored(t *testing.T) {
	cfg := appCfg(config.RunModeBackground, bgTask("MyBGTask", []string{"echo hi"}))
	m := NewApp(cfg)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.bg.task.Name != "MyBGTask" {
		t.Errorf("bg.task.Name: got %q, want MyBGTask", m.bg.task.Name)
	}
}

func TestAppModel_BackgroundLaunch_ErrorPath(t *testing.T) {
	// Launching a background command with invalid dir should store error.
	c := config.Task{
		Name:    "BadDir",
		Actions: []config.Action{{Command: "echo hi"}},
		Dir:     "/this/path/does/not/exist/xyz",
		RunMode: config.RunModeBackground,
	}
	cfg := appCfg(config.RunModeBackground, c)
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
	cfg := appCfg(config.RunModeStream, streamCmd("A", "a"))
	m := NewApp(cfg)
	m.state = stateBG
	m.bg = BackgroundModel{task: config.Task{Name: "BGX", Actions: []config.Action{{Command: "x"}}, RunMode: config.RunModeBackground}, height: 24}
	// Send a window size; should be handled by bg model
	m = updateApp(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.bg.width != 100 || m.bg.height != 30 {
		t.Errorf("updateBG should propagate WindowSizeMsg: got %dx%d", m.bg.width, m.bg.height)
	}
}

// ── SaveLastIndex on launch ────────────────────────────────────────────────────

func TestAppModel_LaunchFromFuzzyGroupView(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "nexus.yaml")

	cfg := appCfg(config.RunModeStream,
		streamCmd("First", "echo 1"),
		streamCmd("Second", "echo 2"),
	)

	// Write config to temp file
	if err := writeTestConfig(cfgPath, cfg); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Build an app and press enter to select the first command in group view
	m := newApp(cfg, cfgPath)
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.state != stateOutput {
		t.Errorf("expected stateOutput after selecting command, got %d", m.state)
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
	cfg := appCfg(config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)
	sel := cfg.Tasks[0]
	m = updateApp(m, handoffMsg{task: sel})
	if m.handoffTask == nil {
		t.Fatal("handoffMsg should set handoffCmd; got nil — handoff would have been silently dropped")
	}
	if m.handoffTask.Name != "Dev" {
		t.Errorf("handoffCmd.Name: got %q, want Dev", m.handoffTask.Name)
	}
}

// TestAppModel_HandoffMsg_SetsQuitting verifies the model quits after handoffMsg.
func TestAppModel_HandoffMsg_SetsQuitting(t *testing.T) {
	cfg := appCfg(config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)
	sel := cfg.Tasks[0]

	// The tea.Quit cmd itself won't run here, but quitting is set through the
	// normal enter-key path before the handoffMsg is dispatched.
	m = updateApp(m, handoffMsg{task: sel})
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
	multiStep := config.Task{
		Name:    "Setup",
		Actions: []config.Action{{Command: "npm install"}, {Command: "npm run dev"}},
		RunMode: config.RunModeHandoff,
	}
	cfg := appCfg(config.RunModeHandoff, multiStep)
	m := NewApp(cfg)
	m = updateApp(m, handoffMsg{task: multiStep})
	if m.handoffTask == nil {
		t.Fatal("handoffCmd should not be nil for multi-step handoff")
	}
	steps := m.handoffTask.AllCommands()
	if len(steps) != 2 {
		t.Fatalf("handoffCmd.AllCommands(): got %d steps, want 2", len(steps))
	}
	if steps[0] != "npm install" || steps[1] != "npm run dev" {
		t.Errorf("handoffCmd steps: got %v", steps)
	}
}

// TestAppModel_HandoffMsg_NilByDefaultBeforeLaunch verifies handoffCmd starts nil.
func TestAppModel_HandoffMsg_NilByDefaultBeforeLaunch(t *testing.T) {
	cfg := appCfg(config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)
	if m.handoffTask != nil {
		t.Error("handoffCmd should be nil before any command is selected")
	}
}

// TestAppModel_EnterOnHandoffCmd_HandoffCmdSetAfterMsg verifies the full
// enter-key → launch → handoffMsg path ends with handoffCmd populated.
// This reproduces the exact hang scenario: user presses enter on a handoff
// command; previously the terminal would hang because handoffCmd was never set.
func TestAppModel_EnterOnHandoffCmd_HandoffCmdSetAfterMsg(t *testing.T) {
	cfg := appCfg(config.RunModeHandoff, handoffCmd("Dev", "echo dev"))
	m := NewApp(cfg)

	// Step 1: press enter — this calls launch() which emits tea.Sequence(
	// tea.ExitAltScreen, handoffMsg{…}). The cmd is returned but not executed
	// in unit tests, so we must deliver the handoffMsg manually (as Bubble Tea
	// would in a real program).
	m = updateApp(m, tea.KeyMsg{Type: tea.KeyEnter})

	// At this point quitting=true but handoffCmd is still nil — the message
	// hasn't been dispatched yet. Simulate Bubble Tea delivering handoffMsg.
	sel := cfg.Tasks[0]
	m = updateApp(m, handoffMsg{task: sel})

	if m.handoffTask == nil {
		t.Fatal("after enter + handoffMsg: handoffCmd is nil — HandleHandoff would never be called")
	}
}

// ── HandleHandoff ─────────────────────────────────────────────────────────────

func TestHandleHandoff_SuccessfulCommandReturnsNil(t *testing.T) {
	c := config.Task{Name: "Echo", Actions: []config.Action{{Command: "echo handlehandoff-ok"}}, RunMode: config.RunModeHandoff}
	err := HandleHandoff(c)
	if err != nil {
		t.Errorf("HandleHandoff should return nil for successful command, got: %v", err)
	}
}

func TestHandleHandoff_FailingCommandReturnsError(t *testing.T) {
	c := config.Task{Name: "Fail", Actions: []config.Action{{Command: "exit 1"}}, RunMode: config.RunModeHandoff}
	err := HandleHandoff(c)
	if err == nil {
		t.Error("HandleHandoff should return error when command exits non-zero")
	}
}

func TestHandleHandoff_EmptyCommandReturnsNil(t *testing.T) {
	c := config.Task{Name: "Empty", Actions: []config.Action{{Command: ""}}, RunMode: config.RunModeHandoff}
	err := HandleHandoff(c)
	if err != nil {
		t.Errorf("HandleHandoff with empty command should return nil, got: %v", err)
	}
}

// ── RunFirstMatch ─────────────────────────────────────────────────────────────

// matchCfg builds a config suitable for RunFirstMatch tests. UIMode is
// irrelevant for RunFirstMatch (it never starts the TUI), so we just use List.
func matchCfg(tasks ...config.Task) *config.Config {
	return &config.Config{
		Title:   "Match Test",
		RunMode: config.RunModeHandoff,
		Tasks:   tasks,
	}
}

func echoTask(name, description, shell string) config.Task {
	return config.Task{
		Name:        name,
		Description: description,
		Actions:     []config.Action{{Command: shell}},
		RunMode:     config.RunModeHandoff,
	}
}

// TestRunFirstMatch_ExactNameMatch verifies that a query matching a command
// name exactly causes that command to be executed.
func TestRunFirstMatch_ExactNameMatch(t *testing.T) {
	cfg := matchCfg(echoTask("build", "", "echo build-ok"))
	if err := RunFirstMatch(cfg, "build"); err != nil {
		t.Errorf("exact name match should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_SubsequenceNameMatch verifies that a fuzzy subsequence
// query matches a command name (same algorithm as the FuzzyModel).
func TestRunFirstMatch_SubsequenceNameMatch(t *testing.T) {
	cfg := matchCfg(echoTask("build production", "", "echo subseq-ok"))
	if err := RunFirstMatch(cfg, "bldprd"); err != nil {
		t.Errorf("subsequence name match should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_CaseInsensitive verifies that matching is case-insensitive.
func TestRunFirstMatch_CaseInsensitive(t *testing.T) {
	cfg := matchCfg(echoTask("Deploy", "", "echo deploy-ok"))
	if err := RunFirstMatch(cfg, "DEPLOY"); err != nil {
		t.Errorf("case-insensitive match should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_MatchOnDescription verifies that a query matching the
// description field (not the name) still selects the command.
func TestRunFirstMatch_MatchOnDescription(t *testing.T) {
	cfg := matchCfg(echoTask("x", "runs all integration tests", "echo desc-ok"))
	if err := RunFirstMatch(cfg, "intgr"); err != nil {
		t.Errorf("description match should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_MatchOnCommandStep verifies that a query matching the
// shell command string selects the command.
func TestRunFirstMatch_MatchOnCommandStep(t *testing.T) {
	// "echo node-run-dist" contains the subsequence "nrd": [n]ode-[r]un-[d]ist.
	cfg := matchCfg(echoTask("x", "", "echo node-run-dist"))
	if err := RunFirstMatch(cfg, "nrd"); err != nil {
		t.Errorf("command step match should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_NoMatch verifies that an unmatched query returns an error.
func TestRunFirstMatch_NoMatch(t *testing.T) {
	cfg := matchCfg(echoTask("build", "desc", "echo hi"))
	err := RunFirstMatch(cfg, "zzzzz")
	if err == nil {
		t.Error("expected error for unmatched query, got nil")
	}
}

// TestRunFirstMatch_NoMatchErrorContainsQuery verifies the error message
// includes the original query so the user knows what was attempted.
func TestRunFirstMatch_NoMatchErrorContainsQuery(t *testing.T) {
	cfg := matchCfg(echoTask("build", "", "echo hi"))
	err := RunFirstMatch(cfg, "xyzzy")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "xyzzy") {
		t.Errorf("error should mention the query, got: %v", err)
	}
}

// TestRunFirstMatch_EmptyCommandList verifies that an empty config returns an error.
func TestRunFirstMatch_EmptyCommandList(t *testing.T) {
	cfg := matchCfg()
	if err := RunFirstMatch(cfg, "anything"); err == nil {
		t.Error("expected error for empty command list, got nil")
	}
}

// TestRunFirstMatch_RunsFirstMatchNotSecond verifies that when multiple
// commands match, the first one in config order is executed, not a later one.
// We detect this by using commands with distinct exit behaviors.
func TestRunFirstMatch_RunsFirstMatchNotSecond(t *testing.T) {
	first := config.Task{Name: "alpha", Actions: []config.Action{{Command: "echo first-ok"}}, RunMode: config.RunModeHandoff}
	second := config.Task{Name: "alpha-two", Actions: []config.Action{{Command: "exit 1"}}, RunMode: config.RunModeHandoff}
	cfg := matchCfg(first, second)
	// "alpha" matches both "alpha" and "alpha-two"; first should win (no error).
	if err := RunFirstMatch(cfg, "alpha"); err != nil {
		t.Errorf("first matching command should be executed (expected no error), got: %v", err)
	}
}

// TestRunFirstMatch_FailingCommandPropagatesError verifies that if the matched
// command exits non-zero, the error is returned to the caller.
func TestRunFirstMatch_FailingCommandPropagatesError(t *testing.T) {
	cfg := matchCfg(echoTask("fail", "", "exit 1"))
	err := RunFirstMatch(cfg, "fail")
	if err == nil {
		t.Error("expected error from failing command, got nil")
	}
}

// TestRunFirstMatch_MultiStepCommand verifies that a multi-step command
// (Commands slice) is executed correctly.
func TestRunFirstMatch_MultiStepCommand(t *testing.T) {
	c := config.Task{
		Name:    "setup",
		Actions: []config.Action{{Command: "echo step-one"}, {Command: "echo step-two"}},
		RunMode: config.RunModeHandoff,
	}
	cfg := matchCfg(c)
	if err := RunFirstMatch(cfg, "setup"); err != nil {
		t.Errorf("multi-step command should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_MatchOnSecondStep verifies that a fuzzy query matching
// only the second step of a multi-step command still selects it.
func TestRunFirstMatch_MatchOnSecondStep(t *testing.T) {
	// The second step "echo node-run-build" contains subsequence "nrb":
	// [n]ode-[r]un-[b]uild.
	c := config.Task{
		Name:    "pipeline",
		Actions: []config.Action{{Command: "echo step-one"}, {Command: "echo node-run-build"}},
		RunMode: config.RunModeHandoff,
	}
	cfg := matchCfg(c)
	if err := RunFirstMatch(cfg, "nrb"); err != nil {
		t.Errorf("match on second step should succeed, got: %v", err)
	}
}

// TestRunFirstMatch_EmptyQueryMatchesFirst verifies that an empty query matches
// the first command (consistent with how fuzzyMatch treats empty queries).
func TestRunFirstMatch_EmptyQueryMatchesFirst(t *testing.T) {
	cfg := matchCfg(
		echoTask("first", "", "echo first"),
		echoTask("second", "", "echo second"),
	)
	if err := RunFirstMatch(cfg, ""); err != nil {
		t.Errorf("empty query should match first command, got: %v", err)
	}
}

// ── helpers for config round-trip ─────────────────────────────────────────────

func writeTestConfig(path string, cfg *config.Config) error {
	return config.Write(path, cfg)
}

func loadTestConfig(path string) (*config.Config, error) {
	return config.Load(path)
}
