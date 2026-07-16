//go:build linux

package system

import (
	"bufio"
	"fmt"
	"strings"
)

// ServiceManagerName labels the service backend in the UI.
const ServiceManagerName = "systemd"

func errUnsupportedAction(action string) error {
	return fmt.Errorf("unsupported action: %s", action)
}

func listServices() ([]ServiceInfo, error) {
	output, err := runCmd("systemctl",
		"list-units",
		"--type=service",
		"--all",
		"--no-pager",
		"--no-legend",
		"--plain",
	)
	if err != nil {
		return nil, fmt.Errorf("run systemctl list-units: %w", err)
	}

	return parseSystemctlListUnits(string(output)), nil
}

func controlService(action, serviceName string) error {
	// "--" terminates option parsing so a crafted unit name cannot be
	// interpreted as a systemctl flag.
	if _, err := runCmd("systemctl", action, "--", serviceName); err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, serviceName, err)
	}
	return nil
}

func getServiceLogs(serviceName string, lines int) (string, error) {
	output, err := runCmd("journalctl",
		"-u", serviceName,
		"-n", fmt.Sprintf("%d", lines),
		"--no-pager",
	)
	if err != nil {
		return "", fmt.Errorf("journalctl %s: %w", serviceName, err)
	}

	return string(output), nil
}

// serviceManagerAvailable reports whether systemd can be reached.
func serviceManagerAvailable() bool {
	_, err := runCmd("systemctl", "--version")
	return err == nil
}

func parseSystemctlListUnits(output string) []ServiceInfo {
	var services []ServiceInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		services = append(services, ServiceInfo{
			Name:        fields[0],
			LoadState:   fields[1],
			ActiveState: fields[2],
			SubState:    fields[3],
			Description: strings.Join(fields[4:], " "),
		})
	}

	return services
}
