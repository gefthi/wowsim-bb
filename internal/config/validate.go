package config

import (
	"fmt"

	"wotlk-destro-sim/internal/runes"
)

func (cfg *Config) validate() error {
	if err := cfg.Player.validate(); err != nil {
		return err
	}
	return nil
}

func (p *Player) validate() error {
	return validateMysticEnchants(&p.MysticEnchants)
}

func validateMysticEnchants(me *MysticEnchantConfig) error {
	if me == nil {
		return nil
	}
	active := map[string]struct{}{}
	check := func(names []string, limit int, expected runes.Rarity) error {
		if limit > 0 && len(names) > limit {
			return fmt.Errorf("mystic_enchants: %s selections exceed limit (%d > %d)", expected, len(names), limit)
		}
		for i, raw := range names {
			name := runes.Normalize(raw)
			names[i] = name
			rarity, ok := runes.RarityOf(name)
			if !ok {
				return fmt.Errorf("mystic_enchants: unknown rune '%s'", raw)
			}
			if rarity != expected {
				return fmt.Errorf("mystic_enchants: rune '%s' is %s but listed under %s", name, rarity, expected)
			}
			if _, dup := active[name]; dup {
				return fmt.Errorf("mystic_enchants: rune '%s' selected more than once", name)
			}
			active[name] = struct{}{}
		}
		return nil
	}
	if err := check(me.Equipped.Legendary, me.Limits.Legendary, runes.RarityLegendary); err != nil {
		return err
	}
	if err := check(me.Equipped.Epic, me.Limits.Epic, runes.RarityEpic); err != nil {
		return err
	}
	if err := check(me.Equipped.Rare, me.Limits.Rare, runes.RarityRare); err != nil {
		return err
	}
	me.active = active
	return nil
}
