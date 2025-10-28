# WotLK Destruction Warlock Simulator

A fast, accurate, cast-by-cast combat simulator for Destruction Warlock on a custom WotLK private server.

## Phase 1 MVP - Current Features

✅ **Core Simulation Engine**
- Hit/miss system (boss vs equal level targets)
- Mana tracking with Life Tap
- Random crit rolls
- GCD management (fixed 1.5s)
- Cooldown tracking (Chaos Bolt, Conflagrate)
- Basic rotation priority system
- 1000 iterations for statistical accuracy

✅ **Spells Implemented**
- Immolate (direct + DoT damage)
- Incinerate (with Immolate bonus)
- Chaos Bolt (12s cooldown)
- Conflagrate (10s cooldown, 25% bonus crit)
- Life Tap (mana generation)

✅ **Talents Included**
- Emberstorm (+15% fire/shadow damage)
- Improved Immolate (+30% damage)
- Aftermath (+6% DoT damage)
- Fire and Brimstone (+10% on Immolated targets, +25% Conflag crit)
- Ruin (2.0x crit multiplier)
- Devastation (+5% crit)
- Backlash (+1% crit)

✅ **External YAML Configuration**
- All spell data, talents, and constants in editable YAML files
- No recompilation needed for balance changes

## Not Yet Implemented

- ❌ Backdraft system (Phase 2)
- ❌ Mystic Enchants (Phase 3)
- ❌ Haste mechanics (Phase 5)
- ❌ Stat weights calculation (Phase 4)
- ❌ Web UI (Phase 2+)

## Requirements

- Go 1.21 or higher

## Installation

```bash
# Clone or download this project
cd wotlk-destro-sim

# Install dependencies
go mod download
```

## Running the Simulator

```bash
# Run from project root
go run cmd/simulator/main.go
```

## Configuration

### Character Stats

Edit `cmd/simulator/main.go` to modify character stats:

```go
charStats := character.Stats{
    SpellPower: 800,
    CritPct:    25.0,  // 25% crit
    HastePct:   0.0,   // No haste in Phase 1
    Spirit:     200,
    HitPct:     17.0,  // Hit capped for boss
    MaxMana:    8000,
}
```

### Simulation Parameters

```go
simConfig := engine.SimulationConfig{
    Duration:   5 * time.Minute, // Fight duration
    Iterations: 1000,             // Number of iterations
    IsBoss:     true,             // Boss (17% hit cap) vs Equal Level (4% miss)
}
```

### Spell Data & Talents

Edit the YAML files in `configs/`:
- `constants.yaml` - Server constants (stat conversions, GCD, hit caps)
- `spells.yaml` - All spell data (damage, costs, coefficients)
- `talents.yaml` - Talent modifiers

No recompilation needed after editing YAML files!

## Example Output

```
========================================
Simulation Results
========================================
Duration: 300s
Iterations: 1000

Total DPS: 2,456.78
Total Damage: 737,034

Spell Breakdown:
----------------------------------------
Immolate:      1 casts  |  12,543 damage (1.7%)  |  12,543 avg
Incinerate:   72 casts  |  524,123 damage (71.1%)  |  7,279 avg
Chaos Bolt:   25 casts  |  145,678 damage (19.8%)  |  5,827 avg
Conflagrate:  30 casts  |  54,690 damage (7.4%)  |  1,823 avg
Life Tap:     12 casts

Statistics:
----------------------------------------
Total Casts: 140
Misses:      0 (0.0%)
Crits:       35 (25.0%)
========================================
```

## Project Structure

```
wotlk-destro-sim/
├── cmd/
│   └── simulator/      # Main program
├── internal/
│   ├── character/      # Character stats and state
│   ├── config/         # YAML configuration loader
│   ├── engine/         # Simulation engine and rotation logic
│   └── spells/         # Spell casting and damage calculation
├── configs/            # YAML configuration files
│   ├── constants.yaml
│   ├── spells.yaml
│   └── talents.yaml
├── go.mod
└── README.md
```

## Rotation Priority (Phase 1)

1. Maintain Immolate (recast if < 3s remaining)
2. Conflagrate on CD
3. Chaos Bolt on CD
4. Life Tap if mana < 30%
5. Incinerate (filler)
6. Life Tap if OOM

## Development Roadmap

- **Phase 1 (CURRENT)**: Core simulation engine ✅
- **Phase 2**: Backdraft system
- **Phase 3**: Mystic Enchants (ME) system
- **Phase 4**: Stat weights calculation
- **Phase 5**: Haste implementation
- **Phase 6**: Polish & React UI
- **Phase 7**: Action Priority Lists (APL) for user-configurable rotations

## Design Document

See `warlock-simcraft-design.md` for complete design specifications.

## License

MIT License - See LICENSE file for details
