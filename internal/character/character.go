package character

import (
	"time"

	"wotlk-destro-sim/internal/effects"
	"wotlk-destro-sim/internal/runes"
)

// Stats represents character statistics
type Stats struct {
	Intellect  float64
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
	Value     float64
}

// Debuff represents an active debuff on target
type Debuff struct {
	Active            bool
	ExpiresAt         time.Duration
	TickInterval      time.Duration
	LastTick          time.Duration
	TickDamage        float64
	BaseTickDamage    float64
	SPTickDamage      float64
	TickCritChance    float64
	TicksRemaining    int
	TotalTicks        int
	SnapshotDotDamage float64
	TickHandle        EventHandle
}

// EventHandle allows simulation systems to cancel scheduled events without
// depending on the underlying scheduler implementation.
type EventHandle interface {
	Cancel()
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
	EmpoweredImp      Buff
	LifeTapBuff       Buff // Glyph of Life Tap bonus
	CursedShadows     Buff
	ShadowTrance      Buff
	CataclysmicBurst  *effects.Aura
	InnerFlame        struct {
		Active bool
	}
	HeatingUp          *effects.Aura
	DecisiveDecimation struct {
		Active bool
	}
	ChaosManifesting struct {
		FireExpiresAt   time.Duration
		ShadowExpiresAt time.Duration
	}
	GuldansChosen *effects.Aura

	// Debuffs on target
	Immolate        Debuff
	Corruption      Debuff
	CurseOfAgony    Debuff
	CurseOfElements Debuff

	// Cooldowns
	ChaosBolt   Cooldown
	Conflagrate Cooldown
	Shadowburn  Cooldown
	Shadowfury  Cooldown

	// GCD
	GCD effects.Timer

	// Combat state
	CurrentTime time.Duration
	IsCasting   bool
	CastEndsAt  time.Duration

	// Soul Leech tracking (for HoT ticks)
	SoulLeechLastTick time.Duration

	// Mystic Enchants buffs
	PureShadow   *effects.Aura
	DuskTillDawn *effects.Aura

	// Nightfall tracking
	ShadowTranceFreeCast      bool
	ShadowTranceLeechFraction float64
	NightfallStacks           int
}

// NewCharacter creates a new character with given stats
func NewCharacter(stats Stats) *Character {
	char := &Character{
		Stats: stats,
		Resources: Resources{
			CurrentMana: stats.MaxMana,
		},
	}
	heatingDuration := time.Duration(runes.HeatingUpDurationSec * float64(time.Second))
	char.HeatingUp = effects.NewAura("Heating Up", heatingDuration, runes.HeatingUpMaxStacks)
	gcDuration := time.Duration(runes.GuldansChosenDurationSec * float64(time.Second))
	char.GuldansChosen = effects.NewAura("Gul'dan's Chosen", gcDuration, 1)
	char.CataclysmicBurst = effects.NewAura("Cataclysmic Burst", 0, runes.CataclysmicBurstMaxStacks)
	char.PureShadow = effects.NewAura("Pure Shadow", time.Duration(runes.PureShadowDurationSec*float64(time.Second)), runes.PureShadowMaxStacks)
	char.DuskTillDawn = effects.NewAura("Dusk till Dawn", time.Duration(runes.DuskTillDawnDurationSec*float64(time.Second)), runes.DuskTillDawnMaxStacks)
	char.GCD.ForceReady(0)
	return char
}

// IsGCDReady checks if GCD is ready
func (c *Character) IsGCDReady() bool {
	return c.GCD.Ready(c.CurrentTime)
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
