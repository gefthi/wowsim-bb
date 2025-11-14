# Wowsim-BB Backend Vision

## Purpose
Capture the long-term direction for our Go backend so we can evolve from the current MVP into a maintainable simulator that rivals (and eventually surpasses) the original wotlk backend while staying tailored to our custom server rules and Mystic Enchants (MEs).

## Guiding Principles
- **Modularity over monoliths**: Every new spell/rune/talent extension should live in its own file or package with clear responsibilities.
- **Data-driven first**: YAML (and, optionally, protobuf when it adds value) remains the authoritative source for numeric knobs, keeping balance changes out of binaries.
- **Deterministic combat loop**: A single event queue drives everything (casts, dots, pet swings, rune expirations) so we can reason about timing and logging.
- **Composable FX system**: Buffs, debuffs, glyphs, talents, and MEs use the same lifecycle primitives. No one-off booleans hidden in the character struct.
- **Incremental delivery**: Ship improvements feature-by-feature to avoid rewrites that stall gameplay validation.

## Current Pain Points
| Area | Issue |
| --- | --- |
| Spell implementation | Per-spell files exist under `internal/spells/`, but rune/talent toggles still share helpers—keep isolating effects inside each module. |
| Buff/debuff state | All state lives on `character.Character`; adding a new ME means adding more fields, manual timers, and duplicated uptime accounting. |
| Time flow | The loop in `internal/engine/engine.go` advances in fixed steps and re-checks conditions. Concurrency (dots, pets, HoTs) will be fragile. |
| Extensibility | No clear plug-in points for additional specs, pets, or target debuffs. The MVP is intentionally narrow but hard to extend. |
| Testing/logging | Logging is ad-hoc and there’s no unit-level validation for spell math. Hard to iterate quickly on new runes. |

## Target Architecture

### 1. Spell Modules & Registry
- Introduce `internal/spells/<spellname>.go` files that expose a `Register` function returning a `Spell` struct (costs, cast config, callbacks).
- Keep YAML as the source for coefficients but deserialize into per-spell config structs to avoid repeated map lookups.
- Each spell handles its own rune/talent hooks by attaching auras or casting-time modifiers rather than branching in a shared engine.

### 2. Aura & Effect Framework
- Create `internal/effects` with an `Aura` type mirroring the essential parts of `sim/core/aura.go` (Activate, Refresh, SetStacks, callbacks).
- Buff state (Heating Up, Cataclysmic Burst, rune procs, Soul Leech HoT) lives in reusable aura instances tracked by the character.
- Provide stat dependency helpers so effects can grant “Spirit → Spell Power” or haste multipliers temporarily.

### 3. Event-Queue Combat Loop
- Replace the manual `for CurrentTime < Duration { ... }` loop with a priority queue of `PendingAction`s (casts finishing, gcd unlocks, dot ticks, pet swings).
- Dot/pet modules enqueue their own future actions, decoupling timekeeping from the main loop.
- Logging hooks fire as events execute, enabling deterministic combat logs for tests.

### 4. Configuration & Serialization
- Continue loading numeric data from YAML so designers can patch coefficients quickly.
- Add optional protobuf schemas for “simulation requests” if/when we expose an API endpoint or need compatibility with other tooling. Until then, YAML + JSON is sufficient.
- Encapsulate ME loadouts via a registry that validates slot limits and exposes boolean helpers (`HasRune`, `GetStacks`), so spells/effects don’t need to know parsing details.

### 5. Testing & Tooling
- Introduce unit tests per spell (e.g., `spells/immolate_test.go`) using deterministic RNG seeds.
- Add golden tests for the rotation compiler to protect the YAML APL pipeline.
- Wire the combat log mode into CI to catch regressions when modifying the event queue.

## Change Plan (Incremental)
1. **Foundation**
   - Add new aura/effect package and migrate one mechanic (Heating Up) as a proof of concept.
   - Introduce a simple timer struct for cooldowns/GCD to replace manual duration tracking.
2. **Spell Modularization (DONE, keep iterating)**
   - Maintain the per-spell modules (`internal/spells/<spell>.go`) fed by YAML configs.
   - Continue pushing rune/talent modifiers into those modules, backed by the aura framework.
3. **Event Queue**
   - Implement a minimal pending-action queue and port dot ticking + life tap GCD handling to it.
   - Once stable, migrate remaining spells/cooldowns to the queue.
4. **Testing & Tooling**
   - Create baseline unit tests for immolate/incinerate/chaos bolt damage calculations.
   - Add rotation compiler tests and scripted combat-log comparisons.
5. **Optional Protobuf Layer**
   - If/when we need API compatibility, define protobufs for sim requests/responses. This is a later step and only pursued if we expose the sim beyond local CLI usage.

## Vision
By adopting a modular spell registry, shared aura system, and deterministic event queue, wowsim-bb becomes a sustainable platform for experimenting with custom WotLK mechanics. Designers can adjust YAML data without touching Go code, engineers can add new MEs or talents without fighting global state, and future specs (Affliction, Demonology, pets) plug into the same primitives. The resulting codebase stays lightweight compared to the original wotlk sim but inherits its best architectural ideas, giving us the confidence to iterate rapidly without regressions.
