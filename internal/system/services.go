package system

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type ServiceInfo struct {
	Name        string
	LoadState   string
	ActiveState string
	SubState    string
	Description string
}

func ListServices() ([]ServiceInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("services are available only on Linux (systemd)")
	}

	cmd := exec.Command(
		"systemctl",
		"list-units",
		"--type=service",
		"--all",
		"--no-pager",
		"--no-legend",
		"--plain",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run systemctl list-units: %w", err)
	}

	services := parseSystemctlListUnits(string(output))
	return services, nil
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

		name := fields[0]
		loadState := fields[1]
		activeState := fields[2]
		subState := fields[3]
		description := strings.Join(fields[4:], " ")

		services = append(services, ServiceInfo{
			Name:        name,
			LoadState:   loadState,
			ActiveState: activeState,
			SubState:    subState,
			Description: description,
		})
	}

	return services
}