package engine

import (
	"strings"
	"time"

	"wotlk-destro-sim/internal/apl"
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
	"wotlk-destro-sim/internal/spells"
)

type rotationContext struct {
	sim  *Simulator
	char *character.Character
}

type buffState struct {
	pyroActive          bool
	pyroExpires         time.Duration
	backdraftActive     bool
	backdraftCharges    int
	soulActive          bool
	empImpActive        bool
	lifeTapActive       bool
	lifeTapExpires      time.Duration
	heatingStacks       int
	heatingExpires      time.Duration
	catBurstStacks      int
	guldansActive       bool
	guldansExpires      time.Duration
	shadowTranceActive  bool
	shadowTranceExpires time.Duration
}

func (c *rotationContext) BuffActive(name string) bool {
	lower := strings.ToLower(name)
	if lower == "life_tap_buff" && !c.sim.Config.Player.HasRune(runes.RuneGlyphOfLifeTap) {
		return true
	}
	if lower == "decisive_decimation" {
		return c.char.DecisiveDecimation.Active
	}
	if lower == "dusk_till_dawn" {
		return c.char.DuskTillDawn != nil && c.char.DuskTillDawn.ActiveAt(c.char.CurrentTime)
	}
	buff := c.getBuff(lower)
	if buff == nil {
		return false
	}
	if !buff.Active {
		return false
	}
	return buff.ExpiresAt > c.char.CurrentTime
}

func (c *rotationContext) BuffRemaining(name string) time.Duration {
	lower := strings.ToLower(name)
	if lower == "life_tap_buff" && !c.sim.Config.Player.HasRune(runes.RuneGlyphOfLifeTap) {
		return time.Hour
	}
	if lower == "decisive_decimation" {
		if c.char.DecisiveDecimation.Active {
			return time.Hour
		}
		return 0
	}
	if lower == "dusk_till_dawn" {
		if c.char.DuskTillDawn == nil {
			return 0
		}
		return c.char.DuskTillDawn.Remaining(c.char.CurrentTime)
	}
	buff := c.getBuff(lower)
	if buff == nil {
		return 0
	}
	if !buff.Active || buff.ExpiresAt <= c.char.CurrentTime {
		return 0
	}
	return buff.ExpiresAt - c.char.CurrentTime
}

func (c *rotationContext) BuffCharges(name string) int {
	if strings.ToLower(name) == "dusk_till_dawn" {
		if c.char.DuskTillDawn == nil {
			return 0
		}
		return c.char.DuskTillDawn.Stacks()
	}
	buff := c.getBuff(name)
	if buff == nil {
		return 0
	}
	return buff.Charges
}

func (c *rotationContext) DebuffActive(name string) bool {
	switch strings.ToLower(name) {
	case "immolate":
		return c.char.Immolate.Active && c.char.Immolate.ExpiresAt > c.char.CurrentTime
	case "curse_of_the_elements":
		return c.char.CurseOfElements.Active && c.char.CurseOfElements.ExpiresAt > c.char.CurrentTime
	case "corruption":
		return c.char.Corruption.Active && c.char.Corruption.ExpiresAt > c.char.CurrentTime
	case "curse_of_agony":
		return c.char.CurseOfAgony.Active && c.char.CurseOfAgony.ExpiresAt > c.char.CurrentTime
	default:
		return false
	}
}

func (c *rotationContext) DebuffRemaining(name string) time.Duration {
	switch strings.ToLower(name) {
	case "immolate":
		if c.char.Immolate.Active && c.char.Immolate.ExpiresAt > c.char.CurrentTime {
			return c.char.Immolate.ExpiresAt - c.char.CurrentTime
		}
	case "curse_of_the_elements":
		if c.char.CurseOfElements.Active && c.char.CurseOfElements.ExpiresAt > c.char.CurrentTime {
			return c.char.CurseOfElements.ExpiresAt - c.char.CurrentTime
		}
	case "corruption":
		if c.char.Corruption.Active && c.char.Corruption.ExpiresAt > c.char.CurrentTime {
			return c.char.Corruption.ExpiresAt - c.char.CurrentTime
		}
	case "curse_of_agony":
		if c.char.CurseOfAgony.Active && c.char.CurseOfAgony.ExpiresAt > c.char.CurrentTime {
			return c.char.CurseOfAgony.ExpiresAt - c.char.CurrentTime
		}
	}
	return 0
}

func (c *rotationContext) ResourcePercent(resource string) float64 {
	switch strings.ToLower(resource) {
	case "mana":
		if c.char.Stats.MaxMana <= 0 {
			return 0
		}
		return c.char.Resources.CurrentMana / c.char.Stats.MaxMana
	default:
		return 0
	}
}

func (c *rotationContext) CooldownReady(name string) bool {
	switch strings.ToLower(name) {
	case "conflagrate":
		return c.char.IsCooldownReady(&c.char.Conflagrate)
	case "chaos_bolt":
		return c.char.IsCooldownReady(&c.char.ChaosBolt)
	case "shadowburn":
		return c.char.IsCooldownReady(&c.char.Shadowburn)
	case "shadowfury":
		return c.char.IsCooldownReady(&c.char.Shadowfury)
	default:
		return true
	}
}

func (c *rotationContext) CooldownRemaining(name string) time.Duration {
	switch strings.ToLower(name) {
	case "conflagrate":
		if c.char.IsCooldownReady(&c.char.Conflagrate) {
			return 0
		}
		if c.char.Conflagrate.ReadyAt > c.char.CurrentTime {
			return c.char.Conflagrate.ReadyAt - c.char.CurrentTime
		}
	case "chaos_bolt":
		if c.char.IsCooldownReady(&c.char.ChaosBolt) {
			return 0
		}
		if c.char.ChaosBolt.ReadyAt > c.char.CurrentTime {
			return c.char.ChaosBolt.ReadyAt - c.char.CurrentTime
		}
	case "shadowburn":
		if c.char.IsCooldownReady(&c.char.Shadowburn) {
			return 0
		}
		if c.char.Shadowburn.ReadyAt > c.char.CurrentTime {
			return c.char.Shadowburn.ReadyAt - c.char.CurrentTime
		}
	case "shadowfury":
		if c.char.IsCooldownReady(&c.char.Shadowfury) {
			return 0
		}
		if c.char.Shadowfury.ReadyAt > c.char.CurrentTime {
			return c.char.Shadowfury.ReadyAt - c.char.CurrentTime
		}
	}
	return 0
}

func (c *rotationContext) getBuff(name string) *character.Buff {
	switch strings.ToLower(name) {
	case "pyroclasm":
		return &c.char.Pyroclasm
	case "backdraft":
		return &c.char.Backdraft
	case "improved_soul_leech", "soul_leech":
		return &c.char.ImprovedSoulLeech
	case "life_tap_buff":
		return &c.char.LifeTapBuff
	case "shadow_trance":
		return &c.char.ShadowTrance
	default:
		return nil
	}
}

func spellFromName(name string) (spells.SpellType, bool) {
	switch strings.ToLower(name) {
	case "immolate":
		return spells.SpellImmolate, true
	case "conflagrate":
		return spells.SpellConflagrate, true
	case "chaos_bolt":
		return spells.SpellChaosBolt, true
	case "incinerate":
		return spells.SpellIncinerate, true
	case "life_tap":
		return spells.SpellLifeTap, true
	case "soul_fire":
		return spells.SpellSoulFire, true
	case "curse_of_the_elements":
		return spells.SpellCurseOfElements, true
	case "shadow_bolt":
		return spells.SpellShadowBolt, true
	case "shadowburn":
		return spells.SpellShadowburn, true
	case "corruption":
		return spells.SpellCorruption, true
	case "curse_of_agony":
		return spells.SpellCurseOfAgony, true
	case "shadowfury":
		return spells.SpellShadowfury, true
	case "shadow_crash":
		return spells.SpellShadowCrash, true
	default:
		return 0, false
	}
}

func spellTypeName(spell spells.SpellType) string {
	switch spell {
	case spells.SpellImmolate:
		return "Immolate"
	case spells.SpellConflagrate:
		return "Conflagrate"
	case spells.SpellChaosBolt:
		return "Chaos Bolt"
	case spells.SpellIncinerate:
		return "Incinerate"
	case spells.SpellLifeTap:
		return "Life Tap"
	case spells.SpellSoulFire:
		return "Soul Fire"
	case spells.SpellCurseOfElements:
		return "Curse of the Elements"
	case spells.SpellShadowBolt:
		return "Shadow Bolt"
	case spells.SpellShadowburn:
		return "Shadowburn"
	case spells.SpellShadowfury:
		return "Shadowfury"
	case spells.SpellCorruption:
		return "Corruption"
	case spells.SpellCurseOfAgony:
		return "Curse of Agony"
	case spells.SpellShadowCrash:
		return "Shadow Crash"
	default:
		return "Unknown"
	}
}

func captureBuffState(char *character.Character) buffState {
	catStacks := 0
	if char.CataclysmicBurst != nil {
		catStacks = char.CataclysmicBurst.Stacks()
	}
	heatStacks := 0
	heatExpires := time.Duration(0)
	if char.HeatingUp != nil {
		heatStacks = char.HeatingUp.Stacks()
		heatExpires = char.HeatingUp.ExpiresAt()
	}
	guldansActive := false
	guldansExpires := time.Duration(0)
	if char.GuldansChosen != nil {
		guldansActive = char.GuldansChosen.ActiveAt(char.CurrentTime)
		guldansExpires = char.GuldansChosen.ExpiresAt()
	}
	return buffState{
		pyroActive:          char.Pyroclasm.Active,
		pyroExpires:         char.Pyroclasm.ExpiresAt,
		backdraftActive:     char.Backdraft.Active,
		backdraftCharges:    char.Backdraft.Charges,
		soulActive:          char.ImprovedSoulLeech.Active,
		empImpActive:        char.EmpoweredImp.Active,
		lifeTapActive:       char.LifeTapBuff.Active,
		lifeTapExpires:      char.LifeTapBuff.ExpiresAt,
		heatingStacks:       heatStacks,
		heatingExpires:      heatExpires,
		catBurstStacks:      catStacks,
		guldansActive:       guldansActive,
		guldansExpires:      guldansExpires,
		shadowTranceActive:  char.ShadowTrance.Active,
		shadowTranceExpires: char.ShadowTrance.ExpiresAt,
	}
}

func (s *Simulator) executeRotation(char *character.Character, result *SimulationResult, spellEngine *spells.Engine) bool {
	if s.Rotation == nil || len(s.Rotation.Actions) == 0 {
		return false
	}
	ctx := &rotationContext{sim: s, char: char}
	for _, action := range s.Rotation.Actions {
		if action == nil {
			continue
		}
		if action.Condition != nil && !action.Condition.Eval(ctx) {
			continue
		}
		switch action.Type {
		case apl.ActionCastSpell:
			spell, ok := spellFromName(action.Spell)
			if !ok {
				continue
			}
			if s.tryCast(char, spell, result, spellEngine) {
				return true
			}
		case apl.ActionMacro:
			for _, step := range action.Steps {
				if step == nil {
					continue
				}
				if step.Condition != nil && !step.Condition.Eval(ctx) {
					continue
				}
				switch step.Type {
				case apl.ActionCastSpell:
					spell, ok := spellFromName(step.Spell)
					if !ok {
						continue
					}
					if s.tryCast(char, spell, result, spellEngine) {
						return true
					}
				case apl.ActionWait:
					if step.Duration <= 0 {
						continue
					}
					s.wait(char, step.Duration, result, spellEngine)
					return true
				default:
					continue
				}
			}
		case apl.ActionWait:
			if action.Duration <= 0 {
				continue
			}
			s.wait(char, action.Duration, result, spellEngine)
			return true
		default:
			// use_item not implemented yet
			continue
		}
	}
	return false
}
