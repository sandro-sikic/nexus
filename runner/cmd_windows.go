package runner

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

// newWindowsCmd builds an exec.Cmd that runs shellCmd via cmd.exe without
// Go's argument quoting. Setting SysProcAttr.CmdLine passes the raw string
// directly to the Windows process creation API, so characters like colons in
// remote paths (e.g. scp host:/path) are not mangled.
func newWindowsCmd(shellCmd string) *exec.Cmd {
	c := exec.Command("cmd")
	c.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:       "/C " + shellCmd,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	return c
}

// killProcess terminates a process on Windows.
// It uses taskkill to terminate the process and its children.
func killProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	pid := strconv.Itoa(cmd.Process.Pid)

	// Use taskkill to kill the process tree
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", pid)
	if err := killCmd.Run(); err != nil {
		// Taskkill failed, try direct kill
		fmt.Fprintf(cmd.Stderr, "taskkill failed: %v, trying direct kill\n", err)
	}

	// Also try direct kill
	return cmd.Process.Kill()
}

// setProcessGroup is a no-op on Windows since we handle it in newWindowsCmd.
func setProcessGroup(cmd *exec.Cmd) {
	// Already set in newWindowsCmd via CREATE_NEW_PROCESS_GROUP
}
