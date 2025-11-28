# Architecture Overview

## Stack & Data
- Language: Go 1.21+
- Config: YAML for constants, player stats, spells, talents, rotations (APL)
- CLI: `cmd/simulator` (main sim), `cmd/aplvalidate` (rotation validator)
- RNG: per-iteration seeds; optional combat log mode (`-log-combat`)

## Core Concepts
- Event queue drives time (casts, GCD unlocks, DoT ticks, pet casts). DoT ticks are scheduled events, not polled.
- Character state: stats, mana, GCD timer, cooldowns, buffs/debuffs, pet state.
- Effects: shared aura/timer helpers in `internal/effects`; used by Heating Up, Gul'dan's Chosen, Cataclysmic Burst, Backdraft timers, etc.
- Spells: modular files under `internal/spells/` with shared helpers in `core.go` (hit/crit rolls, spell power, PvE Power multiplier, Fire and Brimstone checks, target modifiers).
- APL: YAML rotation compiled by `internal/apl`, executed by engine; validate with `go run ./cmd/aplvalidate -rotation configs/rotations/destruction-default.yaml`.

## Mechanics Implemented
- Spells: Immolate (direct + DoT snapshot), Incinerate (Immolate bonus), Chaos Bolt, Conflagrate (Immolate-driven), Life Tap.
- Talents: Emberstorm, Improved Immolate, Aftermath, Fire and Brimstone (Incinerate/Chaos Bolt), Ruin, Devastation, Backlash, Pyroclasm, Backdraft (charges, uptime), Improved Soul Leech (instant + HoT buff), Demonic Power, Empowered Imp.
- Mystic Enchants: selection via config with implemented effects (Destruction Mastery, Heating Up, Cataclysmic Burst, Gul'dan's Chosen, Glyph of Conflagrate, Glyph of Life Tap, Suppression, hooks for Agent of Chaos, etc.).
- PvE Power: currently hardcoded 1.25 multiplier in spell damage; pending config-driven value.
- Pet: Imp Firebolt with talent/rune hooks; basic crit/damage and mana handling.

## Reporting
- Results include per-spell damage stats, miss/crit rates, Backdraft uptime/avg charges, mana info; combat log mode prints event log for a single iteration.

## Extension Points / Risks
- Move remaining buffs/debuffs to the aura framework (Backdraft state, Chaos Manifesting, Decisive Decimation).
- Extend event scheduling to Soul Leech HoT ticks and future pets.
- Make PvE Power configurable and decide pet applicability.
- Add unit tests per spell (deterministic seeds) and validation on YAML inputs.
