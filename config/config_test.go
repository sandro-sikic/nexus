package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"nexus/config"
)

// writeTemp writes YAML content to a temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "nexus-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

// ── Load: happy path ──────────────────────────────────────────────────────────

func TestLoad_FullConfig(t *testing.T) {
	yaml := `
title: "Test Nexus"
run_mode: handoff
tasks:
  - name: Build
    description: Build the project
    actions:
      - command: "make build"
        background: false
    dir: /tmp
    run_mode: stream
    group: CI
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Title != "Test Nexus" {
		t.Errorf("title: got %q, want %q", cfg.Title, "Test Nexus")
	}
	if cfg.RunMode != config.RunModeHandoff {
		t.Errorf("run_mode: got %q, want %q", cfg.RunMode, config.RunModeHandoff)
	}
	if len(cfg.Tasks) != 1 {
		t.Fatalf("tasks: got %d, want 1", len(cfg.Tasks))
	}

	task := cfg.Tasks[0]
	if task.Name != "Build" {
		t.Errorf("task.name: got %q, want %q", task.Name, "Build")
	}
	if task.Description != "Build the project" {
		t.Errorf("task.description: got %q", task.Description)
	}
	if len(task.Actions) != 1 {
		t.Fatalf("task.actions: got %d, want 1", len(task.Actions))
	}
	if task.Actions[0].Command != "make build" {
		t.Errorf("task.actions[0].command: got %q", task.Actions[0].Command)
	}
	if task.Actions[0].Background != false {
		t.Errorf("task.actions[0].background: got %v, want false", task.Actions[0].Background)
	}
	if task.Dir != "/tmp" {
		t.Errorf("task.dir: got %q", task.Dir)
	}
	if task.RunMode != config.RunModeStream {
		t.Errorf("task.run_mode: got %q, want stream", task.RunMode)
	}
	if task.Group != "CI" {
		t.Errorf("task.group: got %q, want CI", task.Group)
	}
}

// ── Defaults ─────────────────────────────────────────────────────────────────

func TestLoad_DefaultTitle(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "tasks: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "Nexus" {
		t.Errorf("default title: got %q, want %q", cfg.Title, "Nexus")
	}
}

func TestLoad_DefaultRunMode(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "tasks: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RunMode != config.RunModeStream {
		t.Errorf("default run_mode: got %q, want stream", cfg.RunMode)
	}
}

func TestLoad_TaskInheritsTopLevelRunMode(t *testing.T) {
	yaml := `
run_mode: background
tasks:
  - name: Test
    actions:
      - command: "echo hi"
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Tasks[0].RunMode != config.RunModeBackground {
		t.Errorf("inherited run_mode: got %q, want background", cfg.Tasks[0].RunMode)
	}
}

func TestLoad_TaskRunModeNotOverriddenWhenExplicit(t *testing.T) {
	yaml := `
run_mode: stream
tasks:
  - name: Deploy
    actions:
      - command: "make deploy"
    run_mode: handoff
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Tasks[0].RunMode != config.RunModeHandoff {
		t.Errorf("explicit run_mode: got %q, want handoff", cfg.Tasks[0].RunMode)
	}
}

func TestLoad_MultipleTasksInheritance(t *testing.T) {
	yaml := `
run_mode: background
tasks:
  - name: A
    actions:
      - command: echo a
  - name: B
    actions:
      - command: echo b
    run_mode: stream
  - name: C
    actions:
      - command: echo c
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cases := []struct {
		idx  int
		want config.RunMode
	}{
		{0, config.RunModeBackground},
		{1, config.RunModeStream},
		{2, config.RunModeBackground},
	}
	for _, tc := range cases {
		if got := cfg.Tasks[tc.idx].RunMode; got != tc.want {
			t.Errorf("tasks[%d].run_mode: got %q, want %q", tc.idx, got, tc.want)
		}
	}
}

func TestLoad_EmptyTaskList(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "tasks: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tasks) != 0 {
		t.Errorf("expected empty tasks, got %d", len(cfg.Tasks))
	}
}

func TestLoad_NoTasksKey(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "title: bare"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Tasks != nil {
		t.Errorf("expected nil tasks, got %v", cfg.Tasks)
	}
}

// ── Error cases ───────────────────────────────────────────────────────────────

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	_, err := config.Load(writeTemp(t, "title: [invalid yaml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_MalformedTasks(t *testing.T) {
	// tasks should be a list, not a string — YAML should error
	_, err := config.Load(writeTemp(t, "tasks: not-a-list"))
	if err == nil {
		t.Fatal("expected error for malformed tasks, got nil")
	}
}

// ── Constants ─────────────────────────────────────────────────────────────────

func TestConstants(t *testing.T) {
	// Ensure string values stay stable — they're used in YAML files.
	if string(config.RunModeStream) != "stream" {
		t.Errorf("RunModeStream = %q, want stream", config.RunModeStream)
	}
	if string(config.RunModeHandoff) != "handoff" {
		t.Errorf("RunModeHandoff = %q, want handoff", config.RunModeHandoff)
	}
	if string(config.RunModeBackground) != "background" {
		t.Errorf("RunModeBackground = %q, want background", config.RunModeBackground)
	}
}

// ── Optional fields ───────────────────────────────────────────────────────────

func TestLoad_TaskOptionalFields(t *testing.T) {
	yaml := `
tasks:
  - name: Minimal
    actions:
      - command: echo hi
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	task := cfg.Tasks[0]
	if task.Dir != "" {
		t.Errorf("dir should default empty, got %q", task.Dir)
	}
	if task.Group != "" {
		t.Errorf("group should default empty, got %q", task.Group)
	}
	if task.Description != "" {
		t.Errorf("description should default empty, got %q", task.Description)
	}
}

// ── Multi-action (Actions) ────────────────────────────────────────────────────

func TestTask_Actions_SingleAction(t *testing.T) {
	task := config.Task{Actions: []config.Action{{Command: "echo hi"}}}
	cmds := task.AllCommands()
	if len(cmds) != 1 || cmds[0] != "echo hi" {
		t.Errorf("AllCommands(): got %v, want [echo hi]", cmds)
	}
}

func TestTask_Actions_MultipleActions(t *testing.T) {
	task := config.Task{Actions: []config.Action{
		{Command: "cd app"},
		{Command: "npm run dev"},
	}}
	cmds := task.AllCommands()
	if len(cmds) != 2 {
		t.Fatalf("AllCommands() len: got %d, want 2", len(cmds))
	}
	if cmds[0] != "cd app" || cmds[1] != "npm run dev" {
		t.Errorf("AllCommands(): got %v", cmds)
	}
}

func TestTask_Actions_Empty(t *testing.T) {
	task := config.Task{}
	cmds := task.AllCommands()
	if len(cmds) != 0 {
		t.Errorf("AllCommands() on empty task: got %v, want []", cmds)
	}
}

func TestLoad_MultipleActions(t *testing.T) {
	yaml := `
tasks:
  - name: Setup
    actions:
      - command: "cd app"
      - command: "npm install"
      - command: "npm run dev"
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tasks) != 1 {
		t.Fatalf("tasks: got %d, want 1", len(cfg.Tasks))
	}
	cmds := cfg.Tasks[0].AllCommands()
	if len(cmds) != 3 {
		t.Fatalf("AllCommands() len: got %d, want 3", len(cmds))
	}
	if cmds[0] != "cd app" || cmds[1] != "npm install" || cmds[2] != "npm run dev" {
		t.Errorf("AllCommands(): got %v", cmds)
	}
}

func TestLoad_MultipleActions_RoundTrip(t *testing.T) {
	path := writeTemp(t, `
tasks:
  - name: Multi
    actions:
      - command: echo one
      - command: echo two
    run_mode: stream
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := config.Write(path, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	reloaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}

	cmds := reloaded.Tasks[0].AllCommands()
	if len(cmds) != 2 || cmds[0] != "echo one" || cmds[1] != "echo two" {
		t.Errorf("round-trip AllCommands(): got %v", cmds)
	}
}
