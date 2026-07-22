package ui

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestRunOnceSkipsWhileInFlight(t *testing.T) {
	var running sync.WaitGroup
	running.Add(1)

	release := make(chan struct{})
	var starts atomic.Int32

	r := &autoRefresher{
		refresh: func() {
			starts.Add(1)
			running.Done() // signal that the first refresh has begun
			<-release      // and block until the test lets it finish
		},
	}

	// First tick starts a refresh and blocks it.
	r.runOnce()
	running.Wait()

	// Ticks that arrive while it is still running must be dropped, not stacked.
	for i := 0; i < 5; i++ {
		r.runOnce()
	}

	if got := starts.Load(); got != 1 {
		t.Fatalf("refresh started %d times while one was in flight; want 1", got)
	}

	// Let it finish, then a later tick runs normally again.
	close(release)
	waitFor(t, func() bool { return !r.inFlight.Load() })

	next := make(chan struct{})
	r.refresh = func() { close(next) }
	r.runOnce()

	select {
	case <-next:
	case <-time.After(time.Second):
		t.Fatal("a tick after the in-flight one finished did not run")
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met within a second")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSetTextIfChanged(t *testing.T) {
	test.NewApp()

	entry := widget.NewMultiLineEntry()
	entry.SetText("line 1\nline 2")

	// A no-op write is skipped: this is what keeps follow mode from resetting
	// the scroll position every second.
	setTextIfChanged(entry, "line 1\nline 2")
	if entry.Text != "line 1\nline 2" {
		t.Fatalf("unexpected text after no-op write: %q", entry.Text)
	}

	// A real change still goes through.
	setTextIfChanged(entry, "line 1\nline 2\nline 3")
	if entry.Text != "line 1\nline 2\nline 3" {
		t.Errorf("changed text was not written: %q", entry.Text)
	}
}
