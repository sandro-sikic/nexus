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
commands:
  - name: Build
    description: Build the project
    command: "make build"
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
	if len(cfg.Commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(cfg.Commands))
	}

	cmd := cfg.Commands[0]
	if cmd.Name != "Build" {
		t.Errorf("cmd.name: got %q, want %q", cmd.Name, "Build")
	}
	if cmd.Description != "Build the project" {
		t.Errorf("cmd.description: got %q", cmd.Description)
	}
	if cmd.Command != "make build" {
		t.Errorf("cmd.command: got %q", cmd.Command)
	}
	if cmd.Dir != "/tmp" {
		t.Errorf("cmd.dir: got %q", cmd.Dir)
	}
	if cmd.RunMode != config.RunModeStream {
		t.Errorf("cmd.run_mode: got %q, want stream", cmd.RunMode)
	}
	if cmd.Group != "CI" {
		t.Errorf("cmd.group: got %q, want CI", cmd.Group)
	}
}

// ── Defaults ─────────────────────────────────────────────────────────────────

func TestLoad_DefaultTitle(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "commands: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "Nexus" {
		t.Errorf("default title: got %q, want %q", cfg.Title, "Nexus")
	}
}

func TestLoad_DefaultRunMode(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "commands: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RunMode != config.RunModeStream {
		t.Errorf("default run_mode: got %q, want stream", cfg.RunMode)
	}
}

func TestLoad_CommandInheritsTopLevelRunMode(t *testing.T) {
	yaml := `
run_mode: background
commands:
  - name: Test
    command: "echo hi"
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Commands[0].RunMode != config.RunModeBackground {
		t.Errorf("inherited run_mode: got %q, want background", cfg.Commands[0].RunMode)
	}
}

func TestLoad_CommandRunModeNotOverriddenWhenExplicit(t *testing.T) {
	yaml := `
run_mode: stream
commands:
  - name: Deploy
    command: "make deploy"
    run_mode: handoff
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Commands[0].RunMode != config.RunModeHandoff {
		t.Errorf("explicit run_mode: got %q, want handoff", cfg.Commands[0].RunMode)
	}
}

func TestLoad_MultipleCommandsInheritance(t *testing.T) {
	yaml := `
run_mode: background
commands:
  - name: A
    command: echo a
  - name: B
    command: echo b
    run_mode: stream
  - name: C
    command: echo c
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
		if got := cfg.Commands[tc.idx].RunMode; got != tc.want {
			t.Errorf("commands[%d].run_mode: got %q, want %q", tc.idx, got, tc.want)
		}
	}
}

func TestLoad_EmptyCommandList(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "commands: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Commands) != 0 {
		t.Errorf("expected empty commands, got %d", len(cfg.Commands))
	}
}

func TestLoad_NoCommandsKey(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "title: bare"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Commands != nil {
		t.Errorf("expected nil commands, got %v", cfg.Commands)
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

func TestLoad_MalformedCommands(t *testing.T) {
	// commands should be a list, not a string — YAML should error
	_, err := config.Load(writeTemp(t, "commands: not-a-list"))
	if err == nil {
		t.Fatal("expected error for malformed commands, got nil")
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

func TestLoad_CommandOptionalFields(t *testing.T) {
	yaml := `
commands:
  - name: Minimal
    command: echo hi
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := cfg.Commands[0]
	if cmd.Dir != "" {
		t.Errorf("dir should default empty, got %q", cmd.Dir)
	}
	if cmd.Group != "" {
		t.Errorf("group should default empty, got %q", cmd.Group)
	}
	if cmd.Description != "" {
		t.Errorf("description should default empty, got %q", cmd.Description)
	}
}

// ── Multi-command (Steps) ────────────────────────────────────────────────────

func TestCommand_Steps_SingleCommand(t *testing.T) {
	cmd := config.Command{Command: "echo hi"}
	steps := cmd.Steps()
	if len(steps) != 1 || steps[0] != "echo hi" {
		t.Errorf("Steps(): got %v, want [echo hi]", steps)
	}
}

func TestCommand_Steps_MultipleCommands(t *testing.T) {
	cmd := config.Command{Commands: []string{"cd app", "npm run dev"}}
	steps := cmd.Steps()
	if len(steps) != 2 {
		t.Fatalf("Steps() len: got %d, want 2", len(steps))
	}
	if steps[0] != "cd app" || steps[1] != "npm run dev" {
		t.Errorf("Steps(): got %v", steps)
	}
}

func TestCommand_Steps_CommandsFieldTakesPrecedence(t *testing.T) {
	// When both Command and Commands are set, Commands wins.
	cmd := config.Command{
		Command:  "single",
		Commands: []string{"first", "second"},
	}
	steps := cmd.Steps()
	if len(steps) != 2 || steps[0] != "first" {
		t.Errorf("Steps() should use Commands when both set: got %v", steps)
	}
}

func TestCommand_Steps_Empty(t *testing.T) {
	cmd := config.Command{}
	steps := cmd.Steps()
	if len(steps) != 0 {
		t.Errorf("Steps() on empty command: got %v, want []", steps)
	}
}

func TestLoad_MultiCommandField(t *testing.T) {
	yaml := `
commands:
  - name: Setup
    commands:
      - "cd app"
      - "npm install"
      - "npm run dev"
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(cfg.Commands))
	}
	steps := cfg.Commands[0].Steps()
	if len(steps) != 3 {
		t.Fatalf("Steps() len: got %d, want 3", len(steps))
	}
	if steps[0] != "cd app" || steps[1] != "npm install" || steps[2] != "npm run dev" {
		t.Errorf("Steps(): got %v", steps)
	}
}

func TestLoad_MultiCommand_RoundTrip(t *testing.T) {
	path := writeTemp(t, `
commands:
  - name: Multi
    commands:
      - echo one
      - echo two
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

	steps := reloaded.Commands[0].Steps()
	if len(steps) != 2 || steps[0] != "echo one" || steps[1] != "echo two" {
		t.Errorf("round-trip Steps(): got %v", steps)
	}
}
