# Quickstart

## Run the Simulator
```bash
go run cmd/simulator/main.go
```

Combat log mode (1 iteration, 60s, prints events):
```bash
go run cmd/simulator/main.go -log-combat
```

## Validate Rotations (APL)
```bash
go run ./cmd/aplvalidate -rotation configs/rotations/destruction-default.yaml
```

## Configure
- `configs/player.yaml`: stats (spell power, crit, haste, spirit, hit, max mana), target type/level, iterations/duration, pet summon, mystic enchants.
- `configs/spells.yaml`, `configs/talents.yaml`, `configs/constants.yaml`: numeric tuning.
- `configs/rotations/`: YAML APLs; edit and re-validate without recompiling.

## Notes
- PvE Power currently baked as a 1.25 multiplier in `internal/spells/core.go` (move to config soon).
- Requires Go 1.21+.
