package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"wotlk-destro-sim/internal/apl"
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/engine"
)

func main() {
	logCombat := flag.Bool("log-combat", false, "Enable combat log mode (forces 1 iteration, 60s duration)")
	seedBase := flag.Int64("seed-base", 0, "Base RNG seed (0 = random)")
	flag.Parse()

	fmt.Println("WotLK Destruction Warlock Simulator - Phase 3")
	fmt.Println("==================================================")
	fmt.Println()

	// Load configuration (now includes player.yaml)
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	rotationDir := "./configs/rotations"
	rotationFile := cfg.Player.Rotation
	if rotationFile == "" {
		rotationFile = "destruction-default.yaml"
	}
	rotRaw, err := apl.LoadRotation(rotationDir, rotationFile)
	if err != nil {
		log.Fatalf("Failed to load rotation %s: %v", filepath.Join(rotationDir, rotationFile), err)
	}
	compiledRotation, err := apl.Compile(rotRaw)
	if err != nil {
		log.Fatalf("Failed to compile rotation: %v", err)
	}

	// Create character from config
	charStats := character.Stats{
		Intellect:  cfg.Player.Stats.Intellect,
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
	fmt.Printf("  Intellect: %.0f\n", char.Stats.Intellect)
	fmt.Printf("  Spell Power: %.0f\n", char.Stats.SpellPower)
	fmt.Printf("  Crit: %.1f%%\n", char.Stats.CritPct)
	fmt.Printf("  Haste: %.1f%%\n", char.Stats.HastePct)
	fmt.Printf("  Spirit: %.0f\n", char.Stats.Spirit)
	fmt.Printf("  Hit: %.1f%%\n", char.Stats.HitPct)
	fmt.Printf("  Max Mana: %.0f\n", char.Stats.MaxMana)
	if cfg.Player.Pet.Summon != "" {
		fmt.Printf("  Pet: %s\n", cfg.Player.Pet.Summon)
	} else {
		fmt.Println("  Pet: none")
	}
	fmt.Println()

	// Configure simulation from YAML
	isBoss := cfg.Player.Target.Type == "boss"
	simConfig := engine.SimulationConfig{
		Duration:   time.Duration(cfg.Player.Simulation.DurationSeconds) * time.Second,
		Iterations: cfg.Player.Simulation.Iterations,
		IsBoss:     isBoss,
	}

	var logWriter io.Writer
	if *logCombat {
		fmt.Println("Combat log mode enabled: forcing 1 iteration; using configured duration.")
		simConfig.Iterations = 1
		logWriter = os.Stdout
	}

	baseSeed := *seedBase
	if baseSeed == 0 {
		baseSeed = time.Now().UnixNano()
	}

	fmt.Printf("Simulation Config:\n")
	fmt.Printf("  Fight Duration: %.0f seconds\n", simConfig.Duration.Seconds())
	fmt.Printf("  Iterations: %d\n", simConfig.Iterations)
	fmt.Printf("  Target: %s (Level %d)\n",
		map[bool]string{true: "Boss", false: "Equal Level"}[simConfig.IsBoss],
		cfg.Player.Target.Level)
	fmt.Printf("  Base Seed: %d\n", baseSeed)
	fmt.Println()

	fmt.Println("Running simulation...")
	fmt.Println()

	// Create and run simulator
	sim := engine.NewSimulator(cfg, simConfig, compiledRotation, baseSeed, *logCombat, logWriter)
	result := sim.Run(char)

	// Print results
	result.PrintResults()
}
