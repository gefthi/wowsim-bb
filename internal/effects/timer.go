package effects

import "time"

// Timer tracks a ready-at timestamp for cooldown-style mechanics.
type Timer struct {
	readyAt time.Duration
}

// Ready returns true if the timer is ready at the provided time.
func (t *Timer) Ready(now time.Duration) bool {
	return now >= t.readyAt
}

// Remaining returns the remaining duration until the timer is ready.
func (t *Timer) Remaining(now time.Duration) time.Duration {
	if now >= t.readyAt {
		return 0
	}
	return t.readyAt - now
}

// Reset sets the timer to become ready after the provided cooldown duration.
func (t *Timer) Reset(now time.Duration, cooldown time.Duration) {
	t.readyAt = now + cooldown
}

// ForceReady immediately marks the timer ready at the supplied time.
func (t *Timer) ForceReady(now time.Duration) {
	t.readyAt = now
}

// ReadyAt returns the current ready timestamp.
func (t *Timer) ReadyAt() time.Duration {
	return t.readyAt
}
