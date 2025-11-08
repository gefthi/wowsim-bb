package main

import (
	"fmt"
	"log"
	"time"

	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/engine"
)

func main() {
	fmt.Println("WotLK Destruction Warlock Simulator - Phase 3")
	fmt.Println("==================================================")
	fmt.Println()

	// Load configuration (now includes player.yaml)
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create character from config
	charStats := character.Stats{
		SpellPower: cfg.Player.Stats.SpellPower,
		CritPct:    cfg.Player.Stats.CritPercent,
		HastePct:   cfg.Player.Stats.HastePercent,
		Spirit:     cfg.Player.Stats.Spirit,
		HitPct:     cfg.Player.Stats.HitPercent,
		MaxMana:    cfg.Player.Stats.MaxMana,
	}

	char := character.NewCharacter(charStats)

	fmt.Printf("Character: %s (Level %d)\n", cfg.Player.Character.Name, cfg.Player.Character.Level)
	fmt.Println("Character Stats:")
	fmt.Printf("  Spell Power: %.0f\n", char.Stats.SpellPower)
	fmt.Printf("  Crit: %.1f%%\n", char.Stats.CritPct)
	fmt.Printf("  Haste: %.1f%%\n", char.Stats.HastePct)
	fmt.Printf("  Spirit: %.0f\n", char.Stats.Spirit)
	fmt.Printf("  Hit: %.1f%%\n", char.Stats.HitPct)
	fmt.Printf("  Max Mana: %.0f\n", char.Stats.MaxMana)
	fmt.Println()

	// Configure simulation from YAML
	isBoss := cfg.Player.Target.Type == "boss"
	simConfig := engine.SimulationConfig{
		Duration:   time.Duration(cfg.Player.Simulation.DurationSeconds) * time.Second,
		Iterations: cfg.Player.Simulation.Iterations,
		IsBoss:     isBoss,
	}

	fmt.Printf("Simulation Config:\n")
	fmt.Printf("  Fight Duration: %.0f seconds\n", simConfig.Duration.Seconds())
	fmt.Printf("  Iterations: %d\n", simConfig.Iterations)
	fmt.Printf("  Target: %s (Level %d)\n",
		map[bool]string{true: "Boss", false: "Equal Level"}[simConfig.IsBoss],
		cfg.Player.Target.Level)
	fmt.Println()

	fmt.Println("Running simulation...")
	fmt.Println()

	// Create and run simulator
	sim := engine.NewSimulator(cfg, simConfig, time.Now().UnixNano())
	result := sim.Run(char)

	// Print results
	result.PrintResults()
}
