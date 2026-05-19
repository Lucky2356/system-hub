package system

import (
	"fmt"
	"strings"
)

type SafeCommand struct {
	Key         string
	Title       string
	Description string
	NeedsArg    bool
	ArgHint     string
	BuildArgs   func(arg string) ([]string, error)
}

func GetSafeCommands() []SafeCommand {
	return []SafeCommand{
		{
			Key:         "systemctl-list-services",
			Title:       "systemctl list services",
			Description: "Показать список systemd-сервисов",
			NeedsArg:    false,
			BuildArgs: func(arg string) ([]string, error) {
				return []string{"list-units", "--type=service", "--all", "--no-pager"}, nil
			},
		},
		{
			Key:         "systemctl-status",
			Title:       "systemctl status <service>",
			Description: "Показать статус одного systemd-сервиса",
			NeedsArg:    true,
			ArgHint:     "Например: nginx.service",
			BuildArgs: func(arg string) ([]string, error) {
				arg = strings.TrimSpace(arg)
				if arg == "" {
					return nil, fmt.Errorf("service name is required")
				}
				return []string{"status", arg, "--no-pager"}, nil
			},
		},
		{
			Key:         "docker-ps",
			Title:       "docker ps -a",
			Description: "Показать все контейнеры Docker",
			NeedsArg:    false,
			BuildArgs: func(arg string) ([]string, error) {
				return []string{"ps", "-a"}, nil
			},
		},
		{
			Key:         "docker-images",
			Title:       "docker images",
			Description: "Показать Docker images",
			NeedsArg:    false,
			BuildArgs: func(arg string) ([]string, error) {
				return []string{"images"}, nil
			},
		},
		{
			Key:         "journalctl-tail",
			Title:       "journalctl -n 100",
			Description: "Показать последние системные логи",
			NeedsArg:    false,
			BuildArgs: func(arg string) ([]string, error) {
				return []string{"-n", "100", "--no-pager"}, nil
			},
		},
		{
			Key:         "journalctl-service",
			Title:       "journalctl -u <service> -n 100",
			Description: "Показать последние логи конкретного сервиса",
			NeedsArg:    true,
			ArgHint:     "Например: docker.service",
			BuildArgs: func(arg string) ([]string, error) {
				arg = strings.TrimSpace(arg)
				if arg == "" {
					return nil, fmt.Errorf("service name is required")
				}
				return []string{"-u", arg, "-n", "100", "--no-pager"}, nil
			},
		},
		{
			Key:         "ss-tulpn",
			Title:       "ss -tulpn",
			Description: "Показать listening ports и процессы",
			NeedsArg:    false,
			BuildArgs: func(arg string) ([]string, error) {
				return []string{"-tulpn"}, nil
			},
		},
	}
}

func RunSafeCommand(key, arg string) (string, error) {
	var selected *SafeCommand

	commands := GetSafeCommands()
	for i := range commands {
		if commands[i].Key == key {
			selected = &commands[i]
			break
		}
	}

	if selected == nil {
		return "", fmt.Errorf("unknown command: %s", key)
	}

	args, err := selected.BuildArgs(arg)
	if err != nil {
		return "", err
	}

	bin := detectBinaryForCommand(key)

	output, err := runCmd(bin, args...)
	if err != nil {
		if len(output) == 0 {
			return "", err
		}
		return string(output), err
	}

	if len(output) == 0 {
		return "No output", nil
	}

	return string(output), nil
}

func detectBinaryForCommand(key string) string {
	switch {
	case strings.HasPrefix(key, "systemctl"):
		return "systemctl"
	case strings.HasPrefix(key, "docker"):
		return "docker"
	case strings.HasPrefix(key, "journalctl"):
		return "journalctl"
	case strings.HasPrefix(key, "ss-"):
		return "ss"
	default:
		return ""
	}
}