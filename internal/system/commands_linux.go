//go:build linux

package system

// platformCommands are the systemd/journald diagnostics available on Linux.
func platformCommands() []SafeCommand {
	return []SafeCommand{
		{
			Key:         "systemctl-list-services",
			Title:       "systemctl list-units --type=service",
			Description: "Показать список systemd-сервисов",
			Binary:      "systemctl",
			BuildArgs: func(string) ([]string, error) {
				return []string{"list-units", "--type=service", "--all", "--no-pager"}, nil
			},
		},
		{
			Key:         "systemctl-status",
			Title:       "systemctl status <сервис>",
			Description: "Показать статус одного systemd-сервиса",
			NeedsArg:    true,
			ArgHint:     "Например: nginx.service",
			Binary:      "systemctl",
			BuildArgs: func(arg string) ([]string, error) {
				name, err := requireServiceArg(arg)
				if err != nil {
					return nil, err
				}
				return []string{"status", "--no-pager", "--", name}, nil
			},
		},
		{
			Key:         "journalctl-tail",
			Title:       "journalctl -n 100",
			Description: "Показать последние системные логи",
			Binary:      "journalctl",
			BuildArgs: func(string) ([]string, error) {
				return []string{"-n", "100", "--no-pager"}, nil
			},
		},
		{
			Key:         "journalctl-service",
			Title:       "journalctl -u <сервис> -n 100",
			Description: "Показать последние логи конкретного сервиса",
			NeedsArg:    true,
			ArgHint:     "Например: docker.service",
			Binary:      "journalctl",
			BuildArgs: func(arg string) ([]string, error) {
				name, err := requireServiceArg(arg)
				if err != nil {
					return nil, err
				}
				return []string{"-u", name, "-n", "100", "--no-pager"}, nil
			},
		},
		{
			Key:         "ss-tulpn",
			Title:       "ss -tulpn",
			Description: "Показать слушающие порты и процессы",
			Binary:      "ss",
			BuildArgs: func(string) ([]string, error) {
				return []string{"-tulpn"}, nil
			},
		},
	}
}
