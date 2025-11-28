package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

// CastCurseOfElements applies the Curse of the Elements debuff to the target.
func (e *Engine) CastCurseOfElements(char *character.Character) CastResult {
	result := CastResult{
		Spell:     SpellCurseOfElements,
		CastTime:  0,
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: 0,
		DidHit:    true,
	}

	e.applyHasteTimes(char, &result)

	// Apply a long-duration debuff (covers typical fight lengths).
	duration := 300 * time.Second
	char.CurseOfElements.Active = true
	char.CurseOfElements.ExpiresAt = char.CurrentTime + duration

	return result
}
