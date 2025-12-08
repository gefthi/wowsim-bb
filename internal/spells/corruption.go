package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

// CastCorruption applies Corruption DoT.
func (e *Engine) CastCorruption(char *character.Character) CastResult {
	spellData := e.Config.Spells.Corruption

	result := CastResult{
		Spell:     SpellCorruption,
		CastTime:  0,
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

	// Snapshot total DoT damage, then derive per-tick.
	tickCount := spellData.DotTicks
	if tickCount <= 0 {
		tickCount = 1
	}
	dotSnapshot := e.CalculateSpellDamage(spellData.DotDamage, spellData.SPCoefficientDot, char)
	dotSnapshot = e.applyShadowTargetModifiers(dotSnapshot, char)

	e.applyCorruptionSnapshot(char, dotSnapshot)
	e.addPureShadowStack(char)

	return result
}
