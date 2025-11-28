package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastSoulFire casts Soul Fire. Only meaningful when Decisive Decimation is active.
func (e *Engine) CastSoulFire(char *character.Character) CastResult {
	spellData := e.Config.Spells.SoulFire

	result := CastResult{
		Spell:     SpellSoulFire,
		ManaSpent: spellData.ManaCost,
	}

	castTime := time.Duration(spellData.CastTime * float64(time.Second))
	gcd := time.Duration(e.Config.Constants.GCD.Base * float64(time.Second))

	if char.DecisiveDecimation.Active {
		castTime = time.Duration(float64(castTime) * (1.0 - runes.DecisiveDecimationCastReduction))
		char.DecisiveDecimation.Active = false
	}

	result.CastTime = castTime
	result.GCDTime = gcd

	e.applyBackdraft(char, &result, true)

	char.SpendMana(spellData.ManaCost)

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	// Roll damage
	base := spellData.BaseDamageMin
	if spellData.BaseDamageMax > spellData.BaseDamageMin {
		base += (spellData.BaseDamageMax - spellData.BaseDamageMin) * e.Rng.Float64()
	}
	damage := e.CalculateSpellDamage(base, spellData.SPCoefficient, char)
	damage = e.applyFireTargetModifiers(damage, char)

	if e.consumeEmpoweredImp(char) || e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage

	return result
}
