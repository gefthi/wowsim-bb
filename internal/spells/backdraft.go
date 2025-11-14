package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

func (e *Engine) backdraftEnabled() bool {
	return e.Config.Talents.Backdraft.Points > 0 &&
		e.Config.Talents.Backdraft.Charges > 0
}

func (e *Engine) isBackdraftActive(char *character.Character) bool {
	if !e.backdraftEnabled() {
		return false
	}
	if !char.Backdraft.Active || char.Backdraft.Charges <= 0 {
		char.Backdraft.Active = false
		char.Backdraft.Charges = 0
		return false
	}
	if char.CurrentTime >= char.Backdraft.ExpiresAt {
		char.Backdraft.Active = false
		char.Backdraft.Charges = 0
		return false
	}
	return true
}

func (e *Engine) applyBackdraft(char *character.Character, result *CastResult, consumesCharge bool) {
	if !e.isBackdraftActive(char) {
		return
	}
	bd := e.Config.Talents.Backdraft
	if result.CastTime > 0 && bd.CastTimeReduction > 0 {
		result.CastTime = time.Duration(float64(result.CastTime) * (1.0 - bd.CastTimeReduction))
	}
	if result.GCDTime > 0 && bd.GCDReduction > 0 {
		result.GCDTime = time.Duration(float64(result.GCDTime) * (1.0 - bd.GCDReduction))
		minGCD := time.Duration(e.Config.Constants.GCD.Minimum * float64(time.Second))
		if minGCD > 0 && result.GCDTime < minGCD {
			result.GCDTime = minGCD
		}
	}
	if consumesCharge && !e.shouldSkipBackdraftConsumption(char) {
		char.Backdraft.Charges--
		if char.Backdraft.Charges <= 0 {
			char.Backdraft.Active = false
			char.Backdraft.Charges = 0
		}
	}
}

func (e *Engine) activateBackdraft(char *character.Character) {
	if !e.backdraftEnabled() {
		return
	}
	char.Backdraft.Active = true
	char.Backdraft.Charges = e.Config.Talents.Backdraft.Charges
	char.Backdraft.ExpiresAt = char.CurrentTime + time.Duration(e.Config.Talents.Backdraft.Duration*float64(time.Second))
}

func (e *Engine) shouldSkipBackdraftConsumption(char *character.Character) bool {
	if !e.Config.Player.HasRune(runes.RuneGuldansChosen) || char.GuldansChosen == nil {
		return false
	}
	return char.GuldansChosen.ActiveAt(char.CurrentTime)
}
