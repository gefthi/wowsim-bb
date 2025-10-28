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
	fmt.Println("WotLK Destruction Warlock Simulator - Phase 1 MVP")
	fmt.Println("==================================================")
	fmt.Println()
	
	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create character with example stats
	// These are placeholder values - will be configurable in Phase 2
	charStats := character.Stats{
		SpellPower: 800,
		CritPct:    25.0,  // 25% crit
		HastePct:   0.0,   // No haste in Phase 1
		Spirit:     200,
		HitPct:     17.0,  // Hit capped for boss
		MaxMana:    8000,
	}
	
	char := character.NewCharacter(charStats)
	
	fmt.Println("Character Stats:")
	fmt.Printf("  Spell Power: %.0f\n", char.Stats.SpellPower)
	fmt.Printf("  Crit: %.1f%%\n", char.Stats.CritPct)
	fmt.Printf("  Haste: %.1f%%\n", char.Stats.HastePct)
	fmt.Printf("  Spirit: %.0f\n", char.Stats.Spirit)
	fmt.Printf("  Hit: %.1f%%\n", char.Stats.HitPct)
	fmt.Printf("  Max Mana: %.0f\n", char.Stats.MaxMana)
	fmt.Println()
	
	// Configure simulation
	simConfig := engine.SimulationConfig{
		Duration:   5 * time.Minute, // 5 minute fight
		Iterations: 1000,             // 1000 iterations for good average
		IsBoss:     true,             // Boss target (17% hit cap)
	}
	
	fmt.Printf("Simulation Config:\n")
	fmt.Printf("  Fight Duration: %.0f seconds\n", simConfig.Duration.Seconds())
	fmt.Printf("  Iterations: %d\n", simConfig.Iterations)
	fmt.Printf("  Target: %s\n", map[bool]string{true: "Boss (+3 levels)", false: "Equal Level"}[simConfig.IsBoss])
	fmt.Println()
	
	fmt.Println("Running simulation...")
	fmt.Println()
	
	// Create and run simulator
	sim := engine.NewSimulator(cfg, simConfig, time.Now().UnixNano())
	result := sim.Run(char)
	
	// Print results
	result.PrintResults()
}
