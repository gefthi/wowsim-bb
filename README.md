# WotLK Destruction Warlock Simulator

A fast, accurate, cast-by-cast combat simulator for Destruction Warlock on a custom WotLK private server.

## Phase 3 - Current Features

✅ **Core Simulation Engine**
- Hit/miss system (boss vs equal level targets)
- Mana tracking with Life Tap
- Random crit rolls
- GCD management (fixed 1.5s)
- Cooldown tracking (Chaos Bolt, Conflagrate)
- Basic rotation priority system
- 1000 iterations for statistical accuracy

✅ **Spells & Talents**
- Immolate (direct + DoT damage)
- Incinerate (with Immolate bonus + Fire and Brimstone gating)
- Chaos Bolt (12s cooldown + Fire and Brimstone gating)
- Conflagrate (Pyroclasm proc source, bonus crit)
- Life Tap (mana generation, fewer casts thanks to Soul Leech)
- Pyroclasm (Conflagrate crit → +6% fire/shadow damage, uptime tracked)
- Improved Soul Leech (instant mana + HoT ticks processed every 5s)
- Devastation & Backlash now point-based and configurable
- **Backdraft** (Conflagrate → 3 charges, -30% cast time/GCD, uptime + avg charges tracked)

✅ **Enhanced Reporting & Tooling**
- Per-spell min/avg/max damage plus crit & miss percentages
- Buff uptime tracking for Pyroclasm, Backdraft, and Improved Soul Leech (with Backdraft avg charges)
- Accurate Soul Leech HoT ticks (1% max mana every 5s)
- External YAML configuration for player stats (`configs/player.yaml`), spells, talents, and server constants
- **New** YAML Action Priority Lists (`configs/rotations/`) compiled at runtime so you can edit the rotation without rebuilding
- Unique RNG seed per iteration for varied results

## Not Yet Implemented

- ❌ Mystic Enchants (Phase 4)
- ❌ Stat weights calculation (Phase 5)
- ❌ Haste mechanics (Phase 6)
- ❌ Web UI + APL tools (Phase 7+)
 - ❌ APL validator + extended predicates/actions (design + first implementation in `doc/apl-schema.md`)

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

### Character & Simulation Settings

Edit `configs/player.yaml` to change stats, targets, and runtime parameters:

```yaml
character:
  name: "Destruction Warlock"
  level: 60
stats:
  spell_power: 800
  crit_percent: 25.0
  haste_percent: 0.0
  spirit: 200
  hit_percent: 17.0
  max_mana: 8000
target:
  type: boss        # or equal_level
  level: 83
simulation:
  duration_seconds: 300
  iterations: 1000
```

Changes to any YAML file take effect immediately — no recompilation required.

### Spell Data & Talents

Edit the YAML files in `configs/`:
- `constants.yaml` - Server constants (stat conversions, GCD, hit caps)
- `spells.yaml` - All spell data (damage, costs, coefficients)
- `talents.yaml` - Talent modifiers
 - `rotations/destruction-default.yaml` - Default YAML APL (editable priority list)

No recompilation needed after editing YAML files!

> Tip: set `points: 0` (or `enabled: false`) on a talent such as `improved_soul_leech` to disable it entirely if your current build doesn't use it.

## Example Output

```
========================================
Simulation Results
========================================
Duration: 300s
Iterations: 1000

Total DPS: 1,226.80
Total Damage: 368,041

Spell Breakdown (average per iteration):
--------------------------------------------------------------------------
Spell         |       Damage |  Share |     Avg |     Min |     Max |   Crit% |   Miss%
--------------------------------------------------------------------------
Immolate      |       106,892 |  29.0% |    4,454 |    3,331 |    7,062 |   30.9% |    0.0%
Incinerate    |        90,623 |  24.6% |    1,944 |    1,381 |    3,175 |   30.6% |    0.0%
Chaos Bolt    |        80,238 |  21.8% |    3,345 |    2,304 |    5,576 |   31.1% |    0.0%
Conflagrate   |        90,288 |  24.5% |    3,762 |    2,403 |    4,807 |   56.5% |    0.0%
--------------------------------------------------------------------------
Life Tap casts (avg): 8.3

Buff Uptimes:
----------------------------------------
Pyroclasm:           133.6s (44.5%)
Improved Soul Leech: 259.6s (86.5%)
Backdraft:           169.2s (56.4%) | avg charges 1.83

Statistics:
----------------------------------------
Total Casts: 126.9
Misses:      0.0 (0.0%)
Crits:       42.7 (33.7%)
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
│   ├── talents.yaml
│   └── player.yaml
├── go.mod
└── README.md
```

## Rotation Priority (Phase 3)

1. Maintain Immolate (recast if < 3s remaining)
2. Conflagrate on CD
3. Chaos Bolt on CD
4. Life Tap if mana < 30%
5. Incinerate (filler)
6. Life Tap if OOM

## Development Roadmap

- **Phase 1 (DONE)**: Core simulation engine ✅
- **Phase 2 (DONE)**: Pyroclasm, Improved Soul Leech, hit/RNG fixes ✅
- **Phase 2.5 (DONE)**: Enhanced statistics, buff uptimes, Soul Leech HoT ✅
- **Phase 3 (DONE)**: Backdraft system ✅
- **Phase 4**: Mystic Enchants (ME) system
- **Phase 5**: Stat weights calculation
- **Phase 6**: Haste implementation
- **Phase 7**: Polish & React UI + Action Priority Lists (APL)

## Design Document

See `warlock-simcraft-design.md` for complete design specifications.

## License

MIT License - See LICENSE file for details
