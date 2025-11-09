package engine

import (
	"strings"
	"time"

	"wotlk-destro-sim/internal/apl"
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/spells"
)

type rotationContext struct {
	sim  *Simulator
	char *character.Character
}

type buffState struct {
	pyroActive       bool
	pyroExpires      time.Duration
	backdraftActive  bool
	backdraftCharges int
	soulActive       bool
}

func (c *rotationContext) BuffActive(name string) bool {
	buff := c.getBuff(name)
	if buff == nil {
		return false
	}
	if !buff.Active {
		return false
	}
	return buff.ExpiresAt > c.char.CurrentTime
}

func (c *rotationContext) BuffRemaining(name string) time.Duration {
	buff := c.getBuff(name)
	if buff == nil {
		return 0
	}
	if !buff.Active || buff.ExpiresAt <= c.char.CurrentTime {
		return 0
	}
	return buff.ExpiresAt - c.char.CurrentTime
}

func (c *rotationContext) BuffCharges(name string) int {
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
	default:
		return "Unknown"
	}
}

func captureBuffState(char *character.Character) buffState {
	return buffState{
		pyroActive:       char.Pyroclasm.Active,
		pyroExpires:      char.Pyroclasm.ExpiresAt,
		backdraftActive:  char.Backdraft.Active,
		backdraftCharges: char.Backdraft.Charges,
		soulActive:       char.ImprovedSoulLeech.Active,
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
					s.advanceTime(char, step.Duration, result, spellEngine)
					return true
				default:
					continue
				}
			}
		case apl.ActionWait:
			if action.Duration <= 0 {
				continue
			}
			s.advanceTime(char, action.Duration, result, spellEngine)
			return true
		default:
			// use_item not implemented yet
			continue
		}
	}
	return false
}
