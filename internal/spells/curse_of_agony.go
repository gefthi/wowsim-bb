package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

// CastCurseOfAgony applies Curse of Agony DoT.
func (e *Engine) CastCurseOfAgony(char *character.Character) CastResult {
	spellData := e.Config.Spells.CurseOfAgony

	result := CastResult{
		Spell:     SpellCurseOfAgony,
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

	// Snapshot base and SP contributions separately for ramped ticks.
	baseSnapshot := e.CalculateSpellDamage(spellData.DotDamage, 0, char)
	baseSnapshot = e.applyShadowTargetModifiers(baseSnapshot, char)
	spSnapshot := e.CalculateSpellDamage(0, spellData.SPCoefficientDot, char)
	spSnapshot = e.applyShadowTargetModifiers(spSnapshot, char)

	e.applyCurseOfAgonySnapshot(char, baseSnapshot, spSnapshot)
	return result
}
