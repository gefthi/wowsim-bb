# APL Schema (Rotation YAML)

## File Layout
- Rotations live in `configs/rotations/`; support `imports` for reuse.

## Top-Level Keys
```yaml
name: "Default - No Rune"
description: "Baseline Destruction rotation"
imports: [rotations/shared/base.yaml]
variables:
  life_tap_threshold: 0.30
rotation:
  - action: cast_spell
    spell: immolate
    when: {...}
```

## Actions
- `cast_spell` (spell)
- `use_item` (item)
- `wait` (duration_seconds)
- `macro` (steps: [actions])

## Conditions (`when`)
- Combinators: `all`, `any`, `not`
- Base literals: `true`, `false`
- Predicates (comparators support `lt`, `lte`, `gt`, `gte`):
  - `buff_active` {buff, min_remaining?, max_remaining?}
  - `debuff_active` {debuff, min_remaining?, max_remaining?}
  - `dot_remaining` {spell, lt_seconds?, lte_seconds?, gt_seconds?, gte_seconds?}
  - `cooldown_ready` {spell/item}
  - `cooldown_remaining` {spell/item, lt_seconds?, lte_seconds?, gt_seconds?, gte_seconds?}
  - `resource_percent` {resource, lt?, lte?, gt?, gte?}
  - `charges` {buff, lt?, lte?, gt?, gte?}
  - (Use `all`/`any`/`not` to compose)

Identifiers (spells/buffs/resources) must exist in `internal/apl/names.go`.

## Execution Model
- Evaluate list top→bottom each decision; first passing action executes, then restart at top.
- On failure (e.g., OOM), fall through to next entry.

## Validation
```bash
go run ./cmd/aplvalidate -rotation configs/rotations/destruction-default.yaml
```

## Known Identifiers (current set)
- Spells: `immolate`, `conflagrate`, `chaos_bolt`, `incinerate`, `life_tap`, `inferno`, `curse_of_doom`, `curse_of_agony`, `curse_of_the_elements`
- Buffs: `pyroclasm`, `backdraft`, `guldans_chosen`, `cataclysmic_burst`, `heating_up`, `improved_soul_leech`, `soul_leech`, `life_tap_buff`, `shadow_trance`, `demonic_soul`
- Debuffs: `immolate`, `curse_of_doom`, `curse_of_agony`, `curse_of_the_elements`
- Resources: `mana`, `health`, `soul_shards`

Add new identifiers in `internal/apl/names.go` if you extend the system.
