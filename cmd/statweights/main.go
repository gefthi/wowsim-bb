package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
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

type sweepConfig struct {
	stat         string
	start        float64
	stop         float64
	step         float64
	avgSeeds     int
	concurrency  int
	includeDelta bool
	outputDir    string
}

type sweepPointResult struct {
	index int
	value float64
	dps   float64
}

func main() {
	configDir := flag.String("config-dir", "./configs", "Path to config directory")
	rotationFlag := flag.String("rotation", "", "Rotation file name (defaults to player.yaml value)")
	iterations := flag.Int("iterations", 0, "Iterations (0 = use player.yaml)")
	seedBase := flag.Int64("seed-base", 0, "Base RNG seed (0 = random)")
	verbose := flag.Bool("verbose", false, "Show plus/minus DPS columns")
	sweepStat := flag.String("stat", "", "Stat to sweep (crit|haste|sp). If set, runs sweep mode instead of central-diff weights.")
	sweepStart := flag.Float64("start", math.NaN(), "Sweep start (percent for crit/haste, raw for spell power). Defaults depend on stat.")
	sweepStop := flag.Float64("stop", math.NaN(), "Sweep stop (percent for crit/haste, raw for spell power). Defaults depend on stat.")
	sweepStep := flag.Float64("step", math.NaN(), "Sweep step (percent for crit/haste, raw for spell power). Defaults depend on stat.")
	sweepConcurrency := flag.Int("concurrency", 0, "Concurrent sims for sweep (0 = num CPU).")
	sweepAvgSeeds := flag.Int("avg-seeds", 1, "Number of seeds to average per sweep point (>=1).")
	includeDelta := flag.Bool("deltas", true, "Include DPS-per-point delta column in sweep CSV.")
	outputDir := flag.String("output-dir", "output/stat_curves", "Directory for sweep CSV output.")
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

	if *sweepStat != "" {
		sweepCfg, err := buildSweepConfig(*sweepStat, *sweepStart, *sweepStop, *sweepStep, *sweepConcurrency, *sweepAvgSeeds, *includeDelta, *outputDir, baseStats)
		if err != nil {
			log.Fatalf("Sweep config error: %v", err)
		}
		if err := runSweep(cfg, simCfg, compiledRotation, baseStats, baseSeed, sweepCfg); err != nil {
			log.Fatalf("Sweep failed: %v", err)
		}
		return
	}

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

	var (
		spWeight   float64
		critWeight float64
		hasteWeight float64
		hitWeight  float64
		spDelta    float64
		critDelta  float64
		hasteDelta float64
		hitDelta   float64
	)
	for _, res := range results {
		if res.delta.name == "Spell Power" {
			spWeight = res.weight
			spDelta = res.delta.delta
		}
		if res.delta.name == "Crit" {
			critWeight = res.weight
			critDelta = res.delta.delta
		}
		if res.delta.name == "Haste" {
			hasteWeight = res.weight
			hasteDelta = res.delta.delta
		}
		if res.delta.name == "Hit" {
			hitWeight = res.weight
			hitDelta = res.delta.delta
		}
	}
	if spWeight != 0 {
		// Weight calculations are per-unit already; keep explicit SP-per-point for clarity.
		spPerPoint := spWeight
		if spDelta != 0 {
			spPerPoint = spWeight // weight is already per 1 SP
		}

		nw := tabWriter()
		fmt.Fprintf(nw, "\nNormalized (SP = 1.0)\n")
		fmt.Fprintf(nw, "Stat\tDelta\tWeight vs SP\n")
		for _, res := range results {
			fmt.Fprintf(nw, "%s\t%+.0f %s\t%.3f\n",
				res.delta.name, res.delta.delta, res.delta.unit, res.weight/spPerPoint)
		}
		fmt.Fprintf(nw, "%s\t%+d %s\t%.3f\n", "Spirit", 1, "Spirit", 0.6) // Hardcoded: 1 Spirit worth 0.6 SP
		nw.Flush()

		// Pawn string assumes 1% crit = 14 rating, 1% haste = 10 rating, 1% hit = 10 rating.
		const critRatingPerPercent = 14.0
		const hasteRatingPerPercent = 10.0
		const hitRatingPerPercent = 10.0

		critPerRating := 0.0
		if critWeight != 0 && critDelta != 0 {
			critPerRating = (critWeight / critDelta) / critRatingPerPercent / spPerPoint
		}
		hastePerRating := 0.0
		if hasteWeight != 0 && hasteDelta != 0 {
			hastePerRating = (hasteWeight / hasteDelta) / hasteRatingPerPercent / spPerPoint
		}
		hitPerRating := 0.0
		if hitWeight != 0 && hitDelta != 0 {
			hitPerRating = (hitWeight / hitDelta) / hitRatingPerPercent / spPerPoint
		}

		const spiritWeightVsSP = 0.6

		fmt.Printf("\nPawn: v1: \"StatWeights (Sim)\": SpellPower=1, CritRating=%.2f, HasteRating=%.2f, HitRating=%.2f, Spirit=%.2f\n",
			critPerRating, hastePerRating, hitPerRating, spiritWeightVsSP)
	}
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

func buildSweepConfig(stat string, start, stop, step float64, concurrency, avgSeeds int, includeDelta bool, outputDir string, baseStats character.Stats) (sweepConfig, error) {
	cfg := sweepConfig{
		stat:         stat,
		start:        start,
		stop:         stop,
		step:         step,
		concurrency:  concurrency,
		avgSeeds:     avgSeeds,
		includeDelta: includeDelta,
		outputDir:    outputDir,
	}

	switch cfg.stat {
	case "crit", "Crit", "CRIT":
		cfg.stat = "crit"
		if math.IsNaN(cfg.start) {
			cfg.start = 0
		}
		if math.IsNaN(cfg.stop) {
			cfg.stop = 50
		}
		if math.IsNaN(cfg.step) {
			cfg.step = 0.5
		}
	case "haste", "Haste", "HASTE":
		cfg.stat = "haste"
		if math.IsNaN(cfg.start) {
			cfg.start = 0
		}
		if math.IsNaN(cfg.stop) {
			cfg.stop = 40
		}
		if math.IsNaN(cfg.step) {
			cfg.step = 0.5
		}
	case "sp", "SP", "Sp", "spellpower", "spell_power":
		cfg.stat = "sp"
		if math.IsNaN(cfg.start) {
			cfg.start = baseStats.SpellPower
		}
		if math.IsNaN(cfg.stop) {
			cfg.stop = baseStats.SpellPower + 800
		}
		if math.IsNaN(cfg.step) {
			cfg.step = 10
		}
	default:
		return sweepConfig{}, fmt.Errorf("unsupported stat %q (use crit|haste|sp)", stat)
	}

	if cfg.step <= 0 {
		return sweepConfig{}, fmt.Errorf("step must be > 0 (got %.2f)", cfg.step)
	}
	if cfg.stop <= cfg.start {
		return sweepConfig{}, fmt.Errorf("stop must be > start (start=%.2f, stop=%.2f)", cfg.start, cfg.stop)
	}
	if cfg.concurrency <= 0 {
		cfg.concurrency = runtime.NumCPU()
	}
	if cfg.avgSeeds < 1 {
		cfg.avgSeeds = 1
	}
	return cfg, nil
}

func runSweep(cfg *config.Config, simCfg engine.SimulationConfig, rotation *apl.CompiledRotation, baseStats character.Stats, baseSeed int64, sweepCfg sweepConfig) error {
	values := make([]float64, 0)
	for v := sweepCfg.start; v <= sweepCfg.stop+1e-9; v += sweepCfg.step {
		values = append(values, v)
	}
	if len(values) == 0 {
		return fmt.Errorf("no sweep points generated")
	}

	seeds := make([][]int64, len(values))
	rng := rand.New(rand.NewSource(baseSeed))
	for i := range values {
		seeds[i] = make([]int64, sweepCfg.avgSeeds)
		for j := 0; j < sweepCfg.avgSeeds; j++ {
			seeds[i][j] = rng.Int63()
		}
	}

	jobs := make(chan sweepPointResult, len(values))
	results := make([]sweepPointResult, len(values))
	var wg sync.WaitGroup
	wg.Add(sweepCfg.concurrency)

	for w := 0; w < sweepCfg.concurrency; w++ {
		go func() {
			defer wg.Done()
			for job := range jobs {
				var total float64
				for _, seed := range seeds[job.index] {
					stats := applyStat(baseStats, sweepCfg.stat, job.value)
					total += runDPS(cfg, simCfg, rotation, stats, seed)
				}
				results[job.index] = sweepPointResult{
					index: job.index,
					value: job.value,
					dps:   total / float64(sweepCfg.avgSeeds),
				}
			}
		}()
	}

	for i, v := range values {
		jobs <- sweepPointResult{index: i, value: v}
	}
	close(jobs)
	wg.Wait()

	if err := os.MkdirAll(sweepCfg.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}
	outPath := filepath.Join(sweepCfg.outputDir, fmt.Sprintf("%s.csv", sweepCfg.stat))
	file, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outPath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	header := []string{"stat_value", "dps"}
	if sweepCfg.includeDelta {
		header = append(header, "dps_per_point")
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for i, res := range results {
		record := []string{
			fmt.Sprintf("%.4f", res.value),
			fmt.Sprintf("%.4f", res.dps),
		}
		if sweepCfg.includeDelta {
			if i == 0 {
				record = append(record, "")
			} else {
				prev := results[i-1]
				delta := (res.dps - prev.dps) / (res.value - prev.value)
				record = append(record, fmt.Sprintf("%.6f", delta))
			}
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush csv: %w", err)
	}

	fmt.Printf("Sweep complete (%s): %d points, seeds/point=%d, output=%s\n", sweepCfg.stat, len(results), sweepCfg.avgSeeds, outPath)
	return nil
}

func applyStat(base character.Stats, stat string, value float64) character.Stats {
	s := base
	switch stat {
	case "crit":
		s.CritPct = value
	case "haste":
		s.HastePct = value
	case "sp":
		s.SpellPower = value
	}
	return s
}
