package activity

import (
	"sync"
	"time"
)

type Entry struct {
	Time     time.Time
	Target   string
	Action   string
	Name     string
	Status   string
	Details  string
}

var (
	mu      sync.RWMutex
	entries []Entry
	maxSize = 300
)

func Add(target, action, name, status, details string) {
	mu.Lock()
	defer mu.Unlock()

	entries = append(entries, Entry{
		Time:    time.Now(),
		Target:  target,
		Action:  action,
		Name:    name,
		Status:  status,
		Details: details,
	})

	if len(entries) > maxSize {
		entries = entries[len(entries)-maxSize:]
	}
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
}