package ui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// notifyCooldown is how long the same problem stays quiet after being
// announced. A flapping service would otherwise raise a desktop notification on
// every dashboard refresh.
const notifyCooldown = 10 * time.Minute

var (
	notifyMu   sync.Mutex
	notifiedAt = map[string]time.Time{}
)

// notifyProblems raises a desktop notification for each problem that has not
// been announced recently, and forgets problems that have been resolved so they
// notify again if they come back.
func notifyProblems(problems []string) {
	notifyMu.Lock()
	defer notifyMu.Unlock()

	now := time.Now()
	current := make(map[string]struct{}, len(problems))

	for _, p := range problems {
		current[p] = struct{}{}

		if last, seen := notifiedAt[p]; seen && now.Sub(last) < notifyCooldown {
			continue
		}

		notifiedAt[p] = now
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "System Hub: предупреждение",
			Content: p,
		})
	}

	// Drop resolved problems: if the same one reappears later it is news again.
	for p := range notifiedAt {
		if _, still := current[p]; !still {
			delete(notifiedAt, p)
		}
	}
}
