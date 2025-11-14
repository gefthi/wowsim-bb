package engine

import (
	"math"
	"strings"
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/spells"
)

const (
	impBaseIntellect           = 264.0
	impBaseSpirit              = 260.0
	impIntellectInheritance    = 0.30
	impSpiritInheritance       = 0.30
	impSpellPowerInheritance   = 0.15
	impCritBasePercent         = 0.94
	impCritIntellectPerPercent = 60.0
	impManaPerIntellect        = 9.0
	impSpiritToMp5             = 0.169
	impCastingRegenFraction    = 0.15
	impFireboltManaCost        = 115.0
)

type petController interface {
	reset(owner *character.Character)
	start(sim *Simulator, owner *character.Character, result *SimulationResult, spellEngine *spells.Engine)
}

func (s *Simulator) initializePets() {
	summon := strings.ToLower(strings.TrimSpace(s.Config.Player.Pet.Summon))
	switch summon {
	case "", "none":
		return
	case "imp":
		s.pets = append(s.pets, newImpController(s.Config))
	default:
		// Unknown pet type, ignore for now.
	}
}

func (s *Simulator) resetPets(owner *character.Character) {
	for _, pet := range s.pets {
		pet.reset(owner)
	}
}

func (s *Simulator) startPets(owner *character.Character, result *SimulationResult, spellEngine *spells.Engine) {
	for _, pet := range s.pets {
		pet.start(s, owner, result, spellEngine)
	}
}

type impController struct {
	cfg            *config.Config
	owner          *character.Character
	spellPower     float64
	critChance     float64
	intellect      float64
	spirit         float64
	mana           float64
	manaMax        float64
	mp5Casting     float64
	mp5OOC         float64
	castTime       time.Duration
	lastManaUpdate time.Duration
	event          *scheduledEvent
}

func newImpController(cfg *config.Config) *impController {
	return &impController{
		cfg:      cfg,
		castTime: 2500 * time.Millisecond,
	}
}

func (imp *impController) reset(owner *character.Character) {
	imp.owner = owner
	if owner == nil {
		return
	}
	imp.castTime = 2500 * time.Millisecond
	if imp.cfg != nil {
		reduction := float64(imp.cfg.Talents.DemonicPower.Points) * imp.cfg.Talents.DemonicPower.FireboltCastReduction
		if reduction > 0 {
			cut := time.Duration(reduction * float64(time.Second))
			if cut >= imp.castTime {
				cut = imp.castTime - 500*time.Millisecond
			}
			if cut > 0 {
				imp.castTime -= cut
			}
		}
	}
	playerInt := owner.Stats.Intellect
	if playerInt < 0 {
		playerInt = 0
	}
	playerSpirit := owner.Stats.Spirit
	if playerSpirit < 0 {
		playerSpirit = 0
	}
	imp.intellect = impBaseIntellect + playerInt*impIntellectInheritance
	imp.spirit = impBaseSpirit + playerSpirit*impSpiritInheritance
	imp.spellPower = owner.Stats.SpellPower * impSpellPowerInheritance
	imp.critChance = ((imp.intellect / impCritIntellectPerPercent) + impCritBasePercent) / 100.0
	imp.manaMax = imp.intellect * impManaPerIntellect
	if imp.manaMax <= 0 {
		imp.manaMax = 2000
	}
	imp.mana = imp.manaMax
	imp.mp5OOC = imp.spirit * impSpiritToMp5
	imp.mp5Casting = imp.mp5OOC * impCastingRegenFraction
	imp.lastManaUpdate = owner.CurrentTime
	imp.event = nil
}

func (imp *impController) start(sim *Simulator, owner *character.Character, result *SimulationResult, spellEngine *spells.Engine) {
	if owner == nil {
		return
	}
	imp.scheduleFirebolt(sim, owner, result, spellEngine, owner.CurrentTime)
	if sim.LogEnabled {
		sim.logStaticf("Imp summoned (Firebolt every %.2fs)", imp.castTime.Seconds())
	}
}

func (imp *impController) scheduleFirebolt(sim *Simulator, owner *character.Character, result *SimulationResult, spellEngine *spells.Engine, start time.Duration) {
	castStart := imp.ensureMana(sim, start)
	finish := castStart + imp.castTime
	imp.event = sim.scheduleEvent(finish, func() {
		imp.event = nil
		imp.castFirebolt(sim, owner, result, spellEngine, castStart, finish)
	})
}

func (imp *impController) ensureMana(sim *Simulator, desiredStart time.Duration) time.Duration {
	if desiredStart < imp.lastManaUpdate {
		desiredStart = imp.lastManaUpdate
	}
	imp.regenMana(desiredStart, imp.mp5Casting)
	if imp.mana >= impFireboltManaCost {
		imp.mana -= impFireboltManaCost
		imp.lastManaUpdate = desiredStart
		return desiredStart
	}
	regenPerSecond := imp.mp5OOC / 5.0
	if regenPerSecond <= 0 {
		regenPerSecond = 1
	}
	deficit := impFireboltManaCost - imp.mana
	secondsNeeded := deficit / regenPerSecond
	delay := time.Duration(math.Ceil(secondsNeeded * float64(time.Second)))
	start := desiredStart + delay
	imp.regenMana(start, imp.mp5OOC)
	if imp.mana < impFireboltManaCost {
		imp.mana = impFireboltManaCost
	}
	imp.mana -= impFireboltManaCost
	imp.lastManaUpdate = start
	return start
}

func (imp *impController) regenMana(now time.Duration, mp5 float64) {
	if mp5 <= 0 {
		return
	}
	if now <= imp.lastManaUpdate {
		return
	}
	elapsed := now - imp.lastManaUpdate
	gain := mp5 * (float64(elapsed) / float64(time.Second)) / 5.0
	imp.mana = math.Min(imp.manaMax, imp.mana+gain)
	imp.lastManaUpdate = now
}

func (imp *impController) castFirebolt(sim *Simulator, owner *character.Character, result *SimulationResult, spellEngine *spells.Engine, castStart, castComplete time.Duration) {
	if owner == nil {
		return
	}
	baseMin := 89.0
	baseMax := 101.0
	damage := baseMin + (baseMax-baseMin)*spellEngine.Rng.Float64()
	damage += imp.spellPower * 0.571
	if imp.cfg != nil {
		points := imp.cfg.Talents.EmpoweredImp.Points
		if points > 0 {
			damage *= 1 + float64(points)*imp.cfg.Talents.EmpoweredImp.DamagePerPoint
		}
	}

	didCrit := false
	if spellEngine.Rng.Float64() < imp.critChance {
		didCrit = true
		damage *= sim.Config.Talents.Ruin.CritMultiplier
	}

	castResult := spells.CastResult{
		Spell:    spells.SpellImpFirebolt,
		Damage:   damage,
		DidHit:   true,
		DidCrit:  didCrit,
		CastTime: imp.castTime,
	}
	result.recordSpellCast(spells.SpellImpFirebolt, castResult)

	if sim.LogEnabled {
		outcome := "HIT"
		if didCrit {
			outcome = "CRIT"
		}
		sim.logAt(castComplete, "PET_CAST Firebolt %s damage=%.0f (mana %.0f/%.0f)", outcome, damage, imp.mana, imp.manaMax)
	}

	imp.scheduleFirebolt(sim, owner, result, spellEngine, castComplete)

	if didCrit && imp.cfg != nil {
		points := imp.cfg.Talents.EmpoweredImp.Points
		if points > 0 {
			chance := float64(points) * imp.cfg.Talents.EmpoweredImp.ProcChancePerPoint
			if chance > 1 {
				chance = 1
			}
			if spellEngine.Rng.Float64() < chance {
				owner.EmpoweredImp.Active = true
				duration := time.Duration(imp.cfg.Talents.EmpoweredImp.BuffDuration * float64(time.Second))
				if duration <= 0 {
					duration = 8 * time.Second
				}
				owner.EmpoweredImp.ExpiresAt = castComplete + duration
			}
		}
	}
}
