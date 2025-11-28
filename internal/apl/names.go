package apl

import (
	"fmt"
	"strings"
)

// NOTE: keep these lists in sync with talents/spells we expose to the APL.
var (
	knownSpells = map[string]struct{}{
		"immolate":              {},
		"conflagrate":           {},
		"chaos_bolt":            {},
		"incinerate":            {},
		"life_tap":              {},
		"soul_fire":             {},
		"inferno":               {},
		"curse_of_doom":         {},
		"curse_of_agony":        {},
		"curse_of_the_elements": {},
	}
	knownBuffs = map[string]struct{}{
		"pyroclasm":           {},
		"backdraft":           {},
		"guldans_chosen":      {},
		"cataclysmic_burst":   {},
		"heating_up":          {},
		"decisive_decimation": {},
		"improved_soul_leech": {},
		"soul_leech":          {},
		"life_tap_buff":       {},
		"shadow_trance":       {},
		"demonic_soul":        {},
	}
	knownDebuffs = map[string]struct{}{
		"immolate":              {},
		"curse_of_doom":         {},
		"curse_of_agony":        {},
		"curse_of_the_elements": {},
	}
	knownResources = map[string]struct{}{
		"mana":        {},
		"health":      {},
		"soul_shards": {},
	}
)

// KnownSpells returns the set of valid spell identifiers.
func KnownSpells() map[string]struct{} {
	return copySet(knownSpells)
}

// KnownBuffs returns the set of valid buff identifiers.
func KnownBuffs() map[string]struct{} {
	return copySet(knownBuffs)
}

// KnownDebuffs returns the set of valid debuff identifiers.
func KnownDebuffs() map[string]struct{} {
	return copySet(knownDebuffs)
}

// KnownResources returns the set of valid resource identifiers.
func KnownResources() map[string]struct{} {
	return copySet(knownResources)
}

func copySet(src map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func validateSpellName(name string) (string, error) {
	n := normalizeName(name)
	if n == "" {
		return n, fmt.Errorf("spell name missing")
	}
	if _, ok := knownSpells[n]; !ok {
		return "", fmt.Errorf("unknown spell '%s'", name)
	}
	return n, nil
}

func validateBuffName(name string) (string, error) {
	n := normalizeName(name)
	if n == "" {
		return n, fmt.Errorf("buff name missing")
	}
	if _, ok := knownBuffs[n]; !ok {
		return "", fmt.Errorf("unknown buff '%s'", name)
	}
	return n, nil
}

func validateDebuffName(name string) (string, error) {
	n := normalizeName(name)
	if n == "" {
		return n, fmt.Errorf("debuff name missing")
	}
	if _, ok := knownDebuffs[n]; !ok {
		return "", fmt.Errorf("unknown debuff '%s'", name)
	}
	return n, nil
}

func validateResourceName(name string) (string, error) {
	n := normalizeName(name)
	if n == "" {
		return n, fmt.Errorf("resource name missing")
	}
	if _, ok := knownResources[n]; !ok {
		return "", fmt.Errorf("unknown resource '%s'", name)
	}
	return n, nil
}

func validateCooldownName(name string) (string, error) {
	// cooldown names map to spells for now
	return validateSpellName(name)
}
