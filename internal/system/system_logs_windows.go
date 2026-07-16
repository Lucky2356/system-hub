//go:build windows

package system

import (
	"fmt"
	"strings"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

// SystemLogName labels the log backend in the UI. It is a function, not a
// constant, because the language can change at runtime.
func SystemLogName() string { return i18n.T("Windows Event Log") }

// GetSystemLogs returns the most recent entries from the Windows System event
// log. wevtutil is used rather than PowerShell's Get-WinEvent: it is a small
// native binary, so polling it does not cost a PowerShell startup each time.
func GetSystemLogs(lines int) (string, error) {
	if lines <= 0 {
		lines = 100
	}

	return queryEventLog("System", lines, "")
}

// getServiceLogs approximates journalctl -u: Windows has no per-service log, so
// filter the System log by the service as the event provider.
func getServiceLogs(serviceName string, lines int) (string, error) {
	// The name is already validated by the caller; quotes would break the XPath.
	if strings.ContainsAny(serviceName, `'"[]`) {
		return "", fmt.Errorf("invalid service name %q", serviceName)
	}

	query := fmt.Sprintf("*[System[Provider[@Name='%s']]]", serviceName)

	out, err := queryEventLog("System", lines, query)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return i18n.Tf("No event log entries for service %s.", serviceName), nil
	}
	return out, nil
}

// queryEventLog reads a channel newest-first as plain text. An empty xpath
// reads the whole channel.
func queryEventLog(channel string, lines int, xpath string) (string, error) {
	args := []string{
		"qe", channel,
		fmt.Sprintf("/c:%d", lines),
		"/rd:true", // reverse direction: newest entries first
		"/f:text",
	}
	if xpath != "" {
		args = append(args, "/q:"+xpath)
	}

	output, err := runCmd("wevtutil", args...)
	if err != nil {
		return "", fmt.Errorf("wevtutil %s: %w", channel, err)
	}

	return string(output), nil
}
