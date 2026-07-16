package system

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const defaultTimeout = 30 * time.Second

// binPathCache memoises absolute paths resolved via exec.LookPath so external
// tools are always invoked by their absolute path. This avoids PATH-hijacking,
// where a malicious binary earlier in $PATH (or the current directory on
// Windows) would otherwise be executed instead of the intended system tool.
var binPathCache sync.Map // map[string]string

func resolveBinary(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty command name")
	}
	if v, ok := binPathCache.Load(name); ok {
		return v.(string), nil
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	binPathCache.Store(name, path)
	return path, nil
}

func runCmd(name string, args ...string) ([]byte, error) {
	return runCmdContext(context.Background(), defaultTimeout, name, args...)
}

// runCmdContext runs an external command with a bounded timeout. A timeout of 0
// means no deadline (used for long-running operations such as `docker pull`).
func runCmdContext(parent context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
	binPath, err := resolveBinary(name)
	if err != nil {
		return nil, err
	}

	ctx := parent
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, binPath, args...)
	hideConsoleWindow(cmd)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if ctx.Err() != nil && timeout > 0 {
			return nil, fmt.Errorf("%s: timeout (%v)", name, timeout)
		}
		text := trimOutput(stderr.String())
		if text == "" {
			text = trimOutput(out.String())
		}
		if text == "" {
			return nil, fmt.Errorf("%s %v: %w", name, args, err)
		}
		return nil, fmt.Errorf("%s %v: %s", name, args, text)
	}

	return bytes.TrimSpace(out.Bytes()), nil
}

func trimOutput(s string) string {
	if len(s) > 500 {
		s = s[:500] + "..."
	}
	return s
}

// validateName guards against argument injection: a user-supplied service,
// container or image name that begins with "-" could otherwise be interpreted
// as a command-line flag by systemctl/journalctl/docker. It also rejects
// control characters and internal whitespace.
func validateName(kind, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%s name is empty", kind)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid %s name %q: must not start with '-'", kind, name)
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("invalid %s name: contains control characters", kind)
		}
		if r == ' ' || r == '\t' {
			return fmt.Errorf("invalid %s name %q: must not contain whitespace", kind, name)
		}
	}
	return nil
}
