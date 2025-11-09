package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

func (e *Engine) CheckSoulLeechProc(char *character.Character) {
	if !e.Config.Talents.ImprovedSoulLeech.Enabled || e.Config.Talents.ImprovedSoulLeech.Points <= 0 {
		return
	}

	if e.Rng.Float64() < 0.30 {
		instantMana := char.Stats.MaxMana * e.Config.Talents.ImprovedSoulLeech.InstantManaReturn
		char.GainMana(instantMana)

		char.ImprovedSoulLeech.Active = true
		char.ImprovedSoulLeech.ExpiresAt = char.CurrentTime + time.Duration(e.Config.Talents.ImprovedSoulLeech.HotDuration*float64(time.Second))
		char.SoulLeechLastTick = char.CurrentTime
	}
}
