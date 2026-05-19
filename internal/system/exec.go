package system

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

const defaultTimeout = 30 * time.Second

func runCmd(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%s: timeout (%v)", name, defaultTimeout)
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
