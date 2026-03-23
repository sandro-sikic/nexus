package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned by Load when the config file does not exist.
var ErrNotFound = errors.New("config file not found")

// RunMode controls how a selected command is executed.
type RunMode string

const (
	RunModeStream     RunMode = "stream"
	RunModeHandoff    RunMode = "handoff"
	RunModeBackground RunMode = "background"
)

// Step represents a single command step with optional background execution.
type Step struct {
	Command    string `yaml:"command"`
	Background bool   `yaml:"background,omitempty"` // run this step in background
}

// Command represents a single runnable entry.
type Command struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`            // single shell command
	Commands    []string `yaml:"commands,omitempty"` // multiple shell commands run sequentially (legacy)
	Steps       []Step   `yaml:"steps,omitempty"`    // multiple steps with background support
	Dir         string   `yaml:"dir"`                // working directory (optional)
	RunMode     RunMode  `yaml:"run_mode"`           // overrides top-level default
	Group       string   `yaml:"group"`              // optional group label for display grouping
}

// AllSteps returns the ordered list of shell commands to run.
// If Steps is set, commands are extracted from it. If Commands is set, it is used.
// Otherwise Command is returned as a single-item slice.
func (c Command) AllSteps() []string {
	if len(c.Steps) > 0 {
		cmds := make([]string, len(c.Steps))
		for i, step := range c.Steps {
			cmds[i] = step.Command
		}
		return cmds
	}
	if len(c.Commands) > 0 {
		return c.Commands
	}
	if c.Command != "" {
		return []string{c.Command}
	}
	return nil
}

// HasBackgroundSteps returns true if any step is marked as background.
func (c Command) HasBackgroundSteps() bool {
	for _, step := range c.Steps {
		if step.Background {
			return true
		}
	}
	return false
}

// GetStep returns the Step at the given index if using Steps field, otherwise returns a Step with the command.
func (c Command) GetStep(index int) Step {
	if len(c.Steps) > 0 && index < len(c.Steps) {
		return c.Steps[index]
	}
	cmds := c.AllSteps()
	if index < len(cmds) {
		return Step{Command: cmds[index]}
	}
	return Step{}
}

// Config is the root structure of nexus.yaml.
type Config struct {
	Title    string    `yaml:"title"`
	RunMode  RunMode   `yaml:"run_mode"` // default run mode
	Commands []Command `yaml:"commands"`
}

// Write serialises cfg to a YAML file at path, creating it if necessary.
func Write(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Load reads and parses a YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	if cfg.RunMode == "" {
		cfg.RunMode = RunModeStream
	}
	if cfg.Title == "" {
		cfg.Title = "Nexus"
	}

	// Per-command defaults
	for i := range cfg.Commands {
		if cfg.Commands[i].RunMode == "" {
			cfg.Commands[i].RunMode = cfg.RunMode
		}
	}

	return &cfg, nil
}
