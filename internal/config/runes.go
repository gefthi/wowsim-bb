package config

import "wotlk-destro-sim/internal/runes"

// Active returns true if the given rune is equipped.
func (me *MysticEnchantConfig) Active(name string) bool {
	if me == nil || me.active == nil {
		return false
	}
	_, ok := me.active[runes.Normalize(name)]
	return ok
}

// HasRune is a convenience helper on Player.
func (p *Player) HasRune(name string) bool {
	return p.MysticEnchants.Active(name)
}
