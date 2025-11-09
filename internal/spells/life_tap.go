package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

// CastLifeTap casts Life Tap.
func (e *Engine) CastLifeTap(char *character.Character) CastResult {
	spellData := e.Config.Spells.LifeTap

	result := CastResult{
		Spell:    SpellLifeTap,
		CastTime: 0,
		GCDTime:  time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		DidHit:   true,
	}

	manaGained := spellData.ManaBase + (char.Stats.SpellPower * spellData.SpellpowerCoefficient)
	char.GainMana(manaGained)
	result.ManaGained = manaGained

	return result
}
