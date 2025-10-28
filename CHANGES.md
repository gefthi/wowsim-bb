# Player Configuration - Changes Summary

## What Changed

### ✅ NEW FILE: `configs/player.yaml`
Player stats and simulation settings now in YAML instead of hardcoded.

**Contents:**
- Character info (name, level)
- Stats (spell power, crit, haste, spirit, hit, max mana)
- Target type (boss/equal_level) and level
- Simulation settings (duration, iterations)

### ✅ MODIFIED: `internal/config/config.go`
**Added:**
- `Player` struct to hold player configuration
- Player loading in `LoadConfig()` function

**Changes:**
- Config struct now includes `Player Player` field
- LoadConfig now reads `player.yaml`

### ✅ MODIFIED: `cmd/simulator/main.go`
**Changed:**
- Removed hardcoded character stats
- Now loads stats from `cfg.Player.Stats`
- Simulation duration/iterations from `cfg.Player.Simulation`
- Target type from `cfg.Player.Target`
- Better output formatting (shows character name, target level)

## How to Use

1. Extract the tarball over your existing project
2. Edit `configs/player.yaml` with your character's stats
3. Run: `go run cmd/simulator/main.go`
4. No recompilation needed when changing stats!

## Testing

Before committing, verify:
```bash
go run cmd/simulator/main.go
```

Should output:
```
Character: Destruction Warlock (Level 60)
Character Stats:
  Spell Power: 800
  Crit: 25.0%
  ...
```

## Next Steps (After This Works)

Once committed, we'll add:
1. Pyroclasm talent (proc system)
2. Fix Fire and Brimstone (only Incinerate + Chaos Bolt)
3. Fix Backlash (+3% crit, not +1%)
4. Add Improved Soul Leech (mana return on proc)
