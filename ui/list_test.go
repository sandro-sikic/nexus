package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

// helpers ─────────────────────────────────────────────────────────────────────

func listCfg(cmds ...config.Command) *config.Config {
	return &config.Config{
		Title:    "Test",
		UIMode:   config.UIModeList,
		RunMode:  config.RunModeStream,
		Commands: cmds,
	}
}

func cmd(name, command string) config.Command {
	return config.Command{Name: name, Command: command, RunMode: config.RunModeStream}
}

func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func arrowMsg(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewListModel_InitialState(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "echo a"), cmd("B", "echo b")))

	if m.cursor != 0 {
		t.Errorf("initial cursor: got %d, want 0", m.cursor)
	}
	if m.Selected() != nil {
		t.Error("initial selected should be nil")
	}
	if m.title != "Test" {
		t.Errorf("title: got %q, want Test", m.title)
	}
	if len(m.commands) != 2 {
		t.Errorf("commands len: got %d, want 2", len(m.commands))
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
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b"), cmd("C", "c")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.cursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", m.cursor)
	}
}

func TestListModel_MoveDown_JKey(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(keyMsg("j"))
	if m.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", m.cursor)
	}
}

func TestListModel_MoveUp(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 0 {
		t.Errorf("after down+up: cursor = %d, want 0", m.cursor)
	}
}

func TestListModel_MoveUp_KKey(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(keyMsg("k"))
	if m.cursor != 0 {
		t.Errorf("after down+k: cursor = %d, want 0", m.cursor)
	}
}

func TestListModel_CursorDoesNotGoAboveZero(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m.cursor)
	}
}

func TestListModel_CursorDoesNotGoBeyondLast(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.cursor != 1 {
		t.Errorf("cursor should stop at last index 1, got %d", m.cursor)
	}
}

func TestListModel_NavigateFullList(t *testing.T) {
	cmds := []config.Command{cmd("A", "a"), cmd("B", "b"), cmd("C", "c"), cmd("D", "d")}
	m := NewListModel(listCfg(cmds...))

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
	m := NewListModel(listCfg(cmd("Build", "make build"), cmd("Test", "make test")))
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
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
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
	m := NewListModel(listCfg(cmd("X", "x")))
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
	m := NewListModel(listCfg(cmd("Alpha", "echo alpha")))
	v := m.View()
	if !strings.Contains(v, "Test") {
		t.Errorf("view does not contain title, got:\n%s", v)
	}
}

func TestListModel_ViewContainsCommandNames(t *testing.T) {
	m := NewListModel(listCfg(cmd("BuildIt", "make"), cmd("TestIt", "go test")))
	v := m.View()
	if !strings.Contains(v, "BuildIt") {
		t.Errorf("view missing BuildIt:\n%s", v)
	}
	if !strings.Contains(v, "TestIt") {
		t.Errorf("view missing TestIt:\n%s", v)
	}
}

func TestListModel_ViewContainsCommands(t *testing.T) {
	m := NewListModel(listCfg(cmd("Run", "npm start")))
	v := m.View()
	if !strings.Contains(v, "npm start") {
		t.Errorf("view missing command string:\n%s", v)
	}
}

func TestListModel_ViewContainsDescription(t *testing.T) {
	c := config.Command{Name: "X", Description: "does things", Command: "x", RunMode: config.RunModeStream}
	m := NewListModel(listCfg(c))
	v := m.View()
	if !strings.Contains(v, "does things") {
		t.Errorf("view missing description:\n%s", v)
	}
}

func TestListModel_ViewContainsHelpText(t *testing.T) {
	m := NewListModel(listCfg(cmd("A", "a")))
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
	m := NewListModel(listCfg(cmd("A", "a"), cmd("B", "b")))
	before := m.cursor
	m, _ = m.Update(keyMsg("x"))
	if m.cursor != before {
		t.Errorf("unknown key changed cursor: %d -> %d", before, m.cursor)
	}
}
