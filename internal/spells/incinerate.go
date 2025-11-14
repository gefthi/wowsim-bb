package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastIncinerate casts Incinerate.
func (e *Engine) CastIncinerate(char *character.Character) CastResult {
	spellData := e.Config.Spells.Incinerate

	result := CastResult{
		Spell:     SpellIncinerate,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	e.applyBackdraft(char, &result, true)
	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)

	if char.Immolate.Active {
		immolateBonus := spellData.ImmolateBonusMin + e.Rng.Float64()*(spellData.ImmolateBonusMax-spellData.ImmolateBonusMin)
		baseDamage += immolateBonus
	}

	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)
	damage = e.ApplyFireAndBrimstone(damage, char, SpellIncinerate)
	damage = e.applyFireTargetModifiers(damage, char)

	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}
	if e.Config.Player.HasRune(runes.RuneGlyphOfIncinerate) {
		damage *= runes.GlyphOfIncinerateDamageMultiplier
	}

	result.Damage = damage

	if result.DidHit {
		e.handleCataclysmicBurstIncinerate(char)
		e.CheckSoulLeechProc(char)
	}

	return result
}
