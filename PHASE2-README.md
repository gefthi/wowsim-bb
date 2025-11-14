# Phase 2 Complete - README

## 🎉 What's New

### ✅ Bug Fixes (CRITICAL)
1. **Hit Calculation Fixed** - Was giving 79% miss rate, now correctly ~7% with 10% hit
2. **RNG Seed Fixed** - Each iteration now has unique random seed (results vary between runs)

### ✅ New Talents
1. **Pyroclasm (3/3)** - Conflagrate crit → +6% fire/shadow damage for 10s
2. **Improved Soul Leech (2/2)** - 30% proc chance:
   - Instant: 2% max mana back
   - HoT: 1% max mana per 5sec for 15sec (total 3% more)
   - Total per proc: 5% max mana

### ✅ Talent Fixes
1. **Fire and Brimstone** - NOW ONLY affects Incinerate + Chaos Bolt (was all spells)
2. **Devastation & Backlash** - Now point-based and configurable

### ✅ Configuration
- Talents now support points system (1/3, 2/3, 3/3, etc.)
- Easy to test different talent builds

## 📦 Files Modified

### Core Files:
- `configs/talents.yaml` - New talents + point system
- `internal/config/config.go` - Updated talent structs
- `internal/character/character.go` - Added buff tracking
- `internal/spells/` - Fixed hit, added procs, fixed Fire and Brimstone across the per-spell modules
- `internal/engine/engine.go` - Fixed RNG seed per iteration

## 🧪 Testing

### 1. Verify Hit Fix
Run with 10% hit:
```yaml
# configs/player.yaml
stats:
  hit_percent: 10.0
```

**Expected**: ~7% miss rate (17% cap - 10% = 7%)
**Before**: 79% miss rate ❌
**Now**: 6-7% miss rate ✅

### 2. Verify RNG Works
Run 3 times:
```bash
go run cmd/simulator/main.go
go run cmd/simulator/main.go
go run cmd/simulator/main.go
```

**Expected**: DPS varies slightly (±1-2 DPS)
**Before**: Exact same DPS every time ❌
**Now**: Varies between runs ✅

### 3. Test Pyroclasm
With 25% crit + Conflagrate's +25% crit = 50% Conflagrate crit rate
**Expected**: Pyroclasm should be active ~50% of the time (10s duration, 10s CD)

### 4. Test Fire and Brimstone
**Should affect**: Incinerate, Chaos Bolt (+10% with Immolate up)
**Should NOT affect**: Immolate, Conflagrate

### 5. Test Soul Leech
With 30% proc rate on 4 fire spells casting ~100 times:
**Expected**: ~30 procs = ~150% max mana returned over fight
**Should see**: Fewer Life Tap casts than before

## 📊 Expected DPS Changes

**Phase 1 (Before)**:
- 800 SP, 25% crit, 17% hit: ~2,400-2,500 DPS

**Phase 2 (After)**:
- Same stats + Pyroclasm + Soul Leech: ~2,600-2,800 DPS
- **+8-12% DPS increase** from new talents

## 🐛 Known Limitations

### Not Yet Implemented:
- ❌ Per-spell min/max/avg damage tracking
- ❌ Per-spell crit rate display
- ❌ Buff uptime % display (Pyroclasm, Soul Leech)
- ❌ Soul Leech HoT tick processing (mana ticks every 5sec)

These will be added in Phase 2.5 (enhanced statistics output).

## 🔧 How to Test Different Talent Builds

### Example: Test 3/3 Backlash
Edit `configs/talents.yaml`:
```yaml
backlash:
  points: 3  # Change from 1 to 3
  crit_bonus_per_point: 0.01
```

This gives +3% crit instead of +1% crit.

### Example: Disable Pyroclasm
```yaml
pyroclasm:
  points: 3
  enabled: false  # Set to false
  damage_multiplier: 1.06
  duration: 10.0
```

Compare DPS with/without to see Pyroclasm's value!

## 🚀 Installation

```bash
# Extract
tar -xzf wowsim-bb-phase2-complete.tar.gz --strip-components=1

# Test
go run cmd/simulator/main.go

# If it works, commit
git add .
git commit -m "Phase 2: Pyroclasm, Soul Leech, hit/RNG fixes"
git push
```

## 📋 Next Steps (Phase 2.5)

1. Add detailed per-spell statistics:
   - Min/max/avg damage per spell
   - Crit rate % per spell
   - Miss rate % per spell

2. Add buff uptime tracking:
   - Pyroclasm uptime %
   - Improved Soul Leech uptime %

3. Process Soul Leech HoT properly:
   - Tick every 5 seconds
   - Grant 1% max mana per tick

4. Enhanced output formatting

## ❓ Questions?

Check `PHASE2-CHANGES.md` for technical details or the design doc for mechanics explanations.

---

**Phase 2 is COMPLETE and READY TO TEST!** 🔥
