package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

// CastConflagrate casts Conflagrate.
func (e *Engine) CastConflagrate(char *character.Character) CastResult {
	spellData := e.Config.Spells.Conflagrate

	result := CastResult{
		Spell:     SpellConflagrate,
		CastTime:  0,
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

	immolateDotDamage := char.Immolate.SnapshotDotDamage
	if !(char.Immolate.Active && immolateDotDamage > 0) {
		immolateSpellData := e.Config.Spells.Immolate
		immolateDotDamage = e.CalculateSpellDamage(immolateSpellData.DotDamage, immolateSpellData.SPCoefficientDot, char)
		immolateDotDamage *= e.Config.Talents.ImprovedImmolate.DamageMultiplier
		immolateDotDamage *= e.Config.Talents.Aftermath.DotDamageMultiplier
	}
	immolateDotDamage *= e.cataclysmicBurstMultiplier(char)

	baseDamage := immolateDotDamage * spellData.ImmolateDotPercentage
	baseDamage *= e.Config.Talents.Emberstorm.DamageMultiplier
	baseDamage = e.applyFireTargetModifiers(baseDamage, char)

	bonusCrit := e.Config.Talents.FireAndBrimstone.ConflagrateCritBonus
	if e.consumeEmpoweredImp(char) || e.RollCrit(char, bonusCrit) {
		result.DidCrit = true
		baseDamage *= e.Config.Talents.Ruin.CritMultiplier
		if e.Config.Talents.Pyroclasm.Points > 0 {
			char.Pyroclasm.Active = true
			char.Pyroclasm.ExpiresAt = char.CurrentTime + time.Duration(e.Config.Talents.Pyroclasm.Duration*float64(time.Second))
		}
	}

	conflagDot := baseDamage * spellData.ConflagDotPercentage
	result.Damage = baseDamage + conflagDot

	e.applyHeatingUpStack(char)
	e.CheckSoulLeechProc(char)

	if e.Config.Player.HasRune(runes.RuneDecisiveDecimation) {
		char.DecisiveDecimation.Active = true
	}

	if e.Config.Player.HasRune(runes.RuneCataclysmicBurst) && char.CataclysmicBurst != nil {
		char.CataclysmicBurst.Clear(char.CurrentTime)
	}

	if !e.Config.Player.HasRune(runes.RuneGlyphOfConflagrate) {
		if char.Immolate.TickHandle != nil {
			char.Immolate.TickHandle.Cancel()
			char.Immolate.TickHandle = nil
		}
		char.Immolate.Active = false
		char.Immolate.TicksRemaining = 0
		char.Immolate.TickDamage = 0
		char.Immolate.TickCritChance = 0
		char.Immolate.SnapshotDotDamage = 0
		char.Immolate.ExpiresAt = char.CurrentTime
	}

	e.activateBackdraft(char)
	char.Conflagrate.ReadyAt = char.CurrentTime + time.Duration(spellData.Cooldown*float64(time.Second))

	return result
}
