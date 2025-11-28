package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"
	"time"

	"wotlk-destro-sim/internal/apl"
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/engine"
)

type statDelta struct {
	name  string
	unit  string
	delta float64
	apply func(*character.Stats, float64)
}

type weightResult struct {
	delta    statDelta
	weight   float64
	dpsPlus  float64
	dpsMinus float64
}

func main() {
	configDir := flag.String("config-dir", "./configs", "Path to config directory")
	rotationFlag := flag.String("rotation", "", "Rotation file name (defaults to player.yaml value)")
	iterations := flag.Int("iterations", 0, "Iterations (0 = use player.yaml)")
	seedBase := flag.Int64("seed-base", 0, "Base RNG seed (0 = random)")
	verbose := flag.Bool("verbose", false, "Show plus/minus DPS columns")
	flag.Parse()

	cfg, err := config.LoadConfig(*configDir)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	rotationFile := *rotationFlag
	if rotationFile == "" {
		rotationFile = cfg.Player.Rotation
		if rotationFile == "" {
			rotationFile = "destruction-default.yaml"
		}
	}
	rotationDir := filepath.Join(*configDir, "rotations")
	rotRaw, err := apl.LoadRotation(rotationDir, rotationFile)
	if err != nil {
		log.Fatalf("Failed to load rotation %s: %v", filepath.Join(rotationDir, rotationFile), err)
	}
	compiledRotation, err := apl.Compile(rotRaw)
	if err != nil {
		log.Fatalf("Failed to compile rotation: %v", err)
	}

	simCfg := simulationConfigFromPlayer(cfg)
	if *iterations > 0 {
		simCfg.Iterations = *iterations
	}

	baseSeed := *seedBase
	if baseSeed == 0 {
		baseSeed = time.Now().UnixNano()
	}

	baseStats := statsFromPlayer(cfg)
	baselineDPS := runDPS(cfg, simCfg, compiledRotation, baseStats, baseSeed)

	fmt.Printf("Stat Weights (central diff, shared seed %d)\n", baseSeed)
	fmt.Printf("Rotation: %s\n", rotationFile)
	fmt.Printf("Iterations: %d, Duration: %.0fs\n\n", simCfg.Iterations, simCfg.Duration.Seconds())
	fmt.Printf("Baseline DPS: %.2f\n\n", baselineDPS)

	deltas := []statDelta{
		{name: "Spell Power", unit: "SP", delta: 10, apply: func(s *character.Stats, d float64) { s.SpellPower += d }},
		{name: "Crit", unit: "% crit", delta: 1, apply: func(s *character.Stats, d float64) { s.CritPct += d }},
		{name: "Hit", unit: "% hit", delta: 1, apply: func(s *character.Stats, d float64) { s.HitPct += d }},
		{name: "Haste", unit: "% haste", delta: 1, apply: func(s *character.Stats, d float64) { s.HastePct += d }},
	}

	w := tabWriter()
	if *verbose {
		fmt.Fprintf(w, "Stat\tDelta\tDPS/Unit\tPlus DPS\tMinus DPS\n")
	} else {
		fmt.Fprintf(w, "Stat\tDelta\tDPS/Unit\n")
	}
	results := make([]weightResult, len(deltas))
	var wg sync.WaitGroup
	for i, sd := range deltas {
		wg.Add(1)
		go func(i int, sd statDelta) {
			defer wg.Done()
			plus := baseStats
			minus := baseStats
			sd.apply(&plus, sd.delta)
			sd.apply(&minus, -sd.delta)

			dpsPlus := runDPS(cfg, simCfg, compiledRotation, plus, baseSeed)
			dpsMinus := runDPS(cfg, simCfg, compiledRotation, minus, baseSeed)

			weight := (dpsPlus - dpsMinus) / (2 * sd.delta)
			results[i] = weightResult{
				delta:    sd,
				weight:   weight,
				dpsPlus:  dpsPlus,
				dpsMinus: dpsMinus,
			}
		}(i, sd)
	}
	wg.Wait()

	for _, res := range results {
		if *verbose {
			fmt.Fprintf(w, "%s\t%+.0f %s\t%.2f\t%.2f\t%.2f\n",
				res.delta.name, res.delta.delta, res.delta.unit, res.weight, res.dpsPlus, res.dpsMinus)
		} else {
			fmt.Fprintf(w, "%s\t%+.0f %s\t%.2f\n",
				res.delta.name, res.delta.delta, res.delta.unit, res.weight)
		}
	}
	w.Flush()
}

func runDPS(cfg *config.Config, simCfg engine.SimulationConfig, rotation *apl.CompiledRotation, stats character.Stats, seed int64) float64 {
	char := character.NewCharacter(stats)
	sim := engine.NewSimulator(cfg, simCfg, rotation, seed, false, nil)
	result := sim.Run(char)
	return result.TotalDPS
}

func statsFromPlayer(cfg *config.Config) character.Stats {
	return character.Stats{
		Intellect:  cfg.Player.Stats.Intellect,
		SpellPower: cfg.Player.Stats.SpellPower,
		CritPct:    cfg.Player.Stats.CritPercent,
		HastePct:   cfg.Player.Stats.HastePercent,
		Spirit:     cfg.Player.Stats.Spirit,
		HitPct:     cfg.Player.Stats.HitPercent,
		MaxMana:    cfg.Player.Stats.MaxMana,
	}
}

func simulationConfigFromPlayer(cfg *config.Config) engine.SimulationConfig {
	return engine.SimulationConfig{
		Duration:   time.Duration(cfg.Player.Simulation.DurationSeconds) * time.Second,
		Iterations: cfg.Player.Simulation.Iterations,
		IsBoss:     cfg.Player.Target.Type == "boss",
	}
}

// tabWriter creates a tab-aligned writer for consistent table output.
func tabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}
