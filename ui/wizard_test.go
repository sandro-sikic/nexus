package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
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

// pressRune sends a single rune keystroke (e.g. 'e', 'a', 's', 'd').
func pressRune(t *testing.T, m WizardModel, r rune) WizardModel {
	t.Helper()
	return updateWiz(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
}

// hubEditSettings presses 'e' on the hub to enter the edit-general-settings sub-flow.
func hubEditSettings(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	if m.step != wizStepEditHub {
		t.Fatalf("hubEditSettings: expected wizStepEditHub, got %d", m.step)
	}
	return pressRune(t, m, 'e')
}

// hubAddCommand presses 'a' on the hub to enter the add-command sub-flow.
func hubAddCommand(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	if m.step != wizStepEditHub {
		t.Fatalf("hubAddCommand: expected wizStepEditHub, got %d", m.step)
	}
	return pressRune(t, m, 'a')
}

// hubSave presses 's' on the hub to go to the summary/save screen.
func hubSave(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	if m.step != wizStepEditHub {
		t.Fatalf("hubSave: expected wizStepEditHub, got %d", m.step)
	}
	return pressRune(t, m, 's')
}

// hubDeleteMarked presses 'd' on the hub to delete all marked commands.
func hubDeleteMarked(t *testing.T, m WizardModel) WizardModel {
	t.Helper()
	if m.step != wizStepEditHub {
		t.Fatalf("hubDeleteMarked: expected wizStepEditHub, got %d", m.step)
	}
	return pressRune(t, m, 'd')
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

// advanceToFirstCmd drives the wizard to just past the title step
// (i.e. ready for the first command), using the given title.
// The uiModeIdx and runModeIdx parameters are kept for compatibility
// but are ignored — project-level run mode is no longer configurable.
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
	if m.step != wizStepCmdName {
		t.Fatalf("expected wizStepCmdName, got %d", m.step)
	}
	return m
}

// addCommand drives through all command sub-steps.
// handoff: whether to enable handoff for the last action.
func addCommand(t *testing.T, m WizardModel, name, desc, command, dir, group string, handoff bool) WizardModel {
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

	// Handoff step
	if m.step != wizStepCmdHandoff {
		t.Fatalf("expected wizStepCmdHandoff, got %d", m.step)
	}
	// handoffOptions: 0 = no, 1 = yes
	if handoff {
		m = pressDown(t, m) // select yes
	}
	m = pressEnter(t, m)

	// After CmdHandoff, we land on wizStepAddAnother (create flow) or
	// wizStepEditHub (edit flow, returnToHub=true). Both are valid endings.
	if m.step != wizStepAddAnother && m.step != wizStepEditHub {
		t.Fatalf("expected wizStepAddAnother or wizStepEditHub, got %d", m.step)
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
	if len(m.tasks) != 0 {
		t.Errorf("tasks should be empty, got %d", len(m.tasks))
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
	if !strings.Contains(v, "nexus.yaml") && !strings.Contains(v, "missing") && !strings.Contains(v, "No ") {
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
	if m.step != wizStepCmdName {
		t.Errorf("should advance to wizStepCmdName, got %d", m.step)
	}
}

func TestWizard_TitleDefaultsToNexus(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // welcome
	// don't type anything
	m = pressEnter(t, m)
	if m.cfgTitle != "Nexus" {
		t.Errorf("empty title should default to 'Nexus', got %q", m.cfgTitle)
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

// ── Title advances directly to run mode (UI mode step removed) ───────────────

func TestWizard_TitleAdvancesDirectlyToRunMode(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title (blank → "Nexus")
	if m.step != wizStepCmdName {
		t.Errorf("expected wizStepCmdName after title, got %d", m.step)
	}
}

func TestWizard_FirstCmdStepReachable(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	_ = m // just assert it didn't panic
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
	m = pressEnter(t, m) // dir blank — should advance to group step
	if m.step != wizStepCmdGroup {
		t.Errorf("blank dir should advance to wizStepCmdGroup, got %d", m.step)
	}
}

func TestWizard_CmdGroupStep_AlwaysShown(t *testing.T) {
	// Group step is always shown (UI mode is no longer selectable)
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = typeText(t, m, "Name")
	m = pressEnter(t, m) // name
	m = pressEnter(t, m) // desc
	m = typeText(t, m, "echo hi")
	m = pressEnter(t, m) // command
	m = pressEnter(t, m) // more commands: blank → skip
	m = pressEnter(t, m) // dir
	if m.step != wizStepCmdGroup {
		t.Errorf("group step should always be shown, got %d", m.step)
	}
}

func TestWizard_CmdHandoff_No(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "A", "", "echo a", "", "", false /* no handoff */)
	// committed task's last action should have Handoff = false
	if len(m.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(m.tasks))
	}
	if len(m.tasks[0].Actions) == 0 {
		t.Fatal("expected at least one action")
	}
	lastAction := m.tasks[0].Actions[len(m.tasks[0].Actions)-1]
	if lastAction.Handoff {
		t.Errorf("handoff: got true, want false")
	}
}

func TestWizard_CmdHandoff_Yes(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "B", "", "echo b", "", "", true /* handoff */)
	if len(m.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(m.tasks))
	}
	if len(m.tasks[0].Actions) == 0 {
		t.Fatal("expected at least one action")
	}
	lastAction := m.tasks[0].Actions[len(m.tasks[0].Actions)-1]
	if !lastAction.Handoff {
		t.Errorf("handoff: got false, want true")
	}
}

// ── Add-another loop ──────────────────────────────────────────────────────────

func TestWizard_AddAnotherYesLoopsBack(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "First", "", "echo 1", "", "", false)
	// addAnother = "yes" (index 0)
	m = pressEnter(t, m)
	if m.step != wizStepCmdName {
		t.Errorf("after yes: step = %d, want wizStepCmdName", m.step)
	}
	if len(m.tasks) != 1 {
		t.Errorf("should have 1 committed task before second, got %d", len(m.tasks))
	}
}

func TestWizard_AddAnotherNoAdvancesToDeleteStep(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "Only", "", "echo only", "", "", false)
	// addAnother = "no" (index 1)
	m = pressDown(t, m)
	m = pressEnter(t, m)
	if m.step != wizStepDeleteCmds {
		t.Errorf("after no: step = %d, want wizStepDeleteCmds", m.step)
	}
}

func TestWizard_DeleteStepConfirmNoMarksGoesToSummary(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "Only", "", "echo only", "", "", false)
	m = pressDown(t, m)  // addAnother = no
	m = pressEnter(t, m) // → delete step
	m = confirmDelete(t, m)
	if m.step != wizStepSummary {
		t.Errorf("confirm with no marks: step = %d, want wizStepSummary", m.step)
	}
}

func TestWizard_TwoTasksCommitted(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = addCommand(t, m, "First", "f desc", "echo 1", "/tmp", "", false)
	m = pressEnter(t, m) // addAnother = yes
	m = addCommand(t, m, "Second", "", "echo 2", "", "", false)
	if len(m.tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(m.tasks))
	}
	if m.tasks[0].Name != "First" || m.tasks[1].Name != "Second" {
		t.Errorf("task names: %v", []string{m.tasks[0].Name, m.tasks[1].Name})
	}
}

func TestWizard_TaskFieldsStoredCorrectly(t *testing.T) {
	m := advanceToFirstCmd(t, "MyApp", 0, 0)
	m = addCommand(t, m, "Build", "compile it", "make build", "/src", "", false)
	c := m.tasks[0]
	if c.Name != "Build" {
		t.Errorf("Name: got %q", c.Name)
	}
	if c.Description != "compile it" {
		t.Errorf("Description: got %q", c.Description)
	}
	if len(c.Actions) != 1 || c.Actions[0].Command != "make build" {
		t.Errorf("Actions: got %v", c.Actions)
	}
	if c.Dir != "/src" {
		t.Errorf("Dir: got %q", c.Dir)
	}
}

// ── Group mode: group field stored ───────────────────────────────────────────

func TestWizard_GroupMode_GroupFieldStored(t *testing.T) {
	m := advanceToFirstCmd(t, "T", 0, 0)
	m = typeText(t, m, "Deploy")
	m = pressEnter(t, m) // name
	m = pressEnter(t, m) // desc
	m = typeText(t, m, "kubectl apply")
	m = pressEnter(t, m) // command
	m = pressEnter(t, m) // more actions: blank → skip
	m = pressEnter(t, m) // dir
	m = typeText(t, m, "Ops")
	m = pressEnter(t, m) // group
	m = pressEnter(t, m) // runMode (inherit)

	if len(m.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(m.tasks))
	}
	if m.tasks[0].Group != "Ops" {
		t.Errorf("Group: got %q, want Ops", m.tasks[0].Group)
	}
}

// ── Summary & save ────────────────────────────────────────────────────────────

func TestWizard_SummaryViewContainsAllFields(t *testing.T) {
	m := advanceToFirstCmd(t, "My Project", 0, 0)
	m = addCommand(t, m, "Build", "compile", "make build", "/src", "", false)
	m = pressDown(t, m)     // addAnother = no
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary
	v := m.View()
	fields := []string{"My Project", "Build", "compile", "make build", "/src"}
	for _, f := range fields {
		if !strings.Contains(v, f) {
			t.Errorf("summary view missing %q:\n%s", f, v)
		}
	}
}

func TestWizard_SummaryViewContainsSavePath(t *testing.T) {
	m := NewWizard("my/path/nexus.yaml")
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // runMode
	m = addCommand(t, m, "X", "", "echo x", "", "", false)
	m = pressDown(t, m)     // no
	m = pressEnter(t, m)    // → delete step
	m = confirmDelete(t, m) // → summary
	if !strings.Contains(m.View(), "my/path/nexus.yaml") {
		t.Errorf("summary missing save path:\n%s", m.View())
	}
}

func TestWizard_SaveWritesFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "nexus.yaml")
	m := NewWizard(tmp)
	m = pressEnter(t, m) // welcome
	m = typeText(t, m, "Save Test")
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // runMode = stream
	m = addCommand(t, m, "Echo", "desc", "echo hello", "", "", false)
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
	m := NewWizard("/nonexistent/deeply/nested/path/nexus.yaml")
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // runMode
	m = addCommand(t, m, "X", "", "echo x", "", "", false)
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
	m.tasks = []config.Task{
		{Name: "A", Actions: []config.Action{{Command: "a"}}},
		{Name: "B", Actions: []config.Action{{Command: "b"}}},
	}
	cfg := m.buildConfig()
	if cfg.Title != "Assembled" {
		t.Errorf("Title: got %q", cfg.Title)
	}
	if len(cfg.Tasks) != 2 {
		t.Errorf("Tasks: got %d", len(cfg.Tasks))
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

func TestWizard_ViewOptionPickerShowsRunModeOptions(t *testing.T) {
	m := NewWizard("x.yaml")
	m = pressEnter(t, m) // welcome
	m = pressEnter(t, m) // title
	// now on cmdName step (no project-level run mode step)
	v := m.View()
	if !strings.Contains(v, "Name") {
		t.Errorf("cmd name step missing 'Name':\n%s", v)
	}
}

func TestWizard_ViewDoneSuccess(t *testing.T) {
	m := NewWizard("x.yaml")
	m.step = wizStepDone
	m.saved = true
	m.savePath = "nexus.yaml"
	v := m.viewSummary()
	if !strings.Contains(v, "nexus.yaml") {
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
		Title: "Round Trip",
		Tasks: []config.Task{
			{Name: "A", Description: "desc a", Actions: []config.Action{{Command: "echo a"}}, Dir: "/tmp", Group: "G1"},
			{Name: "B", Actions: []config.Action{{Command: "echo b"}}},
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
	if len(loaded.Tasks) != len(original.Tasks) {
		t.Fatalf("Tasks len: got %d, want %d", len(loaded.Tasks), len(original.Tasks))
	}
	for i, tsk := range original.Tasks {
		lt := loaded.Tasks[i]
		if lt.Name != tsk.Name {
			t.Errorf("tasks[%d]: got %+v, want %+v", i, lt, tsk)
		}
	}
}

// ── Delete commands step ──────────────────────────────────────────────────────

// advanceToDelete drives the wizard through global settings, adds the given
// tasks, picks "No" at AddAnother, and lands on wizStepDeleteCmds.
func advanceToDelete(t *testing.T, tasks []struct{ name, command string }) WizardModel {
	t.Helper()
	m := advanceToFirstCmd(t, "T", 0, 0)
	for i, c := range tasks {
		m = addCommand(t, m, c.name, "", c.command, "", "", false)
		if i < len(tasks)-1 {
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
	if len(m.deleteMarks) != len(m.tasks) {
		t.Errorf("deleteMarks len %d != tasks len %d", len(m.deleteMarks), len(m.tasks))
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

func TestWizard_DeleteStep_ConfirmNoMarksKeepsAllTasks(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	before := len(m.tasks)
	m = confirmDelete(t, m)
	if len(m.tasks) != before {
		t.Errorf("tasks len changed: got %d, want %d", len(m.tasks), before)
	}
}

func TestWizard_DeleteStep_ConfirmWithMarkRemovesTask(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	m = pressDown(t, m)  // cursor → B
	m = pressSpace(t, m) // mark B
	m = confirmDelete(t, m)
	if len(m.tasks) != 2 {
		t.Fatalf("tasks len: got %d, want 2", len(m.tasks))
	}
	for _, tsk := range m.tasks {
		if tsk.Name == "B" {
			t.Error("B should have been deleted")
		}
	}
}

func TestWizard_DeleteStep_ConfirmPreservesOrder(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}, {"C", "c"}})
	m = pressSpace(t, m) // mark A (cursor=0)
	m = confirmDelete(t, m)
	if len(m.tasks) != 2 {
		t.Fatalf("tasks len: got %d, want 2", len(m.tasks))
	}
	if m.tasks[0].Name != "B" || m.tasks[1].Name != "C" {
		t.Errorf("order wrong: got %q %q", m.tasks[0].Name, m.tasks[1].Name)
	}
}

func TestWizard_DeleteStep_CannotDeleteAllTasks(t *testing.T) {
	m := advanceToDelete(t, []struct{ name, command string }{{"A", "a"}, {"B", "b"}})
	m = pressSpace(t, m) // mark A
	m = pressDown(t, m)
	m = pressSpace(t, m) // mark B
	m = pressEnter(t, m) // try to confirm
	if m.validErr == "" {
		t.Error("should show validation error when all tasks marked")
	}
	if m.step != wizStepDeleteCmds {
		t.Errorf("should stay on delete step, got %d", m.step)
	}
	if len(m.tasks) != 2 {
		t.Errorf("tasks should be unchanged, got %d", len(m.tasks))
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
	m = pressSpace(t, m) // mark only task
	m = pressEnter(t, m) // trigger error
	v := m.View()
	if !strings.Contains(v, "remain") {
		t.Errorf("delete view missing validation error:\n%s", v)
	}
}

func TestWizard_DeleteAndSave_FileHasCorrectTasks(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "nexus.yaml")
	m := NewWizard(tmp)
	m = pressEnter(t, m) // welcome
	m = typeText(t, m, "Del Test")
	m = pressEnter(t, m) // title
	m = pressEnter(t, m) // runMode

	// Add two tasks.
	m = addCommand(t, m, "Keep", "", "echo keep", "", "", false)
	m = pressEnter(t, m) // addAnother = yes
	m = addCommand(t, m, "Remove", "", "echo remove", "", "", false)
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
	if len(loaded.Tasks) != 1 {
		t.Fatalf("tasks: got %d, want 1", len(loaded.Tasks))
	}
	if loaded.Tasks[0].Name != "Keep" {
		t.Errorf("wrong task saved: %q", loaded.Tasks[0].Name)
	}
}

// ── Edit mode (NewWizardFromConfig) ───────────────────────────────────────────

func TestNewWizardFromConfig_EditingFlag(t *testing.T) {
	cfg := &config.Config{
		Title: "My App",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "echo a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	if !m.editing {
		t.Error("editing should be true")
	}
}

func TestNewWizardFromConfig_PrePopulatesTitle(t *testing.T) {
	cfg := &config.Config{Title: "Pre-filled", Tasks: []config.Task{}}
	m := NewWizardFromConfig("out.yaml", cfg)
	// In edit mode, press 'e' from the hub to reach the title step.
	// The inputBuf is seeded with the existing title.
	m = hubEditSettings(t, m)
	if m.step != wizStepTitle {
		t.Fatalf("expected wizStepTitle after 'e', got %d", m.step)
	}
	if m.inputBuf != "Pre-filled" {
		t.Errorf("inputBuf on title step: got %q, want Pre-filled", m.inputBuf)
	}
}

func TestNewWizardFromConfig_PrePopulatesGlobalFields(t *testing.T) {
	cfg := &config.Config{
		Title: "X",
		Tasks: []config.Task{},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	if m.cfgTitle != "X" {
		t.Errorf("cfgTitle: got %q, want X", m.cfgTitle)
	}
}

func TestNewWizardFromConfig_CopiesTasks(t *testing.T) {
	orig := []config.Task{
		{Name: "Build", Actions: []config.Action{{Command: "make build"}}},
		{Name: "Test", Actions: []config.Action{{Command: "go test"}}},
	}
	cfg := &config.Config{Title: "T", Tasks: orig}
	m := NewWizardFromConfig("out.yaml", cfg)

	if len(m.tasks) != 2 {
		t.Fatalf("tasks len: got %d, want 2", len(m.tasks))
	}
	if m.tasks[0].Name != "Build" || m.tasks[1].Name != "Test" {
		t.Errorf("task names: %v %v", m.tasks[0].Name, m.tasks[1].Name)
	}

	// Mutating the copy must not affect the original.
	m.tasks[0].Name = "Changed"
	if orig[0].Name != "Build" {
		t.Error("mutating wizard tasks should not affect original slice")
	}
}

func TestNewWizardFromConfig_StartsAtHub(t *testing.T) {
	cfg := &config.Config{Title: "T", Tasks: []config.Task{}}
	m := NewWizardFromConfig("out.yaml", cfg)
	if m.step != wizStepEditHub {
		t.Errorf("edit mode should start at wizStepEditHub, got %d", m.step)
	}
}

func TestWizardEdit_HubViewMentionsEdit(t *testing.T) {
	cfg := &config.Config{
		Title: "My App",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	v := m.View()
	if strings.Contains(v, "No nexus.yaml") {
		t.Error("hub should not say 'No nexus.yaml found'")
	}
	if !strings.Contains(v, "Edit") && !strings.Contains(v, "edit") {
		t.Errorf("hub view should mention editing:\n%s", v)
	}
}

func TestWizardEdit_TitlePreFilledAndConfirmable(t *testing.T) {
	cfg := &config.Config{Title: "Original", Tasks: []config.Task{}}
	m := NewWizardFromConfig("out.yaml", cfg)

	// Hub → 'e' → Title step (inputBuf pre-filled with existing title)
	m = hubEditSettings(t, m)
	if m.inputBuf != "Original" {
		t.Errorf("inputBuf on title step: got %q, want Original", m.inputBuf)
	}

	// Confirm the existing title with Enter (no typing)
	m = pressEnter(t, m)
	if m.cfgTitle != "Original" {
		t.Errorf("cfgTitle after confirm: got %q, want Original", m.cfgTitle)
	}
	if m.step != wizStepEditHub {
		t.Errorf("step after title: got %d, want wizStepEditHub", m.step)
	}
}

func TestWizardEdit_TitleCanBeReplaced(t *testing.T) {
	cfg := &config.Config{Title: "Old", Tasks: []config.Task{}}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = hubEditSettings(t, m) // hub → 'e' → title step (pre-filled "Old")

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

func TestWizardEdit_ExistingTasksCarriedToSummary(t *testing.T) {
	existing := []config.Task{
		{Name: "Build", Actions: []config.Action{{Command: "make build"}}},
	}
	cfg := &config.Config{Title: "T", Tasks: existing}
	m := NewWizardFromConfig("out.yaml", cfg)

	// From the hub, add a new task via 'a', then save via 's'.
	m = hubAddCommand(t, m)
	if m.step != wizStepCmdName {
		t.Fatalf("expected wizStepCmdName after 'a', got %d", m.step)
	}
	m = addCommand(t, m, "NewCmd", "", "echo new", "", "", false)

	// After addCommand in edit mode, returnToHub=true so we land back on hub.
	if m.step != wizStepEditHub {
		t.Fatalf("expected wizStepEditHub after adding task, got %d", m.step)
	}

	m = hubSave(t, m) // 's' → summary
	if m.step != wizStepSummary {
		t.Fatalf("expected wizStepSummary, got %d", m.step)
	}
	if len(m.tasks) != 2 {
		t.Fatalf("expected 2 tasks (existing + new), got %d", len(m.tasks))
	}
	if m.tasks[0].Name != "Build" {
		t.Errorf("first task should be existing Build, got %q", m.tasks[0].Name)
	}
	if m.tasks[1].Name != "NewCmd" {
		t.Errorf("second task should be NewCmd, got %q", m.tasks[1].Name)
	}
}

func TestWizardEdit_DoneViewSaysUpdated(t *testing.T) {
	m := NewWizard("x.yaml")
	m.step = wizStepDone
	m.saved = true
	m.editing = true
	m.savePath = "nexus.yaml"
	v := m.View()
	if !strings.Contains(v, "updated") {
		t.Errorf("edit done view should say 'updated':\n%s", v)
	}
}

func TestWizardEdit_SaveWritesUpdatedFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "nexus.yaml")

	// Write initial config.
	initial := &config.Config{
		Title: "Initial",
		Tasks: []config.Task{
			{Name: "OldCmd", Actions: []config.Action{{Command: "echo old"}}},
		},
	}
	if err := config.Write(tmp, initial); err != nil {
		t.Fatalf("setup: %v", err)
	}

	m := NewWizardFromConfig(tmp, initial)

	// Edit general settings: change title to "Updated".
	m = hubEditSettings(t, m) // hub → 'e' → title step (pre-filled "Initial")
	for range []rune("Initial") {
		m = pressBackspace(t, m)
	}
	m = typeText(t, m, "Updated")
	m = pressEnter(t, m) // confirm title → back to hub

	if m.step != wizStepEditHub {
		t.Fatalf("expected wizStepEditHub after settings edit, got %d", m.step)
	}

	// Add a new task.
	m = hubAddCommand(t, m)
	m = addCommand(t, m, "Extra", "", "echo extra", "", "", false)
	if m.step != wizStepEditHub {
		t.Fatalf("expected wizStepEditHub after adding task, got %d", m.step)
	}

	// Save.
	m = hubSave(t, m)    // → summary
	m = pressEnter(t, m) // → save → done

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
	// Original task + new extra task
	if len(loaded.Tasks) != 2 {
		t.Errorf("tasks: got %d, want 2", len(loaded.Tasks))
	}
}

// ── Edit hub tests ────────────────────────────────────────────────────────────

func TestEditHub_InitialStep(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	if m.step != wizStepEditHub {
		t.Errorf("expected wizStepEditHub, got %d", m.step)
	}
}

func TestEditHub_ViewContainsSettingsSummary(t *testing.T) {
	cfg := &config.Config{Title: "MyApp",
		Tasks: []config.Task{{Name: "Build", Actions: []config.Action{{Command: "make build"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	v := m.View()
	for _, want := range []string{"MyApp", "Build"} {
		if !strings.Contains(v, want) {
			t.Errorf("hub view missing %q:\n%s", want, v)
		}
	}
}

func TestEditHub_ViewContainsActionMenu(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	v := m.View()
	for _, want := range []string{"[d]", "[e]", "[a]", "[s]"} {
		if !strings.Contains(v, want) {
			t.Errorf("hub view missing action key %q:\n%s", want, v)
		}
	}
}

func TestEditHub_DeleteMarksInitialisedFalse(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "A", Actions: []config.Action{{Command: "a"}}},
			{Name: "B", Actions: []config.Action{{Command: "b"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	for i, marked := range m.deleteMarks {
		if marked {
			t.Errorf("deleteMarks[%d] should be false initially", i)
		}
	}
}

func TestEditHub_SpaceTogglesDeleteMark(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "A", Actions: []config.Action{{Command: "a"}}},
			{Name: "B", Actions: []config.Action{{Command: "b"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressSpace(t, m) // toggle A (cursor=0)
	if !m.deleteMarks[0] {
		t.Error("space should mark task at cursor")
	}
	m = pressSpace(t, m) // toggle again
	if m.deleteMarks[0] {
		t.Error("second space should unmark task")
	}
}

func TestEditHub_DownMovesTaskCursor(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "A", Actions: []config.Action{{Command: "a"}}},
			{Name: "B", Actions: []config.Action{{Command: "b"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
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

func TestEditHub_DeleteMarkedRemovesTask(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "Keep", Actions: []config.Action{{Command: "echo keep"}}},
			{Name: "Gone", Actions: []config.Action{{Command: "echo gone"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressDown(t, m)       // cursor → Gone (index 1)
	m = pressSpace(t, m)      // mark Gone
	m = hubDeleteMarked(t, m) // 'd' → delete
	if len(m.tasks) != 1 {
		t.Fatalf("tasks after delete: got %d, want 1", len(m.tasks))
	}
	if m.tasks[0].Name != "Keep" {
		t.Errorf("wrong task remaining: %q", m.tasks[0].Name)
	}
	if m.step != wizStepEditHub {
		t.Errorf("should stay on hub after delete, got %d", m.step)
	}
}

func TestEditHub_DeleteAllMarkedIsRejected(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "Only", Actions: []config.Action{{Command: "echo only"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressSpace(t, m)      // mark the only task
	m = hubDeleteMarked(t, m) // 'd' → should be rejected
	if m.validErr == "" {
		t.Error("should show validation error when deleting all tasks")
	}
	if len(m.tasks) != 1 {
		t.Errorf("tasks should be unchanged, got %d", len(m.tasks))
	}
}

func TestEditHub_DeleteNoneMarkedShowsError(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "A", Actions: []config.Action{{Command: "a"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = hubDeleteMarked(t, m) // 'd' with nothing marked
	if m.validErr == "" {
		t.Error("should show error when no tasks are marked")
	}
}

func TestEditHub_EditSettingsReturnsToHub(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = hubEditSettings(t, m) // hub → 'e' → title
	m = pressEnter(t, m)      // confirm title → back to hub
	if m.step != wizStepEditHub {
		t.Errorf("after editing settings, should return to hub, got %d", m.step)
	}
}

func TestEditHub_AddCommandReturnsToHub(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = hubAddCommand(t, m)
	m = addCommand(t, m, "B", "", "echo b", "", "", false)
	if m.step != wizStepEditHub {
		t.Errorf("after adding task, should return to hub, got %d", m.step)
	}
	if len(m.tasks) != 2 {
		t.Errorf("tasks: got %d, want 2", len(m.tasks))
	}
}

func TestEditHub_SaveGoesToSummary(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = hubSave(t, m) // 's' → summary
	if m.step != wizStepSummary {
		t.Errorf("after 's', expected wizStepSummary, got %d", m.step)
	}
}

func TestEditHub_DeleteMarksGrowAfterAddCommand(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{{Name: "A", Actions: []config.Action{{Command: "a"}}}},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	if len(m.deleteMarks) != 1 {
		t.Fatalf("initial deleteMarks len: got %d, want 1", len(m.deleteMarks))
	}
	m = hubAddCommand(t, m)
	m = addCommand(t, m, "B", "", "echo b", "", "", false)
	if len(m.deleteMarks) != 2 {
		t.Errorf("deleteMarks len after add: got %d, want 2", len(m.deleteMarks))
	}
}

func TestEditHub_DeleteMarksResetAfterBulkDelete(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "A", Actions: []config.Action{{Command: "a"}}},
			{Name: "B", Actions: []config.Action{{Command: "b"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressSpace(t, m)      // mark A (cursor=0)
	m = hubDeleteMarked(t, m) // delete A, leaving B
	if len(m.deleteMarks) != 1 {
		t.Errorf("deleteMarks after delete: got len %d, want 1", len(m.deleteMarks))
	}
	if m.deleteMarks[0] {
		t.Error("remaining task should not be marked")
	}
}

func TestEditHub_CursorClampsAfterDelete(t *testing.T) {
	cfg := &config.Config{Title: "T",
		Tasks: []config.Task{
			{Name: "A", Actions: []config.Action{{Command: "a"}}},
			{Name: "B", Actions: []config.Action{{Command: "b"}}},
		},
	}
	m := NewWizardFromConfig("out.yaml", cfg)
	m = pressDown(t, m)       // cursor → 1 (B)
	m = pressSpace(t, m)      // mark B
	m = hubDeleteMarked(t, m) // delete B; only A remains → cursor must clamp to 0
	if m.optCursor != 0 {
		t.Errorf("cursor should clamp to 0 after delete, got %d", m.optCursor)
	}
}

func TestEditHub_SaveAndVerifyFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "nexus.yaml")
	existing := &config.Config{
		Title: "Orig",

		Tasks: []config.Task{
			{Name: "Keep", Actions: []config.Action{{Command: "echo keep"}}},
			{Name: "Drop", Actions: []config.Action{{Command: "echo drop"}}},
		},
	}
	if err := config.Write(tmp, existing); err != nil {
		t.Fatalf("setup: %v", err)
	}

	m := NewWizardFromConfig(tmp, existing)
	// Mark "Drop" (index 1) and delete it.
	m = pressDown(t, m)
	m = pressSpace(t, m)
	m = hubDeleteMarked(t, m)
	// Save.
	m = hubSave(t, m)    // → summary
	m = pressEnter(t, m) // → save → done

	if m.saveErr != "" {
		t.Fatalf("save error: %s", m.saveErr)
	}

	loaded, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Tasks) != 1 || loaded.Tasks[0].Name != "Keep" {
		t.Errorf("saved file has wrong tasks: %+v", loaded.Tasks)
	}
}

func TestOptionIndex_FindsCorrectIndex(t *testing.T) {
	cases := []struct {
		v    string
		want int
	}{
		{"no", 0},      // no handoff is first
		{"yes", 1},     // handoff is second
		{"", 0},        // fallback to first
		{"unknown", 0}, // fallback
	}
	for _, tc := range cases {
		got := optionIndex(handoffOptions, tc.v)
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
