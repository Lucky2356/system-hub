package system

import (
	"context"
	"testing"
	"time"
)

// recordedCall captures one invocation made through the command runner.
type recordedCall struct {
	name    string
	args    []string
	timeout time.Duration
}

// fakeRunner installs a stub command runner for the duration of the test and
// returns the slice that collects the calls. Output/err define what the stub
// returns for every call.
func fakeRunner(t *testing.T, output []byte, err error) *[]recordedCall {
	t.Helper()

	calls := &[]recordedCall{}
	original := runner

	runner = func(_ context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
		*calls = append(*calls, recordedCall{
			name:    name,
			args:    append([]string(nil), args...),
			timeout: timeout,
		})
		return output, err
	}

	t.Cleanup(func() {
		runner = original
		// Listings are cached; drop them so tests do not leak state into
		// each other.
		InvalidateServicesCache()
		InvalidateContainersCache()
	})

	return calls
}

// argvEquals compares a recorded call against an expected command line.
func argvEquals(got recordedCall, name string, args ...string) bool {
	if got.name != name || len(got.args) != len(args) {
		return false
	}
	for i := range args {
		if got.args[i] != args[i] {
			return false
		}
	}
	return true
}
