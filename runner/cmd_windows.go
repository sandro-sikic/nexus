package runner

import (
	"os/exec"
	"syscall"
)

// newWindowsCmd builds an exec.Cmd that runs shellCmd via cmd.exe without
// Go's argument quoting. Setting SysProcAttr.CmdLine passes the raw string
// directly to the Windows process creation API, so characters like colons in
// remote paths (e.g. scp host:/path) are not mangled.
func newWindowsCmd(shellCmd string) *exec.Cmd {
	c := exec.Command("cmd")
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: "/C " + shellCmd}
	return c
}
