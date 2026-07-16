// Package activity records the service and container actions taken in the app.
//
// The log is an audit trail of state-changing operations (start/stop/restart of
// services and containers), so it is persisted to disk rather than kept only in
// memory: "what did I restart yesterday?" is exactly the question it answers.
package activity

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Entry struct {
	Time    time.Time `json:"time"`
	Target  string    `json:"target"`
	Action  string    `json:"action"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
	Details string    `json:"details,omitempty"`
}

const maxSize = 300

var (
	mu       sync.RWMutex
	entries  []Entry
	filePath string // empty until Init; the log then stays in memory only
)

// Init loads the persisted log from path and enables writing new entries to it.
// A missing or corrupt file is not fatal: the app keeps working with whatever
// could be read, since an audit log must never block system management.
func Init(path string) error {
	mu.Lock()
	defer mu.Unlock()

	filePath = path

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	loaded := make([]Entry, 0, maxSize)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip a damaged line rather than lose the whole log
		}
		loaded = append(loaded, e)
	}

	if len(loaded) > maxSize {
		loaded = loaded[len(loaded)-maxSize:]
	}
	entries = loaded

	return scanner.Err()
}

func Add(target, action, name, status, details string) {
	mu.Lock()
	defer mu.Unlock()

	entry := Entry{
		Time:    time.Now(),
		Target:  target,
		Action:  action,
		Name:    name,
		Status:  status,
		Details: details,
	}
	entries = append(entries, entry)

	trimmed := false
	if len(entries) > maxSize {
		entries = entries[len(entries)-maxSize:]
		trimmed = true
	}

	if filePath == "" {
		return
	}

	// Rewrite the whole file once it is trimmed so it cannot grow without
	// bound; otherwise just append the new line.
	if trimmed {
		writeAllLocked()
		return
	}
	appendLocked(entry)
}

func List() []Entry {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]Entry, len(entries))
	copy(result, entries)
	return result
}

func Clear() {
	mu.Lock()
	defer mu.Unlock()

	entries = nil
	if filePath != "" {
		_ = os.Remove(filePath)
	}
}

// appendLocked writes one entry; callers must hold mu.
func appendLocked(e Entry) {
	data, err := json.Marshal(e)
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	_, _ = f.Write(append(data, '\n'))
}

// writeAllLocked rewrites the file from the in-memory entries; callers must
// hold mu.
func writeAllLocked() {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return
	}

	tmp := filePath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return
	}

	w := bufio.NewWriter(f)
	for _, e := range entries {
		data, err := json.Marshal(e)
		if err != nil {
			continue
		}
		if _, err := w.Write(append(data, '\n')); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return
		}
	}

	if err := w.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return
	}

	// Rename is atomic, so a crash cannot leave a half-written log.
	_ = os.Rename(tmp, filePath)
}
