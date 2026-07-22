package system

import (
	"os"
	"testing"
)

func TestKillProcessRejectsInvalidPIDs(t *testing.T) {
	for _, pid := range []int32{0, -1, 1} {
		if err := KillProcess(pid); err == nil {
			t.Errorf("KillProcess(%d) returned nil; it must refuse", pid)
		}
	}
}

func TestKillProcessRefusesToKillItself(t *testing.T) {
	self := int32(os.Getpid())

	err := KillProcess(self)
	if err == nil {
		t.Fatal("KillProcess killed the app's own process; a stray click would look like a crash")
	}
	// The UI branches on this to phrase the refusal, so it must be recognisable
	// rather than a raw OS error.
	if !IsSelfKill(err) {
		t.Errorf("IsSelfKill(%v) = false; the UI cannot tell this apart from a real failure", err)
	}
}
