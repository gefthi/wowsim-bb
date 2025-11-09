package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

func (e *Engine) cataclysmicBurstMultiplier(char *character.Character) float64 {
	if !e.Config.Player.HasRune(runes.RuneCataclysmicBurst) || char.CataclysmicBurst == nil {
		return 1
	}
	stacks := char.CataclysmicBurst.Stacks()
	if stacks <= 0 {
		return 1
	}
	return 1 + runes.CataclysmicBurstStackBonus*float64(stacks)
}

func (e *Engine) handleCataclysmicBurstIncinerate(char *character.Character) {
	if !e.Config.Player.HasRune(runes.RuneCataclysmicBurst) || char.CataclysmicBurst == nil {
		return
	}
	if !char.Immolate.Active {
		return
	}
	char.CataclysmicBurst.AddStacks(char.CurrentTime, 1)
	char.Immolate.ExpiresAt += time.Duration(runes.CataclysmicBurstExtendSec * float64(time.Second))
}
