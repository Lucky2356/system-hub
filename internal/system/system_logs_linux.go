//go:build linux

package system

import (
	"fmt"
)

// SystemLogName labels the log backend in the UI. "journalctl" is a proper noun
// and is not translated.
func SystemLogName() string { return "journalctl" }

// GetSystemLogs returns the most recent journal entries.
func GetSystemLogs(lines int) (string, error) {
	if lines <= 0 {
		lines = 100
	}

	output, err := runCmd("journalctl", "-n", fmt.Sprintf("%d", lines), "--no-pager")
	if err != nil {
		return "", fmt.Errorf("journalctl system logs: %w", err)
	}

	return string(output), nil
}
