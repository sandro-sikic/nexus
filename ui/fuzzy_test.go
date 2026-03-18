package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"nexus/config"
)

// helpers ─────────────────────────────────────────────────────────────────────

func fuzzyCfg(cmds ...config.Command) *config.Config {
	return &config.Config{
		Title:    "Fuzzy Test",
		UIMode:   config.UIModeFuzzy,
		RunMode:  config.RunModeStream,
		Commands: cmds,
	}
}

func runesMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func backspaceMsg() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyBackspace}
}

// ── Construction ──────────────────────────────────────────────────────────────

func TestNewFuzzyModel_InitialState(t *testing.T) {
	cmds := []config.Command{cmd("Alpha", "a"), cmd("Beta", "b")}
	m := NewFuzzyModel(fuzzyCfg(cmds...))

	if m.query != "" {
		t.Errorf("initial query: got %q, want empty", m.query)
	}
	if m.cursor != 0 {
		t.Errorf("initial cursor: got %d, want 0", m.cursor)
	}
	if len(m.filtered) != 2 {
		t.Errorf("initial filtered: got %d, want 2", len(m.filtered))
	}
	if m.Selected() != nil {
		t.Error("initial selected should be nil")
	}
}

func TestFuzzyModel_Init_ReturnsNil(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg())
	if m.Init() != nil {
		t.Error("Init() should return nil")
	}
}

// ── fuzzyMatch ────────────────────────────────────────────────────────────────

func TestFuzzyMatch_ExactMatch(t *testing.T) {
	if !fuzzyMatch("build", "build") {
		t.Error("exact match should succeed")
	}
}

func TestFuzzyMatch_Subsequence(t *testing.T) {
	if !fuzzyMatch("bld", "build") {
		t.Error("subsequence should match")
	}
}

func TestFuzzyMatch_CaseInsensitive(t *testing.T) {
	if !fuzzyMatch("BUILD", "build") {
		t.Error("should be case insensitive")
	}
	if !fuzzyMatch("build", "BUILD") {
		t.Error("should be case insensitive both ways")
	}
}

func TestFuzzyMatch_EmptyQuery(t *testing.T) {
	if !fuzzyMatch("", "anything") {
		t.Error("empty query should always match")
	}
}

func TestFuzzyMatch_EmptyTarget(t *testing.T) {
	if fuzzyMatch("a", "") {
		t.Error("non-empty query against empty target should not match")
	}
}

func TestFuzzyMatch_QueryLongerThanTarget(t *testing.T) {
	if fuzzyMatch("toolong", "to") {
		t.Error("query longer than target should not match")
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	if fuzzyMatch("xyz", "build") {
		t.Error("xyz should not match build")
	}
}

func TestFuzzyMatch_SingleChar(t *testing.T) {
	if !fuzzyMatch("b", "build") {
		t.Error("single char should match if present")
	}
	if fuzzyMatch("z", "build") {
		t.Error("single char not in target should not match")
	}
}

// ── Typing / filtering ────────────────────────────────────────────────────────

func TestFuzzyModel_TypingFilters(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "make build"), cmd("Test", "go test"), cmd("Deploy", "make deploy")))

	m, _ = m.Update(runesMsg("bld"))
	if m.query != "bld" {
		t.Errorf("query: got %q, want bld", m.query)
	}
	if len(m.filtered) != 1 {
		t.Errorf("filtered after 'bld': got %d, want 1", len(m.filtered))
	}
	if m.filtered[0].Name != "Build" {
		t.Errorf("filtered[0]: got %q, want Build", m.filtered[0].Name)
	}
}

func TestFuzzyModel_TypingOneChar(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "build"), cmd("Test", "test")))
	m, _ = m.Update(runesMsg("t"))
	// "t" appears in both "Test" (name) and potentially others
	found := false
	for _, c := range m.filtered {
		if c.Name == "Test" {
			found = true
		}
	}
	if !found {
		t.Error("'Test' should be in filtered results after typing 't'")
	}
}

func TestFuzzyModel_EmptyQueryShowsAll(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a"), cmd("B", "b"), cmd("C", "c")))
	m, _ = m.Update(runesMsg("x"))
	m, _ = m.Update(backspaceMsg())
	if len(m.filtered) != 3 {
		t.Errorf("after backspace to empty: got %d filtered, want 3", len(m.filtered))
	}
}

func TestFuzzyModel_BackspaceRemovesChar(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "build")))
	m, _ = m.Update(runesMsg("bui"))
	if m.query != "bui" {
		t.Errorf("query after typing: got %q", m.query)
	}
	m, _ = m.Update(backspaceMsg())
	if m.query != "bu" {
		t.Errorf("query after backspace: got %q, want bu", m.query)
	}
}

func TestFuzzyModel_BackspaceOnEmptyIgnored(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a")))
	before := len(m.filtered)
	m, _ = m.Update(backspaceMsg())
	if m.query != "" {
		t.Errorf("query should stay empty, got %q", m.query)
	}
	if len(m.filtered) != before {
		t.Error("filtered should not change on empty backspace")
	}
}

func TestFuzzyModel_NoMatchShowsEmpty(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "build"), cmd("Test", "test")))
	m, _ = m.Update(runesMsg("zzzzz"))
	if len(m.filtered) != 0 {
		t.Errorf("expected 0 results for zzzzz, got %d", len(m.filtered))
	}
}

func TestFuzzyModel_MatchesOnDescription(t *testing.T) {
	c := config.Command{Name: "Nexus", Description: "executes things", Command: "run", RunMode: config.RunModeStream}
	m := NewFuzzyModel(fuzzyCfg(c))
	m, _ = m.Update(runesMsg("exec"))
	if len(m.filtered) != 1 {
		t.Errorf("expected match on description, got %d results", len(m.filtered))
	}
}

func TestFuzzyModel_MatchesOnCommand(t *testing.T) {
	c := config.Command{Name: "Nexus", Description: "desc", Command: "npm run dev", RunMode: config.RunModeStream}
	m := NewFuzzyModel(fuzzyCfg(c))
	m, _ = m.Update(runesMsg("nrd"))
	if len(m.filtered) != 1 {
		t.Errorf("expected match on command string, got %d results", len(m.filtered))
	}
}

// ── Cursor clamping after filter ──────────────────────────────────────────────

func TestFuzzyModel_CursorClampedWhenFilterShrinks(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a"), cmd("B", "b"), cmd("C", "c")))
	// Move to last item
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.cursor != 2 {
		t.Fatalf("pre-condition: cursor should be 2, got %d", m.cursor)
	}
	// Filter to only 1 result
	m, _ = m.Update(runesMsg("a"))
	if m.cursor != 0 {
		t.Errorf("cursor should clamp to 0 after filter, got %d", m.cursor)
	}
}

// ── Navigation ────────────────────────────────────────────────────────────────

func TestFuzzyModel_MoveDown(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	if m.cursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", m.cursor)
	}
}

func TestFuzzyModel_MoveUp(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a"), cmd("B", "b")))
	m, _ = m.Update(arrowMsg(tea.KeyDown))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 0 {
		t.Errorf("after down+up: cursor = %d, want 0", m.cursor)
	}
}

func TestFuzzyModel_CursorDoesNotGoAboveZero(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a")))
	m, _ = m.Update(arrowMsg(tea.KeyUp))
	if m.cursor != 0 {
		t.Errorf("cursor went negative: %d", m.cursor)
	}
}

func TestFuzzyModel_CursorDoesNotGoBeyondFiltered(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a"), cmd("B", "b")))
	for i := 0; i < 5; i++ {
		m, _ = m.Update(arrowMsg(tea.KeyDown))
	}
	if m.cursor > 1 {
		t.Errorf("cursor exceeded list: got %d, max 1", m.cursor)
	}
}

// ── Selection ─────────────────────────────────────────────────────────────────

func TestFuzzyModel_SelectWithEnter(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "make"), cmd("Test", "test")))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selection, got nil")
	}
	if sel.Name != "Build" {
		t.Errorf("selected: got %q, want Build", sel.Name)
	}
}

func TestFuzzyModel_SelectAfterFiltering(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "make"), cmd("Test", "go test"), cmd("Deploy", "deploy")))
	m, _ = m.Update(runesMsg("dep"))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selection after filtering")
	}
	if sel.Name != "Deploy" {
		t.Errorf("selected: got %q, want Deploy", sel.Name)
	}
}

func TestFuzzyModel_NoSelectOnEmptyFiltered(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a")))
	m, _ = m.Update(runesMsg("zzz"))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Selected() != nil {
		t.Error("should not select when filtered is empty")
	}
}

// ── WindowSize ────────────────────────────────────────────────────────────────

func TestFuzzyModel_WindowSizeUpdate(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.width != 100 || m.height != 30 {
		t.Errorf("after resize: got %dx%d, want 100x30", m.width, m.height)
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func TestFuzzyModel_ViewContainsTitle(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a")))
	if !strings.Contains(m.View(), "Fuzzy Test") {
		t.Error("view missing title")
	}
}

func TestFuzzyModel_ViewContainsQuery(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a")))
	m, _ = m.Update(runesMsg("bld"))
	if !strings.Contains(m.View(), "bld") {
		t.Error("view missing query string")
	}
}

func TestFuzzyModel_ViewShowsNoMatches(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("A", "a")))
	m, _ = m.Update(runesMsg("zzz"))
	if !strings.Contains(m.View(), "no matches") {
		t.Error("view should show 'no matches'")
	}
}

func TestFuzzyModel_ViewContainsItemNames(t *testing.T) {
	m := NewFuzzyModel(fuzzyCfg(cmd("Build", "make"), cmd("Test", "go test")))
	v := m.View()
	if !strings.Contains(v, "Build") || !strings.Contains(v, "Test") {
		t.Errorf("view missing command names:\n%s", v)
	}
}

// ── max helper ────────────────────────────────────────────────────────────────

func TestMax(t *testing.T) {
	cases := []struct{ a, b, want int }{
		{1, 2, 2},
		{2, 1, 2},
		{0, 0, 0},
		{-1, 1, 1},
		{5, 5, 5},
	}
	for _, tc := range cases {
		if got := max(tc.a, tc.b); got != tc.want {
			t.Errorf("max(%d,%d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
