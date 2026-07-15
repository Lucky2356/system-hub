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

var servicesCache listCache[ServiceInfo]

// ListServices returns the systemd service list. Results are cached for a
// short TTL (see listCacheTTL) because dashboard refresh consults the list
// several times per tick; use InvalidateServicesCache after mutating actions.
func ListServices() ([]ServiceInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("services are available only on Linux (systemd)")
	}

	return servicesCache.get(func() ([]ServiceInfo, error) {
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
	})
}

func InvalidateServicesCache() {
	servicesCache.invalidate()
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

	if err := validateName("service", serviceName); err != nil {
		return err
	}

	switch action {
	case "start", "stop", "restart", "enable", "disable":
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	// "--" terminates option parsing so a crafted unit name cannot be
	// interpreted as a systemctl flag.
	_, err := runCmd("systemctl", action, "--", serviceName)
	if err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, serviceName, err)
	}

	InvalidateServicesCache()
	return nil
}

func GetServiceLogs(serviceName string, lines int) (string, error) {
	if runtime.GOOS != "linux" {
		return "", errors.New("logs are available only on Linux (journalctl)")
	}

	serviceName = strings.TrimSpace(serviceName)
	if err := validateName("service", serviceName); err != nil {
		return "", err
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
