# Project State
Last Updated: 2025-11-28

## Completed
- Core sim engine with event queue (cast timeline, GCD, cooldowns, logging hooks)
- Spells: Immolate (direct+DoT), Incinerate, Chaos Bolt, Conflagrate, Life Tap, Soul Fire (Decisive buff aware)
- Talents/runes: Emberstorm, Improved Immolate, Aftermath, Fire and Brimstone, Ruin, Devastation, Backlash, Pyroclasm, Improved Soul Leech (instant + HoT buff), Backdraft (-30% cast/GCD, tracked uptime/avg charges), Demonic Power, Empowered Imp, Decisive Decimation (buff applied by Conflagrate)
 - Mystic Enchants support: selection in `configs/player.yaml`; implemented effects include Destruction Mastery, Cataclysmic Burst, Heating Up, Gul'dan's Chosen, Agent of Chaos (hasteable Immolate ticks + CB CDR, direct penalty), Chaos Manifesting, Decisive Decimation, Inner Flame, Endless Flames (Pyroclasm duration), Glyphs of Life Tap / Conflagrate / Chaos Bolt / Incinerate / Immolate, Demonic Aegis, Suppression, Improved Imp, Curse of the Elements debuff (10% damage multiplier)
- Haste now applied to casts/GCD (respecting min GCD); DoT haste gated behind Agent of Chaos; Immolate tick scheduling fixed to honor Cataclysmic extensions without gaps
- Data-driven config: YAML for constants, player stats, spells, talents, runes; rotation via YAML APL with loader/compiler/validator
- Modular spells, shared aura/timer helpers in `internal/effects`, per-spell files under `internal/spells/`
- CLI: `go run cmd/simulator` (optional `-log-combat` uses configured duration) with seed flag; APL validator `go run ./cmd/aplvalidate`; stat weights helper `go run ./cmd/statweights`

## In Progress
- Migrate remaining buffs/debuffs to aura framework (Backdraft state, Chaos Manifesting)
- Extend event queue to other periodic effects (Improved Soul Leech HoT ticks, pets)
- Add per-spell/unit tests now that modules are isolated

## Planned / Next
- Phase 4: Mystic Enchants expansion and rune-specific interactions
- Phase 5: Stat weights refinement (delta tuning, variance controls)
- Phase 6: Haste mechanics (remaining: pet applicability, tick breakpoints)
- Phase 7: UI/APL tooling polish
- Next session: minimal UI to edit `configs/player.yaml`; rotation editor (APL operators/conditions) with guardrails captured in `doc/UI_REQUIREMENTS.md`
- Consider pet system expansion (affects PvE Power decision)

## Known Issues / TODO
- PvE Power: currently a hardcoded 1.25 multiplier in `internal/spells/core.go`; move to config (`configs/player.yaml`) and thread through calculations.
- Decide whether PvE Power applies to pets; if yes, apply in `internal/engine/pets.go`.
- Improve validation around YAML configs (ranges, names) and add regression tests for damage math.

## Recent Decisions
- Event queue drives time; DoT ticks scheduled, not polled.
- Keep data out of binaries (YAML for stats/spells/talents/rotations).
- Use per-spell modules plus shared effect/aura helpers for extensibility.
- APL lives in YAML, compiled at runtime; validator shipped as CLI.

## Where to Look
- Old docs archived in `doc/old_doc/` for deep dives (design doc, phase summaries, evaluations).
- Rotations live in `configs/rotations/`; validate with `go run ./cmd/aplvalidate -rotation <file>`.
- Session briefs go under `doc/sessions/` (see `SESSION_PLAYBOOK.md`).
