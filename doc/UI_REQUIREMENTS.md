# UI Requirements

Source of truth remains YAML; UI is a guard-railed editor that reads/writes the existing files.

## Player Config (configs/player.yaml)
- Pet: dropdown (currently only Imp; future pets appear disabled with “coming soon”).
- Rotation: dropdown populated from `configs/rotations/*.yaml`; button to open/edit selected rotation.
- Mystic Enchants: checkbox groups per quality; only implemented enchants shown; enforce slot/quantity limits (disable beyond cap).
- Stats/Simulation: numeric inputs with min/max; show derived hit/crit summaries for feedback.
- PvE Power: display read-only “1.25 (hardcoded)” until configurable.

## Rotations (configs/rotations/*.yaml)
- Templates/presets to load/fork (Default, Decisive, etc.); “Save as new file” to avoid overwriting presets; backups on save (timestamped copy).
- Imports: dropdown to add existing rotation files.
- Action builder (no free text):
  - Action types: `cast_spell`, `wait`, `use_item`, `macro` (with sub-steps).
  - Spell/item fields use dropdowns sourced from known identifiers (`internal/apl/names.go`).
  - Condition builder with combinators (`all`/`any`/`not`) and predicates (`buff_active`, `debuff_active`, `dot_remaining`, `cooldown_ready/remaining`, `resource_percent`, `charges`), each with dropdown comparators/fields.
- Validation: button to run `cmd/aplvalidate` on the current file and surface pass/fail; pre-save client-side schema guardrails to block unknown identifiers/keys.
- UX: inline diff vs last save, “open file” link from player config, and explicit warning if template differs from disk before overwriting.
