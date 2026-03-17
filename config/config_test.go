package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"runner/config"
)

// writeTemp writes YAML content to a temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "runner-*.yaml")
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
title: "Test Runner"
ui_mode: fuzzy
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

	if cfg.Title != "Test Runner" {
		t.Errorf("title: got %q, want %q", cfg.Title, "Test Runner")
	}
	if cfg.UIMode != config.UIModeFuzzy {
		t.Errorf("ui_mode: got %q, want %q", cfg.UIMode, config.UIModeFuzzy)
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
	if cfg.Title != "Runner" {
		t.Errorf("default title: got %q, want %q", cfg.Title, "Runner")
	}
}

func TestLoad_DefaultUIMode(t *testing.T) {
	cfg, err := config.Load(writeTemp(t, "commands: []"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.UIMode != config.UIModeList {
		t.Errorf("default ui_mode: got %q, want list", cfg.UIMode)
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
	if string(config.UIModeList) != "list" {
		t.Errorf("UIModeList = %q, want list", config.UIModeList)
	}
	if string(config.UIModeFuzzy) != "fuzzy" {
		t.Errorf("UIModeFuzzy = %q, want fuzzy", config.UIModeFuzzy)
	}
	if string(config.UIModeGroup) != "group" {
		t.Errorf("UIModeGroup = %q, want group", config.UIModeGroup)
	}
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
