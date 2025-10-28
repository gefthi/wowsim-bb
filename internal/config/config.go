package config

import (
	"os"
	
	"gopkg.in/yaml.v3"
)

// Constants holds server constants
type Constants struct {
	Server struct {
		Level int    `yaml:"level"`
		Name  string `yaml:"name"`
	} `yaml:"server"`
	StatConversions struct {
		CritRatingPerPercent  int `yaml:"crit_rating_per_percent"`
		HasteRatingPerPercent int `yaml:"haste_rating_per_percent"`
	} `yaml:"stat_conversions"`
	HitMechanics struct {
		BossHitCap          int `yaml:"boss_hit_cap"`
		EqualLevelMissChance int `yaml:"equal_level_miss_chance"`
	} `yaml:"hit_mechanics"`
	GCD struct {
		Base    float64 `yaml:"base"`
		Minimum float64 `yaml:"minimum"`
	} `yaml:"gcd"`
}

// Spells holds all spell data
type Spells struct {
	Immolate struct {
		DirectDamage       float64 `yaml:"direct_damage"`
		DotDamage          float64 `yaml:"dot_damage"`
		DotDuration        float64 `yaml:"dot_duration"`
		DotTicks           int     `yaml:"dot_ticks"`
		CastTime           float64 `yaml:"cast_time"`
		ManaCost           float64 `yaml:"mana_cost"`
		SPCoefficientDirect float64 `yaml:"sp_coefficient_direct"`
		SPCoefficientDot   float64 `yaml:"sp_coefficient_dot"`
	} `yaml:"immolate"`
	Incinerate struct {
		BaseDamageMin      float64 `yaml:"base_damage_min"`
		BaseDamageMax      float64 `yaml:"base_damage_max"`
		ImmolateBonusMin   float64 `yaml:"immolate_bonus_min"`
		ImmolateBonusMax   float64 `yaml:"immolate_bonus_max"`
		CastTime           float64 `yaml:"cast_time"`
		ManaCost           float64 `yaml:"mana_cost"`
		SPCoefficient      float64 `yaml:"sp_coefficient"`
	} `yaml:"incinerate"`
	ChaosBolt struct {
		BaseDamageMin float64 `yaml:"base_damage_min"`
		BaseDamageMax float64 `yaml:"base_damage_max"`
		CastTime      float64 `yaml:"cast_time"`
		Cooldown      float64 `yaml:"cooldown"`
		ManaCost      float64 `yaml:"mana_cost"`
		SPCoefficient float64 `yaml:"sp_coefficient"`
	} `yaml:"chaos_bolt"`
	Conflagrate struct {
		ImmolateDotPercentage   float64 `yaml:"immolate_dot_percentage"`
		ConflagDotPercentage    float64 `yaml:"conflag_dot_percentage"`
		CastTime                float64 `yaml:"cast_time"`
		Cooldown                float64 `yaml:"cooldown"`
		ManaCost                float64 `yaml:"mana_cost"`
		SPCoefficient           float64 `yaml:"sp_coefficient"`
	} `yaml:"conflagrate"`
	LifeTap struct {
		CastTime                float64 `yaml:"cast_time"`
		Cooldown                float64 `yaml:"cooldown"`
		ManaCost                float64 `yaml:"mana_cost"`
		HealthBase              float64 `yaml:"health_base"`
		SpiritMultiplier        float64 `yaml:"spirit_multiplier"`
		ManaBase                float64 `yaml:"mana_base"`
		SpellpowerCoefficient   float64 `yaml:"spellpower_coefficient"`
		ImprovedLifetapPerRank  float64 `yaml:"improved_lifetap_per_rank"`
	} `yaml:"life_tap"`
}

// Talents holds talent modifiers
type Talents struct {
	Emberstorm struct {
		DamageMultiplier float64 `yaml:"damage_multiplier"`
	} `yaml:"emberstorm"`
	ImprovedImmolate struct {
		DamageMultiplier float64 `yaml:"damage_multiplier"`
	} `yaml:"improved_immolate"`
	Aftermath struct {
		DotDamageMultiplier float64 `yaml:"dot_damage_multiplier"`
	} `yaml:"aftermath"`
	FireAndBrimstone struct {
		ImmolateTargetDamage float64 `yaml:"immolate_target_damage"`
		ConflagrateCritBonus float64 `yaml:"conflagrate_crit_bonus"`
	} `yaml:"fire_and_brimstone"`
	Ruin struct {
		CritMultiplier float64 `yaml:"crit_multiplier"`
	} `yaml:"ruin"`
	ShadowAndFlame struct {
		BonusSPPercentage float64 `yaml:"bonus_sp_percentage"`
	} `yaml:"shadow_and_flame"`
	Devastation struct {
		CritBonus float64 `yaml:"crit_bonus"`
	} `yaml:"devastation"`
	Backlash struct {
		CritBonus float64 `yaml:"crit_bonus"`
	} `yaml:"backlash"`
}

// Player holds player character configuration
type Player struct {
	Character struct {
		Name  string `yaml:"name"`
		Level int    `yaml:"level"`
	} `yaml:"character"`
	Stats struct {
		SpellPower   float64 `yaml:"spell_power"`
		CritPercent  float64 `yaml:"crit_percent"`
		HastePercent float64 `yaml:"haste_percent"`
		Spirit       float64 `yaml:"spirit"`
		HitPercent   float64 `yaml:"hit_percent"`
		MaxMana      float64 `yaml:"max_mana"`
	} `yaml:"stats"`
	Target struct {
		Type  string `yaml:"type"`
		Level int    `yaml:"level"`
	} `yaml:"target"`
	Simulation struct {
		DurationSeconds int `yaml:"duration_seconds"`
		Iterations      int `yaml:"iterations"`
	} `yaml:"simulation"`
}

// Config holds all configuration
type Config struct {
	Constants Constants
	Spells    Spells
	Talents   Talents
	Player    Player
}

// LoadConfig loads all YAML configuration files
func LoadConfig(configDir string) (*Config, error) {
	cfg := &Config{}
	
	// Load constants
	data, err := os.ReadFile(configDir + "/constants.yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg.Constants); err != nil {
		return nil, err
	}
	
	// Load spells
	data, err = os.ReadFile(configDir + "/spells.yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg.Spells); err != nil {
		return nil, err
	}
	
	// Load talents
	data, err = os.ReadFile(configDir + "/talents.yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg.Talents); err != nil {
		return nil, err
	}
	
	// Load player
	data, err = os.ReadFile(configDir + "/player.yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg.Player); err != nil {
		return nil, err
	}
	
	return cfg, nil
}
