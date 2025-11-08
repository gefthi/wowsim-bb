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

func (c *rotationContext) BuffActive(name string) bool {
	return false
}

func (c *rotationContext) BuffRemaining(name string) time.Duration {
	return 0
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

func (c *rotationContext) BuffRemainingDuration(name string) time.Duration {
	return 0
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
				if step.Type == apl.ActionCastSpell {
					spell, ok := spellFromName(step.Spell)
					if !ok {
						continue
					}
					if s.tryCast(char, spell, result, spellEngine) {
						return true
					}
				}
			}
		default:
			// wait/use_item not implemented yet
			continue
		}
	}
	return false
}
