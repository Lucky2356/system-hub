package system

import (
	"strings"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

const (
	DefaultRAMAlertPercent  = 85.0
	DefaultDiskAlertPercent = 90.0
)

func BuildProblems(stats Stats, favoriteServices []string, favoriteContainers []string) []string {
	var problems []string

	if stats.RAMPercent >= DefaultRAMAlertPercent {
		problems = append(problems, i18n.Tf("High RAM usage: %.1f%%", stats.RAMPercent))
	}

	if stats.DiskPercent >= DefaultDiskAlertPercent {
		problems = append(problems, i18n.Tf("High disk usage: %.1f%%", stats.DiskPercent))
	}

	if !stats.ServiceManagerAvailable {
		problems = append(problems, i18n.Tf("%s is unavailable", ServiceManagerName()))
	} else {
		serviceProblems := checkFavoriteServices(favoriteServices)
		problems = append(problems, serviceProblems...)
	}

	if !stats.DockerAvailable {
		problems = append(problems, i18n.Tf("%s is unavailable", "Docker"))
	} else {
		containerProblems := checkFavoriteContainers(favoriteContainers)
		problems = append(problems, containerProblems...)
	}

	// An empty slice means "healthy". Returning a human-readable sentinel here
	// used to make callers treat "no problems" as a problem — the dashboard
	// even raised a desktop notification announcing it.
	return problems
}

func checkFavoriteServices(favorites []string) []string {
	if len(favorites) == 0 {
		return nil
	}

	services, err := ListServices()
	if err != nil {
		return []string{i18n.Tf("Unable to check favorite services: %s", err.Error())}
	}

	serviceMap := make(map[string]ServiceInfo, len(services))
	for _, svc := range services {
		serviceMap[strings.TrimSpace(svc.Name)] = svc
	}

	var problems []string

	for _, name := range favorites {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		svc, ok := serviceMap[name]
		if !ok {
			problems = append(problems, i18n.Tf("Favorite service not found: %s", name))
			continue
		}

		if !strings.EqualFold(strings.TrimSpace(svc.ActiveState), "active") {
			problems = append(problems, i18n.Tf("Service %s is %s", name, svc.ActiveState))
		}
	}

	return problems
}

func checkFavoriteContainers(favorites []string) []string {
	if len(favorites) == 0 {
		return nil
	}

	containers, err := ListDockerContainers()
	if err != nil {
		return []string{i18n.Tf("Unable to check favorite containers: %s", err.Error())}
	}

	containerMap := make(map[string]DockerContainerInfo, len(containers))
	for _, c := range containers {
		containerMap[strings.TrimSpace(c.Names)] = c
	}

	var problems []string

	for _, name := range favorites {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		c, ok := containerMap[name]
		if !ok {
			problems = append(problems, i18n.Tf("Favorite container not found: %s", name))
			continue
		}

		if !strings.EqualFold(strings.TrimSpace(c.State), "running") {
			problems = append(problems, i18n.Tf("Container %s is %s", name, c.State))
		}
	}

	return problems
}
