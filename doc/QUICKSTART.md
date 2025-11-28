# Quickstart

## Run the Simulator
```bash
go run cmd/simulator/main.go
```

Combat log mode (1 iteration, 60s, prints events):
```bash
go run cmd/simulator/main.go -log-combat
```

Deterministic seeds (helpful for comparisons/stat weights):
```bash
go run cmd/simulator/main.go -seed-base 12345
```

## Validate Rotations (APL)
```bash
go run ./cmd/aplvalidate -rotation configs/rotations/destruction-default.yaml
```

## Stat Weights (central diff)
```bash
go run ./cmd/statweights -seed-base 12345
```
Defaults use `configs/player.yaml` rotation/iterations; override with `-rotation` or `-iterations`.

## Configure
- `configs/player.yaml`: stats (spell power, crit, haste, spirit, hit, max mana), target type/level, iterations/duration, pet summon, mystic enchants.
- `configs/spells.yaml`, `configs/talents.yaml`, `configs/constants.yaml`: numeric tuning.
- `configs/rotations/`: YAML APLs; edit and re-validate without recompiling.

## Notes
- PvE Power currently baked as a 1.25 multiplier in `internal/spells/core.go` (move to config soon).
- Requires Go 1.21+.
