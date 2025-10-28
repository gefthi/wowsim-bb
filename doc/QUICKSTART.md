# Quick Start Guide - WotLK Destro Sim Phase 1

## What You Have

✅ **Complete Phase 1 MVP** - Fully functional command-line simulator!

## Files Included

- Complete Go project structure
- All YAML configuration files (spells, talents, constants)
- Core simulation engine with rotation logic
- Character system with mana tracking
- Spell casting with hit/miss/crit
- README with full documentation

## How to Run (3 Steps)

### 1. Extract the archive
```bash
tar -xzf wotlk-destro-sim.tar.gz
cd wotlk-destro-sim
```

### 2. Install Go dependencies
```bash
go mod download
```

### 3. Run the simulator!
```bash
go run cmd/simulator/main.go
```

That's it! You'll see simulation results printed to console.

## Customizing Your Character

Edit `cmd/simulator/main.go` around line 22:

```go
charStats := character.Stats{
    SpellPower: 800,   // YOUR spell power here
    CritPct:    25.0,  // YOUR crit % here
    Spirit:     200,   // YOUR spirit here
    HitPct:     17.0,  // YOUR hit % here
    MaxMana:    8000,  // YOUR max mana here
}
```

Then run again!

## Editing Spell Data

Want to change spell damage or mana costs? Edit the YAML files in `configs/`:

- `spells.yaml` - All spell numbers
- `talents.yaml` - Talent modifiers  
- `constants.yaml` - Server settings

No recompilation needed - just run again!

## What Phase 1 Does

✅ Simulates basic Destro rotation  
✅ Tracks mana and uses Life Tap  
✅ Rolls for hits, misses, crits  
✅ Runs 1000 iterations for accuracy  
✅ Outputs total DPS and spell breakdown  

## What's Coming Next

**Phase 2**: Backdraft system (cast time & GCD reduction)  
**Phase 3**: Mystic Enchants (all 11 MEs with slot selection)  
**Phase 4**: Stat weights calculation  
**Phase 5**: Haste mechanics  
**Phase 6**: React web UI + polish  
**Phase 7**: Action Priority Lists (user-editable rotations)  

## Testing Your Setup

Expected DPS range with example stats (800 SP, 25% crit, hit capped):
- **2,000 - 2,800 DPS** for a 5-minute boss fight

If you see this range, everything is working correctly!

## Questions?

Check the full README.md or the design document (warlock-simcraft-design.md) for details.

---

**Ready to simulate!** 🔥
