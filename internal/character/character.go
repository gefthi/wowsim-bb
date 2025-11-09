package character

import "time"

// Stats represents character statistics
type Stats struct {
	SpellPower float64
	CritPct    float64 // Percentage (e.g., 25.5 for 25.5%)
	HastePct   float64 // Percentage
	Spirit     float64
	HitPct     float64 // Percentage
	MaxMana    float64
}

// Resources tracks current resources
type Resources struct {
	CurrentMana float64
}

// Buff represents an active buff
type Buff struct {
	Active    bool
	ExpiresAt time.Duration
	Charges   int // For Backdraft
}

// Debuff represents an active debuff on target
type Debuff struct {
	Active            bool
	ExpiresAt         time.Duration
	TickInterval      time.Duration
	LastTick          time.Duration
	TickDamage        float64
	TickCritChance    float64
	TicksRemaining    int
	SnapshotDotDamage float64
}

// Cooldown tracks spell cooldowns
type Cooldown struct {
	ReadyAt time.Duration
}

// Character represents the player character
type Character struct {
	Stats     Stats
	Resources Resources

	// Buffs
	Backdraft         Buff // Not implemented in Phase 1
	Pyroclasm         Buff // Phase 2: +6% fire/shadow damage
	ImprovedSoulLeech Buff // Phase 2: Mana regen over time

	// Debuffs on target
	Immolate Debuff

	// Cooldowns
	ChaosBolt   Cooldown
	Conflagrate Cooldown

	// GCD
	GCDReadyAt time.Duration

	// Combat state
	CurrentTime time.Duration
	IsCasting   bool
	CastEndsAt  time.Duration

	// Soul Leech tracking (for HoT ticks)
	SoulLeechLastTick time.Duration
}

// NewCharacter creates a new character with given stats
func NewCharacter(stats Stats) *Character {
	return &Character{
		Stats: stats,
		Resources: Resources{
			CurrentMana: stats.MaxMana,
		},
	}
}

// IsGCDReady checks if GCD is ready
func (c *Character) IsGCDReady() bool {
	return c.CurrentTime >= c.GCDReadyAt
}

// IsCooldownReady checks if a specific cooldown is ready
func (c *Character) IsCooldownReady(cd *Cooldown) bool {
	return c.CurrentTime >= cd.ReadyAt
}

// HasMana checks if character has enough mana for a spell
func (c *Character) HasMana(cost float64) bool {
	return c.Resources.CurrentMana >= cost
}

// SpendMana deducts mana for a spell cast
func (c *Character) SpendMana(cost float64) {
	c.Resources.CurrentMana -= cost
	if c.Resources.CurrentMana < 0 {
		c.Resources.CurrentMana = 0
	}
}

// GainMana adds mana (from Life Tap)
func (c *Character) GainMana(amount float64) {
	c.Resources.CurrentMana += amount
	if c.Resources.CurrentMana > c.Stats.MaxMana {
		c.Resources.CurrentMana = c.Stats.MaxMana
	}
}

// AdvanceTime moves simulation time forward
func (c *Character) AdvanceTime(duration time.Duration) {
	c.CurrentTime += duration
}
