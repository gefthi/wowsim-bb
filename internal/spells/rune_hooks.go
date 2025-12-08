package spells

import (
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/runes"
)

func (e *Engine) addPureShadowStack(char *character.Character) {
	if !e.Config.Player.HasRune(runes.RunePureShadow) || char.PureShadow == nil {
		return
	}
	char.PureShadow.AddStacks(char.CurrentTime, 1)
}

func (e *Engine) pureShadowMultiplier(char *character.Character, spell SpellType) float64 {
	if !e.Config.Player.HasRune(runes.RunePureShadow) || char.PureShadow == nil || !char.PureShadow.ActiveAt(char.CurrentTime) {
		return 1
	}
	stacks := char.PureShadow.Stacks()
	switch spell {
	case SpellShadowBolt:
		return 1 + runes.PureShadowShadowBoltBonusPerStack*float64(stacks)
	case SpellShadowfury:
		return 1 + runes.PureShadowShadowfuryBonusPerStack*float64(stacks)
	case SpellShadowburn:
		return 1 // no direct bonus defined
	default:
		return 1
	}
}

func (e *Engine) addDuskTillDawnStack(char *character.Character) {
	if !e.Config.Player.HasRune(runes.RuneDuskTillDawn) || char.DuskTillDawn == nil {
		return
	}
	char.DuskTillDawn.AddStacks(char.CurrentTime, 1)
}

func (e *Engine) consumeDuskTillDawn(char *character.Character) (stacks int) {
	if !e.Config.Player.HasRune(runes.RuneDuskTillDawn) || char.DuskTillDawn == nil {
		return 0
	}
	stacks = char.DuskTillDawn.Stacks()
	if stacks > 0 {
		char.DuskTillDawn.Clear(char.CurrentTime)
	}
	return stacks
}

// applyCorruptionSnapshot sets up the Corruption debuff using a provided snapshot total damage.
func (e *Engine) applyCorruptionSnapshot(char *character.Character, snapshotDamage float64) {
	spellData := e.Config.Spells.Corruption
	tickCount := spellData.DotTicks
	if tickCount <= 0 {
		tickCount = 1
	}
	baseTickDamage := snapshotDamage / float64(tickCount)
	char.Corruption.Active = true
	char.Corruption.ExpiresAt = char.CurrentTime + time.Duration(spellData.DotDuration*float64(time.Second))
	char.Corruption.TickInterval = time.Duration((spellData.DotDuration / float64(tickCount)) * float64(time.Second))
	char.Corruption.LastTick = char.CurrentTime
	char.Corruption.TickDamage = baseTickDamage
	char.Corruption.BaseTickDamage = baseTickDamage
	char.Corruption.SPTickDamage = 0
	char.Corruption.TickCritChance = 0
	char.Corruption.TicksRemaining = tickCount
	char.Corruption.TotalTicks = tickCount
	char.Corruption.SnapshotDotDamage = snapshotDamage
}

// applyCurseOfAgonySnapshot sets up the Curse of Agony debuff using provided base/SP snapshot totals.
func (e *Engine) applyCurseOfAgonySnapshot(char *character.Character, baseSnapshot, spSnapshot float64) {
	spellData := e.Config.Spells.CurseOfAgony
	tickCount := spellData.DotTicks
	if tickCount <= 0 {
		tickCount = 1
	}
	baseTickDamage := baseSnapshot / float64(tickCount)
	spTickDamage := spSnapshot / float64(tickCount)
	stageMultiplier := 0.5 // first 4 ticks
	char.CurseOfAgony.Active = true
	char.CurseOfAgony.ExpiresAt = char.CurrentTime + time.Duration(spellData.DotDuration*float64(time.Second))
	char.CurseOfAgony.TickInterval = time.Duration((spellData.DotDuration / float64(tickCount)) * float64(time.Second))
	char.CurseOfAgony.LastTick = char.CurrentTime
	char.CurseOfAgony.TickDamage = baseTickDamage*stageMultiplier + spTickDamage
	char.CurseOfAgony.BaseTickDamage = baseTickDamage
	char.CurseOfAgony.SPTickDamage = spTickDamage
	char.CurseOfAgony.TickCritChance = 0
	char.CurseOfAgony.TicksRemaining = tickCount
	char.CurseOfAgony.TotalTicks = tickCount
	char.CurseOfAgony.SnapshotDotDamage = baseSnapshot + spSnapshot
}
