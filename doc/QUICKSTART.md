# Quickstart

## Run the Simulator
```bash
go run cmd/simulator/main.go
```

Combat log mode (1 iteration, 60s, prints events):
```bash
go run cmd/simulator/main.go -log-combat
```
Combat log mode forces 1 iteration but uses the configured fight duration; add `-seed-base` to fix RNG.

**Simulator flags**
- `-log-combat` enable combat log (forces 1 iteration, uses configured duration)
- `-seed-base` set RNG seed (0 = random)

## Validate Rotations (APL)
```bash
go run ./cmd/aplvalidate -rotation configs/rotations/destruction-default.yaml
```

**APL validate flags**
- `-rotation` path to rotation YAML (default `configs/rotations/destruction-default.yaml`)

## Stat Weights (central diff)
```bash
go run ./cmd/statweights -seed-base 12345
```
Defaults use `configs/player.yaml` rotation/iterations; override with `-rotation` or `-iterations`.

**Statweights flags**
- Core: `-config-dir` (default `./configs`), `-rotation` (defaults to player.yaml), `-iterations` (override player.yaml), `-seed-base`, `-verbose` (adds +/- DPS columns)
- Sweep mode (set `-stat` to enable; supports `crit|haste|sp`): `-start`, `-stop`, `-step`, `-concurrency` (0 = num CPU), `-avg-seeds` (seeds per point), `-deltas` (include DPS-per-point column), `-output-dir` (default `output/stat_curves`)
- Output includes SP-normalized weights and a Pawn string (uses 1% crit = 14 rating; 1% haste = 10 rating; 1% hit = 10 rating; Spirit hardcoded to 0.6 SP)

## Configure
- `configs/player.yaml`: stats (spell power, crit, haste, spirit, hit, max mana), target type/level, iterations/duration, pet summon, mystic enchants.
- `configs/spells.yaml`, `configs/talents.yaml`, `configs/constants.yaml`: numeric tuning.
- `configs/rotations/`: YAML APLs; edit and re-validate without recompiling.

## Notes
- PvE Power currently baked as a 1.25 multiplier in `internal/spells/core.go` (move to config soon).
- Requires Go 1.21+.
