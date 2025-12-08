package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

// CastShadowCrash is a placeholder until numbers are finalized.
func (e *Engine) CastShadowCrash(char *character.Character) CastResult {
	spellData := e.Config.Spells.ShadowCrash

	result := CastResult{
		Spell:     SpellShadowCrash,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	e.applyHasteTimes(char, &result)
	e.applyBackdraft(char, &result, true)
	if result.ManaSpent > 0 {
		char.SpendMana(result.ManaSpent)
	}

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// TODO: fill in damage once Shadow Crash numbers are available.
	result.Damage = 0
	return result
}
