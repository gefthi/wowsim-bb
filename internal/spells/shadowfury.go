package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastShadowfury casts Shadowfury. Mainly used to trigger Backdraft via Unstable Void.
func (e *Engine) CastShadowfury(char *character.Character) CastResult {
	spellData := e.Config.Spells.ShadowFury

	result := CastResult{
		Spell:     SpellShadowfury,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	e.applyHasteTimes(char, &result)
	char.SpendMana(spellData.ManaCost)

	if spellData.Cooldown > 0 {
		char.Shadowfury.ReadyAt = char.CurrentTime + time.Duration(spellData.Cooldown*float64(time.Second))
	}

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)
	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)
	damage = e.applyShadowTargetModifiers(damage, char)
	damage *= e.pureShadowMultiplier(char, SpellShadowfury)

	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}
	result.Damage = damage

	if e.Config.Player.HasRune(runes.RuneUnstableVoid) {
		e.activateBackdraft(char)
	}
	e.addPureShadowStack(char)

	return result
}
