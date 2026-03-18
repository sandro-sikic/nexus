// This file exists only to satisfy the compiler on non-Windows platforms.
// The real implementation lives in cmd_windows.go.

//go:build !windows

package runner

import "os/exec"

func newWindowsCmd(_ string) *exec.Cmd {
	// Never called on non-Windows; panic to catch any accidental misuse.
	panic("newWindowsCmd called on non-Windows platform")
}

// Ensure exec is used so the import is not flagged as unused if this file
// is ever compiled in isolation.
var _ = exec.Command
