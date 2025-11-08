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

	// Add talent bonuses (Devastation + Backlash) - now point-based
	totalCrit += float64(e.Config.Talents.Devastation.Points) * e.Config.Talents.Devastation.CritBonusPerPoint * 100.0
	totalCrit += float64(e.Config.Talents.Backlash.Points) * e.Config.Talents.Backlash.CritBonusPerPoint * 100.0

	// Add any spell-specific bonus (e.g., Conflagrate from Fire and Brimstone)
	totalCrit += bonusCrit * 100.0

	roll := e.Rng.Float64() * 100.0
	return roll < totalCrit
}

// CalculateSpellDamage calculates base damage with spell power and buffs
func (e *Engine) CalculateSpellDamage(baseDamage, spCoefficient float64, char *character.Character) float64 {
	// Base damage + (SP * coefficient)
	damage := baseDamage + (char.Stats.SpellPower * spCoefficient)

	// Apply Emberstorm (15% fire/shadow damage)
	damage *= e.Config.Talents.Emberstorm.DamageMultiplier

	// Apply Pyroclasm if active (+6% fire/shadow damage)
	if e.Config.Talents.Pyroclasm.Enabled && char.Pyroclasm.Active && char.CurrentTime < char.Pyroclasm.ExpiresAt {
		damage *= e.Config.Talents.Pyroclasm.DamageMultiplier
	}

	return damage
}

// ApplyFireAndBrimstone applies damage bonus if Immolate is on target (Incinerate and Chaos Bolt only!)
func (e *Engine) ApplyFireAndBrimstone(damage float64, char *character.Character, spellType SpellType) float64 {
	if !char.Immolate.Active {
		return damage
	}

	// Fire and Brimstone ONLY applies to Incinerate and Chaos Bolt
	if spellType == SpellIncinerate && e.Config.Talents.FireAndBrimstone.AppliesToIncinerate {
		return damage * e.Config.Talents.FireAndBrimstone.DamageMultiplier
	}
	if spellType == SpellChaosBolt && e.Config.Talents.FireAndBrimstone.AppliesToChaosBolt {
		return damage * e.Config.Talents.FireAndBrimstone.DamageMultiplier
	}

	return damage
}

func (e *Engine) backdraftEnabled() bool {
	return e.Config.Talents.Backdraft.Enabled &&
		e.Config.Talents.Backdraft.Points > 0 &&
		e.Config.Talents.Backdraft.Charges > 0
}

func (e *Engine) isBackdraftActive(char *character.Character) bool {
	if !e.backdraftEnabled() {
		return false
	}
	if !char.Backdraft.Active {
		return false
	}
	if char.Backdraft.Charges <= 0 {
		char.Backdraft.Active = false
		char.Backdraft.Charges = 0
		return false
	}
	if char.CurrentTime >= char.Backdraft.ExpiresAt {
		char.Backdraft.Active = false
		char.Backdraft.Charges = 0
		return false
	}
	return true
}

func (e *Engine) applyBackdraft(char *character.Character, result *CastResult, consumesCharge bool) {
	if !e.isBackdraftActive(char) {
		return
	}
	bd := e.Config.Talents.Backdraft
	if result.CastTime > 0 && bd.CastTimeReduction > 0 {
		result.CastTime = time.Duration(float64(result.CastTime) * (1.0 - bd.CastTimeReduction))
	}
	if result.GCDTime > 0 && bd.GCDReduction > 0 {
		result.GCDTime = time.Duration(float64(result.GCDTime) * (1.0 - bd.GCDReduction))
		minGCD := time.Duration(e.Config.Constants.GCD.Minimum * float64(time.Second))
		if minGCD > 0 && result.GCDTime < minGCD {
			result.GCDTime = minGCD
		}
	}
	if consumesCharge {
		char.Backdraft.Charges--
		if char.Backdraft.Charges <= 0 {
			char.Backdraft.Active = false
			char.Backdraft.Charges = 0
		}
	}
}

func (e *Engine) activateBackdraft(char *character.Character) {
	if !e.backdraftEnabled() {
		return
	}
	char.Backdraft.Active = true
	char.Backdraft.Charges = e.Config.Talents.Backdraft.Charges
	char.Backdraft.ExpiresAt = char.CurrentTime + time.Duration(e.Config.Talents.Backdraft.Duration*float64(time.Second))
}

// CheckSoulLeechProc checks for Soul Leech proc (30% chance on fire/shadow damage)
func (e *Engine) CheckSoulLeechProc(char *character.Character) {
	if !e.Config.Talents.ImprovedSoulLeech.Enabled || e.Config.Talents.ImprovedSoulLeech.Points <= 0 {
		return
	}

	if e.Rng.Float64() < 0.30 {
		// Instant mana return (2% of max mana)
		instantMana := char.Stats.MaxMana * e.Config.Talents.ImprovedSoulLeech.InstantManaReturn
		char.GainMana(instantMana)

		// Activate HoT buff
		char.ImprovedSoulLeech.Active = true
		char.ImprovedSoulLeech.ExpiresAt = char.CurrentTime + time.Duration(e.Config.Talents.ImprovedSoulLeech.HotDuration*float64(time.Second))
		char.SoulLeechLastTick = char.CurrentTime
	}
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

	e.applyBackdraft(char, &result, true)

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

	// Check for Soul Leech proc (30% chance on fire damage)
	e.CheckSoulLeechProc(char)

	// Apply Immolate debuff
	char.Immolate.Active = true
	char.Immolate.ExpiresAt = char.CurrentTime + time.Duration(spellData.DotDuration*float64(time.Second))
	if spellData.DotTicks > 0 {
		char.Immolate.TickInterval = time.Duration(spellData.DotDuration*float64(time.Second)) / time.Duration(spellData.DotTicks)
	} else {
		char.Immolate.TickInterval = 0
	}
	char.Immolate.LastTick = char.CurrentTime
	if spellData.DotTicks > 0 {
		char.Immolate.TickDamage = dotDamage / float64(spellData.DotTicks)
	} else {
		char.Immolate.TickDamage = dotDamage
	}
	
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

	e.applyBackdraft(char, &result, true)

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

	// Apply Fire and Brimstone if Immolate is up (affects Incinerate)
	damage = e.ApplyFireAndBrimstone(damage, char, SpellIncinerate)

	// Roll crit
	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage

	// Check for Soul Leech proc (30% chance on fire damage)
	e.CheckSoulLeechProc(char)

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

	e.applyBackdraft(char, &result, true)

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// Random base damage
	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)

	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)

	// Apply Fire and Brimstone (affects Chaos Bolt)
	damage = e.ApplyFireAndBrimstone(damage, char, SpellChaosBolt)

	// Roll crit
	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage

	// Check for Soul Leech proc (30% chance on fire/shadow damage)
	e.CheckSoulLeechProc(char)

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

	e.applyBackdraft(char, &result, true)

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

		// Pyroclasm: On Conflagrate crit, gain +6% fire/shadow damage
		if e.Config.Talents.Pyroclasm.Enabled {
			char.Pyroclasm.Active = true
			char.Pyroclasm.ExpiresAt = char.CurrentTime + time.Duration(e.Config.Talents.Pyroclasm.Duration*float64(time.Second))
		}
	}

	// Add Conflagrate DoT (40% of hit)
	conflagDot := baseDamage * spellData.ConflagDotPercentage
	totalDamage := baseDamage + conflagDot

	result.Damage = totalDamage

	// Check for Soul Leech proc (30% chance on fire damage)
	e.CheckSoulLeechProc(char)

	// Activate Backdraft (3 charges, 15s duration) on hit
	if result.DidHit {
		e.activateBackdraft(char)
	}

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
