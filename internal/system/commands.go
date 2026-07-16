package system

import (
	"fmt"
	"strings"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

// SafeCommand is a read-only diagnostic command the user can run from the
// Commands tab. Arguments are built by BuildArgs rather than typed freely, so
// no user input can turn into an extra flag or a shell fragment.
type SafeCommand struct {
	Key         string
	Title       string
	Description string
	NeedsArg    bool
	ArgHint     string
	Binary      string
	BuildArgs   func(arg string) ([]string, error)
}

// GetSafeCommands returns the commands available on this platform. The list is
// platform-specific: offering systemctl and journalctl on Windows only produced
// "not found" errors.
func GetSafeCommands() []SafeCommand {
	return append(platformCommands(), dockerCommands()...)
}

// dockerCommands work anywhere Docker is installed.
func dockerCommands() []SafeCommand {
	return []SafeCommand{
		{
			Key:         "docker-ps",
			Title:       "docker ps -a",
			Description: i18n.T("Show all Docker containers"),
			Binary:      "docker",
			BuildArgs: func(string) ([]string, error) {
				return []string{"ps", "-a"}, nil
			},
		},
		{
			Key:         "docker-images",
			Title:       "docker images",
			Description: i18n.T("Show Docker images"),
			Binary:      "docker",
			BuildArgs: func(string) ([]string, error) {
				return []string{"images"}, nil
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

	output, err := runCmd(selected.Binary, args...)
	if err != nil {
		if len(output) == 0 {
			return "", err
		}
		return string(output), err
	}

	if len(output) == 0 {
		return i18n.T("No output"), nil
	}

	return string(output), nil
}

// requireServiceArg validates a user-supplied service name for the commands
// that take one.
func requireServiceArg(arg string) (string, error) {
	arg = strings.TrimSpace(arg)
	if err := validateName("service", arg); err != nil {
		return "", err
	}
	return arg, nil
}
