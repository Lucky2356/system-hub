//go:build !windows

package system

import "os/exec"

// hideConsoleWindow is a no-op outside Windows: there is no console window to
// suppress when spawning a child process.
func hideConsoleWindow(cmd *exec.Cmd) {}
