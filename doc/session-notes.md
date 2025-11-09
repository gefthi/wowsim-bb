# Session Wrap-Up

## Branch
`feature/backend-vision`

## What Changed
- **Effects package**: Added `internal/effects/aura.go`/`timer.go` to replace ad-hoc buff fields and GCD bookkeeping. Heating Up, Gul'dan's Chosen, and Cataclysmic Burst now use shared auras; GCD uses the timer.
- **Event queue**: `internal/engine/events.go` plus new helpers (`scheduleEvent`, `wait`, `executeImmolateTick`, etc.) give us a pending-action loop. Immolate ticks are now scheduled events instead of being polled each time slice, which sets the pattern for future dots/pets.
- **Spell modularization**: The old `internal/spells/spells.go` was replaced with per-spell files:
  - Shared helpers live in `internal/spells/core.go`.
  - Individual casts reside in dedicated files (`immolate.go`, `incinerate.go`, `chaos_bolt.go`, `conflagrate.go`, `life_tap.go`).
  - Rune/talent helpers sit alongside them (`backdraft.go`, `cataclysmic_burst.go`, `guldans_chosen.go`, `soul_leech.go`).
- **Engine integration**: `internal/engine/engine.go` now drives time through the event queue, schedules Immolate ticks, and uses the new per-spell APIs. Rotation waits were updated to call `wait`.

## Current State
- Rotation and sim run with the new queue; Immolate extensions (Cataclysmic Burst) grant extra ticks correctly.
- Heating Up, Gul'dan's Chosen, and Cataclysmic Burst use aura instances; Backdraft/GCD timers rely on the new helpers.
- Spells compile from their modular files; `go test ./...` succeeds when using a writable `GOCACHE`.

## Next Logical Steps
1. Migrate remaining buffs/debuffs (Backdraft state, Chaos Manifesting, Decisive Decimation, etc.) to the aura framework.
2. Extend the event queue to other periodic effects (Improved Soul Leech HoT, future pets).
3. Add per-spell unit tests now that files are isolated.
4. Continue modularization for the remaining spells or future specs using the same pattern.

Keeping this doc up-to-date each session should make it easy to resume work without re-reading the entire repo.

