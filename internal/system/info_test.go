package system

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{5 * 1024 * 1024 * 1024, "5.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}
	for _, tt := range tests {
		if got := FormatBytes(tt.in); got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		seconds uint64
		want    string
	}{
		{0, "0m"},
		{59, "0m"},
		{60, "1m"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{90000, "1d 1h 0m"},
	}
	for _, tt := range tests {
		if got := FormatUptime(tt.seconds); got != tt.want {
			t.Errorf("FormatUptime(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		value  float64
		places int
		want   float64
	}{
		{1.234, 1, 1.2},
		{1.25, 1, 1.3},
		{99.99, 1, 100.0},
		{0.0, 2, 0.0},
	}
	for _, tt := range tests {
		if got := round(tt.value, tt.places); got != tt.want {
			t.Errorf("round(%v, %d) = %v, want %v", tt.value, tt.places, got, tt.want)
		}
	}
}
