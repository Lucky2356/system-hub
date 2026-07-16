package system

import (
	"sort"
	"strings"
)

// ServiceInfo describes one service in the vocabulary systemd uses. The Windows
// provider maps its own states onto the same words (active/running,
// inactive/dead, activating, ...) so the UI — including StatusColor — does not
// need to know which platform produced the entry.
type ServiceInfo struct {
	Name        string
	LoadState   string
	ActiveState string
	SubState    string
	Description string
}

var servicesCache listCache[ServiceInfo]

// ListServices returns the host's services. Results are cached for a short TTL
// (see listCacheTTL) because a dashboard refresh consults the list several
// times per tick; use InvalidateServicesCache after mutating actions.
//
// The platform-specific work lives in listServices (services_linux.go /
// services_windows.go).
func ListServices() ([]ServiceInfo, error) {
	return servicesCache.get(listServices)
}

func InvalidateServicesCache() {
	servicesCache.invalidate()
}

// ControlService applies action ("start", "stop", "restart", "enable",
// "disable") to a service.
func ControlService(action string, serviceName string) error {
	action = strings.TrimSpace(strings.ToLower(action))
	serviceName = strings.TrimSpace(serviceName)

	if err := validateName("service", serviceName); err != nil {
		return err
	}

	switch action {
	case "start", "stop", "restart", "enable", "disable":
	default:
		return errUnsupportedAction(action)
	}

	if err := controlService(action, serviceName); err != nil {
		return err
	}

	InvalidateServicesCache()
	return nil
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

// GetServiceLogs returns the most recent log lines for a service.
func GetServiceLogs(serviceName string, lines int) (string, error) {
	serviceName = strings.TrimSpace(serviceName)
	if err := validateName("service", serviceName); err != nil {
		return "", err
	}

	if lines <= 0 {
		lines = 50
	}

	return getServiceLogs(serviceName, lines)
}
