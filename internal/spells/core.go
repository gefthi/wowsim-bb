package spells

import (
	"math/rand"
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/runes"
)

// SpellType identifies different spells.
type SpellType int

const (
	pvePowerMultiplier       = 1.25
	CurseOfElementsMultiplier = 1.10
)

const (
	SpellCurseOfElements SpellType = iota
	SpellImmolate
	SpellIncinerate
	SpellChaosBolt
	SpellConflagrate
	SpellLifeTap
	SpellSoulFire
	SpellImpFirebolt
)

// CastResult represents the result of a spell cast.
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

// Engine handles spell casting and damage calculation.
type Engine struct {
	Config *config.Config
	Rng    *rand.Rand

	// Target type for hit calculation
	IsBossTarget bool
}

// NewEngine creates a new spell engine.
func NewEngine(cfg *config.Config, seed int64, isBoss bool) *Engine {
	return &Engine{
		Config:       cfg,
		Rng:          rand.New(rand.NewSource(seed)),
		IsBossTarget: isBoss,
	}
}

// RollHit determines if a spell hits.
func (e *Engine) RollHit(char *character.Character) bool {
	hitCap := float64(e.Config.Constants.HitMechanics.EqualLevelMissChance)
	if e.IsBossTarget {
		hitCap = float64(e.Config.Constants.HitMechanics.BossHitCap)
	}

	hitPct := char.Stats.HitPct
	if e.Config.Player.HasRune(runes.RuneSuppression) {
		hitPct += runes.SuppressionHitBonus
	}

	missChance := hitCap - hitPct
	if missChance <= 0 {
		return true
	}

	roll := e.Rng.Float64() * 100.0
	return roll >= missChance
}

// totalCritChancePercent returns crit chance (0-100) including talents/bonuses.
func (e *Engine) totalCritChancePercent(char *character.Character, bonusCrit float64) float64 {
	totalCrit := char.Stats.CritPct
	totalCrit += float64(e.Config.Talents.Devastation.Points) * e.Config.Talents.Devastation.CritBonusPerPoint * 100.0
	totalCrit += float64(e.Config.Talents.Backlash.Points) * e.Config.Talents.Backlash.CritBonusPerPoint * 100.0
	totalCrit += bonusCrit * 100.0
	if totalCrit < 0 {
		return 0
	}
	return totalCrit
}

// snapshotCritChance returns crit probability (0-1) locked at cast time.
func (e *Engine) snapshotCritChance(char *character.Character, bonusCrit float64) float64 {
	chance := e.totalCritChancePercent(char, bonusCrit)
	if chance <= 0 {
		return 0
	}
	if chance >= 100 {
		return 1
	}
	return chance / 100.0
}

// RollCrit determines if a spell crits.
func (e *Engine) RollCrit(char *character.Character, bonusCrit float64) bool {
	chance := e.totalCritChancePercent(char, bonusCrit)
	if chance <= 0 {
		return false
	}
	if chance >= 100 {
		return true
	}
	roll := e.Rng.Float64() * 100.0
	return roll < chance
}

// CalculateSpellDamage calculates base damage with spell power and buffs.
func (e *Engine) CalculateSpellDamage(baseDamage, spCoefficient float64, char *character.Character) float64 {
	damage := baseDamage + (e.effectiveSpellPower(char) * spCoefficient)
	damage *= pvePowerMultiplier
	damage *= e.Config.Talents.Emberstorm.DamageMultiplier
	if e.Config.Talents.Pyroclasm.Points > 0 && char.Pyroclasm.Active && char.CurrentTime < char.Pyroclasm.ExpiresAt {
		damage *= e.Config.Talents.Pyroclasm.DamageMultiplier
	}
	if e.Config.Player.HasRune(runes.RuneDestructionMastery) {
		damage *= runes.DestructionMasteryGlobalBonus
	}
	return damage
}

// ApplyFireAndBrimstone applies damage bonus if Immolate is on target (Incinerate and Chaos Bolt only).
func (e *Engine) ApplyFireAndBrimstone(damage float64, char *character.Character, spellType SpellType) float64 {
	if !char.Immolate.Active {
		return damage
	}
	if spellType == SpellIncinerate && e.Config.Talents.FireAndBrimstone.AppliesToIncinerate {
		return damage * e.Config.Talents.FireAndBrimstone.DamageMultiplier
	}
	if spellType == SpellChaosBolt && e.Config.Talents.FireAndBrimstone.AppliesToChaosBolt {
		return damage * e.Config.Talents.FireAndBrimstone.DamageMultiplier
	}
	return damage
}

func (e *Engine) fireTargetMultiplier(char *character.Character) float64 {
	mult := 1.0
	if e.Config.Player.HasRune(runes.RuneHeatingUp) && char.HeatingUp != nil {
		active := char.HeatingUp.ActiveAt(char.CurrentTime)
		mult *= runes.HeatingUpMultiplier(active, char.HeatingUp.Stacks(), char.HeatingUp.ExpiresAt(), char.CurrentTime)
	}
	if char.CurseOfElements.Active && char.CurseOfElements.ExpiresAt > char.CurrentTime {
		mult *= CurseOfElementsMultiplier
	}
	return mult
}

func (e *Engine) applyFireTargetModifiers(damage float64, char *character.Character) float64 {
	if damage <= 0 {
		return damage
	}
	return damage * e.fireTargetMultiplier(char)
}

func (e *Engine) applyHeatingUpStack(char *character.Character) {
	if !e.Config.Player.HasRune(runes.RuneHeatingUp) || char.HeatingUp == nil {
		return
	}
	char.HeatingUp.AddStacks(char.CurrentTime, 1)
}

func (e *Engine) agentOfChaosHasteMultiplier(char *character.Character) float64 {
	if !e.Config.Player.HasRune(runes.RuneAgentOfChaos) {
		return 1
	}
	mult := 1.0 + (char.Stats.HastePct / 100.0)
	if mult <= 0 {
		return 1
	}
	return mult
}

func (e *Engine) effectiveSpellPower(char *character.Character) float64 {
	if char == nil {
		return 0
	}
	sp := char.Stats.SpellPower
	if e.Config.Player.HasRune(runes.RuneDemonicAegis) {
		sp += char.Stats.Spirit * runes.DemonicAegisSpiritBonusPerPoint
	}
	if bonus := e.Config.Talents.ShadowAndFlame.BonusSPPercentage; bonus > 0 {
		sp *= 1 + bonus
	}
	if char.LifeTapBuff.Active {
		if char.LifeTapBuff.ExpiresAt > char.CurrentTime {
			sp += char.LifeTapBuff.Value
		} else {
			char.LifeTapBuff.Active = false
			char.LifeTapBuff.Value = 0
		}
	}
	return sp
}

func (e *Engine) consumeEmpoweredImp(char *character.Character) bool {
	if e.Config.Talents.EmpoweredImp.Points <= 0 {
		return false
	}
	if !char.EmpoweredImp.Active {
		return false
	}
	if char.EmpoweredImp.ExpiresAt <= char.CurrentTime {
		char.EmpoweredImp.Active = false
		return false
	}
	char.EmpoweredImp.Active = false
	return true
}
