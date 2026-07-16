package system

import (
	"errors"
	"fmt"
	"runtime"
)

func GetSystemLogs(lines int) (string, error) {
	if runtime.GOOS != "linux" {
		return "", errors.New("system logs are available only on Linux")
	}

	if lines <= 0 {
		lines = 100
	}

	output, err := runCmd("journalctl", "-n", fmt.Sprintf("%d", lines), "--no-pager")
	if err != nil {
		return "", fmt.Errorf("journalctl system logs: %w", err)
	}

	return string(output), nil
}
