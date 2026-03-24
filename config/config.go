package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned by Load when the config file does not exist.
var ErrNotFound = errors.New("config file not found")

// Action represents a single executable shell command.
type Action struct {
	Command    string `yaml:"command"`
	Background bool   `yaml:"background,omitempty"` // run this action in background
}

// Task represents a named task with multiple actions to execute.
type Task struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Actions     []Action `yaml:"actions"`           // list of actions to execute sequentially
	Dir         string   `yaml:"dir"`               // working directory (optional)
	Group       string   `yaml:"group"`             // optional group label for display grouping
	Handoff     bool     `yaml:"handoff,omitempty"` // hand off terminal for last action
}

// AllCommands returns all shell commands as strings for display purposes.
func (t Task) AllCommands() []string {
	cmds := make([]string, len(t.Actions))
	for i, a := range t.Actions {
		cmds[i] = a.Command
	}
	return cmds
}

// HasBackgroundActions returns true if any action is marked as background.
func (t Task) HasBackgroundActions() bool {
	for _, a := range t.Actions {
		if a.Background {
			return true
		}
	}
	return false
}

// HasHandoff returns true if the task has handoff enabled.
func (t Task) HasHandoff() bool {
	return t.Handoff
}

// LastAction returns the last action of the task, or nil if no actions.
func (t Task) LastAction() *Action {
	if len(t.Actions) == 0 {
		return nil
	}
	return &t.Actions[len(t.Actions)-1]
}

// Config is the root structure of nexus.yaml.
type Config struct {
	Title string `yaml:"title"`
	Tasks []Task `yaml:"tasks"`
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
	if cfg.Title == "" {
		cfg.Title = "Nexus"
	}

	return &cfg, nil
}
