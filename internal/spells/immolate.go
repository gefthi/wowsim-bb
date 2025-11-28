package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastImmolate casts Immolate.
func (e *Engine) CastImmolate(char *character.Character) CastResult {
	spellData := e.Config.Spells.Immolate

	result := CastResult{
		Spell:     SpellImmolate,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	e.applyHasteTimes(char, &result)
	baseTickCount := spellData.DotTicks
	if baseTickCount <= 0 {
		baseTickCount = 1
	}
	dotDuration := spellData.DotDuration
	tickCount := baseTickCount
	if e.Config.Player.HasRune(runes.RuneAgentOfChaos) {
		dotDuration += runes.AgentOfChaosExtraDurationSec
		tickCount += runes.AgentOfChaosExtraTicks
	}
	agentHasteMult := e.agentOfChaosHasteMultiplier(char)
	effectiveDuration := dotDuration
	if agentHasteMult != 1 {
		effectiveDuration = dotDuration / agentHasteMult
	}

	e.applyBackdraft(char, &result, true)

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	forceCrit := e.consumeEmpoweredImp(char)
	directDamage := e.CalculateSpellDamage(spellData.DirectDamage, spellData.SPCoefficientDirect, char)
	directDamage *= e.Config.Talents.ImprovedImmolate.DamageMultiplier
	if e.Config.Player.HasRune(runes.RuneDestructionMastery) {
		directDamage *= runes.DestructionMasteryImmolateBonus
	}
	if e.Config.Player.HasRune(runes.RuneAgentOfChaos) {
		directDamage *= runes.AgentOfChaosDirectDamagePenalty
	}

	directCrit := forceCrit || e.RollCrit(char, 0)
	if directCrit {
		directDamage *= e.Config.Talents.Ruin.CritMultiplier
	}
	directDamage = e.applyFireTargetModifiers(directDamage, char)

	dotSnapshot := e.CalculateSpellDamage(spellData.DotDamage, spellData.SPCoefficientDot, char)
	dotSnapshot *= e.Config.Talents.ImprovedImmolate.DamageMultiplier
	dotSnapshot *= e.Config.Talents.Aftermath.DotDamageMultiplier
	if e.Config.Player.HasRune(runes.RuneDestructionMastery) {
		dotSnapshot *= runes.DestructionMasteryImmolateBonus
	}
	if e.Config.Player.HasRune(runes.RuneAgentOfChaos) && baseTickCount > 0 {
		dotSnapshot *= float64(tickCount) / float64(baseTickCount)
	}

	baseTickDamage := 0.0
	if tickCount > 0 {
		baseTickDamage = dotSnapshot / float64(tickCount)
	} else {
		baseTickDamage = dotSnapshot
	}
	tickCritChance := e.snapshotCritChance(char, 0)

	result.DidCrit = directCrit
	result.Damage = directDamage

	e.CheckSoulLeechProc(char)

	char.Immolate.Active = true
	char.Immolate.ExpiresAt = char.CurrentTime + time.Duration(effectiveDuration*float64(time.Second))
	if tickCount > 0 {
		intervalSeconds := dotDuration / float64(tickCount)
		if agentHasteMult != 1 {
			intervalSeconds /= agentHasteMult
		}
		char.Immolate.TickInterval = time.Duration(intervalSeconds * float64(time.Second))
	} else {
		char.Immolate.TickInterval = 0
	}
	char.Immolate.LastTick = char.CurrentTime
	char.Immolate.TickDamage = baseTickDamage
	char.Immolate.TickCritChance = tickCritChance
	char.Immolate.TicksRemaining = tickCount
	char.Immolate.SnapshotDotDamage = dotSnapshot

	return result
}
