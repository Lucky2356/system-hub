package activity

import (
	"os"
	"path/filepath"
	"testing"
)

// reset returns the package to a clean in-memory state between tests.
func reset(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		mu.Lock()
		entries = nil
		filePath = ""
		mu.Unlock()
	})

	mu.Lock()
	entries = nil
	filePath = ""
	mu.Unlock()
}

func TestAddPersistsAndReloads(t *testing.T) {
	reset(t)
	path := filepath.Join(t.TempDir(), "activity.jsonl")

	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}

	Add("service", "restart", "nginx.service", "success", "")
	Add("docker", "stop", "web", "failed", "permission denied")

	// Simulate a restart: wipe memory and load from disk.
	mu.Lock()
	entries = nil
	mu.Unlock()

	if err := Init(path); err != nil {
		t.Fatalf("Init after restart: %v", err)
	}

	got := List()
	if len(got) != 2 {
		t.Fatalf("expected 2 entries after reload, got %d", len(got))
	}
	if got[0].Name != "nginx.service" || got[0].Action != "restart" {
		t.Errorf("unexpected first entry: %+v", got[0])
	}
	if got[1].Status != "failed" || got[1].Details != "permission denied" {
		t.Errorf("unexpected second entry: %+v", got[1])
	}
	if got[0].Time.IsZero() {
		t.Error("entry timestamp should round-trip")
	}
}

func TestInitWithMissingFileIsNotAnError(t *testing.T) {
	reset(t)
	path := filepath.Join(t.TempDir(), "does-not-exist.jsonl")

	if err := Init(path); err != nil {
		t.Errorf("a missing activity log must not be an error, got %v", err)
	}
	if len(List()) != 0 {
		t.Error("expected an empty log")
	}
}

func TestInitSkipsCorruptLines(t *testing.T) {
	reset(t)
	path := filepath.Join(t.TempDir(), "activity.jsonl")

	content := `{"time":"2026-01-01T10:00:00Z","target":"service","action":"start","name":"a","status":"success"}
this line is not json
{"time":"2026-01-01T11:00:00Z","target":"docker","action":"stop","name":"b","status":"success"}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// A damaged line must cost only that line, not the whole audit trail.
	got := List()
	if len(got) != 2 {
		t.Fatalf("expected the 2 readable entries, got %d: %+v", len(got), got)
	}
	if got[0].Name != "a" || got[1].Name != "b" {
		t.Errorf("unexpected entries: %+v", got)
	}
}

func TestAddTrimsToMaxSize(t *testing.T) {
	reset(t)
	path := filepath.Join(t.TempDir(), "activity.jsonl")

	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}

	for i := 0; i < maxSize+50; i++ {
		Add("service", "start", "svc", "success", "")
	}

	if got := len(List()); got != maxSize {
		t.Errorf("in-memory entries = %d, want %d", got, maxSize)
	}

	// The file must be rewritten on trim, not grow forever.
	mu.Lock()
	entries = nil
	mu.Unlock()
	if err := Init(path); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := len(List()); got != maxSize {
		t.Errorf("persisted entries = %d, want %d", got, maxSize)
	}
}

func TestClearRemovesFile(t *testing.T) {
	reset(t)
	path := filepath.Join(t.TempDir(), "activity.jsonl")

	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}
	Add("service", "start", "svc", "success", "")

	Clear()

	if len(List()) != 0 {
		t.Error("expected an empty log after Clear")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected the activity file to be removed, stat err = %v", err)
	}
}

func TestAddWithoutInitStaysInMemory(t *testing.T) {
	reset(t)

	// Init is never called: the app must still record actions, just not persist.
	Add("service", "start", "svc", "success", "")

	if len(List()) != 1 {
		t.Error("entries should be recorded in memory even without a file")
	}
}
