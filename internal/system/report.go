package system

import (
	"fmt"
	"strings"
	"time"
)

type DiagnosticsReportParams struct {
	FavoriteServices   []string
	FavoriteContainers []string
	LogLines           int
}

func BuildDiagnosticsReport(params DiagnosticsReportParams) (string, error) {
	var sections []string

	now := time.Now().Format("2006-01-02 15:04:05")
	sections = append(sections, "System Hub Diagnostics Report")
	sections = append(sections, "Generated: "+now)
	sections = append(sections, strings.Repeat("=", 72))

	stats, err := GetStats()
	if err != nil {
		sections = append(sections, "Stats: error: "+err.Error())
	} else {
		sections = append(sections, buildStatsSection(stats))
		sections = append(sections, buildProblemsSection(stats, params.FavoriteServices, params.FavoriteContainers))
	}

	sections = append(sections, buildFavoritesSection(params.FavoriteServices, params.FavoriteContainers))

	services, err := ListServices()
	if err != nil {
		sections = append(sections, "Services list: error: "+err.Error())
	} else {
		sections = append(sections, buildServicesSection(services, params.FavoriteServices))
	}

	containers, err := ListDockerContainers()
	if err != nil {
		sections = append(sections, "Docker containers list: error: "+err.Error())
	} else {
		sections = append(sections, buildContainersSection(containers, params.FavoriteContainers))
	}

	logLines := params.LogLines
	if logLines <= 0 {
		logLines = 100
	}

	for _, name := range params.FavoriteServices {
		logs, err := GetServiceLogs(name, logLines)
		if err != nil {
			sections = append(sections, fmt.Sprintf("Service logs [%s]: error: %v", name, err))
			continue
		}
		sections = append(sections, buildNamedSection("Service logs: "+name, logs))
	}

	for _, name := range params.FavoriteContainers {
		logs, err := GetDockerContainerLogs(name, logLines)
		if err != nil {
			sections = append(sections, fmt.Sprintf("Container logs [%s]: error: %v", name, err))
			continue
		}
		sections = append(sections, buildNamedSection("Container logs: "+name, logs))
	}

	return strings.Join(sections, "\n\n"), nil
}

func buildStatsSection(stats Stats) string {
	var lines []string
	lines = append(lines, "Stats")
	lines = append(lines, strings.Repeat("-", 72))
	lines = append(lines, fmt.Sprintf("CPU: %.1f%%", stats.CPUPercent))
	lines = append(lines, fmt.Sprintf("RAM: %.1f%% (%s / %s)", stats.RAMPercent, FormatBytes(stats.RAMUsed), FormatBytes(stats.RAMTotal)))
	lines = append(lines, fmt.Sprintf("Disk: %.1f%% (%s / %s)", stats.DiskPercent, FormatBytes(stats.DiskUsed), FormatBytes(stats.DiskTotal)))

	if stats.UptimeKnown {
		lines = append(lines, "Uptime: "+FormatUptime(stats.UptimeSeconds))
	} else {
		lines = append(lines, "Uptime: unavailable")
	}

	if stats.SystemdAvailable {
		lines = append(lines, "systemd: available")
	} else {
		lines = append(lines, "systemd: unavailable")
	}

	if stats.ServiceCountKnown {
		lines = append(lines, fmt.Sprintf("Services count: %d", stats.ServiceCount))
	} else {
		lines = append(lines, "Services count: unavailable")
	}

	if stats.DockerAvailable {
		lines = append(lines, "Docker: available")
	} else {
		lines = append(lines, "Docker: unavailable")
	}

	if stats.DockerCountKnown {
		lines = append(lines, fmt.Sprintf("Containers count: %d", stats.DockerContainerCount))
		lines = append(lines, fmt.Sprintf("Running containers: %d", stats.DockerRunningCount))
	} else {
		lines = append(lines, "Containers count: unavailable")
		lines = append(lines, "Running containers: unavailable")
	}

	return strings.Join(lines, "\n")
}

func buildProblemsSection(stats Stats, favoriteServices []string, favoriteContainers []string) string {
	problems := BuildProblems(stats, favoriteServices, favoriteContainers)

	lines := []string{
		"Problems",
		strings.Repeat("-", 72),
	}
	if len(problems) == 0 {
		lines = append(lines, "- none")
	}
	for _, p := range problems {
		lines = append(lines, "- "+p)
	}

	return strings.Join(lines, "\n")
}

func buildFavoritesSection(favoriteServices []string, favoriteContainers []string) string {
	lines := []string{
		"Favorites",
		strings.Repeat("-", 72),
		"Favorite services:",
	}

	if len(favoriteServices) == 0 {
		lines = append(lines, "- none")
	} else {
		for _, s := range favoriteServices {
			lines = append(lines, "- "+s)
		}
	}

	lines = append(lines, "", "Favorite containers:")

	if len(favoriteContainers) == 0 {
		lines = append(lines, "- none")
	} else {
		for _, c := range favoriteContainers {
			lines = append(lines, "- "+c)
		}
	}

	return strings.Join(lines, "\n")
}

func buildServicesSection(services []ServiceInfo, favorites []string) string {
	favSet := make(map[string]bool, len(favorites))
	for _, f := range favorites {
		favSet[f] = true
	}

	lines := []string{
		"Services",
		strings.Repeat("-", 72),
	}

	for _, s := range services {
		prefix := " "
		if favSet[s.Name] {
			prefix = "*"
		}
		lines = append(lines, fmt.Sprintf("%s %s | load=%s | active=%s | sub=%s | %s",
			prefix, s.Name, s.LoadState, s.ActiveState, s.SubState, s.Description))
	}

	return strings.Join(lines, "\n")
}

func buildContainersSection(containers []DockerContainerInfo, favorites []string) string {
	favSet := make(map[string]bool, len(favorites))
	for _, f := range favorites {
		favSet[f] = true
	}

	lines := []string{
		"Docker Containers",
		strings.Repeat("-", 72),
	}

	for _, c := range containers {
		prefix := " "
		if favSet[c.Names] {
			prefix = "*"
		}
		lines = append(lines, fmt.Sprintf("%s %s | image=%s | state=%s | status=%s | id=%s",
			prefix, c.Names, c.Image, c.State, c.Status, c.ID))
	}

	return strings.Join(lines, "\n")
}

func buildNamedSection(title, body string) string {
	return title + "\n" + strings.Repeat("-", 72) + "\n" + body
}
