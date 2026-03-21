package system

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type DockerContainerInfo struct {
	ID      string
	Names   string
	Image   string
	State   string
	Status  string
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

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
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