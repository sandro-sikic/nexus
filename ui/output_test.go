package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
	"nexus/runner"
)

// ── OutputModel construction ──────────────────────────────────────────────────

func TestNewOutputModel_InitialState(t *testing.T) {
	c := config.Command{Name: "Echo", Command: "echo hi", RunMode: config.RunModeStream}
	m := NewOutputModel(c)

	if m.cmd.Name != "Echo" {
		t.Errorf("cmd.Name: got %q, want Echo", m.cmd.Name)
	}
	if m.done {
		t.Error("initial done should be false")
	}
	if m.err != nil {
		t.Errorf("initial err should be nil, got %v", m.err)
	}
	if len(m.lines) != 0 {
		t.Errorf("initial lines should be empty, got %d", len(m.lines))
	}
	if m.offset != 0 {
		t.Errorf("initial offset should be 0, got %d", m.offset)
	}
	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions: got %dx%d, want 80x24", m.width, m.height)
	}
}

// ── OutputModel.Update: WindowSize ────────────────────────────────────────────

func TestOutputModel_WindowSizeUpdate(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 || m.height != 40 {
		t.Errorf("after resize: got %dx%d, want 120x40", m.width, m.height)
	}
}

// ── OutputModel.Update: outputDoneMsg ─────────────────────────────────────────

func TestOutputModel_DoneMsgSetsDone(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m, _ = m.Update(outputDoneMsg{})
	if !m.done {
		t.Error("outputDoneMsg should set done=true")
	}
	if m.err != nil {
		t.Errorf("done without error: err should be nil, got %v", m.err)
	}
}

func TestOutputModel_DoneMsgWithError(t *testing.T) {
	m := NewOutputModel(config.Command{})
	fakeErr := &testError{"something went wrong"}
	m, _ = m.Update(outputDoneMsg{err: fakeErr})
	if !m.done {
		t.Error("should be done")
	}
	if m.err == nil {
		t.Error("err should be set")
	}
	if m.err.Error() != "something went wrong" {
		t.Errorf("err: got %q, want %q", m.err.Error(), "something went wrong")
	}
}

// ── OutputModel.Update: line message ─────────────────────────────────────────

func TestOutputModel_LineMessageAppendsLine(t *testing.T) {
	m := NewOutputModel(config.Command{})
	ch := make(chan runner.LogLine, 1)
	lineMsg := struct {
		line  runner.LogLine
		lines chan runner.LogLine
	}{runner.LogLine{Text: "hello", IsErr: false}, ch}

	m, _ = m.Update(lineMsg)
	if len(m.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(m.lines))
	}
	if m.lines[0].Text != "hello" {
		t.Errorf("line text: got %q, want hello", m.lines[0].Text)
	}
}

func TestOutputModel_LineMessageAutoScrolls(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.height = 10 // visible = 10-6 = 4

	ch := make(chan runner.LogLine, 10)
	// Add 10 lines — offset should advance to keep last visible
	for i := 0; i < 10; i++ {
		lineMsg := struct {
			line  runner.LogLine
			lines chan runner.LogLine
		}{runner.LogLine{Text: "line"}, ch}
		m, _ = m.Update(lineMsg)
	}

	visible := m.height - 6 // 4
	expected := len(m.lines) - visible
	if m.offset != expected {
		t.Errorf("auto-scroll offset: got %d, want %d", m.offset, expected)
	}
}

// ── OutputModel.Update: scroll keys ──────────────────────────────────────────

func TestOutputModel_ScrollUp(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.offset = 5
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.offset != 4 {
		t.Errorf("scroll up: offset = %d, want 4", m.offset)
	}
}

func TestOutputModel_ScrollUp_KKey(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.offset = 3
	m, _ = m.Update(keyMsg("k"))
	if m.offset != 2 {
		t.Errorf("k scroll: offset = %d, want 2", m.offset)
	}
}

func TestOutputModel_ScrollUpDoesNotGoBelowZero(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.offset = 0
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.offset != 0 {
		t.Errorf("scroll up at 0: offset = %d, want 0", m.offset)
	}
}

func TestOutputModel_ScrollDown(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.height = 10
	// Add enough lines so scrolling is possible
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, runner.LogLine{Text: "x"})
	}
	m.offset = 0
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.offset != 1 {
		t.Errorf("scroll down: offset = %d, want 1", m.offset)
	}
}

func TestOutputModel_ScrollDown_JKey(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.height = 10
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, runner.LogLine{Text: "x"})
	}
	m.offset = 0
	m, _ = m.Update(keyMsg("j"))
	if m.offset != 1 {
		t.Errorf("j scroll: offset = %d, want 1", m.offset)
	}
}

func TestOutputModel_ScrollDownBounded(t *testing.T) {
	m := NewOutputModel(config.Command{})
	m.height = 10
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, runner.LogLine{Text: "x"})
	}
	// Scroll way past the end
	for i := 0; i < 20; i++ {
		m, _ = m.Update(arrowMsg(tea.KeyDown))
	}
	visible := m.height - 6
	maxOffset := len(m.lines) - visible
	if m.offset > maxOffset {
		t.Errorf("scroll exceeded max: offset=%d, maxOffset=%d", m.offset, maxOffset)
	}
}

// ── OutputModel.Update: streamStartMsg ───────────────────────────────────────

func TestOutputModel_StreamStartMsgReturnsCmd(t *testing.T) {
	m := NewOutputModel(config.Command{})
	ch := make(chan runner.LogLine, 1)
	_, cmd := m.Update(streamStartMsg{lines: ch})
	if cmd == nil {
		t.Error("streamStartMsg should return a non-nil tea.Cmd")
	}
}

// ── OutputModel.View ──────────────────────────────────────────────────────────

func TestOutputModel_ViewContainsCmdName(t *testing.T) {
	c := config.Command{Name: "MyCmd", Command: "echo hi"}
	m := NewOutputModel(c)
	v := m.View()
	if !strings.Contains(v, "MyCmd") {
		t.Errorf("view missing cmd name:\n%s", v)
	}
}

func TestOutputModel_ViewContainsCmdString(t *testing.T) {
	c := config.Command{Name: "X", Command: "npm run dev"}
	m := NewOutputModel(c)
	v := m.View()
	if !strings.Contains(v, "npm run dev") {
		t.Errorf("view missing command string:\n%s", v)
	}
}

func TestOutputModel_ViewShowsRunningHelpWhenNotDone(t *testing.T) {
	m := NewOutputModel(config.Command{Name: "X", Command: "y"})
	v := m.View()
	if !strings.Contains(v, "running") {
		t.Errorf("view should show 'running' when not done:\n%s", v)
	}
}

func TestOutputModel_ViewShowsDoneHelpWhenDone(t *testing.T) {
	m := NewOutputModel(config.Command{Name: "X", Command: "y"})
	m, _ = m.Update(outputDoneMsg{})
	v := m.View()
	if !strings.Contains(v, "back") && !strings.Contains(v, "esc") && !strings.Contains(v, "q") {
		t.Errorf("view should show 'go back' help when done:\n%s", v)
	}
}

func TestOutputModel_ViewShowsErrorWhenFailed(t *testing.T) {
	m := NewOutputModel(config.Command{Name: "X", Command: "y"})
	m, _ = m.Update(outputDoneMsg{err: &testError{"boom"}})
	v := m.View()
	if !strings.Contains(v, "boom") {
		t.Errorf("view should show error message:\n%s", v)
	}
}

func TestOutputModel_ViewRendersLines(t *testing.T) {
	m := NewOutputModel(config.Command{Name: "X", Command: "y"})
	m.lines = []runner.LogLine{
		{Text: "output-line-one", IsErr: false},
		{Text: "output-line-two", IsErr: false},
	}
	v := m.View()
	if !strings.Contains(v, "output-line-one") || !strings.Contains(v, "output-line-two") {
		t.Errorf("view missing output lines:\n%s", v)
	}
}

func TestOutputModel_ViewRendersErrorLines(t *testing.T) {
	m := NewOutputModel(config.Command{Name: "X", Command: "y"})
	m.lines = []runner.LogLine{{Text: "err-output", IsErr: true}}
	v := m.View()
	if !strings.Contains(v, "err-output") {
		t.Errorf("view missing stderr line:\n%s", v)
	}
}

// ── BackgroundModel construction ──────────────────────────────────────────────

func TestNewBackgroundModel_InitialState(t *testing.T) {
	c := config.Command{Name: "BG", Command: "echo bg", RunMode: config.RunModeBackground}
	m := BackgroundModel{
		cmd:    c,
		width:  80,
		height: 24,
	}
	if m.done {
		t.Error("initial done should be false")
	}
	if m.offset != 0 {
		t.Error("initial offset should be 0")
	}
}

// ── BackgroundModel.Update ────────────────────────────────────────────────────

func TestBackgroundModel_WindowSizeUpdate(t *testing.T) {
	m := BackgroundModel{width: 80, height: 24}
	m, _ = m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	if m.width != 200 || m.height != 60 {
		t.Errorf("after resize: got %dx%d, want 200x60", m.width, m.height)
	}
}

func TestBackgroundModel_DoneMsgSetsDone(t *testing.T) {
	m := BackgroundModel{}
	m, _ = m.Update(outputDoneMsg{})
	if !m.done {
		t.Error("outputDoneMsg should set done=true on BackgroundModel")
	}
}

func TestBackgroundModel_LineMessageAppendsLine(t *testing.T) {
	m := BackgroundModel{height: 24}
	ch := make(chan runner.LogLine, 1)
	lineMsg := struct {
		line  runner.LogLine
		lines chan runner.LogLine
	}{runner.LogLine{Text: "bg-line"}, ch}
	m, _ = m.Update(lineMsg)
	if len(m.lines) != 1 || m.lines[0].Text != "bg-line" {
		t.Errorf("line not appended: %v", m.lines)
	}
}

func TestBackgroundModel_ScrollUp(t *testing.T) {
	m := BackgroundModel{offset: 5}
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.offset != 4 {
		t.Errorf("scroll up: offset = %d, want 4", m.offset)
	}
}

func TestBackgroundModel_ScrollUpAtZeroIgnored(t *testing.T) {
	m := BackgroundModel{offset: 0}
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.offset != 0 {
		t.Errorf("scroll at 0 should stay 0, got %d", m.offset)
	}
}

func TestBackgroundModel_ScrollDown(t *testing.T) {
	m := BackgroundModel{height: 10}
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, runner.LogLine{Text: "x"})
	}
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.offset != 1 {
		t.Errorf("scroll down: offset = %d, want 1", m.offset)
	}
}

// ── BackgroundModel.View ──────────────────────────────────────────────────────

func TestBackgroundModel_ViewContainsCmdName(t *testing.T) {
	m := BackgroundModel{cmd: config.Command{Name: "MyBG", Command: "x"}, height: 24}
	v := m.View()
	if !strings.Contains(v, "MyBG") {
		t.Errorf("view missing cmd name:\n%s", v)
	}
}

func TestBackgroundModel_ViewShowsRunningStatus(t *testing.T) {
	m := BackgroundModel{cmd: config.Command{Name: "X", Command: "y"}, height: 24}
	v := m.View()
	if !strings.Contains(v, "running") {
		t.Errorf("view should show 'running' when not done:\n%s", v)
	}
}

func TestBackgroundModel_ViewShowsDoneStatus(t *testing.T) {
	m := BackgroundModel{cmd: config.Command{Name: "X", Command: "y"}, height: 24, done: true}
	v := m.View()
	if !strings.Contains(v, "done") {
		t.Errorf("view should show 'done' when finished:\n%s", v)
	}
}

func TestBackgroundModel_ViewRendersLines(t *testing.T) {
	m := BackgroundModel{
		cmd:    config.Command{Name: "X", Command: "y"},
		height: 24,
		lines:  []runner.LogLine{{Text: "bg-output-line"}},
	}
	v := m.View()
	if !strings.Contains(v, "bg-output-line") {
		t.Errorf("view missing output line:\n%s", v)
	}
}

// ── OutputModel.View: multi-step command display ───────────────────────────────

func TestOutputModel_ViewMultiStepShowsAllSteps(t *testing.T) {
	c := config.Command{
		Name:     "MultiStepCmd",
		Commands: []string{"npm install", "npm run build"},
		RunMode:  config.RunModeStream,
	}
	m := NewOutputModel(c)
	v := m.View()
	if !strings.Contains(v, "npm install") {
		t.Errorf("view missing first step 'npm install':\n%s", v)
	}
	if !strings.Contains(v, "npm run build") {
		t.Errorf("view missing second step 'npm run build':\n%s", v)
	}
	// Should render numbered step indicators like [1] and [2]
	if !strings.Contains(v, "[1]") || !strings.Contains(v, "[2]") {
		t.Errorf("view missing step numbers [1]/[2]:\n%s", v)
	}
}

func TestOutputModel_ViewSingleStepDoesNotShowNumbered(t *testing.T) {
	c := config.Command{Name: "Single", Command: "echo hi"}
	m := NewOutputModel(c)
	v := m.View()
	// Single-step path should use "$ cmd" without numbered steps
	if !strings.Contains(v, "$ echo hi") {
		t.Errorf("single-step view should show '$ echo hi':\n%s", v)
	}
	if strings.Contains(v, "[1]") {
		t.Errorf("single-step view should not show numbered step [1]:\n%s", v)
	}
}

// ── BackgroundModel.Update: j/k scroll and auto-scroll ────────────────────────

func TestBackgroundModel_ScrollUp_KKey(t *testing.T) {
	m := BackgroundModel{offset: 3}
	m, _ = m.Update(keyMsg("k"))
	if m.offset != 2 {
		t.Errorf("k key scroll up: offset = %d, want 2", m.offset)
	}
}

func TestBackgroundModel_ScrollDown_JKey(t *testing.T) {
	m := BackgroundModel{height: 10}
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, runner.LogLine{Text: "x"})
	}
	m, _ = m.Update(keyMsg("j"))
	if m.offset != 1 {
		t.Errorf("j key scroll down: offset = %d, want 1", m.offset)
	}
}

func TestBackgroundModel_ScrollDownBounded(t *testing.T) {
	m := BackgroundModel{height: 10}
	for i := 0; i < 10; i++ {
		m.lines = append(m.lines, runner.LogLine{Text: "x"})
	}
	// Scroll way past the end
	for i := 0; i < 20; i++ {
		m, _ = m.Update(arrowMsg(tea.KeyDown))
	}
	visible := m.height - 6
	maxOffset := len(m.lines) - visible
	if m.offset > maxOffset {
		t.Errorf("scroll exceeded max: offset=%d, maxOffset=%d", m.offset, maxOffset)
	}
}

func TestBackgroundModel_AutoScrollOnLineMessage(t *testing.T) {
	m := BackgroundModel{height: 10} // visible = 10-6 = 4
	ch := make(chan runner.LogLine, 20)

	// Add enough lines to trigger auto-scroll
	for i := 0; i < 10; i++ {
		lineMsg := struct {
			line  runner.LogLine
			lines chan runner.LogLine
		}{runner.LogLine{Text: "line"}, ch}
		m, _ = m.Update(lineMsg)
	}

	visible := m.height - 6 // 4
	expected := len(m.lines) - visible
	if m.offset != expected {
		t.Errorf("bg auto-scroll offset: got %d, want %d", m.offset, expected)
	}
}

// ── BackgroundModel.View: multi-step display ──────────────────────────────────

func TestBackgroundModel_ViewMultiStepShowsAllSteps(t *testing.T) {
	m := BackgroundModel{
		cmd: config.Command{
			Name:     "BGMulti",
			Commands: []string{"step-a", "step-b"},
		},
		height: 24,
	}
	v := m.View()
	if !strings.Contains(v, "step-a") {
		t.Errorf("bg view missing first step:\n%s", v)
	}
	if !strings.Contains(v, "step-b") {
		t.Errorf("bg view missing second step:\n%s", v)
	}
	if !strings.Contains(v, "[1]") || !strings.Contains(v, "[2]") {
		t.Errorf("bg multi-step view missing step numbers:\n%s", v)
	}
}

// ── BackgroundModel.Update: DoneMsgWithError ──────────────────────────────────

func TestBackgroundModel_DoneMsgSetsErrNotStored(t *testing.T) {
	// BackgroundModel does not store err, just done=true; this verifies no panic.
	m := BackgroundModel{}
	m, _ = m.Update(outputDoneMsg{err: &testError{"bg-error"}})
	if !m.done {
		t.Error("outputDoneMsg with error should still set done=true")
	}
}

// ── testError helper ─────────────────────────────────────────────────────────

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
