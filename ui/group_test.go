package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"runner/config"
)

// helpers ─────────────────────────────────────────────────────────────────────

func groupCfg(cmds ...config.Command) *config.Config {
	return &config.Config{
		Title:    "Group Test",
		UIMode:   config.UIModeGroup,
		RunMode:  config.RunModeStream,
		Commands: cmds,
	}
}

func gcmd(name, command, group string) config.Command {
	return config.Command{Name: name, Command: command, Group: group, RunMode: config.RunModeStream}
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewGroupModel_BuildsEntries(t *testing.T) {
	m := NewGroupModel(groupCfg(
		gcmd("Build", "make build", "CI"),
		gcmd("Test", "go test", "CI"),
		gcmd("Deploy", "kubectl apply", "Ops"),
	))

	// Expected entries: header(CI), Build, Test, header(Ops), Deploy = 5
	if len(m.entries) != 5 {
		t.Errorf("entries: got %d, want 5", len(m.entries))
	}
	if !m.entries[0].isHeader || m.entries[0].group != "CI" {
		t.Errorf("entries[0] should be CI header")
	}
	if m.entries[1].isHeader || m.entries[1].cmd.Name != "Build" {
		t.Errorf("entries[1] should be Build cmd")
	}
	if m.entries[2].isHeader || m.entries[2].cmd.Name != "Test" {
		t.Errorf("entries[2] should be Test cmd")
	}
	if !m.entries[3].isHeader || m.entries[3].group != "Ops" {
		t.Errorf("entries[3] should be Ops header")
	}
	if m.entries[4].isHeader || m.entries[4].cmd.Name != "Deploy" {
		t.Errorf("entries[4] should be Deploy cmd")
	}
}

func TestNewGroupModel_EmptyGroupDefaultsToGeneral(t *testing.T) {
	m := NewGroupModel(groupCfg(cmd("A", "a")))
	// entries: header(General), A
	if len(m.entries) != 2 {
		t.Fatalf("entries: got %d, want 2", len(m.entries))
	}
	if !m.entries[0].isHeader || m.entries[0].group != "General" {
		t.Errorf("expected General header, got %+v", m.entries[0])
	}
}

func TestNewGroupModel_InitialCursorSkipsHeader(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("Item", "cmd", "Group")))
	// entries[0] = header, entries[1] = Item → cursor should be 1
	if m.cursor != 1 {
		t.Errorf("initial cursor: got %d, want 1", m.cursor)
	}
}

func TestNewGroupModel_PreservesGroupOrder(t *testing.T) {
	m := NewGroupModel(groupCfg(
		gcmd("Z-first", "z", "Zzz"),
		gcmd("A-second", "a", "Aaa"),
	))
	// Groups appear in insertion order: Zzz first, Aaa second
	headers := []string{}
	for _, e := range m.entries {
		if e.isHeader {
			headers = append(headers, e.group)
		}
	}
	if len(headers) != 2 || headers[0] != "Zzz" || headers[1] != "Aaa" {
		t.Errorf("group order: got %v, want [Zzz Aaa]", headers)
	}
}

func TestNewGroupModel_DefaultDimensions(t *testing.T) {
	m := NewGroupModel(groupCfg())
	if m.width != 80 || m.height != 24 {
		t.Errorf("dimensions: got %dx%d, want 80x24", m.width, m.height)
	}
}

func TestGroupModel_Init_ReturnsNil(t *testing.T) {
	m := NewGroupModel(groupCfg())
	if m.Init() != nil {
		t.Error("Init() should return nil")
	}
}

// ── Navigation ────────────────────────────────────────────────────────────────

func TestGroupModel_MoveDownSkipsHeaders(t *testing.T) {
	// entries: header(A), cmd1, header(B), cmd2
	m := NewGroupModel(groupCfg(gcmd("cmd1", "c1", "A"), gcmd("cmd2", "c2", "B")))
	// initial cursor = 1 (cmd1)
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	// cursor should skip header at index 2 and land on cmd2 at index 3
	if m.cursor != 3 {
		t.Errorf("after down: cursor = %d, want 3 (skipped header)", m.cursor)
	}
}

func TestGroupModel_MoveUpSkipsHeaders(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("cmd1", "c1", "A"), gcmd("cmd2", "c2", "B")))
	// Move to cmd2
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	// Move back up
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 1 {
		t.Errorf("after down+up: cursor = %d, want 1 (skipped header)", m.cursor)
	}
}

func TestGroupModel_MoveDown_JKey(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("A", "a", "G"), gcmd("B", "b", "G")))
	start := m.cursor
	m, _ = m.Update(keyMsg("j"))
	if m.cursor <= start {
		t.Errorf("j key did not advance cursor: %d -> %d", start, m.cursor)
	}
}

func TestGroupModel_MoveUp_KKey(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("A", "a", "G"), gcmd("B", "b", "G")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	after := m.cursor
	m, _ = m.Update(keyMsg("k"))
	if m.cursor >= after {
		t.Errorf("k key did not move cursor back: %d -> %d", after, m.cursor)
	}
}

func TestGroupModel_CursorDoesNotMoveUpPastFirstItem(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("Only", "cmd", "G")))
	initial := m.cursor
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != initial {
		t.Errorf("cursor should not move up past first item: %d -> %d", initial, m.cursor)
	}
}

func TestGroupModel_CursorDoesNotMoveDownPastLastItem(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("A", "a", "G"), gcmd("B", "b", "G")))
	for i := 0; i < 10; i++ {
		m, _ = m.Update(arrowMsg(tea.KeyDown))
	}
	// Last command entry index
	last := len(m.entries) - 1
	if m.cursor != last {
		t.Errorf("cursor overshot last item: got %d, want %d", m.cursor, last)
	}
}

func TestGroupModel_NavigateMultipleGroups(t *testing.T) {
	m := NewGroupModel(groupCfg(
		gcmd("A", "a", "G1"),
		gcmd("B", "b", "G1"),
		gcmd("C", "c", "G2"),
	))
	// Move down twice should get to C (skipping G2 header)
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.entries[m.cursor].cmd == nil || m.entries[m.cursor].cmd.Name != "C" {
		t.Errorf("expected cursor at C, got cursor=%d entry=%+v", m.cursor, m.entries[m.cursor])
	}
}

// ── Selection ─────────────────────────────────────────────────────────────────

func TestGroupModel_SelectWithEnter(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("Deploy", "kubectl", "Ops")))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selection, got nil")
	}
	if sel.Name != "Deploy" {
		t.Errorf("selected: got %q, want Deploy", sel.Name)
	}
}

func TestGroupModel_SelectWithSpace(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("X", "x", "G")))
	m, _ = m.Update(keyMsg(" "))
	if m.Selected() == nil {
		t.Error("space should select item")
	}
}

func TestGroupModel_SelectSecondCommand(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("First", "f", "G"), gcmd("Second", "s", "G")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selection")
	}
	if sel.Name != "Second" {
		t.Errorf("selected: got %q, want Second", sel.Name)
	}
}

func TestGroupModel_NoSelectionOnEmptyList(t *testing.T) {
	m := NewGroupModel(groupCfg())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Selected() != nil {
		t.Error("entering on empty list should not select")
	}
}

// ── WindowSize ────────────────────────────────────────────────────────────────

func TestGroupModel_WindowSizeUpdate(t *testing.T) {
	m := NewGroupModel(groupCfg())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	if m.width != 160 || m.height != 50 {
		t.Errorf("after resize: got %dx%d, want 160x50", m.width, m.height)
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestGroupModel_ViewContainsTitle(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("A", "a", "G")))
	if !strings.Contains(m.View(), "Group Test") {
		t.Error("view missing title")
	}
}

func TestGroupModel_ViewContainsGroupHeaders(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("Build", "make", "CI"), gcmd("Deploy", "kube", "Ops")))
	v := m.View()
	if !strings.Contains(v, "CI") {
		t.Errorf("view missing CI header:\n%s", v)
	}
	if !strings.Contains(v, "Ops") {
		t.Errorf("view missing Ops header:\n%s", v)
	}
}

func TestGroupModel_ViewContainsCommandNames(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("Build", "make", "CI"), gcmd("Deploy", "kube", "Ops")))
	v := m.View()
	if !strings.Contains(v, "Build") || !strings.Contains(v, "Deploy") {
		t.Errorf("view missing command names:\n%s", v)
	}
}

func TestGroupModel_ViewContainsCommandStrings(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("Build", "make build", "CI")))
	v := m.View()
	if !strings.Contains(v, "make build") {
		t.Errorf("view missing command string:\n%s", v)
	}
}

func TestGroupModel_ViewContainsHelpText(t *testing.T) {
	m := NewGroupModel(groupCfg(gcmd("A", "a", "G")))
	v := m.View()
	if !strings.Contains(v, "navigate") || !strings.Contains(v, "quit") {
		t.Errorf("view missing help text:\n%s", v)
	}
}

func TestGroupModel_ViewDescriptionRendered(t *testing.T) {
	c := config.Command{Name: "X", Description: "my desc", Command: "x", Group: "G", RunMode: config.RunModeStream}
	m := NewGroupModel(groupCfg(c))
	v := m.View()
	if !strings.Contains(v, "my desc") {
		t.Errorf("view missing description:\n%s", v)
	}
}
