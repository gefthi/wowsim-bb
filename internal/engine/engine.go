package engine

import (
	"fmt"
	"math"
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

var spellPrintOrder = []struct {
	Type  spells.SpellType
	Label string
}{
	{spells.SpellImmolate, "Immolate"},
	{spells.SpellIncinerate, "Incinerate"},
	{spells.SpellChaosBolt, "Chaos Bolt"},
	{spells.SpellConflagrate, "Conflagrate"},
}

// SpellStats keeps per-spell performance details
type SpellStats struct {
	Casts     int
	Hits      int
	Crits     int
	Misses    int
	Damage    float64
	MinDamage float64
	MaxDamage float64
}

func newSpellStats() *SpellStats {
	return &SpellStats{
		MinDamage: math.MaxFloat64,
	}
}

func newSpellStatsMap() map[spells.SpellType]*SpellStats {
	stats := make(map[spells.SpellType]*SpellStats, len(spellPrintOrder))
	for _, spell := range spellPrintOrder {
		stats[spell.Type] = newSpellStats()
	}
	return stats
}

func (s *SpellStats) add(other *SpellStats) {
	s.Casts += other.Casts
	s.Hits += other.Hits
	s.Crits += other.Crits
	s.Misses += other.Misses
	s.Damage += other.Damage
	if other.Hits > 0 {
		if s.MinDamage == math.MaxFloat64 || other.MinDamage < s.MinDamage {
			s.MinDamage = other.MinDamage
		}
		if other.MaxDamage > s.MaxDamage {
			s.MaxDamage = other.MaxDamage
		}
	}
}

// SimulationResult holds results from simulation
type SimulationResult struct {
	TotalDPS     float64
	TotalDamage  float64
	Duration     time.Duration
	Iterations   int
	LifeTapCount int

	// Spell breakdown
	SpellBreakdown map[spells.SpellType]*SpellStats

	// Statistics
	MissCount  int
	CritCount  int
	TotalCasts int

	// Mana
	OOMEvents int // Out of mana events

	// Buff uptimes (seconds across all iterations)
	PyroclasmActiveSeconds         float64
	ImprovedSoulLeechActiveSeconds float64
	BackdraftActiveSeconds         float64
	BackdraftChargeSeconds         float64
}

func (r *SimulationResult) recordSpellCast(spell spells.SpellType, castResult spells.CastResult) {
	stats, ok := r.SpellBreakdown[spell]
	if !ok {
		return
	}
	stats.Casts++
	if castResult.DidHit {
		stats.Hits++
		stats.Damage += castResult.Damage
		if castResult.Damage < stats.MinDamage {
			stats.MinDamage = castResult.Damage
		}
		if castResult.Damage > stats.MaxDamage {
			stats.MaxDamage = castResult.Damage
		}
		if castResult.DidCrit {
			stats.Crits++
		}
		r.TotalDamage += castResult.Damage
	} else {
		stats.Misses++
	}
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
		Duration:       s.SimConfig.Duration,
		Iterations:     s.SimConfig.Iterations,
		SpellBreakdown: newSpellStatsMap(),
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

	result := &SimulationResult{
		SpellBreakdown: newSpellStatsMap(),
	}
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
		s.advanceTime(char, time.Duration(s.Config.Constants.GCD.Base*float64(time.Second)), result)
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
	case spells.SpellIncinerate:
		castResult = spellEngine.CastIncinerate(char)
	case spells.SpellChaosBolt:
		castResult = spellEngine.CastChaosBolt(char)
	case spells.SpellConflagrate:
		castResult = spellEngine.CastConflagrate(char)
	case spells.SpellLifeTap:
		castResult = spellEngine.CastLifeTap(char)
		result.LifeTapCount++
	}

	result.recordSpellCast(spell, castResult)

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
	s.advanceTime(char, totalTime, result)

	return true
}

func (s *Simulator) advanceTime(char *character.Character, duration time.Duration, result *SimulationResult) {
	if duration <= 0 {
		return
	}
	start := char.CurrentTime
	end := start + duration

	result.PyroclasmActiveSeconds += s.buffOverlapSeconds(&char.Pyroclasm, start, end)
	result.ImprovedSoulLeechActiveSeconds += s.buffOverlapSeconds(&char.ImprovedSoulLeech, start, end)
	backdraftOverlap := s.buffOverlapSeconds(&char.Backdraft, start, end)
	if backdraftOverlap > 0 {
		result.BackdraftActiveSeconds += backdraftOverlap
		charges := char.Backdraft.Charges
		if charges > 0 {
			result.BackdraftChargeSeconds += backdraftOverlap * float64(charges)
		}
	}

	char.AdvanceTime(duration)
	s.processSoulLeechHoT(char, start, end)
	s.expireBuffs(char)
	char.GCDReadyAt = char.CurrentTime
}

func (s *Simulator) buffOverlapSeconds(buff *character.Buff, start, end time.Duration) float64 {
	if !buff.Active {
		return 0
	}
	if end <= start {
		return 0
	}
	if start >= buff.ExpiresAt {
		return 0
	}
	activeEnd := buff.ExpiresAt
	if activeEnd > end {
		activeEnd = end
	}
	if activeEnd <= start {
		return 0
	}
	overlap := activeEnd - start
	return float64(overlap) / float64(time.Second)
}

func (s *Simulator) processSoulLeechHoT(char *character.Character, start, end time.Duration) {
	if !s.Config.Talents.ImprovedSoulLeech.Enabled || s.Config.Talents.ImprovedSoulLeech.Points <= 0 {
		return
	}
	if !char.ImprovedSoulLeech.Active {
		return
	}
	tickInterval := time.Duration(s.Config.Talents.ImprovedSoulLeech.HotTickInterval * float64(time.Second))
	if tickInterval <= 0 {
		return
	}
	nextTick := char.SoulLeechLastTick + tickInterval
	for nextTick <= end && nextTick <= char.ImprovedSoulLeech.ExpiresAt {
		if nextTick > start {
			mana := char.Stats.MaxMana * s.Config.Talents.ImprovedSoulLeech.HotManaPerTick
			char.GainMana(mana)
		}
		char.SoulLeechLastTick = nextTick
		nextTick += tickInterval
	}
}

func (s *Simulator) expireBuffs(char *character.Character) {
	now := char.CurrentTime
	if char.Pyroclasm.Active && now >= char.Pyroclasm.ExpiresAt {
		char.Pyroclasm.Active = false
	}
	if char.ImprovedSoulLeech.Active && now >= char.ImprovedSoulLeech.ExpiresAt {
		char.ImprovedSoulLeech.Active = false
	}
	if char.Backdraft.Active && (now >= char.Backdraft.ExpiresAt || char.Backdraft.Charges <= 0) {
		char.Backdraft.Active = false
		char.Backdraft.Charges = 0
	}
	if char.Immolate.Active && now >= char.Immolate.ExpiresAt {
		char.Immolate.Active = false
	}
}

// aggregateResult combines results from multiple iterations
func (r *SimulationResult) aggregateResult(iter *SimulationResult) {
	r.TotalDamage += iter.TotalDamage
	r.LifeTapCount += iter.LifeTapCount
	r.MissCount += iter.MissCount
	r.CritCount += iter.CritCount
	r.TotalCasts += iter.TotalCasts
	r.OOMEvents += iter.OOMEvents
	r.PyroclasmActiveSeconds += iter.PyroclasmActiveSeconds
	r.ImprovedSoulLeechActiveSeconds += iter.ImprovedSoulLeechActiveSeconds
	r.BackdraftActiveSeconds += iter.BackdraftActiveSeconds
	r.BackdraftChargeSeconds += iter.BackdraftChargeSeconds

	for spell, stats := range iter.SpellBreakdown {
		if base, ok := r.SpellBreakdown[spell]; ok {
			base.add(stats)
		} else {
			r.SpellBreakdown[spell] = stats
		}
	}
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

	fmt.Println("Spell Breakdown (average per iteration):")
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Printf("%-13s | %12s | %6s | %7s | %7s | %7s | %7s | %7s\n",
		"Spell", "Damage", "Share", "Avg", "Min", "Max", "Crit%", "Miss%")
	fmt.Println("--------------------------------------------------------------------------")
	totalDamage := 0.0
	for _, stats := range r.SpellBreakdown {
		totalDamage += stats.Damage
	}
	for _, entry := range spellPrintOrder {
		stats := r.SpellBreakdown[entry.Type]
		if stats == nil || stats.Casts == 0 {
			continue
		}
		avgDamagePerIter := stats.Damage / float64(r.Iterations)
		damagePct := 0.0
		if totalDamage > 0 {
			damagePct = (stats.Damage / totalDamage) * 100.0
		}
		var avgHit, minHit, maxHit, critPct float64
		if stats.Hits > 0 {
			avgHit = stats.Damage / float64(stats.Hits)
			minHit = stats.MinDamage
			if stats.MinDamage == math.MaxFloat64 {
				minHit = 0
			}
			maxHit = stats.MaxDamage
			if stats.MaxDamage == 0 {
				maxHit = avgHit
			}
			critPct = float64(stats.Crits) / float64(stats.Hits) * 100.0
		}
		missPct := 0.0
		if stats.Casts > 0 {
			missPct = float64(stats.Misses) / float64(stats.Casts) * 100.0
		}
		fmt.Printf("%-13s | %12.0f | %5.1f%% | %7.0f | %7.0f | %7.0f | %6.1f%% | %6.1f%%\n",
			entry.Label, avgDamagePerIter, damagePct, avgHit, minHit, maxHit, critPct, missPct)
	}
	fmt.Println("--------------------------------------------------------------------------")
	if r.LifeTapCount > 0 {
		fmt.Printf("Life Tap casts (avg): %.1f\n", float64(r.LifeTapCount)/float64(r.Iterations))
	}

	fmt.Println()
	fmt.Println("Buff Uptimes:")
	fmt.Println("----------------------------------------")
	fightSeconds := r.Duration.Seconds()
	avgPyroSeconds := r.PyroclasmActiveSeconds / float64(r.Iterations)
	avgSoulSeconds := r.ImprovedSoulLeechActiveSeconds / float64(r.Iterations)
	avgBackdraftSeconds := r.BackdraftActiveSeconds / float64(r.Iterations)
	pyroPct := 0.0
	soulPct := 0.0
	backdraftPct := 0.0
	if fightSeconds > 0 {
		pyroPct = (avgPyroSeconds / fightSeconds) * 100.0
		soulPct = (avgSoulSeconds / fightSeconds) * 100.0
		backdraftPct = (avgBackdraftSeconds / fightSeconds) * 100.0
	}
	avgBackdraftCharges := 0.0
	if avgBackdraftSeconds > 0 {
		avgChargeSeconds := r.BackdraftChargeSeconds / float64(r.Iterations)
		avgBackdraftCharges = avgChargeSeconds / avgBackdraftSeconds
	}
	fmt.Printf("Pyroclasm:           %.1fs (%.1f%%)\n", avgPyroSeconds, pyroPct)
	fmt.Printf("Improved Soul Leech: %.1fs (%.1f%%)\n", avgSoulSeconds, soulPct)
	if avgBackdraftSeconds > 0 {
		fmt.Printf("Backdraft:           %.1fs (%.1f%%) | avg charges %.2f\n", avgBackdraftSeconds, backdraftPct, avgBackdraftCharges)
	} else if r.BackdraftActiveSeconds > 0 {
		fmt.Printf("Backdraft:           %.1fs (%.1f%%)\n", avgBackdraftSeconds, backdraftPct)
	} else {
		fmt.Println("Backdraft:           0.0s (0.0%)")
	}

	fmt.Println()
	fmt.Println("Statistics:")
	fmt.Println("----------------------------------------")
	avgCasts := float64(r.TotalCasts) / float64(r.Iterations)
	fmt.Printf("Total Casts: %.1f\n", avgCasts)
	if r.TotalCasts > 0 {
		fmt.Printf("Misses:      %.1f (%.1f%%)\n",
			float64(r.MissCount)/float64(r.Iterations),
			float64(r.MissCount)/float64(r.TotalCasts)*100.0)
		fmt.Printf("Crits:       %.1f (%.1f%%)\n",
			float64(r.CritCount)/float64(r.Iterations),
			float64(r.CritCount)/float64(r.TotalCasts)*100.0)
	}
	if r.OOMEvents > 0 {
		fmt.Printf("OOM Events:  %.1f\n", float64(r.OOMEvents)/float64(r.Iterations))
	}
	fmt.Println("========================================")
}
