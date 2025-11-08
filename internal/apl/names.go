package apl

import (
	"fmt"
	"strings"
)

// NOTE: keep these lists in sync with talents/spells we expose to the APL.
var (
	knownSpells = map[string]struct{}{
		"immolate":        {},
		"conflagrate":     {},
		"chaos_bolt":      {},
		"incinerate":      {},
		"life_tap":        {},
		"inferno":         {},
		"curse_of_doom":   {},
		"curse_of_agony":  {},
		"curse_of_the_elements": {},
	}
	knownBuffs = map[string]struct{}{
		"pyroclasm":           {},
		"backdraft":           {},
		"improved_soul_leech": {},
		"soul_leech":          {},
		"shadow_trance":       {},
		"demonic_soul":        {},
	}
	knownDebuffs = map[string]struct{}{
		"immolate":        {},
		"curse_of_doom":   {},
		"curse_of_agony":  {},
		"curse_of_the_elements": {},
	}
	knownResources = map[string]struct{}{
		"mana":   {},
		"health": {},
		"soul_shards": {},
	}
)

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
