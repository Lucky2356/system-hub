//go:build linux

package system

import "github.com/Lucky2356/system-hub/internal/i18n"

// platformCommands are the systemd/journald diagnostics available on Linux.
// The list is rebuilt on every call, so a language switch reaches it.
func platformCommands() []SafeCommand {
	return []SafeCommand{
		{
			Key:         "systemctl-list-services",
			Title:       "systemctl list-units --type=service",
			Description: i18n.T("Show the list of systemd services"),
			Binary:      "systemctl",
			BuildArgs: func(string) ([]string, error) {
				return []string{"list-units", "--type=service", "--all", "--no-pager"}, nil
			},
		},
		{
			Key:         "systemctl-status",
			Title:       "systemctl status " + i18n.T("<service>"),
			Description: i18n.T("Show the status of a single systemd service"),
			NeedsArg:    true,
			ArgHint:     i18n.Tf("For example: %s", "nginx.service"),
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
			Description: i18n.T("Show the latest system logs"),
			Binary:      "journalctl",
			BuildArgs: func(string) ([]string, error) {
				return []string{"-n", "100", "--no-pager"}, nil
			},
		},
		{
			Key:         "journalctl-service",
			Title:       "journalctl -u " + i18n.T("<service>") + " -n 100",
			Description: i18n.T("Show the latest logs for a specific service"),
			NeedsArg:    true,
			ArgHint:     i18n.Tf("For example: %s", "docker.service"),
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
			Description: i18n.T("Show listening ports and processes"),
			Binary:      "ss",
			BuildArgs: func(string) ([]string, error) {
				return []string{"-tulpn"}, nil
			},
		},
	}
}
