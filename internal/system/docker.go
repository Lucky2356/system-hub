package system

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type DockerContainerInfo struct {
	ID     string
	Names  string
	Image  string
	State  string
	Status string
}

func ListDockerContainers() ([]DockerContainerInfo, error) {
	cmd := exec.Command(
		"docker",
		"ps",
		"-a",
		"--format",
		"{{json .}}",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return nil, fmt.Errorf("run docker ps: %w", err)
		}
		return nil, fmt.Errorf("run docker ps: %s", text)
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
}

func ControlDockerContainer(action string, containerName string) error {
	action = strings.TrimSpace(strings.ToLower(action))
	containerName = strings.TrimSpace(containerName)

	if containerName == "" {
		return fmt.Errorf("container name is empty")
	}

	switch action {
	case "start", "stop", "restart":
	default:
		return fmt.Errorf("unsupported docker action: %s", action)
	}

	cmd := exec.Command("docker", action, containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return fmt.Errorf("docker %s %s: %w", action, containerName, err)
		}
		return fmt.Errorf("docker %s %s: %s", action, containerName, text)
	}

	return nil
}

func GetDockerContainerLogs(containerName string, lines int) (string, error) {
	containerName = strings.TrimSpace(containerName)
	if containerName == "" {
		return "", fmt.Errorf("container name is empty")
	}

	if lines <= 0 {
		lines = 100
	}

	cmd := exec.Command(
		"docker",
		"logs",
		"--tail",
		fmt.Sprintf("%d", lines),
		containerName,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return "", fmt.Errorf("docker logs %s: %w", containerName, err)
		}
		return "", fmt.Errorf("docker logs %s: %s", containerName, text)
	}

	return string(output), nil
}