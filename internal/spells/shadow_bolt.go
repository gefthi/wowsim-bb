package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastShadowBolt casts Shadow Bolt.
func (e *Engine) CastShadowBolt(char *character.Character) CastResult {
	spellData := e.Config.Spells.ShadowBolt

	// Cursed Shadows buff modifies mana cost and damage.
	if char.CursedShadows.Active && char.CursedShadows.ExpiresAt > char.CurrentTime {
		spellData.ManaCost *= 1 - runes.CursedShadowsManaReduction
	}
	hasShadowTrance := char.ShadowTrance.Active && char.ShadowTrance.ExpiresAt > char.CurrentTime
	freeCast := hasShadowTrance && char.ShadowTranceFreeCast

	result := CastResult{
		Spell:     SpellShadowBolt,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}
	if hasShadowTrance {
		result.CastTime = 0
	}
	if freeCast {
		result.ManaSpent = 0
	}

	e.applyHasteTimes(char, &result)
	e.applyBackdraft(char, &result, true)
	if result.ManaSpent > 0 {
		char.SpendMana(result.ManaSpent)
	}

	if !e.RollHit(char) {
		result.DidHit = false
		if hasShadowTrance {
			char.ShadowTrance.Active = false
			char.ShadowTrance.Charges = 0
			char.ShadowTrance.ExpiresAt = 0
			char.ShadowTranceFreeCast = false
			char.ShadowTranceLeechFraction = 0
		}
		return result
	}
	result.DidHit = true

	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)
	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)
	if char.CursedShadows.Active && char.CursedShadows.ExpiresAt > char.CurrentTime {
		damage *= 1 + runes.CursedShadowsDamageBonus
	}
	damage = e.applyShadowTargetModifiers(damage, char)
	damage *= e.pureShadowMultiplier(char, SpellShadowBolt)

	bonusCrit := 0.0
	if e.Config.Player.HasRune(runes.RunePyroclasmicShadows) && char.Pyroclasm.Active && char.CurrentTime < char.Pyroclasm.ExpiresAt {
		bonusCrit += runes.PyroclasmicShadowsShadowboltCritBonus
	}

	if e.RollCrit(char, bonusCrit) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage
	if result.DidHit && freeCast && char.ShadowTranceLeechFraction > 0 && damage > 0 {
		result.Healing = damage * char.ShadowTranceLeechFraction
	}
	// Procs and stacks
	e.addPureShadowStack(char)
	e.addDuskTillDawnStack(char)
	if hasShadowTrance {
		char.ShadowTrance.Active = false
		char.ShadowTrance.Charges = 0
		char.ShadowTrance.ExpiresAt = 0
		char.ShadowTranceFreeCast = false
		char.ShadowTranceLeechFraction = 0
	}
	if char.CursedShadows.Active {
		char.CursedShadows.Active = false
		char.CursedShadows.ExpiresAt = 0
		char.CursedShadows.Charges = 0
	}
	return result
}
