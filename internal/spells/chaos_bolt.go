package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
)

// CastChaosBolt casts Chaos Bolt.
func (e *Engine) CastChaosBolt(char *character.Character) CastResult {
	spellData := e.Config.Spells.ChaosBolt

	result := CastResult{
		Spell:     SpellChaosBolt,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	e.activateGuldansChosen(char)
	e.applyBackdraft(char, &result, true)

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)

	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)
	damage = e.ApplyFireAndBrimstone(damage, char, SpellChaosBolt)
	damage = e.applyFireTargetModifiers(damage, char)

	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage
	e.CheckSoulLeechProc(char)

	char.ChaosBolt.ReadyAt = char.CurrentTime + time.Duration(spellData.Cooldown*float64(time.Second))

	return result
}
