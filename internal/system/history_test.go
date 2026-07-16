package system

import (
	"testing"
	"time"
)

func TestHistoryKeepsSamplesInChronologicalOrder(t *testing.T) {
	h := &history{samples: make([]Sample, 3)}
	start := time.Now()

	for i := 1; i <= 2; i++ {
		h.push(Stats{CPUPercent: float64(i)}, start.Add(time.Duration(i)*time.Second))
	}

	got := h.snapshot()
	if len(got) != 2 {
		t.Fatalf("snapshot has %d samples, want 2", len(got))
	}
	if got[0].CPUPercent != 1 || got[1].CPUPercent != 2 {
		t.Errorf("samples out of order: %v", []float64{got[0].CPUPercent, got[1].CPUPercent})
	}
}

func TestHistoryDropsOldestOnceFull(t *testing.T) {
	h := &history{samples: make([]Sample, 3)}
	start := time.Now()

	// Five samples into a buffer of three: the ring must wrap and still read
	// oldest-first, which is the part that is easy to get wrong.
	for i := 1; i <= 5; i++ {
		h.push(Stats{CPUPercent: float64(i)}, start.Add(time.Duration(i)*time.Second))
	}

	got := h.snapshot()
	if len(got) != 3 {
		t.Fatalf("snapshot has %d samples, want 3", len(got))
	}

	want := []float64{3, 4, 5}
	for i, w := range want {
		if got[i].CPUPercent != w {
			t.Errorf("sample %d = %v, want %v (full order: %v)", i, got[i].CPUPercent, w, cpuSeries(got))
		}
	}
}

func TestHistoryTurnsNetworkCountersIntoRates(t *testing.T) {
	h := &history{samples: make([]Sample, 4)}
	start := time.Now()

	h.push(Stats{NetRecv: 1000, NetSent: 500}, start)
	h.push(Stats{NetRecv: 3000, NetSent: 1500}, start.Add(2*time.Second))

	got := h.snapshot()

	// The first sample has no predecessor, so it has no rate.
	if got[0].NetRecvPerSec != 0 || got[0].NetSentPerSec != 0 {
		t.Errorf("first sample should have no rate, got rx=%v tx=%v", got[0].NetRecvPerSec, got[0].NetSentPerSec)
	}

	// 2000 bytes over 2 seconds = 1000 B/s.
	if got[1].NetRecvPerSec != 1000 {
		t.Errorf("NetRecvPerSec = %v, want 1000", got[1].NetRecvPerSec)
	}
	if got[1].NetSentPerSec != 500 {
		t.Errorf("NetSentPerSec = %v, want 500", got[1].NetSentPerSec)
	}
}

func TestHistoryIgnoresCounterResets(t *testing.T) {
	h := &history{samples: make([]Sample, 4)}
	start := time.Now()

	h.push(Stats{NetRecv: 5000, NetSent: 5000}, start)
	// An interface disappearing or a resume from sleep rewinds the counter. The
	// unsigned subtraction would wrap into an enormous rate and flatten the
	// whole chart.
	h.push(Stats{NetRecv: 10, NetSent: 10}, start.Add(2*time.Second))

	got := h.snapshot()
	if got[1].NetRecvPerSec != 0 || got[1].NetSentPerSec != 0 {
		t.Errorf("a counter reset must not produce a rate, got rx=%v tx=%v",
			got[1].NetRecvPerSec, got[1].NetSentPerSec)
	}
}

func TestHistorySnapshotIsACopy(t *testing.T) {
	h := &history{samples: make([]Sample, 3)}
	h.push(Stats{CPUPercent: 42}, time.Now())

	got := h.snapshot()
	got[0].CPUPercent = 99

	if again := h.snapshot(); again[0].CPUPercent != 42 {
		t.Errorf("callers can mutate the buffer through the snapshot: %v", again[0].CPUPercent)
	}
}

func cpuSeries(samples []Sample) []float64 {
	out := make([]float64, len(samples))
	for i, s := range samples {
		out[i] = s.CPUPercent
	}
	return out
}
