package runes

import (
	"strings"
	"time"
)

type Rarity string

const (
	RarityLegendary Rarity = "legendary"
	RarityEpic      Rarity = "epic"
	RarityRare      Rarity = "rare"
)

const (
	RuneDestructionMastery = "destruction_mastery"
	RuneCataclysmicBurst   = "cataclysmic_burst"

	RuneInnerFlame         = "inner_flame"
	RuneEndlessFlames      = "endless_flames"
	RuneHeatingUp          = "heating_up"
	RuneDecisiveDecimation = "decisive_decimation"
	RuneChaosManifesting   = "chaos_manifesting"
	RuneGuldansChosen      = "guldans_chosen"
	RuneAgentOfChaos       = "agent_of_chaos"

	RuneGlyphOfLifeTap     = "glyph_of_life_tap"
	RuneGlyphOfConflagrate = "glyph_of_conflagrate"
	RuneDemonicAegis       = "demonic_aegis"
	RuneSuppression        = "suppression"
	RuneGlyphOfChaosBolt   = "glyph_of_chaos_bolt"
	RuneGlyphOfIncinerate  = "glyph_of_incinerate"
	RuneGlyphOfImmolate    = "glyph_of_immolate"
	RuneImprovedImp        = "improved_imp"
)

var runeRarity = map[string]Rarity{
	RuneDestructionMastery: RarityLegendary,
	RuneCataclysmicBurst:   RarityLegendary,

	RuneInnerFlame:         RarityEpic,
	RuneEndlessFlames:      RarityEpic,
	RuneHeatingUp:          RarityEpic,
	RuneDecisiveDecimation: RarityEpic,
	RuneChaosManifesting:   RarityEpic,
	RuneGuldansChosen:      RarityEpic,
	RuneAgentOfChaos:       RarityEpic,

	RuneGlyphOfLifeTap:     RarityRare,
	RuneGlyphOfConflagrate: RarityRare,
	RuneDemonicAegis:       RarityRare,
	RuneSuppression:        RarityRare,
	RuneGlyphOfChaosBolt:   RarityRare,
	RuneGlyphOfIncinerate:  RarityRare,
	RuneGlyphOfImmolate:    RarityRare,
	RuneImprovedImp:        RarityRare,
}

const (
	DestructionMasteryGlobalBonus   = 1.04
	DestructionMasteryImmolateBonus = 1.05

	CataclysmicBurstStackBonus = 0.08
	CataclysmicBurstMaxStacks  = 4
	CataclysmicBurstExtendSec  = 2.0

	HeatingUpStackBonus  = 0.02
	HeatingUpMaxStacks   = 5
	HeatingUpDurationSec = 15.0

	GuldansChosenDurationSec = 4.0

	AgentOfChaosExtraDurationSec    = 3.0
	AgentOfChaosExtraTicks          = 1
	AgentOfChaosDirectDamagePenalty = 0.5
	AgentOfChaosChaosBoltReduceSec  = 0.5

	GlyphOfLifeTapSpiritMultiplier = 0.20
	GlyphOfLifeTapDurationSec      = 40.0

	GlyphOfChaosBoltCooldownReduction = 2.0
	GlyphOfIncinerateDamageMultiplier = 1.05
	DemonicAegisSpiritBonusPerPoint   = 0.09
	SuppressionHitBonus               = 3.0

	DecisiveDecimationCastReduction = 0.30
)

// Normalize returns the canonical lowercase snake_case rune name.
func Normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// RarityOf returns the rarity and whether the rune is known.
func RarityOf(name string) (Rarity, bool) {
	r, ok := runeRarity[name]
	return r, ok
}

// IsKnown returns true if the rune identifier is recognized.
func IsKnown(name string) bool {
	_, ok := runeRarity[name]
	return ok
}

func HeatingUpMultiplier(active bool, stacks int, expiresAt, now time.Duration) float64 {
	if !active || stacks <= 0 {
		return 1
	}
	if expiresAt <= now {
		return 1
	}
	return 1 + HeatingUpStackBonus*float64(stacks)
}

// KnownRunes returns all rune identifiers keyed by name.
func KnownRunes() map[string]Rarity {
	out := make(map[string]Rarity, len(runeRarity))
	for k, v := range runeRarity {
		out[k] = v
	}
	return out
}
