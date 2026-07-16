package system

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type DockerContainerInfo struct {
	ID     string
	Names  string
	Image  string
	State  string
	Status string
}

type DockerImageInfo struct {
	ID         string
	Repository string
	Tag        string
	Size       string
}

// DockerContainerStats is the live resource usage of one running container.
// Values are kept as docker formats them ("0.15%", "12.4MiB / 1.94GiB"): they
// are shown as-is, and re-deriving them would only add rounding of our own.
type DockerContainerStats struct {
	Name       string
	CPUPercent string
	MemUsage   string
	MemPercent string
}

var containersCache listCache[DockerContainerInfo]

// ListDockerContainers returns all containers. Results are cached for a short
// TTL (see listCacheTTL); use InvalidateContainersCache after mutating actions.
func ListDockerContainers() ([]DockerContainerInfo, error) {
	return containersCache.get(func() ([]DockerContainerInfo, error) {
		output, err := runCmd("docker", "ps", "-a", "--format", "{{json .}}")
		if err != nil {
			return nil, fmt.Errorf("run docker ps: %w", err)
		}

		rawText := strings.TrimSpace(string(output))
		if rawText == "" {
			return []DockerContainerInfo{}, nil
		}

		lines := strings.Split(rawText, "\n")
		var containers []DockerContainerInfo

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var raw struct {
				ID     string
				Image  string
				Names  string
				State  string
				Status string
			}

			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				return nil, fmt.Errorf("parse docker output: %w", err)
			}

			containers = append(containers, DockerContainerInfo{
				ID:     raw.ID,
				Names:  raw.Names,
				Image:  raw.Image,
				State:  raw.State,
				Status: raw.Status,
			})
		}

		return containers, nil
	})
}

func InvalidateContainersCache() {
	containersCache.invalidate()
	statsCache.invalidate()
}

var statsCache listCache[DockerContainerStats]

// ListDockerContainerStats returns live CPU and memory usage for the running
// containers, keyed by container name.
//
// --no-stream is what makes this usable at all: without it docker streams
// forever and the command never returns. Even so it samples for about a second
// before printing, which is why it is not folded into ListDockerContainers --
// callers that only need names and states must not pay for it.
func ListDockerContainerStats() ([]DockerContainerStats, error) {
	return statsCache.get(func() ([]DockerContainerStats, error) {
		output, err := runCmd("docker", "stats", "--no-stream", "--format", "{{json .}}")
		if err != nil {
			return nil, fmt.Errorf("run docker stats: %w", err)
		}
		return parseDockerStats(string(output))
	})
}

// DockerStatsByName indexes the stats by container name for row lookups.
func DockerStatsByName() (map[string]DockerContainerStats, error) {
	stats, err := ListDockerContainerStats()
	if err != nil {
		return nil, err
	}

	byName := make(map[string]DockerContainerStats, len(stats))
	for _, s := range stats {
		byName[s.Name] = s
	}
	return byName, nil
}

// parseDockerStats reads the JSON-per-line output of `docker stats`.
func parseDockerStats(output string) ([]DockerContainerStats, error) {
	rawText := strings.TrimSpace(output)
	if rawText == "" {
		return []DockerContainerStats{}, nil
	}

	var stats []DockerContainerStats

	for _, line := range strings.Split(rawText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw struct {
			Name     string
			CPUPerc  string
			MemUsage string
			MemPerc  string
		}

		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("parse docker stats output: %w", err)
		}

		name := strings.TrimSpace(raw.Name)
		if name == "" {
			continue
		}

		stats = append(stats, DockerContainerStats{
			Name:       name,
			CPUPercent: strings.TrimSpace(raw.CPUPerc),
			MemUsage:   strings.TrimSpace(raw.MemUsage),
			MemPercent: strings.TrimSpace(raw.MemPerc),
		})
	}

	return stats, nil
}

func ListDockerContainerNames() ([]string, error) {
	containers, err := ListDockerContainers()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(containers))
	seen := make(map[string]struct{})

	for _, c := range containers {
		name := strings.TrimSpace(c.Names)
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

func CountDockerContainers() (total int, running int, err error) {
	containers, err := ListDockerContainers()
	if err != nil {
		return 0, 0, err
	}

	total = len(containers)

	for _, c := range containers {
		if strings.EqualFold(c.State, "running") {
			running++
		}
	}

	return total, running, nil
}

func ControlDockerContainer(action string, containerName string) error {
	action = strings.TrimSpace(strings.ToLower(action))
	containerName = strings.TrimSpace(containerName)

	if err := validateName("container", containerName); err != nil {
		return err
	}

	switch action {
	case "start", "stop", "restart":
	default:
		return fmt.Errorf("unsupported docker action: %s", action)
	}

	_, err := runCmd("docker", action, "--", containerName)
	if err != nil {
		return fmt.Errorf("docker %s %s: %w", action, containerName, err)
	}

	InvalidateContainersCache()
	return nil
}

func GetDockerContainerLogs(containerName string, lines int) (string, error) {
	containerName = strings.TrimSpace(containerName)
	if err := validateName("container", containerName); err != nil {
		return "", err
	}

	if lines <= 0 {
		lines = 100
	}

	output, err := runCmd("docker", "logs", "--tail", fmt.Sprintf("%d", lines), "--", containerName)
	if err != nil {
		return "", fmt.Errorf("docker logs %s: %w", containerName, err)
	}

	return string(output), nil
}

func ListDockerImages() ([]DockerImageInfo, error) {
	output, err := runCmd("docker", "images", "--format", "{{json .}}")
	if err != nil {
		return nil, fmt.Errorf("run docker images: %w", err)
	}

	rawText := strings.TrimSpace(string(output))
	if rawText == "" {
		return []DockerImageInfo{}, nil
	}

	lines := strings.Split(rawText, "\n")
	var images []DockerImageInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw struct {
			ID         string `json:"ID"`
			Repository string `json:"Repository"`
			Tag        string `json:"Tag"`
			Size       string `json:"Size"`
		}

		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("parse docker images output: %w", err)
		}

		if raw.Repository == "<none>" {
			continue
		}

		images = append(images, DockerImageInfo{
			ID:         raw.ID,
			Repository: raw.Repository,
			Tag:        raw.Tag,
			Size:       raw.Size,
		})
	}

	sort.SliceStable(images, func(i, j int) bool {
		return images[i].Repository < images[j].Repository
	})

	return images, nil
}

func PullDockerImage(imageName string) error {
	imageName = strings.TrimSpace(imageName)
	if err := validateName("image", imageName); err != nil {
		return err
	}

	// Pulling a large image can far exceed the default 30s timeout, so run it
	// without a deadline (still cancellable via process exit).
	_, err := runCmdContext(context.Background(), 0, "docker", "pull", "--", imageName)
	if err != nil {
		return fmt.Errorf("docker pull %s: %w", imageName, err)
	}

	return nil
}

func RemoveDockerImage(imageName string) error {
	imageName = strings.TrimSpace(imageName)
	if err := validateName("image", imageName); err != nil {
		return err
	}

	_, err := runCmd("docker", "rmi", "--", imageName)
	if err != nil {
		return fmt.Errorf("docker rmi %s: %w", imageName, err)
	}

	return nil
}

func GetDockerContainerInspect(containerName string) (string, error) {
	containerName = strings.TrimSpace(containerName)
	if err := validateName("container", containerName); err != nil {
		return "", err
	}

	output, err := runCmd("docker", "inspect", "--", containerName)
	if err != nil {
		return "", fmt.Errorf("docker inspect %s: %w", containerName, err)
	}

	if len(output) == 0 {
		return "No output", nil
	}

	return string(output), nil
}
