package system

import (
	"fmt"
	"strings"
)

const (
	DefaultRAMAlertPercent  = 85.0
	DefaultDiskAlertPercent = 90.0
)

func BuildProblems(stats Stats, favoriteServices []string, favoriteContainers []string) []string {
	var problems []string

	if stats.RAMPercent >= DefaultRAMAlertPercent {
		problems = append(problems, fmt.Sprintf("High RAM usage: %.1f%%", stats.RAMPercent))
	}

	if stats.DiskPercent >= DefaultDiskAlertPercent {
		problems = append(problems, fmt.Sprintf("High disk usage: %.1f%%", stats.DiskPercent))
	}

	if !stats.SystemdAvailable {
		problems = append(problems, "systemd is unavailable")
	} else {
		serviceProblems := checkFavoriteServices(favoriteServices)
		problems = append(problems, serviceProblems...)
	}

	if !stats.DockerAvailable {
		problems = append(problems, "Docker is unavailable")
	} else {
		containerProblems := checkFavoriteContainers(favoriteContainers)
		problems = append(problems, containerProblems...)
	}

	if len(problems) == 0 {
		return []string{"No problems detected"}
	}

	return problems
}

func checkFavoriteServices(favorites []string) []string {
	if len(favorites) == 0 {
		return nil
	}

	services, err := ListServices()
	if err != nil {
		return []string{"Unable to check favorite services: " + err.Error()}
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
			problems = append(problems, "Favorite service not found: "+name)
			continue
		}

		if !strings.EqualFold(strings.TrimSpace(svc.ActiveState), "active") {
			problems = append(problems, fmt.Sprintf("Service %s is %s", name, svc.ActiveState))
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
		return []string{"Unable to check favorite containers: " + err.Error()}
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
			problems = append(problems, "Favorite container not found: "+name)
			continue
		}

		if !strings.EqualFold(strings.TrimSpace(c.State), "running") {
			problems = append(problems, fmt.Sprintf("Container %s is %s", name, c.State))
		}
	}

	return problems
}