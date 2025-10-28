package spells

import (
	"math/rand"
	"time"
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
)

// SpellType identifies different spells
type SpellType int

const (
	SpellImmolate SpellType = iota
	SpellIncinerate
	SpellChaosBolt
	SpellConflagrate
	SpellLifeTap
)

// CastResult represents the result of a spell cast
type CastResult struct {
	Spell      SpellType
	Damage     float64
	DidHit     bool
	DidCrit    bool
	ManaSpent  float64
	ManaGained float64
	CastTime   time.Duration
	GCDTime    time.Duration
}

// Engine handles spell casting and damage calculation
type Engine struct {
	Config *config.Config
	Rng    *rand.Rand

	// Target type for hit calculation
	IsBossTarget bool
}

// NewEngine creates a new spell engine
func NewEngine(cfg *config.Config, seed int64, isBoss bool) *Engine {
	return &Engine{
		Config:       cfg,
		Rng:          rand.New(rand.NewSource(seed)),
		IsBossTarget: isBoss,
	}
}

// RollHit determines if a spell hits
func (e *Engine) RollHit(char *character.Character) bool {
	hitCap := float64(e.Config.Constants.HitMechanics.EqualLevelMissChance)
	if e.IsBossTarget {
		hitCap = float64(e.Config.Constants.HitMechanics.BossHitCap)
	}

	// Calculate actual miss chance
	missChance := hitCap - char.Stats.HitPct
	if missChance <= 0 {
		return true // Hit capped
	}

	// Roll for miss
	roll := e.Rng.Float64() * 100.0
	return roll >= missChance // Hit if roll is above miss chance
}

// RollCrit determines if a spell crits
func (e *Engine) RollCrit(char *character.Character, bonusCrit float64) bool {
	// Base crit from stats
	totalCrit := char.Stats.CritPct

	// Add talent bonuses (Devastation + Backlash)
	totalCrit += e.Config.Talents.Devastation.CritBonus * 100.0 // Convert to percentage
	totalCrit += e.Config.Talents.Backlash.CritBonus * 100.0

	// Add any spell-specific bonus (e.g., Conflagrate from Fire and Brimstone)
	totalCrit += bonusCrit * 100.0

	roll := e.Rng.Float64() * 100.0
	return roll < totalCrit
}

// CalculateSpellDamage calculates base damage with spell power
func (e *Engine) CalculateSpellDamage(baseDamage, spCoefficient float64, char *character.Character) float64 {
	// Base damage + (SP * coefficient)
	damage := baseDamage + (char.Stats.SpellPower * spCoefficient)

	// Apply Emberstorm (15% fire/shadow damage)
	damage *= e.Config.Talents.Emberstorm.DamageMultiplier

	return damage
}

// ApplyFireAndBrimstone applies damage bonus if Immolate is on target
func (e *Engine) ApplyFireAndBrimstone(damage float64, char *character.Character) float64 {
	if char.Immolate.Active {
		return damage * e.Config.Talents.FireAndBrimstone.ImmolateTargetDamage
	}
	return damage
}

// CastImmolate casts Immolate
func (e *Engine) CastImmolate(char *character.Character) CastResult {
	spellData := e.Config.Spells.Immolate

	result := CastResult{
		Spell:     SpellImmolate,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	// Spend mana
	char.SpendMana(spellData.ManaCost)

	// Roll hit
	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// Calculate direct damage
	directDamage := e.CalculateSpellDamage(spellData.DirectDamage, spellData.SPCoefficientDirect, char)

	// Apply Improved Immolate (+30% all damage)
	directDamage *= e.Config.Talents.ImprovedImmolate.DamageMultiplier

	// Calculate DoT damage (full duration)
	dotDamage := e.CalculateSpellDamage(spellData.DotDamage, spellData.SPCoefficientDot, char)

	// Apply Improved Immolate to DoT
	dotDamage *= e.Config.Talents.ImprovedImmolate.DamageMultiplier

	// Apply Aftermath (+6% DoT only)
	dotDamage *= e.Config.Talents.Aftermath.DotDamageMultiplier

	totalDamage := directDamage + dotDamage

	// Roll crit
	if e.RollCrit(char, 0) {
		result.DidCrit = true
		totalDamage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = totalDamage

	// Apply Immolate debuff
	char.Immolate.Active = true
	char.Immolate.ExpiresAt = char.CurrentTime + time.Duration(spellData.DotDuration*float64(time.Second))

	return result
}

// CastIncinerate casts Incinerate
func (e *Engine) CastIncinerate(char *character.Character) CastResult {
	spellData := e.Config.Spells.Incinerate

	result := CastResult{
		Spell:     SpellIncinerate,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// Random base damage
	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)

	// Add Immolate bonus if active
	if char.Immolate.Active {
		immolateBonus := spellData.ImmolateBonusMin + e.Rng.Float64()*(spellData.ImmolateBonusMax-spellData.ImmolateBonusMin)
		baseDamage += immolateBonus
	}

	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)

	// Apply Fire and Brimstone if Immolate is up
	damage = e.ApplyFireAndBrimstone(damage, char)

	// Roll crit
	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage
	return result
}

// CastChaosBolt casts Chaos Bolt
func (e *Engine) CastChaosBolt(char *character.Character) CastResult {
	spellData := e.Config.Spells.ChaosBolt

	result := CastResult{
		Spell:     SpellChaosBolt,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// Random base damage
	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)

	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)

	// Apply Fire and Brimstone
	damage = e.ApplyFireAndBrimstone(damage, char)

	// Roll crit
	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage

	// Set cooldown
	char.ChaosBolt.ReadyAt = char.CurrentTime + time.Duration(spellData.Cooldown*float64(time.Second))

	return result
}

// CastConflagrate casts Conflagrate
func (e *Engine) CastConflagrate(char *character.Character) CastResult {
	spellData := e.Config.Spells.Conflagrate

	result := CastResult{
		Spell:     SpellConflagrate,
		CastTime:  0, // Instant
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// Base damage is 60% of Immolate's DoT
	immolateSpellData := e.Config.Spells.Immolate
	immolateDotDamage := e.CalculateSpellDamage(immolateSpellData.DotDamage, immolateSpellData.SPCoefficientDot, char)
	immolateDotDamage *= e.Config.Talents.ImprovedImmolate.DamageMultiplier
	immolateDotDamage *= e.Config.Talents.Aftermath.DotDamageMultiplier

	baseDamage := immolateDotDamage * spellData.ImmolateDotPercentage

	// Apply Emberstorm
	baseDamage *= e.Config.Talents.Emberstorm.DamageMultiplier

	// Conflagrate has +25% crit from Fire and Brimstone
	bonusCrit := e.Config.Talents.FireAndBrimstone.ConflagrateCritBonus

	// Roll crit
	if e.RollCrit(char, bonusCrit) {
		result.DidCrit = true
		baseDamage *= e.Config.Talents.Ruin.CritMultiplier
	}

	// Add Conflagrate DoT (40% of hit)
	conflagDot := baseDamage * spellData.ConflagDotPercentage
	totalDamage := baseDamage + conflagDot

	result.Damage = totalDamage

	// Set cooldown
	char.Conflagrate.ReadyAt = char.CurrentTime + time.Duration(spellData.Cooldown*float64(time.Second))

	return result
}

// CastLifeTap casts Life Tap
func (e *Engine) CastLifeTap(char *character.Character) CastResult {
	spellData := e.Config.Spells.LifeTap

	result := CastResult{
		Spell:    SpellLifeTap,
		CastTime: 0, // Instant
		GCDTime:  time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		DidHit:   true, // Life Tap always succeeds
	}

	// Calculate mana gained
	// Formula: [827 * (1 + talent_rank * 0.10)] + [spellpower * 0.5 * (1 + talent_rank * 0.10)]
	// We don't have Improved Life Tap, so multiplier = 1.0
	manaGained := spellData.ManaBase + (char.Stats.SpellPower * spellData.SpellpowerCoefficient)

	char.GainMana(manaGained)
	result.ManaGained = manaGained

	return result
}
