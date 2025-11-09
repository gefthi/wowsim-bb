package effects

import "time"

// Aura represents a timed buff/debuff with optional stacking behavior.
type Aura struct {
	Label     string
	Duration  time.Duration
	MaxStacks int

	stacks    int
	active    bool
	expiresAt time.Duration

	OnGain         func(a *Aura, now time.Duration)
	OnExpire       func(a *Aura, now time.Duration)
	OnStacksChange func(a *Aura, now time.Duration, oldStacks, newStacks int)
}

// NewAura returns a ready-to-use aura instance.
func NewAura(label string, duration time.Duration, maxStacks int) *Aura {
	return &Aura{
		Label:     label,
		Duration:  duration,
		MaxStacks: maxStacks,
	}
}

// Active reports whether the aura is currently active (ignores expiration checks).
func (a *Aura) Active() bool {
	return a != nil && a.active
}

// ActiveAt reports whether the aura is active and not expired at the provided time.
func (a *Aura) ActiveAt(now time.Duration) bool {
	if a == nil || !a.active {
		return false
	}
	if a.Duration <= 0 {
		return true
	}
	return now < a.expiresAt
}

// Stacks returns the current stack count.
func (a *Aura) Stacks() int {
	if a == nil {
		return 0
	}
	return a.stacks
}

// ExpiresAt returns the timestamp when the aura will expire.
func (a *Aura) ExpiresAt() time.Duration {
	if a == nil {
		return 0
	}
	return a.expiresAt
}

// Remaining returns remaining duration if active, zero otherwise.
func (a *Aura) Remaining(now time.Duration) time.Duration {
	if a == nil || !a.ActiveAt(now) {
		return 0
	}
	if a.Duration <= 0 {
		return 0
	}
	return a.expiresAt - now
}

// AddStacks increases the stack count and refreshes duration.
func (a *Aura) AddStacks(now time.Duration, delta int) {
	if a == nil || delta == 0 {
		return
	}
	a.ensureActive(now)
	a.refreshExpiry(now)
	old := a.stacks
	newStacks := old + delta
	if newStacks < 0 {
		newStacks = 0
	}
	if a.MaxStacks > 0 && newStacks > a.MaxStacks {
		newStacks = a.MaxStacks
	}
	a.stacks = newStacks
	if a.stacks == 0 {
		a.deactivate(now)
		return
	}
	if a.OnStacksChange != nil && newStacks != old {
		a.OnStacksChange(a, now, old, newStacks)
	}
}

// SetStacks sets the stack count directly and refreshes duration.
func (a *Aura) SetStacks(now time.Duration, stacks int) {
	if a == nil {
		return
	}
	if stacks <= 0 {
		a.deactivate(now)
		return
	}
	if a.MaxStacks > 0 && stacks > a.MaxStacks {
		stacks = a.MaxStacks
	}
	a.ensureActive(now)
	a.refreshExpiry(now)
	old := a.stacks
	a.stacks = stacks
	if a.OnStacksChange != nil && stacks != old {
		a.OnStacksChange(a, now, old, stacks)
	}
}

// Clear removes the aura and resets stacks to zero.
func (a *Aura) Clear(now time.Duration) {
	if a == nil {
		return
	}
	a.deactivate(now)
}

// CheckExpiration removes the aura if it has expired and returns true if it did.
func (a *Aura) CheckExpiration(now time.Duration) bool {
	if a == nil || !a.active || a.Duration <= 0 {
		return false
	}
	if now < a.expiresAt {
		return false
	}
	a.deactivate(now)
	return true
}

func (a *Aura) ensureActive(now time.Duration) {
	if a.active {
		return
	}
	a.active = true
	if a.OnGain != nil {
		a.OnGain(a, now)
	}
}

func (a *Aura) refreshExpiry(now time.Duration) {
	if a.Duration <= 0 {
		return
	}
	a.expiresAt = now + a.Duration
}

func (a *Aura) deactivate(now time.Duration) {
	if !a.active {
		return
	}
	a.active = false
	old := a.stacks
	a.stacks = 0
	a.expiresAt = 0
	if a.OnExpire != nil {
		a.OnExpire(a, now)
	}
	if a.OnStacksChange != nil && old != 0 {
		a.OnStacksChange(a, now, old, 0)
	}
}
