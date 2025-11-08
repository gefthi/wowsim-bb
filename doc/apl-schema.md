# Rotation APL Schema (Draft)

Status: **Design-ready, not implemented yet.**  
Purpose: Capture the YAML data model we will build in Phase 4 so future iterations stay aligned.

## File Layout

```
configs/
  rotations/
    base.yaml
    gulnas-chosen.yaml
    experimental/
      execute.yaml
```

- Each YAML file defines one named rotation.
- Files can `import` others to reuse blocks. Imports are processed depth-first and merged (later entries override earlier ones).

## Top-Level Keys

```yaml
name: "Default - No Rune"
description: "Baseline Destruction rotation without Mystic Enchants."
imports:
  - rotations/shared/baseline.yaml
variables:
  life_tap_threshold: 0.30
  immolate_refresh: 3
rotation:
  - ...
```

| Key         | Type        | Notes                                                        |
|-------------|-------------|--------------------------------------------------------------|
| `name`      | string      | Friendly label for UI/logs                                   |
| `description` | string    | Optional doc text                                            |
| `imports`   | []string    | Relative paths, resolved from `configs/rotations`            |
| `variables` | map         | User-tunable constants referenced via `${var}`               |
| `rotation`  | []Action    | Ordered list evaluated top→bottom each combat decision       |

## Action Definition

```yaml
- action: cast_spell
  spell: immolate
  when: {...}
  tags: [ opener ]

- action: use_item
  item: "Talisman of Flames"
  when:
    cooldown_ready: talisman_of_flames

- action: wait
  duration_seconds: 0.2
  when:
    all:
      - cooldown_remaining:
          spell: chaos_bolt
          lt_seconds: 0.2
      - buff_active:
          buff: backdraft
```

Supported `action` values (initial scope):

| Action        | Required Fields               | Description                                    |
|---------------|-------------------------------|------------------------------------------------|
| `cast_spell`  | `spell`                       | Casts a spell if conditions allow              |
| `use_item`    | `item`                        | Triggers on-use trinket or consumable          |
| `wait`        | `duration_seconds`            | Advances time/GCD deliberately                 |
| `macro`       | `steps` (nested actions)      | Executes sub-actions sequentially              |

Future extensions: `channel`, `pet_command`, `apply_aura`.

## Conditions

Conditions are expressed as small YAML objects. Common keys:

```yaml
when:
  all:
    - buff_active:
        buff: pyroclasm
        min_remaining: 3
    - cooldown_ready: chaos_bolt
```

Available primitives (initial set):

| Predicate          | Parameters                         | Meaning                                                |
|--------------------|------------------------------------|--------------------------------------------------------|
| `buff_active`      | `buff`, optional `min_remaining`   | Player buff up with optional seconds remaining check   |
| `debuff_active` / `dot_remaining` | `debuff`/`spell`, optional `lt_seconds` / `gt_seconds` | Target debuff window checks                            |
| `cooldown_ready`   | `spell` or `item`                  | Off cooldown                                           |
| `cooldown_remaining` | `spell`/`item`, `lt_seconds`     | Time until ready                                       |
| `resource_percent` | `resource` (`mana`, `health`...), comparator (`lt`, `gt`) | Resource threshold                                     |
| `charges`          | `buff`, optional `gte`, `lte`      | e.g., Backdraft charge count                           |
| `time_elapsed`     | `gt_seconds` / `lt_seconds`        | Fight timer gates                                      |
| `true` / `false`   | none                               | Always/never (useful for quick toggles)                |

Combinators:
- `all`: every child true (logical AND).
- `any`: at least one child true (logical OR).
- `not`: negate a single child.

If `when` is omitted on an action, it defaults to `true`.

## Execution Semantics

1. **Decision loop** begins whenever the character is ready to act (GCD finished, current cast done).
2. Evaluate `rotation` list from top to bottom:
   - First action whose `when` resolves true attempts to execute.
   - On success, restart evaluation at the top.
   - On failure (e.g., not enough mana) the engine falls through to the next entry.
3. If nothing succeeds, fallback action (`wait` for GCD or emergency Life Tap) runs to prevent stalls.

## Validation & Debugging

- `sim validate-rotation <file>` – checks imports, unknown spells/items, missing fields, bad variable references.
- `debug_apl: true` (CLI flag) – logs the first N decisions with condition results to help iterate quickly.

## Iteration Plan

1. **Loader**: read YAML, resolve imports, build AST with variables substituted.
2. **Validator**: ensure referenced spells/items/conditions exist.
3. **Executor (Phase 1)**: replace current hardcoded rotation with compiled actions.
4. **Extensions**: add item usage, macro chaining, rune-specific condition helpers.

Everything in this document is a contract for the implementation tasks. We update it after each milestone to keep future sessions aligned.
