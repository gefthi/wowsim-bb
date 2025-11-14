package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
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

	if e.Config.Player.HasRune(runes.RuneGlyphOfLifeTap) {
		bonus := char.Stats.Spirit * runes.GlyphOfLifeTapSpiritMultiplier
		char.LifeTapBuff.Active = true
		char.LifeTapBuff.Value = bonus
		char.LifeTapBuff.ExpiresAt = char.CurrentTime + time.Duration(runes.GlyphOfLifeTapDurationSec*float64(time.Second))
	}

	return result
}
