package engine

import (
	"fmt"
	"time"
	"wotlk-destro-sim/internal/character"
	"wotlk-destro-sim/internal/config"
	"wotlk-destro-sim/internal/spells"
)

// SimulationConfig holds simulation parameters
type SimulationConfig struct {
	Duration   time.Duration // Fight duration
	Iterations int           // Number of iterations to run
	IsBoss     bool          // Boss target (17% hit cap) vs equal level (4% miss)
}

// SimulationResult holds results from simulation
type SimulationResult struct {
	TotalDPS     float64
	TotalDamage  float64
	Duration     time.Duration
	Iterations   int
	
	// Spell breakdown
	ImmolateCount    int
	IncinerateCount  int
	ChaosBoltCount   int
	ConflagrateCount int
	LifeTapCount     int
	
	ImmolateDamage    float64
	IncinerateDamage  float64
	ChaosBoltDamage   float64
	ConflagrateDamage float64
	
	// Statistics
	MissCount    int
	CritCount    int
	TotalCasts   int
	
	// Mana
	OOMEvents    int // Out of mana events
}

// Simulator runs the combat simulation
type Simulator struct {
	Config    *config.Config
	SimConfig SimulationConfig
	BaseSeed  int64
}

// NewSimulator creates a new simulator
func NewSimulator(cfg *config.Config, simCfg SimulationConfig, seed int64) *Simulator {
	return &Simulator{
		Config:    cfg,
		SimConfig: simCfg,
		BaseSeed:  seed,
	}
}

// Run executes the simulation for configured iterations
func (s *Simulator) Run(char *character.Character) *SimulationResult {
	result := &SimulationResult{
		Duration:   s.SimConfig.Duration,
		Iterations: s.SimConfig.Iterations,
	}
	
	// Run multiple iterations with unique seed each
	for i := 0; i < s.SimConfig.Iterations; i++ {
		iterResult := s.runSingleIteration(char, i)
		result.aggregateResult(iterResult)
	}
	
	// Calculate averages
	result.TotalDamage /= float64(s.SimConfig.Iterations)
	result.TotalDPS = result.TotalDamage / s.SimConfig.Duration.Seconds()
	
	return result
}

// runSingleIteration runs one simulation iteration
func (s *Simulator) runSingleIteration(originalChar *character.Character, iteration int) *SimulationResult {
	// Create a fresh copy of character for this iteration
	char := character.NewCharacter(originalChar.Stats)
	
	// Create spell engine with unique seed for this iteration
	spellEngine := spells.NewEngine(s.Config, s.BaseSeed+int64(iteration), s.SimConfig.IsBoss)
	
	result := &SimulationResult{}
	hasImmolate := false
	
	// Combat loop
	for char.CurrentTime < s.SimConfig.Duration {
		// Check if we need to reapply Immolate (< 3s remaining or not active)
		immolateTimeLeft := char.Immolate.ExpiresAt - char.CurrentTime
		if !char.Immolate.Active || immolateTimeLeft < 3*time.Second {
			if !hasImmolate || immolateTimeLeft < 3*time.Second {
				if s.tryCast(char, spells.SpellImmolate, result, spellEngine) {
					hasImmolate = true
					continue
				}
			}
		}
		
		// Priority 2: Conflagrate on CD
		if char.IsCooldownReady(&char.Conflagrate) {
			if s.tryCast(char, spells.SpellConflagrate, result, spellEngine) {
				continue
			}
		}
		
		// Priority 3: Chaos Bolt on CD
		if char.IsCooldownReady(&char.ChaosBolt) {
			if s.tryCast(char, spells.SpellChaosBolt, result, spellEngine) {
				continue
			}
		}
		
		// Priority 4: Life Tap if low mana (< 30%)
		manaThreshold := char.Stats.MaxMana * 0.30
		if char.Resources.CurrentMana < manaThreshold {
			if s.tryCast(char, spells.SpellLifeTap, result, spellEngine) {
				continue
			}
		}
		
		// Priority 5: Incinerate (filler)
		if s.tryCast(char, spells.SpellIncinerate, result, spellEngine) {
			continue
		}
		
		// Priority 6: Life Tap if OOM (can't cast anything else)
		if s.tryCast(char, spells.SpellLifeTap, result, spellEngine) {
			result.OOMEvents++
			continue
		}
		
		// If we somehow can't do anything, advance time by GCD
		char.AdvanceTime(time.Duration(s.Config.Constants.GCD.Base * float64(time.Second)))
	}
	
	return result
}

// tryCast attempts to cast a spell
func (s *Simulator) tryCast(char *character.Character, spell spells.SpellType, result *SimulationResult, spellEngine *spells.Engine) bool {
	// Check if GCD is ready
	if !char.IsGCDReady() {
		return false
	}
	
	// Check mana cost
	var manaCost float64
	switch spell {
	case spells.SpellImmolate:
		manaCost = s.Config.Spells.Immolate.ManaCost
	case spells.SpellIncinerate:
		manaCost = s.Config.Spells.Incinerate.ManaCost
	case spells.SpellChaosBolt:
		manaCost = s.Config.Spells.ChaosBolt.ManaCost
	case spells.SpellConflagrate:
		manaCost = s.Config.Spells.Conflagrate.ManaCost
	case spells.SpellLifeTap:
		manaCost = 0
	}
	
	if manaCost > 0 && !char.HasMana(manaCost) {
		return false
	}
	
	// Cast the spell
	var castResult spells.CastResult
	switch spell {
	case spells.SpellImmolate:
		castResult = spellEngine.CastImmolate(char)
		result.ImmolateCount++
		if castResult.DidHit {
			result.ImmolateDamage += castResult.Damage
		}
	case spells.SpellIncinerate:
		castResult = spellEngine.CastIncinerate(char)
		result.IncinerateCount++
		if castResult.DidHit {
			result.IncinerateDamage += castResult.Damage
		}
	case spells.SpellChaosBolt:
		castResult = spellEngine.CastChaosBolt(char)
		result.ChaosBoltCount++
		if castResult.DidHit {
			result.ChaosBoltDamage += castResult.Damage
		}
	case spells.SpellConflagrate:
		castResult = spellEngine.CastConflagrate(char)
		result.ConflagrateCount++
		if castResult.DidHit {
			result.ConflagrateDamage += castResult.Damage
		}
	case spells.SpellLifeTap:
		castResult = spellEngine.CastLifeTap(char)
		result.LifeTapCount++
	}
	
	// Track statistics
	result.TotalCasts++
	if !castResult.DidHit {
		result.MissCount++
	}
	if castResult.DidCrit {
		result.CritCount++
	}
	
	// Advance time by cast time + GCD
	totalTime := castResult.CastTime + castResult.GCDTime
	char.AdvanceTime(totalTime)
	char.GCDReadyAt = char.CurrentTime
	
	return true
}

// aggregateResult combines results from multiple iterations
func (r *SimulationResult) aggregateResult(iter *SimulationResult) {
	r.TotalDamage += iter.ImmolateDamage + iter.IncinerateDamage + iter.ChaosBoltDamage + iter.ConflagrateDamage
	
	r.ImmolateCount += iter.ImmolateCount
	r.IncinerateCount += iter.IncinerateCount
	r.ChaosBoltCount += iter.ChaosBoltCount
	r.ConflagrateCount += iter.ConflagrateCount
	r.LifeTapCount += iter.LifeTapCount
	
	r.ImmolateDamage += iter.ImmolateDamage
	r.IncinerateDamage += iter.IncinerateDamage
	r.ChaosBoltDamage += iter.ChaosBoltDamage
	r.ConflagrateDamage += iter.ConflagrateDamage
	
	r.MissCount += iter.MissCount
	r.CritCount += iter.CritCount
	r.TotalCasts += iter.TotalCasts
	r.OOMEvents += iter.OOMEvents
}

// PrintResults outputs simulation results
func (r *SimulationResult) PrintResults() {
	fmt.Println("========================================")
	fmt.Println("Simulation Results")
	fmt.Println("========================================")
	fmt.Printf("Duration: %.0fs\n", r.Duration.Seconds())
	fmt.Printf("Iterations: %d\n", r.Iterations)
	fmt.Println()
	
	fmt.Printf("Total DPS: %.2f\n", r.TotalDPS)
	fmt.Printf("Total Damage: %.0f\n", r.TotalDamage)
	fmt.Println()
	
	fmt.Println("Spell Breakdown:")
	fmt.Println("----------------------------------------")
	
	totalDmg := r.ImmolateDamage + r.IncinerateDamage + r.ChaosBoltDamage + r.ConflagrateDamage
	
	if r.ImmolateCount > 0 {
		pct := (r.ImmolateDamage / totalDmg) * 100
		avgDmg := r.ImmolateDamage / float64(r.Iterations) / float64(r.ImmolateCount)
		fmt.Printf("Immolate:    %3d casts  |  %.0f damage (%.1f%%)  |  %.0f avg\n",
			r.ImmolateCount/r.Iterations, r.ImmolateDamage/float64(r.Iterations), pct, avgDmg)
	}
	
	if r.IncinerateCount > 0 {
		pct := (r.IncinerateDamage / totalDmg) * 100
		avgDmg := r.IncinerateDamage / float64(r.Iterations) / float64(r.IncinerateCount)
		fmt.Printf("Incinerate:  %3d casts  |  %.0f damage (%.1f%%)  |  %.0f avg\n",
			r.IncinerateCount/r.Iterations, r.IncinerateDamage/float64(r.Iterations), pct, avgDmg)
	}
	
	if r.ChaosBoltCount > 0 {
		pct := (r.ChaosBoltDamage / totalDmg) * 100
		avgDmg := r.ChaosBoltDamage / float64(r.Iterations) / float64(r.ChaosBoltCount)
		fmt.Printf("Chaos Bolt:  %3d casts  |  %.0f damage (%.1f%%)  |  %.0f avg\n",
			r.ChaosBoltCount/r.Iterations, r.ChaosBoltDamage/float64(r.Iterations), pct, avgDmg)
	}
	
	if r.ConflagrateCount > 0 {
		pct := (r.ConflagrateDamage / totalDmg) * 100
		avgDmg := r.ConflagrateDamage / float64(r.Iterations) / float64(r.ConflagrateCount)
		fmt.Printf("Conflagrate: %3d casts  |  %.0f damage (%.1f%%)  |  %.0f avg\n",
			r.ConflagrateCount/r.Iterations, r.ConflagrateDamage/float64(r.Iterations), pct, avgDmg)
	}
	
	if r.LifeTapCount > 0 {
		fmt.Printf("Life Tap:    %3d casts\n", r.LifeTapCount/r.Iterations)
	}
	
	fmt.Println()
	fmt.Println("Statistics:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Total Casts: %d\n", r.TotalCasts/r.Iterations)
	fmt.Printf("Misses:      %d (%.1f%%)\n", r.MissCount/r.Iterations, float64(r.MissCount)/float64(r.TotalCasts)*100)
	fmt.Printf("Crits:       %d (%.1f%%)\n", r.CritCount/r.Iterations, float64(r.CritCount)/float64(r.TotalCasts)*100)
	if r.OOMEvents > 0 {
		fmt.Printf("OOM Events:  %d\n", r.OOMEvents/r.Iterations)
	}
	fmt.Println("========================================")
}
