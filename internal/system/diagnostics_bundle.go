package system

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lucky2356/system-hub/internal/activity"
)

type DiagnosticsBundleParams struct {
	FavoriteServices   []string
	FavoriteContainers []string
	LogLines           int
	TopProcessLimit    int
}

func BuildDiagnosticsBundle(params DiagnosticsBundleParams) (string, error) {
	var sections []string

	now := time.Now().Format("2006-01-02 15:04:05")
	sections = append(sections, "System Hub Diagnostics Bundle")
	sections = append(sections, "Generated: "+now)
	sections = append(sections, strings.Repeat("=", 80))

	logLines := params.LogLines
	if logLines <= 0 {
		logLines = 100
	}

	topLimit := params.TopProcessLimit
	if topLimit <= 0 {
		topLimit = 5
	}

	stats, err := GetStats()
	if err != nil {
		sections = append(sections, buildNamedSection("Stats", "Error: "+err.Error()))
	} else {
		sections = append(sections, buildStatsSection(stats))
		sections = append(sections, buildProblemsSection(stats, params.FavoriteServices, params.FavoriteContainers))
	}

	sections = append(sections, buildFavoritesSection(params.FavoriteServices, params.FavoriteContainers))

	topProcesses, err := ListTopProcesses(topLimit)
	if err != nil {
		sections = append(sections, buildNamedSection("Top Processes", "Error: "+err.Error()))
	} else {
		sections = append(sections, buildTopProcessesSection(topProcesses))
	}

	entries := activity.List()
	sections = append(sections, buildActivitySection(entries))

	for _, name := range params.FavoriteServices {
		if strings.TrimSpace(name) == "" {
			continue
		}

		logs, err := GetServiceLogs(name, logLines)
		if err != nil {
			sections = append(sections, buildNamedSection("Service Logs: "+name, "Error: "+err.Error()))
			continue
		}

		sections = append(sections, buildNamedSection("Service Logs: "+name, logs))
	}

	for _, name := range params.FavoriteContainers {
		if strings.TrimSpace(name) == "" {
			continue
		}

		logs, err := GetDockerContainerLogs(name, logLines)
		if err != nil {
			sections = append(sections, buildNamedSection("Container Logs: "+name, "Error: "+err.Error()))
			continue
		}

		sections = append(sections, buildNamedSection("Container Logs: "+name, logs))
	}

	return strings.Join(sections, "\n\n"), nil
}

func buildTopProcessesSection(items []ProcessUsageInfo) string {
	lines := []string{
		"Top Processes",
		strings.Repeat("-", 80),
	}

	if len(items) == 0 {
		lines = append(lines, "No process data")
		return strings.Join(lines, "\n")
	}

	for _, item := range items {
		lines = append(lines,
			fmt.Sprintf(
				"%s | PID=%d | CPU=%.1f%% | RAM=%s | RAM%%=%.1f",
				item.ProcessName,
				item.PID,
				item.CPUPercent,
				FormatBytes(item.MemoryBytes),
				item.MemoryPercent,
			),
		)
	}

	return strings.Join(lines, "\n")
}

func buildActivitySection(entries []activity.Entry) string {
	lines := []string{
		"Activity",
		strings.Repeat("-", 80),
	}

	if len(entries) == 0 {
		lines = append(lines, "No activity entries")
		return strings.Join(lines, "\n")
	}

	for _, e := range entries {
		line := fmt.Sprintf(
			"%s | %s | %s | %s | %s",
			e.Time.Format("2006-01-02 15:04:05"),
			e.Target,
			e.Action,
			e.Name,
			e.Status,
		)

		if strings.TrimSpace(e.Details) != "" {
			line += " | " + e.Details
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
