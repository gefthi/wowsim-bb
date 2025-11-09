package spells

import (
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

func (e *Engine) activateGuldansChosen(char *character.Character) {
	if !e.Config.Player.HasRune(runes.RuneGuldansChosen) || char.GuldansChosen == nil {
		return
	}
	char.GuldansChosen.SetStacks(char.CurrentTime, 1)
}
