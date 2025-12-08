# Project Progress

Current scope: Phase 3 complete with Backdraft, detailed reporting, and APL-driven rotations. Legacy docs remain under `doc/` for history; this file is the live status.

## Completed
- Core engine with event queue, haste-aware GCD/casts, cooldowns, logging hooks.
- Spells: Immolate (direct + DoT snapshot), Incinerate (Immolate bonus), Chaos Bolt, Conflagrate (DoT split), Life Tap, Soul Fire; Imp pet with talent/rune hooks.
- Talents/runes: Emberstorm, Improved Immolate, Aftermath, Fire and Brimstone (Incinerate/Chaos Bolt only), Ruin, Devastation, Backlash, Pyroclasm, Backdraft (charges, uptime), Improved Soul Leech (instant + HoT), Demonic Power, Empowered Imp; Mystic Enchants/runes listed in `docs/BUSINESS_RULES.md`.
- Data-driven YAML configs for spells/talents/constants/player; rotations expressed in YAML APL with validator.
- Reporting: per-spell damage stats, crit/miss rates, buff uptimes (Pyroclasm, Soul Leech, Backdraft avg charges), Shadow Trance proc counts, combat log mode, unique seed per iteration.
- Periodic scheduling: Improved Soul Leech HoT ticks and Imp Firebolt chain-casting (with mana regen delays) run through the event queue.

## In Progress
- Refactor Backdraft and Chaos Manifesting to use the shared `effects.Aura` helper (unified stacks/expiration/logging) instead of bespoke timers.
- Add per-spell/unit tests now that modules are isolated and APL is data-driven.

## Next Up
- Phase 4: Broaden Mystic Enchants/rune-specific interactions. Baseline effects are in (Destruction Mastery, Cataclysmic Burst, Heating Up, Agent of Chaos, Gul'dan's Chosen with Backdraft non-consuming windows, Chaos Manifesting, Pure Shadow, Dusk till Dawn, Pyroclasmic Shadows, Unstable Void via Shadowfury, Nightfall/Twilight Reaper, Cursed Shadows); remaining work is adding more ME hooks in `docs/BUSINESS_RULES.md` (Shadow Siphon, Unstable Void’s Shadow Crash) plus better toggles/guardrails.
- Refresh player config UX: make `configs/player.yaml` rune selection friendlier (grouping by rarity/spec, clearer toggles) and update APL samples once new MEs land.
- Phase 5: Stat weights refinement (delta tuning, variance controls).
- Phase 6: Haste mechanics polish (pet applicability, tick breakpoints) and PvE Power configurability.
- Phase 7+: UI/APL tooling polish (rotation editor per `doc/UI_REQUIREMENTS.md`).

## Known Issues / TODO
- PvE Power still hardcoded at 1.25 in `internal/spells/core.go`; move to config and decide pet applicability.
- Tighten YAML validation (ranges/names) and add regression tests for damage math.
- Clarify Improved Soul Leech proc hooks on any future spell additions to keep mana model consistent.
- UI polish: add saved rune/Mystic Enchant profiles for quick swapping; redesign the UI to be visually cohesive and friendly; longer-term, revamp the rotation editor UX to be less clunky.
