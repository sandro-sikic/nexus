package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
)

// helpers ─────────────────────────────────────────────────────────────────────

func listCfg(tasks ...config.Task) *config.Config {
	return &config.Config{
		Title:   "Test",
		RunMode: config.RunModeStream,
		Tasks:   tasks,
	}
}

func task(name string, commands []string) config.Task {
	actions := make([]config.Action, len(commands))
	for i, cmd := range commands {
		actions[i] = config.Action{Command: cmd}
	}
	return config.Task{Name: name, Actions: actions, RunMode: config.RunModeStream}
}

func cmd(name, command string) config.Task {
	return task(name, []string{command})
}

func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func arrowMsg(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewListModel_InitialState(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"echo a"}), task("B", []string{"echo b"})))

	if m.cursor != 0 {
		t.Errorf("initial cursor: got %d, want 0", m.cursor)
	}
	if m.Selected() != nil {
		t.Error("initial selected should be nil")
	}
	if m.title != "Test" {
		t.Errorf("title: got %q, want Test", m.title)
	}
	if len(m.tasks) != 2 {
		t.Errorf("tasks len: got %d, want 2", len(m.tasks))
	}
}

func TestNewListModel_DefaultDimensions(t *testing.T) {
	m := NewListModel(listCfg())
	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions: got %dx%d, want 80x24", m.width, m.height)
	}
}

func TestListModel_Init_ReturnsNil(t *testing.T) {
	m := NewListModel(listCfg())
	if m.Init() != nil {
		t.Error("Init() should return nil")
	}
}

// ── Navigation ────────────────────────────────────────────────────────────────

func TestListModel_MoveDown(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"}), task("C", []string{"c"})))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.cursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", m.cursor)
	}
}

func TestListModel_MoveDown_JKey(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	m, _ = m.Update(keyMsg("j"))
	if m.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", m.cursor)
	}
}

func TestListModel_MoveUp(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 0 {
		t.Errorf("after down+up: cursor = %d, want 0", m.cursor)
	}
}

func TestListModel_MoveUp_KKey(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(keyMsg("k"))
	if m.cursor != 0 {
		t.Errorf("after down+k: cursor = %d, want 0", m.cursor)
	}
}

func TestListModel_CursorDoesNotGoAboveZero(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m.cursor)
	}
}

func TestListModel_CursorDoesNotGoBeyondLast(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.cursor != 1 {
		t.Errorf("cursor should stop at last index 1, got %d", m.cursor)
	}
}

func TestListModel_NavigateFullList(t *testing.T) {
	tasks := []config.Task{task("A", []string{"a"}), task("B", []string{"b"}), task("C", []string{"c"}), task("D", []string{"d"})}
	m := NewListModel(listCfg(tasks...))

	for i := 0; i < 3; i++ {
		m, _ = m.Update(arrowMsg(tea.KeyDown))
	}
	if m.cursor != 3 {
		t.Errorf("expected cursor=3, got %d", m.cursor)
	}

	for i := 0; i < 3; i++ {
		m, _ = m.Update(arrowMsg(tea.KeyUp))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.cursor)
	}
}

// ── Selection ─────────────────────────────────────────────────────────────────

func TestListModel_SelectWithEnter(t *testing.T) {
	m := NewListModel(listCfg(task("Build", []string{"make build"}), task("Test", []string{"make test"})))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selection, got nil")
	}
	if sel.Name != "Build" {
		t.Errorf("selected: got %q, want Build", sel.Name)
	}
}

func TestListModel_SelectSecondItem(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selection")
	}
	if sel.Name != "B" {
		t.Errorf("selected: got %q, want B", sel.Name)
	}
}

func TestListModel_SelectWithSpace(t *testing.T) {
	m := NewListModel(listCfg(task("X", []string{"x"})))
	m, _ = m.Update(keyMsg(" "))
	if m.Selected() == nil {
		t.Error("space should select item")
	}
}

func TestListModel_NoSelectionOnEmpty(t *testing.T) {
	m := NewListModel(listCfg())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Selected() != nil {
		t.Error("entering on empty list should not select")
	}
}

// ── WindowSize ────────────────────────────────────────────────────────────────

func TestListModel_WindowSizeUpdate(t *testing.T) {
	m := NewListModel(listCfg())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 || m.height != 40 {
		t.Errorf("after resize: got %dx%d, want 120x40", m.width, m.height)
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestListModel_ViewContainsTitle(t *testing.T) {
	m := NewListModel(listCfg(task("Alpha", []string{"echo alpha"})))
	v := m.View()
	if !strings.Contains(v, "Test") {
		t.Errorf("view does not contain title, got:\n%s", v)
	}
}

func TestListModel_ViewContainsTaskNames(t *testing.T) {
	m := NewListModel(listCfg(task("BuildIt", []string{"make"}), task("TestIt", []string{"go test"})))
	v := m.View()
	if !strings.Contains(v, "BuildIt") {
		t.Errorf("view missing BuildIt:\n%s", v)
	}
	if !strings.Contains(v, "TestIt") {
		t.Errorf("view missing TestIt:\n%s", v)
	}
}

func TestListModel_ViewContainsCommands(t *testing.T) {
	m := NewListModel(listCfg(task("Run", []string{"npm start"})))
	v := m.View()
	if !strings.Contains(v, "npm start") {
		t.Errorf("view missing command string:\n%s", v)
	}
}

func TestListModel_ViewContainsDescription(t *testing.T) {
	tst := config.Task{Name: "X", Description: "does things", Actions: []config.Action{{Command: "x"}}, RunMode: config.RunModeStream}
	m := NewListModel(listCfg(tst))
	v := m.View()
	if !strings.Contains(v, "does things") {
		t.Errorf("view missing description:\n%s", v)
	}
}

func TestListModel_ViewContainsHelpText(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"})))
	v := m.View()
	if !strings.Contains(v, "navigate") && !strings.Contains(v, "quit") {
		t.Errorf("view missing help text:\n%s", v)
	}
}

func TestListModel_ViewEmptyList(t *testing.T) {
	m := NewListModel(listCfg())
	v := m.View()
	// Should not panic and should contain the title at minimum
	if !strings.Contains(v, "Test") {
		t.Errorf("empty list view missing title:\n%s", v)
	}
}

func TestListModel_UnknownKeyIgnored(t *testing.T) {
	m := NewListModel(listCfg(task("A", []string{"a"}), task("B", []string{"b"})))
	before := m.cursor
	m, _ = m.Update(keyMsg("x"))
	if m.cursor != before {
		t.Errorf("unknown key changed cursor: %d -> %d", before, m.cursor)
	}
}
