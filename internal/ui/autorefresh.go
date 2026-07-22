// Package ui builds the Fyne widgets and tabs. It depends on internal/system
// for all host access; the dependency never points the other way.
package ui

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Lucky2356/system-hub/internal/appstate"
)

// autoRefresher periodically invokes a refresh function. It replaces the
// hand-rolled ticker/stop-channel blocks that were duplicated across tabs, and
// fixes two problems those had: stopping twice panicked with "close of closed
// channel", and the interval was captured once so changing it in Settings had
// no effect until restart.
//
// Start/Stop are idempotent and safe to call from any goroutine.
type autoRefresher struct {
	mu       sync.Mutex
	stop     chan struct{}
	refresh  func()
	interval func() time.Duration

	// inFlight drops a tick whose predecessor has not finished. A dashboard
	// heavy tick lists 315 services, runs docker ps and docker stats (~1s on its
	// own) and scans every process; without this, a slow tick would let the next
	// one start on top of it and the goroutines would pile up. Only the log tab
	// guarded against this before, by hand.
	inFlight atomic.Bool
}

// newAutoRefresher refreshes on the interval configured in Settings, re-read on
// every tick so changes apply live.
func newAutoRefresher(refresh func()) *autoRefresher {
	return &autoRefresher{refresh: refresh, interval: configInterval}
}

// newFixedRefresher refreshes on a constant interval (e.g. log follow mode).
func newFixedRefresher(d time.Duration, refresh func()) *autoRefresher {
	return &autoRefresher{refresh: refresh, interval: func() time.Duration { return d }}
}

func configInterval() time.Duration {
	seconds := appstate.GetConfig().RefreshIntervalSeconds
	if seconds <= 0 {
		seconds = 2
	}
	return time.Duration(seconds) * time.Second
}

func (a *autoRefresher) Start() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stop != nil {
		return // already running
	}

	stop := make(chan struct{})
	a.stop = stop
	go a.loop(stop)
}

func (a *autoRefresher) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stop == nil {
		return // not running
	}
	close(a.stop)
	a.stop = nil
}

// SetEnabled matches the signature of widget.Check.OnChanged.
func (a *autoRefresher) SetEnabled(enabled bool) {
	if enabled {
		a.Start()
	} else {
		a.Stop()
	}
}

func (a *autoRefresher) loop(stop chan struct{}) {
	current := a.interval()
	ticker := time.NewTicker(current)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if next := a.interval(); next != current {
				current = next
				ticker.Reset(next)
			}
			a.runOnce()
		case <-stop:
			return
		}
	}
}

// runOnce refreshes unless the previous refresh is still running, in which case
// the tick is skipped. Skipping is correct here: every refresh reads the full
// current state, so a dropped tick loses nothing — the next one shows the same
// thing a moment later.
func (a *autoRefresher) runOnce() {
	if !a.inFlight.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer a.inFlight.Store(false)
		a.refresh()
	}()
}
