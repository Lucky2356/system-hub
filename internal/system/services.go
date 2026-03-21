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

	return parseSystemctlListUnits(string(output)), nil
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
	case "start", "stop", "restart":
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	cmd := exec.Command("systemctl", action, serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return fmt.Errorf("systemctl %s %s: %w", action, serviceName, err)
		}
		return fmt.Errorf("systemctl %s %s: %s", action, serviceName, text)
	}

	return nil
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

func GetServiceLogs(serviceName string, lines int) (string, error) {
	if runtime.GOOS != "linux" {
		return "", errors.New("logs are available only on Linux (journalctl)")
	}

	if serviceName == "" {
		return "", errors.New("service name is empty")
	}

	if lines <= 0 {
		lines = 50
	}

	cmd := exec.Command(
		"journalctl",
		"-u", serviceName,
		"-n", fmt.Sprintf("%d", lines),
		"--no-pager",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("journalctl: %s", strings.TrimSpace(string(output)))
	}

	return string(output), nil
}