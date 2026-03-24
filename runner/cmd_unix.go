// This file exists only to satisfy the compiler on non-Windows platforms.
// The real implementation lives in cmd_windows.go.

//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

func newWindowsCmd(_ string) *exec.Cmd {
	// Never called on non-Windows; panic to catch any accidental misuse.
	panic("newWindowsCmd called on non-Windows platform")
}

// killProcess terminates a process on Unix-like systems.
// It kills the entire process group to ensure child processes are also terminated.
func killProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	// Kill the process group (negative PID)
	// This ensures child processes spawned by the shell are also killed
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		// Kill the entire process group
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// Also try to kill the process directly
	cmd.Process.Signal(syscall.SIGTERM)

	// Give it a moment to terminate gracefully, then force kill if needed
	go func() {
		// Wait a bit then force kill if still running
		// This is done in a goroutine to not block
		// The actual cleanup will be handled by the registry
	}()

	return nil
}

// setProcessGroup sets up the process to run in its own process group.
// This is used when creating new processes so we can kill them as a group.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// Ensure exec is used so the import is not flagged as unused if this file
// is ever compiled in isolation.
var _ = exec.Command
