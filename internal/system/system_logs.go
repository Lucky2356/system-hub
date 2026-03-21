package system

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func GetSystemLogs(lines int) (string, error) {
	if runtime.GOOS != "linux" {
		return "", errors.New("system logs are available only on Linux")
	}

	if lines <= 0 {
		lines = 100
	}

	cmd := exec.Command(
		"journalctl",
		"-n", fmt.Sprintf("%d", lines),
		"--no-pager",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return "", fmt.Errorf("journalctl system logs: %w", err)
		}
		return "", fmt.Errorf("journalctl system logs: %s", text)
	}

	return string(output), nil
}