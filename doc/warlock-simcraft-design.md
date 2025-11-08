# WotLK Destruction Warlock SimulationCraft - Design Document
*Source of Truth for Development*

---

## 🎯 Project Vision

Build a **fast, accurate, cast-by-cast combat simulator** for Destruction Warlock on a custom WotLK private server. The simulator will model the actual rotation with all buff/debuff interactions, proc chances, and cooldown management to calculate accurate DPS and stat weights.

**Server Context**: Progressive server currently at **Level 60** with WotLK talent system and custom modifications. Uses Level 60 stat conversions (10 rating = 1% for crit/haste) but WotLK spell mechanics and talents.

**Key Philosophy**: Simple, Fast, MVP Approach with Iterative Development

---

## 📋 Working Rules (System Prompt)

These are our development principles that govern ALL decisions:

1. **NO CODE WITHOUT APPROVAL** - We design first, implement only when given explicit green light
2. **RUNES ARE CONFIGURABLE** - Never assume runes are active. The sim helps users discover optimal rune combinations
3. **MVP & ITERATIVE** - Each iteration must be functional. Add complexity gradually, one feature at a time
4. **DESIGN DOCUMENT FIRST** - This artifact is our source of truth for continuity across sessions
5. **GRADUAL QUESTIONS** - Ask important design questions bit by bit, never overwhelm with too many at once
6. **Claude PROPOSES UI/UX** - User focuses on game mechanics, Claude designs the interface
7. **KEEP RULES IN ARTIFACT** - These principles stay documented and guide all conversations
8. **REMEMBER CONTEXT** - Save configuration states between sessions (e.g., "destro-haste setup")

---

## 🏗️ Technical Architecture

### Core Stack Decision

**Simulation Engine**: **Go** (same as wowsims/wotlk)
- **Why**: Fast execution for 1000s of simulation runs
- **Why**: Proven choice by the reference implementation
- **Why**: Excellent for concurrent simulation iterations
- **Why**: Native performance without overhead

**Frontend**: **React + TypeScript**
- **Why**: Simple, modern, widely supported
- **Why**: Easy state management for form inputs and results display
- **Why**: Can reference wowsims UI patterns when needed
- **Why**: TypeScript gives us type safety for stat/configuration objects

**Data Exchange**: JSON (simple, no proto needed for MVP)
- Go simulation engine exposes HTTP endpoint
- React frontend sends configuration, receives results
- Keep it simple - no WebAssembly complexity initially

### Architecture Overview
```
┌─────────────────────────────────────┐
│         React Frontend              │
│  ┌──────────────────────────────┐  │
│  │  Stat Input Form             │  │
│  │  - SP, Crit, Haste, Spirit   │  │
│  │  - Hit, Rune Toggles         │  │
│  │  - Rotation Options          │  │
│  └──────────────────────────────┘  │
│              ↓ JSON                 │
│  ┌──────────────────────────────┐  │
│  │  Configuration Manager       │  │
│  │  - Save/Load Presets         │  │
│  │  - LocalStorage Persistence  │  │
│  └──────────────────────────────┘  │
│              ↓ HTTP POST            │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│       Go Simulation Engine          │
│  ┌──────────────────────────────┐  │
│  │  HTTP Server (localhost)     │  │
│  └──────────────────────────────┘  │
│              ↓                      │
│  ┌──────────────────────────────┐  │
│  │  Config Loader (YAML)        │  │ ← **NEW**
│  │  - Spell data (costs, dmg)   │  │
│  │  - Formulas (Life Tap, etc)  │  │
│  │  - Talent modifiers          │  │
│  │  - Hot-reload support        │  │
│  └──────────────────────────────┘  │
│              ↓                      │
│  ┌──────────────────────────────┐  │
│  │  Simulation Core             │  │
│  │  - Cast Timeline Generator   │  │
│  │  - Buff/Debuff Tracker       │  │
│  │  - Random Crit/Proc System   │  │
│  │  - Damage Calculator         │  │
│  └──────────────────────────────┘  │
│              ↓                      │
│  ┌──────────────────────────────┐  │
│  │  Statistics Aggregator       │  │
│  │  - Run N iterations          │  │
│  │  - Average results           │  │
│  │  - Calculate stat weights    │  │
│  └──────────────────────────────┘  │
│              ↓ JSON                 │
└─────────────────────────────────────┘
                ↓
┌─────────────────────────────────────┐
│         React Frontend              │
│  ┌──────────────────────────────┐  │
│  │  Results Display             │  │
│  │  - Total DPS                 │  │
│  │  - Spell Breakdown           │  │
│  │  - Buff Uptimes              │  │
│  │  - Stat Weights              │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │  Optional Cast Timeline      │  │
│  │  - Detailed log viewer       │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

### External Configuration System (YAML)

**Why**: As the server evolves (leveling, patches, balance changes), spell data will change. Having this in external YAML files means:
- No recompilation needed for data changes
- Human-readable and easy to edit
- Can version control different server patches
- Easy to share configurations with others

**File Structure**:
```
config/
├── spells.yaml          # Spell base data
├── talents.yaml         # Talent modifiers
├── formulas.yaml        # Complex formulas (Life Tap, etc)
└── constants.yaml       # Server constants (stat conversions, GCD, etc)
```

**Example: spells.yaml**
```yaml
immolate:
  direct_damage: 404
  dot_damage: 770
  dot_duration: 15
  cast_time: 1.5  # After Bane talent
  mana_cost: 252
  sp_coefficient_direct: 0.20
  sp_coefficient_dot: 1.00

incinerate:
  base_damage_min: 416
  base_damage_max: 490
  immolate_bonus_min: 104
  immolate_bonus_max: 123
  cast_time: 2.25  # After Emberstorm
  mana_cost: 207
  sp_coefficient: 0.714

chaos_bolt:
  base_damage_min: 1250
  base_damage_max: 1508
  cast_time: 2.0  # After Bane
  cooldown: 12  # Base cooldown (before rune)
  mana_cost: 103
  sp_coefficient: 0.714

conflagrate:
  immolate_dot_percentage: 0.60  # 60% of Immolate DoT
  conflag_dot_percentage: 0.40  # 40% of hit as DoT
  cast_time: 0  # Instant
  cooldown: 10
  mana_cost: 236
  sp_coefficient: 0.60

life_tap:
  cast_time: 0  # Instant (GCD only)
  cooldown: 0
  mana_cost: 0  # Generates mana
```

**Example: formulas.yaml**
```yaml
life_tap:
  health_cost:
    base: 827
    spirit_multiplier: 1.5
    formula: "827 + (spirit * 1.5)"
  
  mana_gain:
    base: 827
    spellpower_coefficient: 0.5
    improved_lifetap_per_rank: 0.10  # 10% per talent rank
    formula: "[827 * (1 + talent_rank * 0.10)] + [spellpower * 0.5 * (1 + talent_rank * 0.10)]"
```

**Example: talents.yaml**
```yaml
emberstorm:
  damage_multiplier: 1.15  # +15% fire/shadow damage

improved_immolate:
  damage_multiplier: 1.30  # +30% all Immolate damage

aftermath:
  dot_damage_multiplier: 1.06  # +6% DoT damage only (for Immolate)

fire_and_brimstone:
  immolate_target_damage: 1.10  # +10% on Immolated targets
  conflagrate_crit_bonus: 0.25  # +25% crit chance

ruin:
  crit_multiplier: 2.0  # Crits do 200% instead of 150%

shadow_and_flame:
  bonus_sp_percentage: 0.20  # Gains +20% of bonus SP
```

**Example: constants.yaml**
```yaml
server:
  level: 60
  name: "Progressive WotLK"

stat_conversions:
  crit_rating_per_percent: 14
  haste_rating_per_percent: 10

hit_mechanics:
  boss_hit_cap: 17  # +3 level difference
  equal_level_miss_chance: 4  # Base miss vs same level

gcd:
  base: 1.5
  minimum: 1.0

spell_power:
  unified: true  # Single stat for all schools
  note: "Some items have legacy shadow power - user should sum them"
```

**Implementation**:
- Load all YAML files on server startup
- Parse into Go structs
- Cache in memory (no disk I/O during simulation)
- Optional: Add file watcher for hot-reload during development
- Go has excellent YAML support: `gopkg.in/yaml.v3`

---

## 🎮 Reference Implementations

### Primary Reference: WoWSims/WotLK
**URL**: https://github.com/wowsims/wotlk

This is a **CRITICAL RESOURCE** - comprehensive WotLK simulation framework with:
- Proven simulation engine architecture in Go
- Event-driven combat system
- Proper GCD, cast time, and cooldown management
- Proc system implementation
- Stat conversion formulas
- Unit test examples

**What We Use From It**:
- Architecture patterns for simulation core
- How to model GCD and cast sequences
- Event-driven buff/debuff system
- Statistical aggregation approaches
- Spell coefficient implementations

**What We Simplify**:
- No paper doll UI (manual stat entry)
- No gear database (work with raw stats)
- Single spec focus (Destruction Warlock only)
- No raid sim complexity
- No WebAssembly initially (simple HTTP server)

### Secondary Reference: SimulationCraft (Original)
**URL**: https://github.com/simulationcraft/simc

The **ORIGINAL** simulation tool that pioneered combat simulation for WoW. Written in C++.

**Key Features to Study**:
- Event-driven simulator architecture
- Action Priority List (APL) system for rotation definition
- Stat weight calculation methodology
- Multi-iteration statistical approach
- Buff/debuff tracking system
- Proc modeling

**Action Priority Lists (APL)**:
- **URL**: https://github.com/simulationcraft/simc/wiki/ActionLists
- **FUTURE FEATURE**: User-configurable rotations via text file
- **Why**: Allow users to modify rotation logic without recompiling
- **Syntax Example**: 
  ```
  actions=/immolate,if=!ticking
  actions+=/conflagrate,if=cooldown_ready
  actions+=/chaos_bolt,if=cooldown_ready
  actions+=/incinerate
  ```
- **Implementation Plan**: Phase 6+ (after core mechanics are solid)
- **Benefits**:
  - Test different rotation strategies easily
  - Share rotation configurations
  - Iterate on optimal play without coding
  - Community can contribute rotation improvements

**Not for MVP**: APL system is complex and not needed initially. We'll hardcode the priority rotation in Phase 1-3, then add APL support in Phase 6+ once the core simulation is proven.

**What to Extract from SimulationCraft**:
- Action priority list parsing concepts (for future)
- Conditional expression syntax ideas
- How they handle spell queuing and timing
- Statistical reporting approaches
- Buff/debuff expiration handling

---

## 🎯 Project Scope - MVP Definition

### Phase 1: Core Simulation Engine (MVP)
**Goal**: Accurate DPS calculation for basic rotation with mana management and hit mechanics

**Must Have**:
- **Hit/miss system** (each spell rolls for hit based on player hit %)
- **Mana tracking** (current mana, max mana, spell costs)
- **Life Tap mechanic** (health → mana conversion, GCD cost)
- Cast-by-cast timeline generation
- Proper GCD handling (NOT affected by haste, only Backdraft)
- Basic spell damage calculation with spell power
- Random crit rolls (using crit chance)
- Cooldown tracking (Chaos Bolt 10s, Conflagrate 10s)
- Simple priority rotation:
  1. If mana < threshold: Life Tap
  2. Cast Immolate once
  3. Conflagrate on CD
  4. Chaos Bolt on CD
  5. Incinerate filler (if enough mana)
  6. Life Tap (if OOM)
- Run 1000 iterations and average results
- Output: Total DPS, miss count, OOM events (if any)

**Not Yet**:
- Stat weights (Phase 5)
- Buff uptimes (Phase 2)
- Detailed logs (Phase 6)
- Haste implementation (Phase 2)
- Proc systems (Pyroclasm - Phase 3)
- Runes (Phase 3)

### Phase 2: Buff/Debuff System
**Goal**: Model Backdraft charges and consumption

**Adds**:
- Backdraft charge tracking (3 charges, 15s duration)
- Backdraft cast time reduction (30%)
- Backdraft GCD reduction (30%)
- Immolate debuff tracking
- Fire and Brimstone damage bonus

**Outputs Add**:
- Backdraft uptime %
- Average Backdraft charges active

### Phase 3: Configurable Runes
**Goal**: Let users test different rune combinations

**Adds**:
- Gul'na's Chosen (4s window of non-consuming Backdraft)
- Pyroclasm proc system (crit Conflag → 6% damage buff, 16s duration)
- Life Tap rune (spirit → spell power conversion)
- Chaos Bolt cooldown reduction rune
- Immolate extension rune (optional infinite uptime)
- Toggle each rune on/off in UI

**Outputs Add**:
- Pyroclasm uptime %
- Effective spell power (showing Life Tap contribution)

### Phase 4: Stat Weights
**Goal**: Answer "What stat is best?"

**Adds**:
- Rerun simulation with +1 to each stat
- Calculate marginal DPS gain
- Output stat weight table

---

## Rotation DSL / APL Plan (Pre-Implementation)

We no longer hardcode priorities – rotations now live in YAML (`internal/apl` + `configs/rotations/destruction-default.yaml`). Remaining APL roadmap items:

1. **YAML-driven** – users edit `configs/rotations/*.yaml`.
2. **Condition-based** – predicates like `buff_active`, `dot_remaining`, `resource_percent`, `cooldown_ready`, `charges`, `time_elapsed`. Nest via `all` / `any`.
3. **Action-rich** – `cast_spell`, `use_item`, `wait`, `macro` (chain multiple actions), future `channel`.
4. **Top-to-bottom evaluation** – loop through the list each decision, execute the first entry whose condition is true, then restart at the top.
5. **Extensible** – rotations may define `variables` (e.g., `life_tap_threshold`) and `imports` to extend a base list for rune-specific tweaks.

### YAML Skeleton

```yaml
variables:
  life_tap_threshold: 0.30
imports:
  - rotations/base.yaml

rotation:
  - action: cast_spell
    spell: immolate
    when:
      any:
        - not_active: immolate
        - dot_remaining:
            spell: immolate
            lt_seconds: 3

  - action: cast_spell
    spell: conflagrate
    when:
      all:
        - cooldown_ready: conflagrate
        - debuff_active:
            debuff: immolate

  - action: cast_spell
    spell: life_tap
    when:
      resource_percent:
        resource: mana
        lt: ${life_tap_threshold}
```

### Engine Notes

- YAML parsed once → compiled into structs with typed predicates/actions for quick runtime evaluation.
- Each `Condition` implements `Eval(state)`; each `Action` implements `Execute(state)`.
- CLI helper (`go run ./cmd/aplvalidate -rotation ...`) validates syntax/names before running the sim.
- Debug combat log (`go run ./cmd/simulator -log-combat`) runs a single 60s iteration and emits a WoW-style log (casts, DoT ticks, buff gains/expirations) for manual verification.
- TODO: future realism tweaks for combat log (apply configurable latency delay + spell travel time before recording hit events).
- Debug flag prints the first N decisions to help users tune their list.

This is the working contract for the Phase 4+ implementation. We proceed in small, verifiable iterations (loader → validation → execution) to keep the sim buildable at every step.

---

## 📊 Simulation Core Design

### Core Concepts (from wowsims reference)

The simulation runs on a **timeline-based event system**:

1. **Timeline**: Ordered list of events (cast finishes, cooldowns ready, buff expires)
2. **Event Loop**: Process events in order, each event may create new events
3. **Current Time**: Simulation time in milliseconds
4. **GCD State**: Track if GCD is active and when it ends
5. **Cast State**: Track current cast and when it finishes

### Key Data Structures

```
Character {
    Stats {
        SpellPower, Crit, Haste, Spirit, Hit
        MaxMana
    }
    Resources {
        CurrentMana
    }
    Buffs {
        Backdraft { charges, expiresAt }
        Pyroclasm { active, expiresAt }
        LifeTap { active, expiresAt }
        GulnasChosen { active, expiresAt }
    }
    Debuffs {
        Immolate { active, expiresAt }
    }
    Cooldowns {
        ChaosBolt { readyAt }
        Conflagrate { readyAt }
    }
    GCD { readyAt }
}

CastEvent {
    spell: SpellType
    startTime: float64
    finishTime: float64
    gcdUntil: float64
    manaCost: int
}

DamageEvent {
    spell: SpellType
    damage: float64
    crit: bool
    time: float64
}
```

### Decision Loop (Priority System)

At each decision point (when GCD available and no cast in progress):
```
1. Check if mana is too low for next spell
   → Cast Life Tap (generates mana, costs health + GCD)
   
2. Check if Immolate needs refreshing (< 3s remaining)
   → Cast Immolate (if enough mana)
   
3. Check if Conflagrate is off CD
   → Cast Conflagrate (instant, if enough mana)
   
4. Check if Chaos Bolt is off CD
   → Cast Chaos Bolt (if enough mana)
   
5. Default: Cast Incinerate (if enough mana)

6. If not enough mana for anything: Life Tap

7. If fight will end before next GCD, stop simulation
```

**Mana Threshold Logic**:
- Keep enough mana for at least 2-3 spells to avoid awkward Life Tap timing
- Prioritize Life Tap before going OOM (out of mana)
- Life Tap is instant (GCD only), so it's a reactive recovery tool

### Damage Calculation Flow

```
For each spell cast:
1. Roll for hit/miss
   - If player hit % < required hit cap: roll random number
   - If miss: record miss event, spell deals 0 damage, mana still spent
   - If hit: continue to step 2
2. Calculate base damage (from spell data)
3. Apply spell power coefficient
4. Apply talent modifiers (Emberstorm, Improved Immolate, etc.)
5. Apply situational bonuses (Fire and Brimstone if Immolate up)
6. Roll for crit (using character crit %)
7. If crit: multiply by crit multiplier (Ruin: 2.0x)
8. Record damage event
9. Check for procs (Pyroclasm on Conflag crit)
```

**Hit Chance Calculation**:
```
Required Hit Cap:
- Boss (+3 levels): 17%
- Equal level: 4% (base miss chance)

Player Hit Chance = min(player_hit_percent, 100%)

For each spell cast:
  roll = random(0, 100)
  if roll > player_hit_percent AND player_hit_percent < required_cap:
    MISS
  else:
    HIT
```

---

## 🔢 Spell Data & Mechanics

### Server-Specific Constants
**IMPORTANT**: This is a progressive server currently at **Level 60** with WotLK talents/mechanics.

- **Crit Rating Conversion**: **14 rating = 1% crit** (same as Level 80 WotLK)
- **Haste Rating Conversion**: **10 rating = 1% haste** (Level 60 - MUCH better than Level 80's ~32.79!)
- **Hit Cap**: **17% vs. Boss** (+3 levels), lower vs. equal level targets
- **GCD Base**: 1.5 seconds
- **GCD Minimum**: 1.0 seconds (with haste cap)
- **Spell Power**: Unified system (affects all spell schools)
  - Some legacy items may have both "Spell Power" and "Shadow Power"
  - User should enter their total effective spell power
  - For MVP: treat as single unified stat

**Hit Mechanics** (CRITICAL):
- **Boss (default)**: 17% spell hit required for hit cap (+3 level difference)
- **Equal Level**: Lower hit requirement (based on level difference)
- User enters their hit % in UI
- UI toggle to select target type: "Boss" vs "Equal Level (60)"
- Each spell cast rolls for hit/miss based on hit chance
- Miss = 0 damage, wasted GCD and mana

**Haste Mechanics** (CRITICAL):
- **Cast Times**: Haste reduces cast times (e.g., 10% haste → Incinerate 2.25s becomes 2.05s)
- **DoT Tick Speed**: **Haste does NOT affect DoTs normally on this server!**
  - **EXCEPTION**: Agent of Chaos ME makes Immolate DoT benefit from haste
  - Without Agent of Chaos: Immolate always ticks at base speed (15s duration, 5 ticks)
  - With Agent of Chaos: Immolate ticks faster with haste (same 5 ticks, shorter duration)
- **GCD**: Haste does NOT affect GCD normally!
  - **ONLY Backdraft reduces GCD** (-30% for 3 charges or 15s duration)
  - Regular haste stat does NOT reduce GCD

*Note: Haste is ~3x more accessible on this server compared to standard WotLK Level 80. This makes haste a very strong stat for cast times!*

### Spell Definitions

**Mana Costs (Level 60 - Server Specific)**:
- Immolate: 252 mana
- Incinerate: 207 mana
- Chaos Bolt: 103 mana
- Conflagrate: 236 mana
- Soul Fire: 133 mana
- Life Tap: 0 mana (generates mana)

#### Immolate
```
Direct Damage: 404 fire
DoT: 770 over 15 seconds (51.33 per tick, 5 ticks)
Cast Time: 1.5s (includes Bane talent)
Mana Cost: 252
SP Coefficient: ~0.20 direct, ~1.00 DoT total
Modifiers:
  - Improved Immolate: +30% damage (affects both direct and DoT)
  - Emberstorm: +15% damage (affects both direct and DoT)
  - Aftermath: +6% damage (DoT ONLY - we have this talent!)
```

#### Incinerate
```
Base: 416-490 (avg 453)
Immolate Bonus: +104-123 (avg 113.5)
Cast Time: 2.25s (includes Emberstorm reduction)
Mana Cost: 207
SP Coefficient: ~0.714
Modifiers:
  - Emberstorm: +15% damage
  - Fire and Brimstone: +10% damage (if target has Immolate)
  - Shadow and Flame: Gains 20% of bonus SP as additional damage
```

#### Chaos Bolt
```
Base: 1250-1508 (avg 1379)
Cast Time: 2.0s (includes Bane)
Cooldown: 12s base (10s with rune)
Mana Cost: 103
SP Coefficient: ~0.714
Modifiers:
  - Emberstorm: +15% damage
  - Fire and Brimstone: +10% damage (if target has Immolate)
  - Shadow and Flame: Gains 20% of bonus SP as additional damage
```

#### Conflagrate
```
Base: 60% of Immolate's total DoT (770 * 0.6 = 462)
Conflag DoT: 40% of hit damage over time
Instant Cast
Cooldown: 10s
Mana Cost: 236
SP Coefficient: ~0.60
Modifiers:
  - Fire and Brimstone: +25% crit chance
  - Does NOT consume Immolate (with Glyph of Conflagrate ME)
Procs:
  - Pyroclasm on crit (with talent)
  - Heating Up debuff on target (with ME): +2% fire damage per stack (max 5)
  - Decisive Decimation (with ME): next Soul Fire buffed
```

#### Soul Fire
```
Base: 808-1014 (avg 911) fire damage
Cast Time: 6s base → 4s (with Bane) → affected by haste → affected by Backdraft
Mana Cost: 133
Soul Shard Cost: 1 (removed with Decisive Decimation ME)
SP Coefficient: 1.15 (115%)
Notes:
  - Generally ONLY used with Decisive Decimation ME active
  - With Decisive Decimation: -40% cast time (stacks with other reductions)
  - Example: 4s → 3.91s (2.3% haste) → ~2.73s (with Backdraft) → ~1.64s (with Decisive Decimation)
  - Very powerful when all buffs align
  - Soul Shard cost ignored since we only cast with Decisive Decimation ME
```

#### Life Tap
```
Converts health to mana (instant cast, GCD only)

Formula (server-specific):
  Health Cost = 827 + (Spirit * 1.5)  [We don't track this - assume player gets healed]
  Mana Gain = [827 * ImprovedLifeTapMultiplier] + [SpellPower * 0.5 * ImprovedLifeTapMultiplier]
  
  Where:
    ImprovedLifeTapMultiplier = 1.0 + (talent_rank * 0.10)
    - Rank 0 (default for Destro): 1.0
    - Rank 1: 1.1 
    - Rank 2: 1.2
    - We DON'T have this talent as Destruction, so multiplier = 1.0
    
  SpellPower = Unified spell power stat (affects all schools)
    Note: Some legacy items may have both "Spell Power" and "Shadow Power"
    For Life Tap: use total effective spell power (base + any shadow-specific bonuses)
    For MVP: treat as unified (user enters total spell power)

No cooldown (but costs GCD: 1.5s base, affected by haste and Backdraft)
Mana Cost: None (generates mana)
Health Cost: Ignored (assume player gets healed - we only track mana)
Notes:
  - Essential for mana management
  - With Life Tap rune: grants 20% spirit → spell power for 40s
  - Formula ready for future Improved Life Tap talent/rune if needed
```

### Talent Modifiers (Always Active)

These are baked into character stats:
- **Bane 5/5**: Already in cast times above
- **Ruin 5/5**: Crit multiplier = 2.0x (instead of 1.5x)
- **Devastation**: +5% crit
- **Backlash**: +1% crit
- **Emberstorm**: +15% Fire/Shadow damage, -0.25s Incinerate
- **Shadow and Flame**: +20% bonus SP effect
- **Backdraft**: -30% cast time AND GCD (3 charges, 15s duration)
- **Fire and Brimstone**: +10% damage on Immolated targets, +25% Conflag crit
- **Soul Leech**: 30% chance for 20% damage → health (not modeled initially)
- **Improved Soul Leech**: Mana return (not modeled initially)

### Mystic Enchants (ME) System

**IMPORTANT**: MEs have slot restrictions! You can only have:
- 1x Artifact (ignored - doesn't affect DPS)
- **1x Legendary** (choose 1)
- **3x Epic** (choose 3)
- **6x Rare** (choose 6)

**All MEs are OPTIONAL and togglable in the UI** - the simulator helps discover optimal combinations.

#### Legendary MEs (1 slot):

**Gul'na's Chosen** (EPIC - CORRECTION: This is actually Epic!)
- Casting Chaos Bolt → 4s window where Backdraft charges NOT consumed

**Destruction Mastery**
- +4% shadow and fire damage (all spells)
- +5% additional Immolate damage (both direct hit and DoT)
- Total Immolate modifier: 1.04 * 1.05 = 1.092 (9.2% total increase)

**Cataclysmic Burst**
- When Incinerate deals damage: extends Immolate duration by 2 seconds
- Also adds stacking buff: +8% Immolate periodic damage (stacks up to 4 times = 32% max)
- Stacks consumed when casting Conflagrate
- Note: This is the "Immolate Extension" functionality

#### Epic MEs (3 slots):

**Gul'na's Chosen** (moved from Legendary)
- Casting Chaos Bolt → 4s window where Backdraft charges NOT consumed

**Endless Flames** (Pyroclasm enhancement)
- Prerequisite: Must have Pyroclasm talent
- Increases Pyroclasm buff duration from 10s → 16s (+6s)
- Reminder: Pyroclasm = Conflag/Searing Pain crit → 6% fire/shadow damage buff

**Heating Up**
- Conflagrate's direct damage (not DoT!) applies "Heating Up!" debuff to target
- +2% fire damage taken per stack
- Stacks up to 5 times (10% max)
- Lasts 15 seconds
- Affects: Immolate, Conflagrate, Incinerate, Chaos Bolt (chaos is both fire and shadow), Soul Fire

**Decisive Decimation**
- When you cast Conflagrate: next Soul Fire has -40% cast time AND no soul shard cost
- Soul Fire base: 6s cast → 4s (with Bane) → affected by haste → affected by Backdraft → -40% from this ME
- Soul Fire damage: 808-1014 fire damage
- Generally only used WITH this ME active (otherwise 4s+ cast time is too long)

**Inner Flame**
- Direct damage with fire or shadow spells: 12% chance to proc
- Proc effect: Next direct fire spell is guaranteed critical strike
- Fire spells: Immolate (direct), Conflagrate (direct), Incinerate, Chaos Bolt, Soul Fire
- Does NOT proc from DoT ticks

**Agent of Chaos**
- Immolate direct damage reduced by 50% (halved)
- Immolate duration increased by 3 seconds (15s → 18s)
- **Immolate DoT NOW benefits from spell haste** (overrides normal mechanic!)
- Each Immolate DoT tick reduces Chaos Bolt cooldown by 0.5 seconds

#### Rare MEs (6 slots):

**Life Tap ME** (moved from Epic)
- When you Life Tap: gain 20% of spirit as spell power for 40 seconds
- Must cast Life Tap every 40s to maintain buff

**Glyph of Chaos Bolt**
- Chaos Bolt cooldown reduced by 2 seconds (12s → 10s)

**Glyph of Conflagrate**
- Conflagrate does NOT consume Immolate from target

**Demonic Aegis**
- Increases Fel Armor spell power bonus
- New formula: 65 + (Spirit × 0.39) instead of 65 + (Spirit × 0.30)

**Suppression**
- +3% spell hit

**Glyph of Incinerate**
- Increases Incinerate damage by 5%

---

## 💾 Data Persistence Strategy

**Problem**: User enters stats, runs sim, closes browser. Comes back later, has to re-enter everything.

**Solution**: LocalStorage for configuration presets

### Preset System
```
Preset = {
    name: string (e.g., "Destro-Haste", "Full BiS", "Test-NoLifeTap")
    stats: {
        spellPower: number
        crit: number (as %)
        haste: number (as %)
        spirit: number
        hit: number (as %)
        maxMana: number (user configures)
    }
    target: {
        type: "boss" | "equal_level"  // Default: "boss"
        level: number  // Default: 60 (equal level) or 63 (boss)
    }
    mystic_enchants: {
        legendary: string | null  // Choose 1: "destruction_mastery", "cataclysmic_burst", null
        epic: string[]  // Choose 3 from: "guldans_chosen", "endless_flames", "heating_up", "decisive_decimation", "inner_flame", "agent_of_chaos"
        rare: string[]  // Choose 6 from: "life_tap", "glyph_cb", "glyph_conflag", "demonic_aegis", "suppression", "glyph_incinerate"
    }
    rotation: {
        syncChaosBoltWithBackdraft: bool
        lifeTapThreshold: number (% of max mana to trigger Life Tap)
        useSoulFire: bool  // Only relevant if Decisive Decimation is active
    }
    fightDuration: number (seconds)
    iterations: number
}
```

### UI Features
- "Save Preset" button → prompts for name
- Dropdown to load saved presets
- "Last Used" preset auto-saved on every sim run
- Export/Import presets as JSON (future: share configurations)

---

## 🎨 UI Design Proposal

### Layout Structure

```
┌─────────────────────────────────────────────────────────┐
│  WotLK Destruction Warlock Simulator                    │
│  [GitHub Link]                              [Help (?)]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────┐  ┌──────────────────────────┐│
│  │   CONFIGURATION      │  │      RESULTS             ││
│  │                      │  │                          ││
│  │ Presets: [Dropdown▾] │  │  Total DPS: 12,456      ││
│  │    [Save] [Delete]   │  │                          ││
│  │                      │  │  ┌────────────────────┐  ││
│  │ Character Stats      │  │  │ Spell Breakdown    │  ││
│  │  Spell Power: [____] │  │  │ • Chaos Bolt  35%  │  ││
│  │  Crit %:      [____] │  │  │ • Incinerate  40%  │  ││
│  │  Haste %:     [____] │  │  │ • Conflagrate 15%  │  ││
│  │  Spirit:      [____] │  │  │ • Immolate    10%  │  ││
│  │  Hit %:       [____] │  │  └────────────────────┘  ││
│  │  Max Mana:    [____] │  │                          ││
│  │                      │  │  ┌────────────────────┐  ││
│  │ Target Type          │  │  │ Buff Uptimes       │  ││
│  │  ◉ Boss (+3 lvl)     │  │  │ • Backdraft   78%  │  ││
│  │  ○ Equal Level (60)  │  │  │ • Pyroclasm   65%  │  ││
│  │                      │  │  └────────────────────┘  ││
│  │  ☐ Life Tap          │  │  │ Buff Uptimes       │  ││
│  │  ☐ Gul'na's Chosen   │  │  │ • Backdraft   78%  │  ││
│  │  ☐ Pyroclasm         │  │  │ • Pyroclasm   65%  │  ││
│  │  ☐ CB CDR            │  │  └────────────────────┘  ││
│  │  ☐ Immolate Extend   │  │                          ││
│  │                      │  │  ┌────────────────────┐  ││
│  │ Rotation Options     │  │  │ Stat Weights       │  ││
│  │  ☐ Sync CB/Backdraft │  │  │ (Phased - Not yet) │  ││
│  │                      │  │  └────────────────────┘  ││
│  │ Sim Settings         │  │                          ││
│  │  Duration: [___] sec │  │  [View Cast Timeline]   ││
│  │  Iterations: [____]  │  │                          ││
│  │                      │  │  [Running... 43%]        ││
│  │  [RUN SIMULATION]    │  │      or                  ││
│  │                      │  │  Sim completed in 2.4s   ││
│  └──────────────────────┘  └──────────────────────────┘│
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Component Breakdown

**ConfigurationPanel** (Left Side)
- Preset management
- Stat inputs (number fields with validation)
- Rune checkboxes (organized by category)
- Rotation toggles
- Sim settings
- Big "Run Simulation" button

**ResultsPanel** (Right Side)
- Total DPS (large, prominent)
- Spell breakdown (pie chart or bars?)
- Buff uptimes (progress bars)
- Stat weights table (phased)
- Export results button
- Cast timeline viewer (expandable/modal)

**Design Principles**:
- Clean, minimal, functional (no fancy animations)
- Mobile-friendly layout (stack vertically on small screens)
- Clear visual hierarchy
- Instant feedback (validation, loading states)
- Tooltips on runes (explain what they do)

---

## 🔍 Open Design Questions

### Must Answer Before Coding:

1. **Haste Rating Conversion** ✅ **ANSWERED**
   - Q: What's the haste rating → % conversion on this server?
   - A: **10 rating = 1% haste** (Level 60 progressive server)
   - Impact: Affects cast times and GCD
   - Priority: High - **CRITICAL FOR ACCURACY**

2. **Mana Management** ✅ **DECISION MADE**
   - Q: Model mana costs and Life Tap timing, or assume infinite mana?
   - A: **MODEL MANA FROM THE START** - This is critical!
   - Reasoning: Destruction IS mana-constrained on this server (unlike MoP+)
   - Life Tap needed for both mana AND the Life Tap rune (spirit → spell power)
   - Assume player won't die from Life Tap (gets healed), but GCD cost is real
   - Priority: **CRITICAL - Include in MVP**

3. **Haste Mechanics** ✅ **ANSWERED** (CORRECTED v1.6)
   - Q: Does haste affect cast times, GCD, and/or DoT tick rates?
   - A: **Haste affects:**
     - Cast times: YES (e.g., 10% haste → 2.25s becomes 2.05s)
     - DoT tick speed: **NO** (except with Agent of Chaos ME for Immolate only!)
     - GCD: NO! **Only Backdraft reduces GCD** (-30%), regular haste does not
   - Impact: Critical for accurate DPS calculations
   - Priority: **CRITICAL - Must implement correctly**

4. **Hit Mechanics** ✅ **DECISION MADE**
   - Q: Should hit mechanics be in MVP?
   - A: **YES - Critical for stat weights from the start**
   - Boss hit cap: 17% (standard, +3 level difference)
   - Equal level hit cap: 4% (base miss chance)
   - UI toggle: "Boss" (default) vs "Equal Level (60)"
   - Each spell rolls for hit/miss
   - Misses = 0 damage but still cost mana and GCD
   - Priority: **CRITICAL - Include in MVP**

5. **Empowered Imp**
   - Q: Model Imp crits → guaranteed player crit mechanic?
   - Recommendation: **Skip for MVP** (pet mechanics add complexity)
   - Impact: Minor DPS gain, not core to rotation
   - Priority: Low

6. **Spell Coefficients**
   - Q: Are the listed coefficients correct for this server?
   - Recommendation: Make them **editable in YAML config** (can tweak if needed)
   - Priority: Medium (use defaults, verify later)

7. **Conflagrate DoT**
   - Q: Does the Conflag DoT (40% of hit) also scale with talents?
   - Recommendation: **Yes, inherits all modifiers**
   - Priority: Medium (affects Conflag value)

---

## 🚀 Development Roadmap

### Milestone 1: Simulation Core (No UI)
**Goal**: Command-line sim that outputs DPS

**Tasks**:
1. Set up Go project structure
2. Implement character stats structure
3. Implement spell data (base damage, coefficients)
4. Build timeline event system
5. Implement GCD and cast time tracking
6. Build basic rotation priority logic
7. Implement damage calculation
8. Add random crit rolls
9. Run 1000 iterations, average results
10. Output total DPS to console

**Deliverable**: `go run main.go` outputs DPS number

**Test**: Verify DPS is reasonable (compare to spreadsheet estimate)

### Milestone 2: Basic UI
**Goal**: React frontend that calls Go backend

**Tasks**:
1. Set up React + TypeScript project
2. Create Go HTTP server (POST /simulate endpoint)
3. Build stat input form
4. Implement "Run Simulation" button
5. Display DPS result
6. Add loading state

**Deliverable**: Working web interface

### Milestone 3: Buff/Debuff System
**Goal**: Model Backdraft and Immolate

**Tasks**:
1. Add buff tracking to character state
2. Implement Backdraft charge system
3. Backdraft cast time reduction
4. Backdraft GCD reduction
5. Fire and Brimstone damage bonus
6. Output buff uptimes

**Deliverable**: Accurate Backdraft modeling

### Milestone 4: Configurable Runes
**Goal**: Toggle runes in UI

**Tasks**:
1. Add rune options to configuration struct
2. UI checkboxes for each rune
3. Implement Gul'na's Chosen logic
4. Implement Pyroclasm proc system
5. Implement Life Tap rune (periodic cast)
6. Output includes rune-specific metrics

**Deliverable**: Users can test different rune combos

### Milestone 5: Stat Weights
**Goal**: Calculate marginal DPS gains

**Tasks**:
1. Run baseline simulation
2. For each stat: run sim with +1 to that stat
3. Calculate DPS difference
4. Output stat weight table
5. Display in UI

**Deliverable**: Stat priority recommendations

### Milestone 6: Polish
**Goal**: Production-ready tool

**Tasks**:
1. Add preset save/load system
2. Implement cast timeline log
3. Add spell breakdown display
4. Create tooltips for runes
5. Add export results as JSON
6. Write documentation
7. Add error handling

**Deliverable**: Complete, polished simulator

### Milestone 7: Action Priority Lists (APL) - FUTURE
**Goal**: User-configurable rotations via text file

**Why**: Allow rotation experimentation without recompiling

**Tasks**:
1. Design APL syntax (inspired by SimulationCraft)
2. Implement APL parser
3. Convert hardcoded rotation to APL internally
4. Add APL text editor to UI
5. Support conditional expressions (if=, target_if=, etc)
6. Allow saving/loading APL files
7. Provide default APLs for common scenarios

**Deliverable**: Users can edit rotations in text format

**Example APL**:
```
# Destruction Warlock - Standard Rotation
actions=/life_tap,if=mana.pct<30
actions+=/immolate,if=!debuff.immolate.up
actions+=/conflagrate,if=cooldown_ready
actions+=/chaos_bolt,if=cooldown_ready&mana>200
actions+=/incinerate
```

**Not a Priority**: This is a nice-to-have feature for Phase 7+. Core mechanics and ME system come first.

---

## 🧪 Testing Strategy

### Unit Tests (Go)
- Spell damage calculations (with various modifiers)
- Buff/debuff application and expiration
- Cooldown tracking
- Crit roll correctness
- Priority logic decision making

### Integration Tests
- Full rotation simulation (fixed seed for reproducibility)
- Verify DPS against manual calculations
- Test each rune individually
- Test rune interactions

### Validation
- Compare to existing spreadsheet calculations
- Sanity check: DPS should be ~10-15k range (ballpark)
- Backdraft uptime should be high (>70%)
- Pyroclasm uptime depends on Conflag crit rate

---

## 📚 Resources & References

### Primary References

**WoWSims/WotLK** - https://github.com/wowsims/wotlk
- Study `sim/core/` for simulation architecture
- Reference `sim/warlock/` for Warlock specifics
- Look at any spec's rotation logic for patterns
- Key files:
  - `sim/core/sim.go`: Main event loop
  - `sim/core/cast.go`: Cast mechanics
  - `sim/core/aura.go`: Buff/debuff system
  - `sim/core/spell.go`: Spell damage calculation
  - `sim/core/character.go`: Character state

**SimulationCraft (Original)** - https://github.com/simulationcraft/simc
- Original combat simulator (C++)
- **Action Priority Lists**: https://github.com/simulationcraft/simc/wiki/ActionLists
- Study for future APL implementation
- Reference for conditional expressions
- Stat weight calculation methodology
- Multi-iteration statistical approaches

### Additional Resources
- WoWHead WotLK Classic Database
- Private server documentation (if available)
- Community spreadsheets (for coefficient verification)

---

## ✅ Next Steps

### Before Writing Any Code:

1. **Answer open questions** (gradually, in conversation)
   - Start with: Haste rating conversion?
   - Then: Mana modeling yes/no?
   - Then: Other questions as needed

2. **Review and approve this design document**
   - User confirms architecture makes sense
   - Agree on MVP scope
   - Approve UI mockup concept

3. **Set up development environment**
   - Initialize Go project
   - Initialize React project
   - Test basic HTTP communication

4. **Begin Milestone 1** (with approval)

---

## 📝 Change Log

**Version 1.0** - Initial design document created
- Established architecture (Go backend + React frontend)
- Defined MVP scope and phased approach
- Documented spell mechanics and formulas
- Created UI mockup
- Listed open design questions
- Outlined development roadmap

**Version 1.1** - Stat conversions clarified
- **CRITICAL**: Confirmed server uses Level 60 stat conversions (10 rating = 1% for crit/haste)
- Updated server-specific constants section
- Added progressive server context to project vision
- Marked haste conversion question as answered

**Version 1.2** - Mana management made critical, crit rating corrected
- **CORRECTED**: Crit rating is 14 = 1% (standard WotLK), NOT 10
- Haste remains 10 = 1% (makes haste ~3x more accessible than standard WotLK!)
- **CRITICAL DECISION**: Mana management MUST be in from MVP start
- Destruction IS mana-constrained on this server
- Added Life Tap spell data
- Updated priority rotation to include mana checks
- Added mana/health fields to UI and data structures
- Updated Phase 1 MVP to include full mana tracking

**Version 1.3** - Spell costs, talents, and external configuration system
- Added exact mana costs for all spells (Immolate: 252, Incinerate: 207, CB: 103, Conflag: 236)
- Added Aftermath talent: +6% DoT damage for Immolate (we have this!)
- Clarified spell power unification (user enters total)
- **NEW DESIGN**: External YAML configuration system
  - Spell data in config files (not hardcoded)
  - Easy to modify without recompilation
  - Supports server evolution (leveling, patches)
  - Human-readable format
  - File structure: spells.yaml, talents.yaml, formulas.yaml, constants.yaml

**Version 1.4** - Haste mechanics fully clarified
- **CRITICAL**: Haste affects cast times AND DoT tick speed (not GCD!)
  - Cast times reduced by haste % (e.g., 10% haste → 2.25s becomes 2.05s)
  - DoT tick speed increased (same number of ticks, duration decreases)
  - **GCD NOT affected by haste** - only Backdraft reduces GCD by 30%
- Removed health tracking (don't care - assume healed)
- Simplified Life Tap implementation
- Updated open design questions with haste answer

**Version 1.5** - Hit mechanics added to MVP
- **CRITICAL**: Hit mechanics must be in MVP from start (affects stat weights)
- Boss hit cap: 17% (+3 level difference)
- Equal level miss chance: 4% (base)
- UI toggle: Target type selection (Boss vs Equal Level)
- Each spell cast rolls for hit/miss
- Misses cost mana and GCD but deal 0 damage
- Added hit mechanics to Phase 1 MVP requirements
- Updated damage calculation flow with hit rolls
- Added target type to preset system
- Updated UI mockup with target type radio buttons

**Version 1.6** - Mystic Enchants (ME) system fully defined + critical haste correction
- **CRITICAL CORRECTION**: Haste does NOT affect DoT tick speed on this server!
  - Exception: Agent of Chaos ME makes Immolate DoT benefit from haste
  - Without Agent of Chaos: Immolate ticks at base speed (15s, 5 ticks)
- **Renamed**: "Runes" → "Mystic Enchants (ME)" to match in-game terminology
- **ME Slot System**: 1 Legendary + 3 Epic + 6 Rare (choose from available options)
- **Legendary MEs defined**: Destruction Mastery, Cataclysmic Burst
- **Epic MEs defined**: Gul'na's Chosen (moved from Legendary!), Endless Flames, Heating Up, Decisive Decimation, Inner Flame, Agent of Chaos
- **Rare MEs defined**: Life Tap (moved from Epic!), Glyph of CB, Glyph of Conflag, Demonic Aegis, Suppression, Glyph of Incinerate
- **Soul Fire spell added**: 808-1014 damage, 4s cast (with Bane), used with Decisive Decimation ME
- **Clarified**: Cataclysmic Burst = the "Immolate Extension" functionality + damage stacking
- Updated preset system to use ME slot selection instead of boolean toggles
- Added rotation option for Soul Fire usage

**Version 1.7** - SimulationCraft references and APL future planning
- **Added**: SimulationCraft (original) as secondary reference (https://github.com/simulationcraft/simc)
- **Added**: Action Priority List (APL) wiki link for future rotation configuration
- **Future Feature**: User-configurable rotations via text file (Phase 7+)
  - Inspired by SimulationCraft's APL system
  - Allows rotation experimentation without recompiling
  - Not in MVP - will implement after core mechanics are solid
- Added Phase 7 milestone for APL implementation
- Updated Resources section with both reference implementations
- Documented why APL is valuable but not immediate priority

---

*This document is living and will be updated as we make decisions and progress through development.*
