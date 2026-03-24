package runner

import (
	"os/exec"
	"sync"
)

// ProcessRegistry tracks all running processes spawned by nexus.
// It provides a mechanism to terminate all processes on application exit.
type ProcessRegistry struct {
	mu        sync.RWMutex
	processes map[*exec.Cmd]bool
}

// NewProcessRegistry creates a new process registry.
func NewProcessRegistry() *ProcessRegistry {
	return &ProcessRegistry{
		processes: make(map[*exec.Cmd]bool),
	}
}

// Register adds a process to the registry.
func (pr *ProcessRegistry) Register(cmd *exec.Cmd) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.processes[cmd] = true
}

// Unregister removes a process from the registry.
func (pr *ProcessRegistry) Unregister(cmd *exec.Cmd) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	delete(pr.processes, cmd)
}

// KillAll terminates all registered processes.
// It attempts to kill each process and returns the last error encountered.
func (pr *ProcessRegistry) KillAll() error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	var lastErr error
	for cmd := range pr.processes {
		if cmd.Process != nil {
			if err := killProcess(cmd); err != nil {
				lastErr = err
			}
		}
		delete(pr.processes, cmd)
	}
	return lastErr
}

// Count returns the number of registered processes.
func (pr *ProcessRegistry) Count() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	return len(pr.processes)
}

// Global registry instance
var globalRegistry = NewProcessRegistry()

// GetGlobalRegistry returns the global process registry.
func GetGlobalRegistry() *ProcessRegistry {
	return globalRegistry
}
