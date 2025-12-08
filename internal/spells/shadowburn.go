package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastShadowburn casts Shadowburn (instant, with cooldown).
func (e *Engine) CastShadowburn(char *character.Character) CastResult {
	spellData := e.Config.Spells.Shadowburn

	result := CastResult{
		Spell:     SpellShadowburn,
		CastTime:  time.Duration(spellData.CastTime * float64(time.Second)),
		GCDTime:   time.Duration(e.Config.Constants.GCD.Base * float64(time.Second)),
		ManaSpent: spellData.ManaCost,
	}

	e.applyBackdraft(char, &result, true)
	char.SpendMana(spellData.ManaCost)

	// Start cooldown
	if spellData.Cooldown > 0 {
		char.Shadowburn.ReadyAt = char.CurrentTime + time.Duration(spellData.Cooldown*float64(time.Second))
	}

	if !e.RollHit(char) {
		result.DidHit = false
		return result
	}
	result.DidHit = true

	baseDamage := spellData.BaseDamageMin + e.Rng.Float64()*(spellData.BaseDamageMax-spellData.BaseDamageMin)
	damage := e.CalculateSpellDamage(baseDamage, spellData.SPCoefficient, char)
	damage = e.applyShadowTargetModifiers(damage, char)
	damage *= e.pureShadowMultiplier(char, SpellShadowburn)
	if e.Config.Player.HasRune(runes.RuneShadowSiphon) {
		// TODO: gate on target health <35% when target health modeling is added.
		damage *= 1 + runes.ShadowSiphonDamageBonus
	}

	if stacks := e.consumeDuskTillDawn(char); stacks > 0 {
		damage *= 1 + runes.DuskTillDawnShadowburnBonusPerStack*float64(stacks)
		if stacks >= runes.DuskTillDawnMaxStacks {
			corruption := e.Config.Spells.Corruption
			dotSnapshot := e.CalculateSpellDamage(corruption.DotDamage, corruption.SPCoefficientDot, char)
			dotSnapshot = e.applyShadowTargetModifiers(dotSnapshot, char)
			e.applyCorruptionSnapshot(char, dotSnapshot)
		}
	}

	if e.RollCrit(char, 0) {
		result.DidCrit = true
		damage *= e.Config.Talents.Ruin.CritMultiplier
	}

	result.Damage = damage
	e.addPureShadowStack(char)
	return result
}
