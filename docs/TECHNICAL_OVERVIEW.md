# Technical Overview

Single-node Go 1.21+ simulator that models a Level 60 Destruction Warlock on a progressive WotLK-like server. Phase 3 status with Backdraft and detailed reporting.

## Layout
- `cmd/simulator`: CLI entry; `-log-combat` prints a single-iteration combat log for debugging.
- `cmd/aplvalidate`: Static validator for YAML Action Priority Lists.
- `internal/engine`: Event queue (casts, GCD unlocks, cooldowns, DoT ticks, pet casts) and rotation executor.
- `internal/spells`: Per-spell modules plus shared math/helpers in `core.go` (hit/crit rolls, coefficients, PvE Power, target mods, Fire and Brimstone gates).
- `internal/effects`: Aura/timer helpers used by buffs like Heating Up, Backdraft timers, Gul'dan's Chosen, etc.
- `configs/*.yaml`: Data-driven constants for spells, talents, player stats, and rotations (APL).

## Mechanics & Calculations
- **Event queue** drives time; DoT ticks (Immolate) are scheduled events, not polled. Cast time/GCD respect haste and the Backdraft floor (base GCD 1.5s, min 1.0s).
- **Hit/crit**: Rolls per cast; hit caps from `configs/constants.yaml` (17% boss cap, 4% equal level). RNG seed is unique per iteration to avoid identical runs.
- **Damage**: `base roll + SP * coefficient`, multiplied by talents/runes (Emberstorm, Fire and Brimstone on Immolated targets, Shadow and Flame bonus SP, PvE Power currently hardcoded 1.25). Crits use Ruin’s 200% multiplier.
- **DoTs**: Immolate snapshots damage multipliers at cast, ticks for 5 ticks over 15s; Conflagrate consumes Immolate damage (60% of DoT) and applies its own DoT (40% of hit).
- **Backdraft**: Conflagrate grants 3 charges for 15s, reducing next Destruction spell cast time and GCD by 30%; charges consumed by all Destruction spells (instants included). Uptime and average charges are tracked.
- **Pyroclasm**: Conflagrate crit can grant +6% fire/shadow damage for 10s (duration extended by Endless Flames ME).
- **Improved Soul Leech**: 30% proc on fire spells; instantly returns 2% max mana and applies a HoT (1% max mana every 5s for 15s), both tracked in uptime/mana reporting.
- **Pet**: Imp Firebolt casting with Demonic Power/Empowered Imp hooks; mana and crit/damage bonuses applied in the same damage math path.

## Config & Tooling
- Edit YAML under `configs/` (spells, talents, constants, player stats, rotations). Changes apply without recompiling.
- Validate rotations: `go run ./cmd/aplvalidate -rotation configs/rotations/destruction-default.yaml`.
- Run sims: `go run ./cmd/simulator` (set `GOCACHE=$(mktemp -d)` for clean runs).
- Combat log mode: `go run ./cmd/simulator -log-combat` forces a 60s single-iteration log for verification.

## Reporting
- Per-spell breakdown: min/avg/max hit, crit %, miss %, shares of total damage, cast counts.
- Buff uptimes: Pyroclasm, Backdraft (with average charges), Improved Soul Leech HoT/instant tracking.
- Resource metrics: Life Tap casts, mana returned from Soul Leech, pet contribution.
