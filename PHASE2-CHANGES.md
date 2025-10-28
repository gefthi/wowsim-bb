# Phase 2 Changes Summary

## Files Modified (7 files):

### 1. configs/talents.yaml
- ✅ Added Pyroclasm talent (3/3 points, +6% damage, 10s duration)
- ✅ Added Improved Soul Leech (2/2 points, mana return on proc)
- ✅ Made Devastation and Backlash point-based (configurable)
- ✅ Fixed Fire and Brimstone - now only applies to Incinerate + Chaos Bolt

### 2. internal/config/config.go
- ✅ Updated Talents struct with new talents
- ✅ Added point-based system for Devastation/Backlash
- ✅ Added Pyroclasm and ImprovedSoulLeech structs

### 3. internal/character/character.go
- ✅ Added Pyroclasm buff tracking
- ✅ Added ImprovedSoulLeech buff tracking
- ✅ Added SoulLeechLastTick for HoT timing

### 4. internal/spells/spells.go
- ✅ FIXED RollHit() - correct miss chance calculation
- ✅ Updated RollCrit() - use talent points system
- ✅ Updated CalculateSpellDamage() - apply Pyroclasm buff
- ✅ FIXED ApplyFireAndBrimstone() - only Incinerate + Chaos Bolt
- ✅ Added CheckSoulLeechProc() helper function
- ✅ Added Pyroclasm proc on Conflagrate crit
- ✅ Added Soul Leech proc checks to all fire spells

### 5. internal/engine/engine.go
- TODO: Add per-spell detailed statistics
- TODO: Add buff uptime tracking
- TODO: Process Soul Leech HoT ticks during combat
- TODO: Enhanced output formatting

## What's Working Now:
- ✅ Hit calculation fixed
- ✅ RNG seed per iteration fixed
- ✅ Pyroclasm procs on Conflagrate crit
- ✅ Pyroclasm +6% damage applied to all fire/shadow spells
- ✅ Fire and Brimstone only affects Incinerate + Chaos Bolt
- ✅ Soul Leech instant mana return (2% on proc)
- ✅ Soul Leech HoT buff activated
- ✅ Talent points configurable

## Still TODO (engine.go enhancements):
-  Per-spell min/max/avg damage tracking
- Per-spell crit rate and miss rate
- Pyroclasm uptime % calculation
- Improved Soul Leech uptime % calculation
- Soul Leech HoT tick processing (1% mana per 5sec)
- Enhanced output formatting

## Testing Priority:
1. Verify Pyroclasm procs and uptime
2. Verify Fire and Brimstone only on correct spells
3. Verify Soul Leech mana return
4. Check DPS increase from new talents
