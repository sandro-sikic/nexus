package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// UIMode controls how commands are presented.
type UIMode string

const (
	UIModeList  UIMode = "list"
	UIModeFuzzy UIMode = "fuzzy"
	UIModeGroup UIMode = "group"
)

// RunMode controls how a selected command is executed.
type RunMode string

const (
	RunModeStream     RunMode = "stream"
	RunModeHandoff    RunMode = "handoff"
	RunModeBackground RunMode = "background"
)

// Command represents a single runnable entry.
type Command struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Command     string  `yaml:"command"`
	Dir         string  `yaml:"dir"`      // working directory (optional)
	RunMode     RunMode `yaml:"run_mode"` // overrides top-level default
	Group       string  `yaml:"group"`    // used when ui_mode = group
}

// Config is the root structure of runner.yaml.
type Config struct {
	Title    string    `yaml:"title"`
	UIMode   UIMode    `yaml:"ui_mode"`
	RunMode  RunMode   `yaml:"run_mode"` // default run mode
	Commands []Command `yaml:"commands"`
}

// Load reads and parses a YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	if cfg.UIMode == "" {
		cfg.UIMode = UIModeList
	}
	if cfg.RunMode == "" {
		cfg.RunMode = RunModeStream
	}
	if cfg.Title == "" {
		cfg.Title = "Runner"
	}

	// Per-command defaults
	for i := range cfg.Commands {
		if cfg.Commands[i].RunMode == "" {
			cfg.Commands[i].RunMode = cfg.RunMode
		}
	}

	return &cfg, nil
}
