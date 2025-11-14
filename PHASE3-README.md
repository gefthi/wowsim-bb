# Phase 3 Complete - Backdraft System

## 🔥 What's New

### ✅ Backdraft Talent
- Conflagrate now grants **3 Backdraft charges** (15s duration).
- Each charge reduces the next Destruction spell's **cast time and GCD by 30%**.
- Charges are consumed on every Destruction spell cast (including instants like Conflagrate).
- Backdraft uptime and **average charges** are tracked in the results output.

### ✅ Configuration
- Added `backdraft` block to `configs/talents.yaml` so charges, duration, and reduction % can be tuned without code changes.
- Talent can be disabled by setting `enabled: false` or `points: 0`.

### ✅ Engine & Reporting
- New helper logic across `internal/spells/*.go` applies cast/GCD reductions, refreshes charges after Conflagrate, and consumes charges per spell.
- `internal/engine/engine.go` now tracks Backdraft uptime seconds and charge-weighted seconds, and reports both uptime % and average charges.

## 📦 Key Files
- `configs/talents.yaml` – configurable Backdraft parameters.
- `internal/config/config.go` – Backdraft struct added to the talent config.
- `internal/spells/` – Backdraft activation & consumption logic baked into each spell module.
- `internal/engine/engine.go` – Buff tracking, charge metrics, and output formatting.
- `README.md` / `cmd/simulator/main.go` – updated to Phase 3 status and sample output.

## 🧪 Testing
```bash
tmp=$(mktemp -d)
GOCACHE=$tmp go run ./cmd/simulator
```

**Expected**: report shows the Backdraft uptime line (≈55% with default stats) and per-spell cast times noticeably faster after Conflagrate usage.

## 🚀 Next Steps (Phase 4)
1. Mystic Enchants toggles (e.g., Gul'na's Chosen).
2. Rune-specific interactions (non-consuming Backdraft windows).
3. Keep expanding reporting to surface rune/buff uptimes.
