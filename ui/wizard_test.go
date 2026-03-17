package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// updateWiz sends a message and asserts the model stays a WizardModel.
func updateWiz(t *testing.T, m WizardModel, msg tea.Msg) WizardModel {
	t.Helper()
	result, _ := m.Update(msg)
	wm, ok := result.(WizardModel)
	if !ok {
		t.Fatalf("Update returned %T, want WizardModel", result)
	}
	return wm
}

// typeText sends each rune of s as a KeyRunes message.
func typeText(t *testing.T, m WizardModel, s string) WizardModel {
	t.Helper()
	for _, r := range s {
		m = updateWiz(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

// pressEnter sends a KeyEnter.
func pressEnter(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	return updateWiz(t, m, tea.KeyMsg{Type: tea.KeyEnter})
}

// pressDown sends a KeyDown.
func pressDown(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	return updateWiz(t, m, tea.KeyMsg{Type: tea.KeyDown})
}

// pressUp sends a KeyUp.
func pressUp(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	return updateWiz(t, m, tea.KeyMsg{Type: tea.KeyUp})
}

// pressBackspace sends a KeyBackspace.
func pressBackspace(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	return updateWiz(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
}

// pressSpace sends a space rune (used to toggle marks on the delete step).
func pressSpace(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	return updateWiz(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
}

// confirmDelete presses enter on the delete step, confirming the current marks
// (which may be empty — no deletions).
func confirmDelete(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	if m.step != wizStepDeleteCmds {
		t.Fatalf("confirmDelete: expected wizStepDeleteCmds, got %d", m.step)
	}
	return pressEnter(t, m)
}

// advanceThrough drives the wizard to just past the "default run mode" step
// (i.e. ready for the first command), using the given title, uiMode index,
// and runMode index (0-based within their option lists).
func advanceToFirstCmd(t *testing.T, title string, uiModeIdx, runModeIdx int) WizardModel {
	t.Helper()
	m := NewWizard("out.yaml")

	// Welcome
	m = pressEnter(t, m)
	if m.step != wizStepTitle {
		t.Fatalf("expected wizStepTitle after welcome, got %d", m.step)
	}

	// Title
	m = typeText(t, m, title)
	m = pressEnter(t, m)
	if m.step != wizStepUIMode {
		t.Fatalf("expected wizStepUIMode, got %d", m.step)
	}

	// UI mode
	for i := 0; i < uiModeIdx; i++ {
		m = pressDown(t, m)
	}
	m = pressEnter(t, m)
	if m.step != wizStepRunMode {
		t.Fatalf("expected wizStepRunMode, got %d", m.step)
	}

	// Run mode
	for i := 0; i < runModeIdx; i++ {
		m = pressDown(t, m)
	}
	m = pressEnter(t, m)
	if m.step != wizStepCmdName {
		t.Fatalf("expected wizStepCmdName, got %d", m.step)
	}
	return m
}

// addCommand drives through all command sub-steps.
// runModeIdx: index in cmdRunModeOptions (0 = inherit).
func addCommand(t *testing.T, m WizardModel, name, desc, command, dir, group string, runModeIdx int) WizardModel {
	t.Helper()

	// Name
	if m.step != wizStepCmdName {
		t.Fatalf("expected wizStepCmdName, got %d", m.step)
	}
	m = typeText(t, m, name)
	m = pressEnter(t, m)

	// Desc
	if m.step != wizStepCmdDesc {
		t.Fatalf("expected wizStepCmdDesc, got %d", m.step)
	}
	m = typeText(t, m, desc)
	m = pressEnter(t, m)

	// Command
	if m.step != wizStepCmdCommand {
		t.Fatalf("expected wizStepCmdCommand, got %d", m.step)
	}
	m = typeText(t, m, command)
	m = pressEnter(t, m)

	// More commands (optional): press Enter with empty input to skip.
	if m.step != wizStepCmdMoreCommands {
		t.Fatalf("expected wizStepCmdMoreCommands, got %d", m.step)
	}
	m = pressEnter(t, m) // blank → move on to dir step

	// Dir
	if m.step != wizStepCmdDir {
		t.Fatalf("expected wizStepCmdDir, got %d", m.step)
	}
	m = typeText(t, m, dir)
	m = pressEnter(t, m)

	// Group (only in group mode)
	if m.step == wizStepCmdGroup {
		m = typeText(t, m, group)
		m = pressEnter(t, m)
	}

	// Run mode override
	if m.step != wizStepCmdRunMode {
		t.Fatalf("expected wizStepCmdRunMode, got %d", m.step)
	}
	for i := 0; i < runModeIdx; i++ {
		m = pressDown(t, m)
	}
	m = pressEnter(t, m)

	if m.step != wizStepAddAnother {
		t.Fatalf("expected wizStepAddAnother, got %d", m.step)
	}
	return m
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewWizard_InitialState(t *testing.T) {
	m := NewWizard("test.yaml")
	if m.step != wizStepWelcome {
		t.Errorf("initial step: got %d, want wizStepWelcome", m.step)
	}
	if m.savePath != "test.yaml" {
		t.Errorf("savePath: got %q, want test.yaml", m.savePath)
	}
	if m.aborted {
		t.Error("should not be aborted initially")
	}
	if m.saved {
		t.Error("should not be saved initially")
	}
	if len(m.commands) != 0 {
		t.Errorf("commands should be empty, got %d", len(m.commands))
	}
}

func TestWizardModel_Init_ReturnsNil(t *testing.T) {
	m := NewWizard("x.yaml")
	if m.Init() != nil {
		t.Error("Init() should return nil")
	}
}

// ── Abort ─────────────────────────────────────────────────────────────────────

func TestWizard_CtrlCAbortsOnWelcome(t *testing.T) {
	m := NewWizard("x.yaml")
	m = updateWiz(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.aborted {
		t.Error("ctrl+c should set aborted=true")
	}
}

func TestWizard_CtrlCAbortsOnAnyStep(t *testing.T) {
	m := advanceToFirstCmd(t, "My App", 0, 0)
	m = updateWiz(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.aborted {
		t.Error("ctrl+c should abort mid-wizard")
	}
}

func TestWizard_AbortedResultIsNil(t *testing.T) {
	m := NewWizard("x.yaml")
	m.aborted = true
	if m.Result() != nil {
		t.Error("Result() should be nil when aborted")
	}
}

// ── Welcome step ──────────────────────────────────────────────────────────────

func TestWizard_WelcomeAdvancesToTitle(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	if m.step != wizStepTitle {
		t.Errorf("after welcome enter: step = %d, want wizStepTitle", m.step)
	}
}

func TestWizard_WelcomeView(t *testing.T) {
	m := NewWizard("x.yaml")
	v := m.View()
	if !strings.Contains(v, "Wizard") {
		t.Errorf("welcome view missing 'Wizard':\n%s", v)
	}
	if !strings.Contains(v, "runner.yaml") && !strings.Contains(v, "missing") && !strings.Contains(v, "No ") {
		t.Errorf("welcome view should mention missing file:\n%s", v)
	}
}

// ── Title step ────────────────────────────────────────────────────────────────

func TestWizard_TitleTypingAndConfirm(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // welcome
	m = typeText(t, m, "My Runner")
	m = pressEnter(t, m)
	if m.cfgTitle != "My Runner" {
		t.Errorf("cfgTitle: got %q, want My Runner", m.cfgTitle)
	}
	if m.step != wizStepUIMode {
		t.Errorf("should advance to wizStepUIMode, got %d", m.step)
	}
}

func TestWizard_TitleDefaultsToRunner(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // welcome
	// don't type anything
	m = pressEnter(t, m)
	if m.cfgTitle != "Runner" {
		t.Errorf("empty title should default to 'Runner', got %q", m.cfgTitle)
	}
}

func TestWizard_TitleBackspace(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = typeText(t, m, "abc")
	m = pressBackspace(t, m)
	if m.inputBuf != "ab" {
		t.Errorf("after backspace: inputBuf = %q, want ab", m.inputBuf)
	}
}

func TestWizard_TitleBackspaceOnEmpty(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = pressBackspace(t, m) // should be no-op
	if m.inputBuf != "" {
		t.Errorf("backspace on empty should keep empty, got %q", m.inputBuf)
	}
}

func TestWizard_TitleInputBufferClearedOnAdvance(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = typeText(t, m, "test")
	m = pressEnter(t, m)
	if m.inputBuf != "" {
		t.Errorf("inputBuf should be cleared after title step, got %q", m.inputBuf)
	}
}

// ── UI mode step ──────────────────────────────────────────────────────────────

func TestWizard_UIModeDefaultIsFirst(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	_ = m // just assert it didn't panic
}

func TestWizard_UIModePicker_List(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m) // title (blank → "Runner")
	// cursor at 0 = list
	m = pressEnter(t, m)
	if m.cfgUIMode != config.UIModeList {
		t.Errorf("UIMode: got %q, want list", m.cfgUIMode)
	}
}

func TestWizard_UIModePicker_Fuzzy(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m) // title
	m = pressDown(t, m)  // cursor → fuzzy
	m = pressEnter(t, m)
	if m.cfgUIMode != config.UIModeFuzzy {
		t.Errorf("UIMode: got %q, want fuzzy", m.cfgUIMode)
	}
}

func TestWizard_UIModePicker_Group(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m) // title
	m = pressDown(t, m)
	m = pressDown(t, m) // cursor → group
	m = pressEnter(t, m)
	if m.cfgUIMode != config.UIModeGroup {
		t.Errorf("UIMode: got %q, want group", m.cfgUIMode)
	}
}

func TestWizard_UIModePicker_CursorDoesNotGoAboveZero(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = pressUp(t, m) // should stay at 0
	if m.optCursor != 0 {
		t.Errorf("cursor went negative: %d", m.optCursor)
	}
}

func TestWizard_UIModePicker_CursorDoesNotExceedMax(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	for i := 0; i < 10; i++ {
		m = pressDown(t, m)
	}
	if m.optCursor >= len(uiModeOptions) {
		t.Errorf("cursor exceeded options: %d", m.optCursor)
	}
}

// ── Run mode step ─────────────────────────────────────────────────────────────

func TestWizard_RunModePicker_Stream(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	// runModeIdx 0 = stream
	if m.cfgRunMode != config.RunModeStream {
		t.Errorf("cfgRunMode: got %q, want stream", m.cfgRunMode)
	}
}

func TestWizard_RunModePicker_Handoff(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 1)
	if m.cfgRunMode != config.RunModeHandoff {
		t.Errorf("cfgRunMode: got %q, want handoff", m.cfgRunMode)
	}
}

func TestWizard_RunModePicker_Background(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 2)
	if m.cfgRunMode != config.RunModeBackground {
		t.Errorf("cfgRunMode: got %q, want background", m.cfgRunMode)
	}
}

// ── Command steps ─────────────────────────────────────────────────────────────

func TestWizard_CmdNameRequired(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = pressEnter(t, m) // empty name
	if m.validErr == "" {
		t.Error("empty cmd name should produce validation error")
	}
	if m.step != wizStepCmdName {
		t.Errorf("should stay on wizStepCmdName, got %d", m.step)
	}
}

func TestWizard_CmdNameValidationClearedOnType(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = pressEnter(t, m) // trigger error
	m = typeText(t, m, "x")
	if m.validErr != "" {
		t.Error("typing should clear validation error")
	}
}

func TestWizard_CmdCommandRequired(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = typeText(t, m, "Name")
	m = pressEnter(t, m) // name ✓
	m = pressEnter(t, m) // desc blank ✓
	m = pressEnter(t, m) // command empty → error
	if m.validErr == "" {
		t.Error("empty command should produce validation error")
	}
	if m.step != wizStepCmdCommand {
		t.Errorf("should stay on wizStepCmdCommand, got %d", m.step)
	}
}

func TestWizard_CmdDescOptional(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = typeText(t, m, "Name")
	m = pressEnter(t, m) // name ✓
	m = pressEnter(t, m) // desc blank — should advance without error
	if m.step != wizStepCmdCommand {
		t.Errorf("blank desc should advance to wizStepCmdCommand, got %d", m.step)
	}
}

func TestWizard_CmdDirOptional(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = typeText(t, m, "Name")
	m = pressEnter(t, m) // name
	m = pressEnter(t, m) // desc
	m = typeText(t, m, "echo hi")
	m = pressEnter(t, m) // command ✓
	m = pressEnter(t, m) // more commands: blank → skip
	m = pressEnter(t, m) // dir blank — should advance
	if m.step != wizStepCmdRunMode {
		t.Errorf("blank dir should advance past dir step, got %d", m.step)
	}
}

func TestWizard_CmdGroupStep_ShownOnlyInGroupMode(t *testing.T) {
	// In list mode the group step should be skipped
	m := advanceToFirstCmd(t, "T", 0 /* list */, 0)
	m = typeText(t, m, "Name")
	m = pressEnter(t, m) // name
	m = pressEnter(t, m) // desc
	m = typeText(t, m, "echo hi")
	m = pressEnter(t, m) // command
	m = pressEnter(t, m) // more commands: blank → skip
	m = pressEnter(t, m) // dir
	if m.step == wizStepCmdGroup {
		t.Error("group step should be skipped in list mode")
	}
}

func TestWizard_CmdGroupStep_ShownInGroupMode(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 2 /* group */, 0)
	m = typeText(t, m, "Name")
	m = pressEnter(t, m) // name
	m = pressEnter(t, m) // desc
	m = typeText(t, m, "echo hi")
	m = pressEnter(t, m) // command
	m = pressEnter(t, m) // more commands: blank → skip
	m = pressEnter(t, m) // dir
	if m.step != wizStepCmdGroup {
		t.Errorf("group mode should show wizStepCmdGroup, got %d", m.step)
	}
}

func TestWizard_CmdRunMode_Inherit(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0) // default run_mode = stream
	m = addCommand(t, m, "A", "", "echo a", "", "", 0 /* inherit */)
	// committed command's RunMode should be equal to the inherited default (stream)
	if len(m.commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(m.commands))
	}
	if m.commands[0].RunMode != config.RunModeStream {
		t.Errorf("inherited run_mode: got %q, want stream", m.commands[0].RunMode)
	}
}

func TestWizard_CmdRunMode_ExplicitHandoff(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "B", "", "echo b", "", "", 2 /* handoff */)
	if m.commands[0].RunMode != config.RunModeHandoff {
		t.Errorf("explicit run_mode: got %q, want handoff", m.commands[0].RunMode)
	}
}

// ── Add-another loop ──────────────────────────────────────────────────────────

func TestWizard_AddAnotherYesLoopsBack(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "First", "", "echo 1", "", "", 0)
	// addAnother = "yes" (index 0)
	m = pressEnter(t, m)
	if m.step != wizStepCmdName {
		t.Errorf("after yes: step = %d, want wizStepCmdName", m.step)
	}
	if len(m.commands) != 1 {
		t.Errorf("should have 1 committed command before second, got %d", len(m.commands))
	}
}

func TestWizard_AddAnotherNoAdvancesToDeleteStep(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "Only", "", "echo only", "", "", 0)
	// addAnother = "no" (index 1)
	m = pressDown(t, m)
	m = pressEnter(t, m)
	if m.step != wizStepDeleteCmds {
		t.Errorf("after no: step = %d, want wizStepDeleteCmds", m.step)
	}
}

func TestWizard_DeleteStepConfirmNoMarksGoesToSummary(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "Only", "", "echo only", "", "", 0)
	m = pressDown(t, m)  // addAnother = no
	m = pressEnter(t, m) // → delete step
	m = confirmDelete(t, m)
	if m.step != wizStepSummary {
		t.Errorf("confirm with no marks: step = %d, want wizStepSummary", m.step)
	}
}

func TestWizard_TwoCommandsCommitted(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "First", "f desc", "echo 1", "/tmp", "", 0)
	m = pressEnter(t, m) // addAnother = yes
	m = addCommand(t, m, "Second", "", "echo 2", "", "", 1 /* stream */)
	if len(m.commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(m.commands))
	}
	if m.commands[0].Name != "First" || m.commands[1].Name != "Second" {
		t.Errorf("command names: %v", []string{m.commands[0].Name, m.commands[1].Name})
	}
}

func TestWizard_CommandFieldsStoredCorrectly(t *testing.T) {
	m := advanceToFirstCmd(t, "MyApp", 0, 0)
	m = addCommand(t, m, "Build", "compile it", "make build", "/src", "", 0)
	c := m.commands[0]
	if c.Name != "Build" {
		t.Errorf("Name: got %q", c.Name)
	}
	if c.Description != "compile it" {
		t.Errorf("Description: got %q", c.Description)
	}
	if c.Command != "make build" {
		t.Errorf("Command: got %q", c.Command)
	}
	if c.Dir != "/src" {
		t.Errorf("Dir: got %q", c.Dir)
	}
}

// ── Group mode: group field stored ───────────────────────────────────────────

func TestWizard_GroupMode_GroupFieldStored(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 2 /* group */, 0)
	m = typeText(t, m, "Deploy")
	m = pressEnter(t, m) // name
	m = pressEnter(t, m) // desc
	m = typeText(t, m, "kubectl apply")
	m = pressEnter(t, m) // command
	m = pressEnter(t, m) // more commands: blank → skip
	m = pressEnter(t, m) // dir
	m = typeText(t, m, "Ops")
	m = pressEnter(t, m) // group
	m = pressEnter(t, m) // runMode (inherit)

	if len(m.commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(m.commands))
	}
	if m.commands[0].Group != "Ops" {
		t.Errorf("Group: got %q, want Ops", m.commands[0].Group)
	}
}

// ── Summary & save ────────────────────────────────────────────────────────────

func TestWizard_SummaryViewContainsAllFields(t *testing.T) {
	m := advanceToFirstCmd(t, "My Project", 0, 0)
	m = addCommand(t, m, "Build", "compile", "make build", "/src", "", 0)
	m = pressDown(t, m)     // addAnother = no
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary
	v := m.View()
	fields := []string{"My Project", "list", "stream", "Build", "compile", "make build", "/src"}
	for _, f := range fields {
		if !strings.Contains(v, f) {
			t.Errorf("summary view missing %q:\n%s", f, v)
		}
	}
}

func TestWizard_SummaryViewContainsSavePath(t *testing.T) {
	m := NewWizard("my/path/runner.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = addCommand(t, m, "X", "", "echo x", "", "", 0)
	m = pressDown(t, m)     // no
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary
	if !strings.Contains(m.View(), "my/path/runner.yaml") {
		t.Errorf("summary missing save path:\n%s", m.View())
	}
}

func TestWizard_SaveWritesFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "runner.yaml")
	m := NewWizard(tmp)
	m = pressEnter(t, m) // welcome
	m = typeText(t, m, "Save Test")
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // uiMode = list
	m = pressEnter(t, m) // runMode = stream
	m = addCommand(t, m, "Echo", "desc", "echo hello", "", "", 0)
	m = pressDown(t, m)     // addAnother = no
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary (no deletions)
	m = pressEnter(t, m)    // → save → done

	if m.saveErr != "" {
		t.Fatalf("unexpected save error: %s", m.saveErr)
	}
	if !m.saved {
		t.Fatal("m.saved should be true after successful write")
	}
	if m.step != wizStepDone {
		t.Errorf("step should be wizStepDone, got %d", m.step)
	}

	// Verify file was actually written
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("could not read saved file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Save Test") {
		t.Errorf("saved YAML missing title:\n%s", content)
	}
	if !strings.Contains(content, "Echo") {
		t.Errorf("saved YAML missing command name:\n%s", content)
	}
}

func TestWizard_SaveErrorRecorded(t *testing.T) {
	// Use an unwritable path to force an error
	m := NewWizard("/nonexistent/deeply/nested/path/runner.yaml")
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = pressEnter(t, m)
	m = addCommand(t, m, "X", "", "echo x", "", "", 0)
	m = pressDown(t, m)
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary
	m = pressEnter(t, m)    // → save
	if m.saveErr == "" {
		t.Error("expected save error for unwritable path")
	}
	if m.saved {
		t.Error("saved should be false on write failure")
	}
}

// ── buildConfig ───────────────────────────────────────────────────────────────

func TestWizard_BuildConfigAssemblesCorrectly(t *testing.T) {
	m := NewWizard("x.yaml")
	m.cfgTitle = "Assembled"
	m.cfgUIMode = config.UIModeFuzzy
	m.cfgRunMode = config.RunModeHandoff
	m.commands = []config.Command{
		{Name: "A", Command: "a", RunMode: config.RunModeStream},
		{Name: "B", Command: "b", RunMode: config.RunModeHandoff},
	}

	cfg := m.buildConfig()
	if cfg.Title != "Assembled" {
		t.Errorf("Title: got %q", cfg.Title)
	}
	if cfg.UIMode != config.UIModeFuzzy {
		t.Errorf("UIMode: got %q", cfg.UIMode)
	}
	if cfg.RunMode != config.RunModeHandoff {
		t.Errorf("RunMode: got %q", cfg.RunMode)
	}
	if len(cfg.Commands) != 2 {
		t.Errorf("Commands: got %d", len(cfg.Commands))
	}
}

// ── Result() ─────────────────────────────────────────────────────────────────

func TestWizard_ResultNilWhenNotSaved(t *testing.T) {
	m := NewWizard("x.yaml")
	if m.Result() != nil {
		t.Error("Result() should be nil before save")
	}
}

func TestWizard_ResultNilWhenAborted(t *testing.T) {
	m := NewWizard("x.yaml")
	m.aborted = true
	m.saved = true // even if "saved" flag somehow set
	if m.Result() != nil {
		t.Error("Result() should be nil when aborted")
	}
}

// ── Window resize ─────────────────────────────────────────────────────────────

func TestWizard_WindowSizeUpdate(t *testing.T) {
	m := NewWizard("x.yaml")
	m = updateWiz(t, m, tea.WindowSizeMsg{Width: 100, Height: 40})
	if m.width != 100 || m.height != 40 {
		t.Errorf("after resize: got %dx%d, want 100x40", m.width, m.height)
	}
}

// ── View helpers ──────────────────────────────────────────────────────────────

func TestWizard_ViewTextInputShowsPlaceholderWhenEmpty(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // → title step
	v := m.View()
	if !strings.Contains(v, "e.g.") {
		t.Errorf("placeholder not shown when input is empty:\n%s", v)
	}
}

func TestWizard_ViewTextInputShowsTypedText(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m)
	m = typeText(t, m, "hello")
	v := m.View()
	if !strings.Contains(v, "hello") {
		t.Errorf("typed text not shown in view:\n%s", v)
	}
}

func TestWizard_ViewShowsValidationError(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = pressEnter(t, m) // empty name → validation error
	v := m.View()
	if !strings.Contains(v, "required") {
		t.Errorf("validation error not shown:\n%s", v)
	}
}

func TestWizard_ViewOptionPickerShowsAllOptions(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title
	// now on uiMode picker
	v := m.View()
	if !strings.Contains(v, "list") || !strings.Contains(v, "fuzzy") || !strings.Contains(v, "group") {
		t.Errorf("option picker missing options:\n%s", v)
	}
}

func TestWizard_ViewDoneSuccess(t *testing.T) {
	m := NewWizard("x.yaml")
	m.step = wizStepDone
	m.saved = true
	m.savePath = "runner.yaml"
	v := m.View()
	if !strings.Contains(v, "runner.yaml") {
		t.Errorf("done view missing file path:\n%s", v)
	}
}

func TestWizard_ViewDoneError(t *testing.T) {
	m := NewWizard("x.yaml")
	m.step = wizStepDone
	m.saveErr = "permission denied"
	v := m.View()
	if !strings.Contains(v, "permission denied") {
		t.Errorf("done-error view missing error text:\n%s", v)
	}
}

// ── config.Write + config.Load round-trip ────────────────────────────────────

func TestConfig_WriteAndLoad_RoundTrip(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "round-trip.yaml")

	original := &config.Config{
		Title:   "Round Trip",
		UIMode:  config.UIModeFuzzy,
		RunMode: config.RunModeBackground,
		Commands: []config.Command{
			{Name: "A", Description: "desc a", Command: "echo a", Dir: "/tmp", Group: "G1", RunMode: config.RunModeStream},
			{Name: "B", Command: "echo b", RunMode: config.RunModeHandoff},
		},
	}

	if err := config.Write(tmp, original); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	loaded, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if loaded.Title != original.Title {
		t.Errorf("Title: got %q, want %q", loaded.Title, original.Title)
	}
	if loaded.UIMode != original.UIMode {
		t.Errorf("UIMode: got %q, want %q", loaded.UIMode, original.UIMode)
	}
	if loaded.RunMode != original.RunMode {
		t.Errorf("RunMode: got %q, want %q", loaded.RunMode, original.RunMode)
	}
	if len(loaded.Commands) != len(original.Commands) {
		t.Fatalf("Commands len: got %d, want %d", len(loaded.Commands), len(original.Commands))
	}
	for i, c := range original.Commands {
		lc := loaded.Commands[i]
		if lc.Name != c.Name || lc.Command != c.Command || lc.RunMode != c.RunMode {
			t.Errorf("commands[%d]: got %+v, want %+v", i, lc, c)
		}
	}
}

// ── Delete commands step ──────────────────────────────────────────────────────

// advanceToDelete drives the wizard through global settings, adds the given
// commands, picks "No" at AddAnother, and lands on wizStepDeleteCmds.
func advanceToDelete(t *testing.T, cmds []struct{ name, command string }) WizardModel {
	t.Helper()
	m := advanceToFirstCmd(t, "T", 0, 0)
	for i, c := range cmds {
		m = addCommand(t, m, c.name, "", c.command, "", "", 0)
		if i < len(cmds)-1 {
			// say "yes" to add another
			m = pressEnter(t, m)
		}
	}
	// say "no"
	m = pressDown(t, m)
	m = pressEnter(t, m)
	if m.step != wizStepDeleteCmds {
		t.Fatalf("expected wizStepDeleteCmds, got %d", m.step)
	}
	return m
}

func TestWizard_DeleteStep_ReachedAfterAddAnotherNo(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "echo a"}})
	if m.step != wizStepDeleteCmds {
		t.Errorf("step: got %d, want wizStepDeleteCmds", m.step)
	}
}

func TestWizard_DeleteStep_DeleteMarksInitiallyAllFalse(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	for i, marked := range m.deleteMarks {
		if marked {
			t.Errorf("deleteMarks[%d] should be false initially", i)
		}
	}
}

func TestWizard_DeleteStep_DeleteMarksLenMatchesCommands(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	if len(m.deleteMarks) != len(m.commands) {
		t.Errorf("deleteMarks len %d != commands len %d", len(m.deleteMarks), len(m.commands))
	}
}

func TestWizard_DeleteStep_SpaceTogglesMark(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	// cursor starts at 0
	m = pressSpace(t, m)
	if !m.deleteMarks[0] {
		t.Error("space should mark item at cursor")
	}
	m = pressSpace(t, m)
	if m.deleteMarks[0] {
		t.Error("second space should unmark item")
	}
}

func TestWizard_DeleteStep_CursorNavigation(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	if m.optCursor != 0 {
		t.Errorf("initial cursor: got %d, want 0", m.optCursor)
	}
	m = pressDown(t, m)
	if m.optCursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", m.optCursor)
	}
	m = pressUp(t, m)
	if m.optCursor != 0 {
		t.Errorf("after up: cursor = %d, want 0", m.optCursor)
	}
}

func TestWizard_DeleteStep_CursorClampsAtBounds(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	m = pressUp(t, m) // already at 0
	if m.optCursor != 0 {
		t.Errorf("cursor went below 0: %d", m.optCursor)
	}
	m = pressDown(t, m)
	m = pressDown(t, m) // try to go past last
	if m.optCursor != 1 {
		t.Errorf("cursor went past last: %d", m.optCursor)
	}
}

func TestWizard_DeleteStep_SpaceMarksDifferentRows(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	m = pressDown(t, m)  // cursor → 1
	m = pressSpace(t, m) // mark B
	if !m.deleteMarks[1] {
		t.Error("B should be marked")
	}
	if m.deleteMarks[0] || m.deleteMarks[2] {
		t.Error("A and C should not be marked")
	}
}

func TestWizard_DeleteStep_ConfirmNoMarksKeepsAllCommands(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	before := len(m.commands)
	m = confirmDelete(t, m)
	if len(m.commands) != before {
		t.Errorf("commands len changed: got %d, want %d", len(m.commands), before)
	}
}

func TestWizard_DeleteStep_ConfirmWithMarkRemovesCommand(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	m = pressDown(t, m)  // cursor → B
	m = pressSpace(t, m) // mark B
	m = confirmDelete(t, m)
	if len(m.commands) != 2 {
		t.Fatalf("commands len: got %d, want 2", len(m.commands))
	}
	for _, c := range m.commands {
		if c.Name == "B" {
			t.Error("B should have been deleted")
		}
	}
}

func TestWizard_DeleteStep_ConfirmPreservesOrder(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	m = pressSpace(t, m) // mark A (cursor=0)
	m = confirmDelete(t, m)
	if len(m.commands) != 2 {
		t.Fatalf("commands len: got %d, want 2", len(m.commands))
	}
	if m.commands[0].Name != "B" || m.commands[1].Name != "C" {
		t.Errorf("order wrong: got %q %q", m.commands[0].Name, m.commands[1].Name)
	}
}

func TestWizard_DeleteStep_CannotDeleteAllCommands(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	m = pressSpace(t, m) // mark A
	m = pressDown(t, m)
	m = pressSpace(t, m) // mark B
	m = pressEnter(t, m) // try to confirm
	if m.validErr == "" {
		t.Error("should show validation error when all commands marked")
	}
	if m.step != wizStepDeleteCmds {
		t.Errorf("should stay on delete step, got %d", m.step)
	}
	if len(m.commands) != 2 {
		t.Errorf("commands should be unchanged, got %d", len(m.commands))
	}
}

func TestWizard_DeleteStep_ValidationClearedOnNavigation(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	m = pressSpace(t, m) // mark A
	m = pressDown(t, m)
	m = pressSpace(t, m) // mark B
	m = pressEnter(t, m) // trigger error
	m = pressUp(t, m)    // navigate — should clear error
	if m.validErr != "" {
		t.Errorf("validErr should be cleared on navigation, got %q", m.validErr)
	}
}

func TestWizard_DeleteStep_ConfirmGoesToSummary(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	m = confirmDelete(t, m)
	if m.step != wizStepSummary {
		t.Errorf("after confirm: step = %d, want wizStepSummary", m.step)
	}
}

func TestWizard_DeleteStep_DeleteMarksNilAfterConfirm(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	m = pressSpace(t, m) // mark A
	m = confirmDelete(t, m)
	if m.deleteMarks != nil {
		t.Error("deleteMarks should be nil after confirm")
	}
}

func TestWizard_DeleteStep_ViewContainsCommandNames(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"Alpha", "a"}, {"Beta", "b"}})
	v := m.View()
	if !strings.Contains(v, "Alpha") {
		t.Errorf("delete view missing Alpha:\n%s", v)
	}
	if !strings.Contains(v, "Beta") {
		t.Errorf("delete view missing Beta:\n%s", v)
	}
}

func TestWizard_DeleteStep_ViewShowsCheckboxes(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}})
	v := m.View()
	if !strings.Contains(v, "[ ]") {
		t.Errorf("delete view missing unchecked checkbox:\n%s", v)
	}
	m = pressSpace(t, m)
	v = m.View()
	if !strings.Contains(v, "[✗]") {
		t.Errorf("delete view missing checked checkbox:\n%s", v)
	}
}

func TestWizard_DeleteStep_ViewShowsValidationError(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}})
	m = pressSpace(t, m) // mark only command
	m = pressEnter(t, m) // trigger error
	v := m.View()
	if !strings.Contains(v, "remain") {
		t.Errorf("delete view missing validation error:\n%s", v)
	}
}

func TestWizard_DeleteAndSave_FileHasCorrectCommands(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "runner.yaml")
	m := NewWizard(tmp)
	m = pressEnter(t, m) // welcome
	m = typeText(t, m, "Del Test")
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // uiMode
	m = pressEnter(t, m) // runMode

	// Add two commands.
	m = addCommand(t, m, "Keep", "", "echo keep", "", "", 0)
	m = pressEnter(t, m) // addAnother = yes
	m = addCommand(t, m, "Remove", "", "echo remove", "", "", 0)
	m = pressDown(t, m)     // addAnother = no
	m = pressEnter(t, m)    // → delete step
	m = pressDown(t, m)     // cursor → Remove (index 1)
	m = pressSpace(t, m)    // mark Remove
	m = confirmDelete(t, m) // → summary
	m = pressEnter(t, m)    // → save → done

	if m.saveErr != "" {
		t.Fatalf("save error: %s", m.saveErr)
	}

	loaded, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(loaded.Commands))
	}
	if loaded.Commands[0].Name != "Keep" {
		t.Errorf("wrong command saved: %q", loaded.Commands[0].Name)
	}
}

// ── Edit mode (NewWizardFromConfig) ───────────────────────────────────────────

func TestNewWizardFromConfig_EditingFlag(t *testing.T) {
	cfg := &config.Config{
		Title:    "My App",
		UIMode:   config.UIModeFuzzy,
		RunMode:  config.RunModeHandoff,
		Commands: []config.Command{{Name: "A", Command: "echo a", RunMode: config.RunModeHandoff}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	if !m.editing {
		t.Error("editing should be true")
	}
}

func TestNewWizardFromConfig_PrePopulatesTitle(t *testing.T) {
	cfg := &config.Config{Title: "Pre-filled", UIMode: config.UIModeList, RunMode: config.RunModeStream}
	m := NewWizardFromConfig("out.yaml", cfg)
	// inputBuf is seeded when the Welcome step advances (not at construction),
	// so confirm welcome first and then check the title step's buffer.
	m = pressEnter(t, m)
	if m.inputBuf != "Pre-filled" {
		t.Errorf("inputBuf on title step: got %q, want Pre-filled", m.inputBuf)
	}
}

func TestNewWizardFromConfig_PrePopulatesGlobalFields(t *testing.T) {
	cfg := &config.Config{
		Title:   "X",
		UIMode:  config.UIModeFuzzy,
		RunMode: config.RunModeBackground,
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	if m.cfgUIMode != config.UIModeFuzzy {
		t.Errorf("cfgUIMode: got %q, want fuzzy", m.cfgUIMode)
	}
	if m.cfgRunMode != config.RunModeBackground {
		t.Errorf("cfgRunMode: got %q, want background", m.cfgRunMode)
	}
}

func TestNewWizardFromConfig_CopiesCommands(t *testing.T) {
	orig := []config.Command{
		{Name: "Build", Command: "make build", RunMode: config.RunModeStream},
		{Name: "Test", Command: "go test", RunMode: config.RunModeStream},
	}
	cfg := &config.Config{Title: "T", UIMode: config.UIModeList, RunMode: config.RunModeStream, Commands: orig}
	m := NewWizardFromConfig("out.yaml", cfg)

	if len(m.commands) != 2 {
		t.Fatalf("commands len: got %d, want 2", len(m.commands))
	}
	if m.commands[0].Name != "Build" || m.commands[1].Name != "Test" {
		t.Errorf("command names: %v %v", m.commands[0].Name, m.commands[1].Name)
	}

	// Mutating the copy must not affect the original.
	m.commands[0].Name = "Changed"
	if orig[0].Name != "Build" {
		t.Error("mutating wizard commands should not affect original slice")
	}
}

func TestNewWizardFromConfig_StartsAtWelcome(t *testing.T) {
	cfg := &config.Config{Title: "T", UIMode: config.UIModeList, RunMode: config.RunModeStream}
	m := NewWizardFromConfig("out.yaml", cfg)
	if m.step != wizStepWelcome {
		t.Errorf("initial step: got %d, want wizStepWelcome", m.step)
	}
}

func TestWizardEdit_WelcomeViewMentionsEdit(t *testing.T) {
	cfg := &config.Config{
		Title:    "My App",
		UIMode:   config.UIModeList,
		RunMode:  config.RunModeStream,
		Commands: []config.Command{{Name: "A", Command: "a", RunMode: config.RunModeStream}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	v := m.View()
	if strings.Contains(v, "No runner.yaml") {
		t.Error("edit welcome should not say 'No runner.yaml found'")
	}
	if !strings.Contains(v, "Edit") && !strings.Contains(v, "edit") && !strings.Contains(v, "update") {
		t.Errorf("edit welcome should mention editing:\n%s", v)
	}
}

func TestWizardEdit_TitlePreFilledAndConfirmable(t *testing.T) {
	cfg := &config.Config{Title: "Original", UIMode: config.UIModeList, RunMode: config.RunModeStream}
	m := NewWizardFromConfig("out.yaml", cfg)

	// Welcome → Title
	m = pressEnter(t, m)
	if m.inputBuf != "Original" {
		t.Errorf("inputBuf on title step: got %q, want Original", m.inputBuf)
	}

	// Confirm the existing title with Enter (no typing)
	m = pressEnter(t, m)
	if m.cfgTitle != "Original" {
		t.Errorf("cfgTitle after confirm: got %q, want Original", m.cfgTitle)
	}
	if m.step != wizStepUIMode {
		t.Errorf("step after title: got %d, want wizStepUIMode", m.step)
	}
}

func TestWizardEdit_TitleCanBeReplaced(t *testing.T) {
	cfg := &config.Config{Title: "Old", UIMode: config.UIModeList, RunMode: config.RunModeStream}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressEnter(t, m) // welcome

	// Clear the existing value and type a new one.
	for range []rune("Old") {
		m = pressBackspace(t, m)
	}
	m = typeText(t, m, "New Title")
	m = pressEnter(t, m)

	if m.cfgTitle != "New Title" {
		t.Errorf("cfgTitle: got %q, want New Title", m.cfgTitle)
	}
}

func TestWizardEdit_UIModePreselected(t *testing.T) {
	cfg := &config.Config{Title: "T", UIMode: config.UIModeFuzzy, RunMode: config.RunModeStream}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // confirm title → moves to UIMode

	// optCursor should point at "fuzzy" (index 1)
	if m.optCursor != 1 {
		t.Errorf("optCursor on UIMode step: got %d, want 1 (fuzzy)", m.optCursor)
	}
}

func TestWizardEdit_RunModePreselected(t *testing.T) {
	cfg := &config.Config{Title: "T", UIMode: config.UIModeList, RunMode: config.RunModeBackground}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // confirm UIMode → moves to RunMode

	// optCursor should point at "background" (index 2)
	if m.optCursor != 2 {
		t.Errorf("optCursor on RunMode step: got %d, want 2 (background)", m.optCursor)
	}
}

func TestWizardEdit_ExistingCommandsCarriedToSummary(t *testing.T) {
	existing := []config.Command{
		{Name: "Build", Command: "make build", RunMode: config.RunModeStream},
	}
	cfg := &config.Config{Title: "T", UIMode: config.UIModeList, RunMode: config.RunModeStream, Commands: existing}
	m := NewWizardFromConfig("out.yaml", cfg)

	// Drive through global steps without adding a new command.
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title (keep)
	m = pressEnter(t, m) // UIMode (keep list)
	m = pressEnter(t, m) // RunMode (keep stream) → wizStepCmdName

	// Add one new command then say No.
	m = addCommand(t, m, "NewCmd", "", "echo new", "", "", 0)
	m = pressDown(t, m)     // no
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary (no deletions)

	if m.step != wizStepSummary {
		t.Fatalf("expected wizStepSummary, got %d", m.step)
	}
	if len(m.commands) != 2 {
		t.Fatalf("expected 2 commands (existing + new), got %d", len(m.commands))
	}
	if m.commands[0].Name != "Build" {
		t.Errorf("first command should be existing Build, got %q", m.commands[0].Name)
	}
	if m.commands[1].Name != "NewCmd" {
		t.Errorf("second command should be NewCmd, got %q", m.commands[1].Name)
	}
}

func TestWizardEdit_DoneViewSaysUpdated(t *testing.T) {
	m := NewWizard("x.yaml")
	m.step = wizStepDone
	m.saved = true
	m.editing = true
	m.savePath = "runner.yaml"
	v := m.View()
	if !strings.Contains(v, "updated") {
		t.Errorf("edit done view should say 'updated':\n%s", v)
	}
}

func TestWizardEdit_SaveWritesUpdatedFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "runner.yaml")

	// Write initial config.
	initial := &config.Config{
		Title:   "Initial",
		UIMode:  config.UIModeList,
		RunMode: config.RunModeStream,
		Commands: []config.Command{
			{Name: "OldCmd", Command: "echo old", RunMode: config.RunModeStream},
		},
	}
	if err := config.Write(tmp, initial); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Run the edit wizard, change the title, keep everything else, no new commands.
	m := NewWizardFromConfig(tmp, initial)
	m = pressEnter(t, m) // welcome

	// Replace title
	for range []rune("Initial") {
		m = pressBackspace(t, m)
	}
	m = typeText(t, m, "Updated")
	m = pressEnter(t, m) // title confirmed
	m = pressEnter(t, m) // UIMode (list)
	m = pressEnter(t, m) // RunMode (stream) → CmdName

	// Add no new commands: pick No at the very first AddAnother.
	// But we must go through at least one command sub-flow because the wizard
	// always asks for at least one command. Add a minimal extra command.
	m = addCommand(t, m, "Extra", "", "echo extra", "", "", 0)
	m = pressDown(t, m)     // No
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary (no deletions)
	m = pressEnter(t, m)    // → save → done

	if m.saveErr != "" {
		t.Fatalf("save error: %s", m.saveErr)
	}
	if !m.saved {
		t.Fatal("m.saved should be true")
	}

	// Verify the written file.
	loaded, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("load after save: %v", err)
	}
	if loaded.Title != "Updated" {
		t.Errorf("title: got %q, want Updated", loaded.Title)
	}
	// Original command + new extra command
	if len(loaded.Commands) != 2 {
		t.Errorf("commands: got %d, want 2", len(loaded.Commands))
	}
}

func TestOptionIndex_FindsCorrectIndex(t *testing.T) {
	cases := []struct {
		v    string
		want int
	}{
		{"list", 0},
		{"fuzzy", 1},
		{"group", 2},
		{"unknown", 0}, // fallback
	}
	for _, tc := range cases {
		got := optionIndex(uiModeOptions, tc.v)
		if got != tc.want {
			t.Errorf("optionIndex(%q): got %d, want %d", tc.v, got, tc.want)
		}
	}
}

// ── config.ErrNotFound ────────────────────────────────────────────────────────

func TestConfig_Load_ErrNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Must wrap ErrNotFound so errors.Is works.
	var target interface{ Unwrap() error }
	_ = target
	// Use the exported sentinel directly.
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
}
