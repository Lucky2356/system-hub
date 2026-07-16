//go:build windows

package system

import (
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW — the child process gets no console window.
// https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
const createNoWindow = 0x08000000

// hideConsoleWindow stops a console window from flashing on screen every time
// the GUI spawns an external tool (docker, sensors, ...). Without this the app
// visibly blinks a cmd window on each poll.
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
