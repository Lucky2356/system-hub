package system

import (
	"bufio"
	"errors"
	"fmt"
	"runtime"
	"sort"
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

func ListServiceNames() ([]string, error) {
	services, err := ListServices()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(services))
	seen := make(map[string]struct{})

	for _, svc := range services {
		name := strings.TrimSpace(svc.Name)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}

		seen[name] = struct{}{}
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

func CountServices() (int, error) {
	services, err := ListServices()
	if err != nil {
		return 0, err
	}
	return len(services), nil
}

func ControlService(action string, serviceName string) error {
	if runtime.GOOS != "linux" {
		return errors.New("service control is available only on Linux (systemd)")
	}

	action = strings.TrimSpace(strings.ToLower(action))
	serviceName = strings.TrimSpace(serviceName)

	if serviceName == "" {
		return errors.New("service name is empty")
	}

	switch action {
	case "start", "stop", "restart", "enable", "disable":
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	_, err := runCmd("systemctl", action, serviceName)
	if err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, serviceName, err)
	}

	return nil
}

func GetServiceLogs(serviceName string, lines int) (string, error) {
	if runtime.GOOS != "linux" {
		return "", errors.New("logs are available only on Linux (journalctl)")
	}

	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "", errors.New("service name is empty")
	}

	if lines <= 0 {
		lines = 50
	}

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
