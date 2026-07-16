package system

import (
	"sync"
	"time"
)

// HistoryCapacity is how many samples are kept. At the default 2s refresh this
// is about five minutes, which is enough to see a spike that has just passed
// without turning the dashboard into a time-series database.
const HistoryCapacity = 150

// Sample is one point of dashboard history.
//
// Network is stored as a rate rather than the raw counters gopsutil reports:
// those are monotonic totals since boot, so plotting them directly would draw a
// straight line that says nothing about current traffic.
type Sample struct {
	At            time.Time
	CPUPercent    float64
	RAMPercent    float64
	NetRecvPerSec float64
	NetSentPerSec float64
}

// history is a ring buffer of recent samples. It is filled from the dashboard's
// light tick, which already has the numbers, so recording history costs no
// extra system calls.
type history struct {
	mu      sync.Mutex
	samples []Sample
	next    int
	filled  bool

	// The previous raw counters, needed to turn totals into a rate.
	lastAt   time.Time
	lastRecv uint64
	lastSent uint64
	hasLast  bool
}

var metricHistory = &history{samples: make([]Sample, HistoryCapacity)}

// RecordSample adds one point to the metric history.
func RecordSample(stats Stats) {
	metricHistory.push(stats, time.Now())
}

// MetricHistory returns the samples in chronological order, oldest first.
func MetricHistory() []Sample {
	return metricHistory.snapshot()
}

// ResetMetricHistory drops all samples. It exists for tests.
func ResetMetricHistory() {
	metricHistory.reset()
}

func (h *history) push(stats Stats, now time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sample := Sample{
		At:         now,
		CPUPercent: stats.CPUPercent,
		RAMPercent: stats.RAMPercent,
	}

	// The first sample has no predecessor, so it has no rate yet; leaving it at
	// zero is honest, and it scrolls out of the window within seconds.
	if h.hasLast {
		if elapsed := now.Sub(h.lastAt).Seconds(); elapsed > 0 {
			// Counters reset when an interface goes away or the machine wakes
			// from sleep. A negative delta would render as a huge downward
			// spike, so treat it as "no data" instead.
			if stats.NetRecv >= h.lastRecv {
				sample.NetRecvPerSec = float64(stats.NetRecv-h.lastRecv) / elapsed
			}
			if stats.NetSent >= h.lastSent {
				sample.NetSentPerSec = float64(stats.NetSent-h.lastSent) / elapsed
			}
		}
	}

	h.lastAt, h.lastRecv, h.lastSent, h.hasLast = now, stats.NetRecv, stats.NetSent, true

	h.samples[h.next] = sample
	h.next = (h.next + 1) % len(h.samples)
	if h.next == 0 {
		h.filled = true
	}
}

func (h *history) snapshot() []Sample {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.filled {
		out := make([]Sample, h.next)
		copy(out, h.samples[:h.next])
		return out
	}

	// Once wrapped, the oldest sample is the one about to be overwritten.
	out := make([]Sample, 0, len(h.samples))
	out = append(out, h.samples[h.next:]...)
	out = append(out, h.samples[:h.next]...)
	return out
}

func (h *history) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.next = 0
	h.filled = false
	h.hasLast = false
}
