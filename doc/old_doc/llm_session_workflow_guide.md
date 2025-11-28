# Complete LLM Session Documentation Workflow

A comprehensive guide for maintaining context across multiple LLM sessions when building complex projects.

---

## Table of Contents

1. [Overview](#overview)
2. [Files You'll Maintain](#files-youll-maintain)
3. [Phase 0: First-Time Project Setup](#phase-0-first-time-project-setup)
4. [Phase 1: First Session](#phase-1-first-session)
5. [Phase 2: Regular Sessions](#phase-2-regular-sessions)
6. [Special Cases](#special-cases)
7. [Document Update Frequency](#document-update-frequency)
8. [Token Budget Summary](#token-budget-summary)
9. [Complete Workflow Checklist](#complete-workflow-checklist)
10. [Minimal Workflow Alternative](#minimal-workflow-alternative)
11. [Advanced Techniques](#advanced-techniques)

---

## Overview

**Problem:** LLMs don't remember previous conversations. Each new session starts from zero.

**Solution:** Structured documentation that provides exactly what the LLM needs to continue work efficiently.

**Key Principle:** Load context incrementally - don't dump entire project history into each session.

---

## Files You'll Maintain

```
project/
├── docs/
│   ├── PROJECT_VISION.md           (created once, rarely updated)
│   ├── TECH_STACK.md               (created once, rarely updated)
│   ├── PROJECT_STATE.md            (updated every session)
│   ├── sessions/
│   │   ├── session_01_integration.md
│   │   ├── session_02_integration.md
│   │   ├── session_03_integration.md
│   │   └── ...
│   └── templates/
│       ├── SESSION_START.md        (template you copy each time)
│       └── SESSION_END.md          (template you copy each time)
├── src/
│   └── [your code]
└── data/
    └── [your data files]
```

---

## Phase 0: First-Time Project Setup

**Time Required:** 30 minutes, ONE TIME

### Step 1: Create Core Documents

#### File: `docs/PROJECT_VISION.md` (500-800 tokens)

```markdown
# WoW Combat Simulator

## Vision
Simulate [Expansion] [Class/Spec] DPS for gear/talent/rotation optimization.
Goal: Match SimulationCraft accuracy within 5%.

## Target Users
- Theorycrafters optimizing builds
- Players choosing between gear pieces
- Me learning simulation techniques

## Core Features
1. Character stat calculation (base stats → final stats)
2. Spell damage simulation (formulas, scaling, crits)
3. Rotation execution (priority-based decision making)
4. Event-driven combat simulation (10k+ iterations)
5. Results analysis (DPS, stat weights, ability breakdown)

## Explicitly NOT Building (v1)
- PvP simulation
- Multi-target scenarios  
- Real-time encounter mechanics
- Raid buffs/debuffs
- Movement optimization

## Success Criteria
- Simulate 10k iterations in < 5 seconds
- Within 5% of SimC for Patchwerk (single target, no movement)
- Support all major talents/glyphs for [spec]

## Why This Approach
- Data-driven: Spells/talents in JSON, easy to update
- Modular: Each system isolated for testing
- Educational: Learning simulation techniques
```

#### File: `docs/TECH_STACK.md` (300-500 tokens)

```markdown
# Technical Stack

## Backend
- **Language**: Python 3.11+
- **Why**: Numerical accuracy, NumPy for performance, clear syntax
- **Libraries**: NumPy (calculations), Pytest (testing)

## Frontend (future)
- **Language**: TypeScript + React
- **Why**: Type safety, component reusability
- **Libraries**: D3.js (visualizations), Recharts (charts)

## Data Storage
- **Format**: JSON files
- **Why**: Human-readable, easy to update when patches change
- **Structure**: Separate files per class/spec

## Architecture Principles
1. **Data-driven**: Game content in JSON, not code
2. **Event-driven**: Discrete event simulation
3. **Stateless functions**: Pure calculations where possible
4. **Test coverage**: 80%+ on core systems

## Hard Constraints
- Must run on single CPU (no distributed computing)
- No external APIs (offline simulation)
- Support [Expansion] formula/mechanics only
- Must handle 50,000+ events per simulation

## Development Environment
- Python venv for dependencies
- Git for version control
- VS Code as IDE
```

#### File: `docs/PROJECT_STATE.md` (200-400 tokens, updated frequently)

```markdown
# Project State

Last Updated: [Date]

## Completed ✅
*[Empty initially]*

## In Progress 🚧
*[Empty initially]*

## Planned ⏳
1. Phase 1: Stats System
2. Phase 2: Spell System  
3. Phase 3: Rotation Engine
4. Phase 4: Event Simulation
5. Phase 5: Results Analysis

## Known Issues
*[Empty initially]*

## Recent Decisions
*[Empty initially]*
```

### Step 2: Create Templates

#### File: `docs/templates/SESSION_START.md`

```markdown
# Session [NUMBER] - [DATE]

## Context Loading

### Project Identity
[Copy from PROJECT_VISION.md - first 3 paragraphs]

### Current State  
[Copy from PROJECT_STATE.md]

### Today's Goal
[FILL THIS: What are you implementing today?]

Requirements:
- [FILL: Requirement 1]
- [FILL: Requirement 2]

Success Criteria:
- [FILL: How you know it's done]

### Relevant Existing Systems
[FILL: Paste integration briefs from previous sessions that you'll use today]

### Reference Data for Today
[FILL: Formulas, data structures, examples needed for today]
```

#### File: `docs/templates/SESSION_END.md`

```markdown
# Session Wrap-Up

## What Was Built
[Ask Claude: List files created/modified with one-line descriptions]

## Integration Brief  
[Ask Claude: Generate brief with classes, methods, signatures, usage examples]

## Test Status
[Ask Claude: What's tested, what needs testing]

## Known Issues / TODOs
[Ask Claude: Anything incomplete or needs attention]

## Updates Needed
- [ ] Update PROJECT_STATE.md (move current phase to completed)
- [ ] Save integration brief to docs/sessions/session_XX_integration.md
- [ ] Update any changed decisions

## Next Session Should
[Ask Claude: Suggest logical next step]
```

---

## Phase 1: First Session

**Note:** First session is different - you're establishing the foundation.

### START of First Session

**What to provide Claude:**

```markdown
# New Project: WoW Combat Simulator

[Paste entire PROJECT_VISION.md]

[Paste entire TECH_STACK.md]

## Today: Session 1 - Stats System

Goal: Implement character stat calculation system.

Given base stats and gear, calculate:
- Final primary stats (Intellect, Stamina)
- Convert rating to percentages (Haste Rating → Haste %)
- Derive stats (Spell Power, Crit Chance)

Reference formulas for [Expansion]:
[Paste formulas - you provide these]

Haste % = Haste Rating / 425 (at level 85)
Crit % = 5% (base) + (Crit Rating / 179.28) + (Intellect / 2000)
Spell Power = Intellect * 1.0 + gear Spell Power

Success criteria:
- Character class with stat properties
- Stat calculator with rating conversions  
- Test with my character stats: [example]
- Results match WoWhead calculator

Files to create:
- src/character.py
- src/stat_calculator.py
- tests/test_stats.py
```

**Token count: ~1,200-1,500**
- Vision: 500-800
- Tech stack: 300-500  
- Today's goal: 400-500

### END of First Session

**Ask Claude:**

```markdown
Session wrap-up time. Please provide:

1. **What we built today**
   - List files created  
   - Main classes and their purposes

2. **Integration brief for next session**
   Include:
   - Classes with key methods and signatures
   - Usage examples (2-3 lines each)
   - Important notes for future integration

3. **Test status**
   - What's tested
   - What needs testing later

4. **Suggested next step**
   - What should Session 2 tackle?
```

**Example Claude Response:**

```markdown
# Integration Brief: Stats System (Session 1)

## What Was Built

### File: src/character.py

```python
class Character:
    def __init__(self, name: str, level: int)
    
    # Base Stats
    self.intellect: int
    self.stamina: int
    
    # Gear Stats (ratings)
    self.haste_rating: int
    self.crit_rating: int
    self.mastery_rating: int
    
    # Calculated Stats (computed from above)
    self.spell_power: float
    self.haste_percent: float
    self.crit_percent: float
```

### File: src/stat_calculator.py

```python
class StatCalculator:
    @staticmethod
    def calculate_haste_percent(haste_rating: int, level: int) -> float
        """Converts Haste Rating to Haste %"""
        # Returns: haste_rating / rating_per_percent
    
    @staticmethod
    def calculate_crit_percent(crit_rating: int, intellect: int, level: int) -> float
        """Converts Crit Rating + Intellect to Crit %"""
        # Returns: base_crit + (crit_rating / rating_per_percent) + (intellect / conversion)
    
    @staticmethod
    def calculate_spell_power(intellect: int, gear_sp: int) -> float
        """Calculates total Spell Power"""
        # Returns: intellect * mult + gear_sp
```

## Usage Examples

### Creating a character and calculating stats

```python
from character import Character
from stat_calculator import StatCalculator

char = Character(name="Pyromancer", level=85)
char.intellect = 5000
char.haste_rating = 2500
char.crit_rating = 1800

# Calculate derived stats
char.haste_percent = StatCalculator.calculate_haste_percent(
    char.haste_rating, 
    char.level
)
# Result: 5.88% haste
```

## Integration Points for Future Sessions

**Spell System** (next session) should:
- Accept a Character object to get stats
- Use `character.spell_power` for damage scaling
- Use `character.crit_percent` for crit chance calculations
- Use `character.haste_percent` for cast time reduction

**Example integration:**
```python
def calculate_spell_damage(spell, character):
    base_damage = spell.base_damage
    sp_scaling = spell.coefficient * character.spell_power
    return base_damage + sp_scaling
```

## Important Notes

- All ratings use level-specific conversion rates
- Spell Power is calculated from Intellect (1:1 ratio at level 85)
- Crit % includes 5% base crit for all casters
- Stats are recalculated whenever gear changes

## Test Coverage
- ✅ Haste rating conversion
- ✅ Crit rating conversion
- ✅ Spell Power calculation
- ✅ Full stat calculation with real character data
- ⏳ Stat caps/diminishing returns (future)

## Next Session Should
Implement Spell System:
- Spell class that reads from JSON
- Damage calculation using character stats
- Crit damage multipliers
- Cast time with haste scaling
```

**Save Claude's response to:**
- `docs/sessions/session_01_integration.md`

**Update manually:**

`docs/PROJECT_STATE.md`:
```markdown
## Completed ✅
- Session 1: Stats System (character.py, stat_calculator.py)
  - Rating conversions working
  - Test coverage: 85%

## In Progress 🚧
*[Empty]*

## Planned ⏳
2. Phase 2: Spell System
3. Phase 3: Rotation Engine
4. Phase 4: Event Simulation
5. Phase 5: Results Analysis

## Recent Decisions
- Using static methods in StatCalculator (no state needed)
- Level-specific rating conversions (stored as constants)
```

---

## Phase 2: Regular Sessions (2 through N)

### START of Each Session

#### Step 1: Copy template

```bash
cp docs/templates/SESSION_START.md session_prep.md
```

#### Step 2: Fill in the template

Open `session_prep.md` and fill in:

```markdown
# Session 4 - Nov 24, 2025

## Context Loading

### Project Identity
Building WoW Combat Simulator for [Expansion] [Spec].
Goal: Match SimC within 5% accuracy.
Tech: Python + NumPy, data-driven design (spells in JSON).

### Current State

✅ Session 1: Stats System (character.py, stat_calculator.py)
   - Rating conversions, Spell Power calculation
   
✅ Session 2: Spell System (spell.py, spell_executor.py)  
   - 12 Fire Mage spells in JSON
   - Damage calculation with SP scaling, crit
   
✅ Session 3: Rotation Engine (rotation_engine.py)
   - Condition parser, priority evaluator
   - Selects highest-priority available spell
   
🚧 Session 4: Event System (simulation.py)
   - Last session: Event queue working
   - Today: Add buff tracking

### Today's Goal

Implement buff tracking in event system.

Requirements:
- Track active buffs with expiration times
- Create buff_expire events automatically  
- Handle buff consumption (Hot Streak consumed by Pyroblast)

Success Criteria:
- Two crit Fireballs → Hot Streak buff applied
- Hot Streak expires after 10s if unused
- Casting Pyroblast consumes Hot Streak immediately
- Tests pass for buff lifecycle

### Relevant Existing Systems

#### Event System (built last session)

From docs/sessions/session_03_integration.md:

```python
class EventQueue:
    def add_event(self, event: Event, time: float) -> None
        """Add event to queue, sorted by time"""
    
    def get_next_event(self) -> Optional[Event]
        """Pop and return earliest event"""
    
    def peek_time(self) -> Optional[float]
        """Look at next event time without removing"""

class Event:
    def __init__(self, time: float, event_type: str, data: dict)
    
    # Properties
    self.time: float          # When event occurs
    self.event_type: str      # "spell_complete", "cooldown_ready", etc.
    self.data: dict           # Event-specific data
```

Usage:
```python
event = Event(
    time=15.5, 
    event_type="spell_complete", 
    data={"spell": "fireball", "result": "crit"}
)
queue.add_event(event, 15.5)
```

#### Spell Result Tracking (from Session 2)

```python
class Simulation:
    def __init__(self):
        self.recent_spell_results = []  # Last N spell results
        
    def track_spell_result(self, spell_name: str, result: str):
        """Track spell outcome for proc detection"""
        self.recent_spell_results.append({
            "spell": spell_name,
            "result": result  # "hit", "crit", "miss"
        })
        # Keep only last 10 results
        if len(self.recent_spell_results) > 10:
            self.recent_spell_results.pop(0)
```

### Reference Data for Today

Hot Streak mechanic (from WoW Cataclysm):
- **Trigger**: Two spell crits in a row
- **Effect**: Next Pyroblast is instant cast (0 cast time)
- **Duration**: 10 seconds
- **Consumption**: Buff removed when Pyroblast casts

Buff data structure to implement:
```python
{
    "hot_streak": {
        "applied_at": 45.3,
        "expires_at": 55.3,
        "effect_type": "instant_cast",
        "consumed_by": "pyroblast"
    }
}
```

Test scenario:
```
Time 10.0: Fireball cast starts
Time 12.5: Fireball completes → crits
Time 12.5: Fireball cast starts  
Time 15.0: Fireball completes → crits → Hot Streak applied
Expected: buff exists, expires at 25.0

Time 15.1: Pyroblast cast starts
Expected: Pyroblast instant (0 cast time), Hot Streak removed
```
```

#### Step 3: Paste into Claude

Copy the filled `session_prep.md` and paste into Claude to start session.

**Token count: ~1,000-1,300**
- Project identity: 200
- Current state: 300
- Today's goal: 200
- Existing systems: 300  
- Reference data: 200

### DURING the Session

**Work normally.** Claude has all the context it needs.

**If you need to reference other systems mid-session:**

```markdown
Wait, I need to show you the spell_executor code we built.

From docs/sessions/session_02_integration.md:

```python
class SpellExecutor:
    def execute_spell(self, spell: Spell, character: Character) -> SpellResult:
        """
        Executes spell, returns result with damage and outcome
        """
        # Damage calculation
        # Crit roll
        # Resource cost
        return SpellResult(damage, outcome, cast_time)
```

Please modify your buff implementation to integrate with this.
The buff should affect cast_time when Hot Streak is active.
```

### END of Each Session

#### Step 1: Ask Claude to generate wrap-up

```markdown
Session wrap-up time. Please fill out this template:

## What Was Built
[List files created/modified with one-line descriptions]

## Integration Brief  
[Generate brief with classes, methods, signatures, usage examples]

## Test Status
[What's tested, what needs testing]

## Known Issues / TODOs
[Anything incomplete or needs attention]

## Next Session Should
[Suggest logical next step]
```

#### Step 2: Save Claude's output

Save the integration brief to `docs/sessions/session_04_integration.md`

#### Step 3: Update PROJECT_STATE.md (manually, 2 minutes)

```markdown
# Project State

Last Updated: Nov 24, 2025

## Completed ✅
- Session 1: Stats System (character.py, stat_calculator.py)
  - Test coverage: 85%
  
- Session 2: Spell System (spell.py, spell_executor.py)
  - 12 spells implemented
  - Test coverage: 78%
  
- Session 3: Rotation Engine (rotation_engine.py)
  - Priority-based action selection
  - Test coverage: 82%
  
- Session 4: Buff Tracking (simulation.py updates)
  - Buff application/expiration/consumption
  - Test coverage: 88%

## In Progress 🚧
- Session 5: Proc System (starting next)

## Planned ⏳
6. Full simulation loop (combat iterations)
7. Results analysis (DPS calculation, stat weights)
8. Multi-target damage (future)

## Known Issues
- None currently

## Recent Decisions
- Buffs stored in dict by name (not list) - allows O(1) lookup
- Buff expiration creates events (not polled) - more efficient
- Hot Streak consumed immediately on Pyroblast cast - matches game behavior
```

---

## Special Cases

### After a Break (Returning After a Week+)

If you haven't worked on the project for a while:

#### START of session:

```markdown
# Resuming: Session X

## Project Refresh

We're building a WoW Combat Simulator for [Expansion] [Spec].
Goal: Match SimC accuracy within 5%.
Tech: Python + NumPy, data-driven (JSON for spells/talents).

## What's Been Built (Summary)

[Paste from PROJECT_STATE.md - Completed section only]

✅ Stats System - character.py, stat_calculator.py
✅ Spell System - spell.py, spell_executor.py  
✅ Rotation Engine - rotation_engine.py
✅ Event System - simulation.py with buff tracking

## Reorientation Questions

Before we continue, can you remind me:

1. What did we build in the last session?
2. What were we planning to build next?
3. Any issues or TODOs pending?

[Wait for Claude's response - it will reconstruct from what you've told it]

---

[After Claude responds]

## Today's Goal

[Specify what you want to do based on Claude's reminder]

Implement proc detection system...
[Continue as normal]
```

### Debugging Session

When something's broken and you need to fix it:

```markdown
# Debug Session - Session X

## Context

We're working on buff tracking in simulation.py.
Hot Streak should proc on 2 crits but isn't triggering.

## What I Expected

Fire two Fireballs that both crit → Hot Streak buff applied to character

## What Actually Happens

Two Fireballs crit (confirmed in logs) but buff_tracker dict stays empty.
No buff_expire event created.

## Relevant Code (where issue is)

```python
def check_hot_streak(self):
    recent = self.recent_spell_results[-2:]
    if len(recent) == 2 and all(r["result"] == "crit" for r in recent):
        # This should apply buff but doesn't
        buff = Buff("hot_streak", duration=10.0)
        self.buff_tracker["hot_streak"] = buff
        print(f"Applied Hot Streak at {self.current_time}")  # This DOES print
```

## Debug Logs

```
Time 12.5: Fireball complete - CRIT
Time 15.0: Fireball complete - CRIT
Applied Hot Streak at 15.0
[Later check]
buff_tracker contents: {}  <-- EMPTY!
```

## Question

Why is the buff disappearing from buff_tracker immediately after being added?
```

For debugging, you DO paste the specific problematic code.

### Multi-File Operation

When working across multiple systems simultaneously:

```markdown
# Session X - Cross-System Integration

## Context

Integrating proc system with buff system and spell executor.

## Systems Involved

### 1. Buff System (session_04_integration.md)
[Paste relevant interface]

### 2. Spell Executor (session_02_integration.md)
[Paste relevant interface]

### 3. Proc System (building today)
[Your requirements]

## Integration Requirements

When spell completes:
1. SpellExecutor tracks result → Simulation.track_spell_result()
2. Simulation checks proc conditions → ProcManager.check_procs()
3. If proc triggers → BuffManager.apply_buff()
4. Next spell cast → SpellExecutor checks BuffManager.has_buff()

## Today's Goal

Build ProcManager that coordinates these systems...
```

---

## Document Update Frequency

| Document | When to Update | Who Updates | Time |
|----------|---------------|-------------|------|
| PROJECT_VISION.md | Rarely (major scope changes) | You manually | 10 min |
| TECH_STACK.md | Rarely (tech stack changes) | You manually | 5 min |
| PROJECT_STATE.md | Every session end | You manually | 2 min |
| session_XX_integration.md | Every session end | Claude generates | Auto |
| SESSION_START.md | Every session start | You fill template | 5 min |
| SESSION_END.md | Every session end | Claude fills | Auto |

**Total overhead per session: ~7 minutes**

---

## Token Budget Summary

### First Session
- **Context:** ~1,200-1,500 tokens
  - Full vision: 500-800
  - Full tech stack: 300-500
  - Today's goal: 400-500
- **Reason:** Setting foundation, establishing patterns

### Regular Sessions  
- **Context:** ~1,000-1,300 tokens
  - Project identity: 200
  - Current state: 300
  - Today's goal: 200
  - Relevant systems: 300
  - Reference data: 200
- **Reason:** Efficient, focused, no repetition

### After Break
- **Context:** ~800-1,000 tokens (reorientation)
- **Then:** ~300-500 tokens (today's goal after reorienting)
- **Reason:** Refresh memory first, then work

**Remaining tokens for actual work:** 198,000+ out of 200,000

---

## Complete Workflow Checklist

### ✅ One-Time Setup (30 min)
- [ ] Create `docs/PROJECT_VISION.md`
- [ ] Create `docs/TECH_STACK.md`
- [ ] Create `docs/PROJECT_STATE.md`
- [ ] Create `docs/templates/SESSION_START.md`
- [ ] Create `docs/templates/SESSION_END.md`
- [ ] Create `docs/sessions/` directory

### ✅ Every Session Start (5 min)
- [ ] Copy `SESSION_START.md` template
- [ ] Fill in: Today's goal and requirements
- [ ] Fill in: Relevant systems (paste from previous session briefs)
- [ ] Fill in: Reference data needed (formulas, examples)
- [ ] Paste completed template into Claude

### ✅ Every Session End (5 min)
- [ ] Ask Claude to generate integration brief
- [ ] Save to `docs/sessions/session_XX_integration.md`
- [ ] Update `PROJECT_STATE.md`:
  - [ ] Move completed work from "In Progress" to "Completed"
  - [ ] Add new work to "In Progress"
  - [ ] Note any important decisions
- [ ] Note Claude's suggested next step for tomorrow

### ✅ Periodic Maintenance (as needed)
- [ ] Update PROJECT_VISION.md if scope changes
- [ ] Update TECH_STACK.md if tech choices change
- [ ] Archive old session briefs (if project grows very large)

---

## Minimal Workflow Alternative

**Can't maintain all those files?** Here's the bare minimum that still works:

### Two-File System

#### File 1: `docs/PROJECT.md` (update every session)

```markdown
# WoW Combat Simulator

**Vision:** Simulate Cataclysm Fire Mage DPS to match SimC within 5%
**Tech:** Python + NumPy, data-driven design (JSON for spells)

## What's Built

✅ **Stats System** (character.py, stat_calculator.py)
   - Rating conversions, Spell Power calculation

✅ **Spell System** (spell.py, spell_executor.py)
   - 12 spells, damage calc, crit rolls

✅ **Rotation Engine** (rotation_engine.py)
   - Priority-based action selection

✅ **Event System** (simulation.py)
   - Event queue, buff tracking

## Currently Building

🚧 Proc detection system

## Next

Build full simulation loop (10k iterations)
```

#### File 2: `docs/LAST_SESSION.md` (overwrite every session)

```markdown
# Last Session: Built Event System Buff Tracking

## What to Use

```python
# From simulation.py
class BuffManager:
    def apply_buff(self, name: str, duration: float)
    def has_buff(self, name: str) -> bool
    def remove_buff(self, name: str)

class EventQueue:
    def add_event(self, event: Event, time: float)
    def get_next_event() -> Event
```

## Usage Example

```python
buff_manager.apply_buff("hot_streak", 10.0)
if buff_manager.has_buff("hot_streak"):
    cast_time = 0  # Instant cast
```

## Next Session Should

Build proc detection system that uses BuffManager
```

#### Workflow

**Session start:** Paste both PROJECT.md + LAST_SESSION.md (total: ~600-800 tokens)

**Session end:** 
```markdown
Update LAST_SESSION.md with:
1. What we built
2. Key interfaces to use
3. What next session should do
```

**Token cost: ~600-800 per session**

**Trade-off:** Less organized, but minimal overhead.

---

## Advanced Techniques

### Multi-Phase Projects

If your project has distinct phases (e.g., backend, then frontend, then optimization):

#### Add Phase Documents

```
docs/
├── PROJECT_VISION.md
├── TECH_STACK.md  
├── PROJECT_STATE.md
├── phases/
│   ├── phase_1_backend.md
│   ├── phase_2_frontend.md
│   └── phase_3_optimization.md
└── sessions/
    └── [integration briefs]
```

#### Phase Document Example: `phases/phase_1_backend.md`

```markdown
# Phase 1: Backend Simulation Engine

## Goal
Build complete simulation engine that can run 10k combat iterations.

## Scope
- Stats system
- Spell system  
- Rotation engine
- Event simulation
- Proc detection
- Basic results (DPS only)

## Out of Scope (for Phase 1)
- Web UI
- Visualization
- Stat weight calculations
- Advanced analysis

## Success Criteria
- Simulate 10k iterations in < 5 seconds
- Single-target Patchwerk results within 5% of SimC
- All core mechanics working (procs, buffs, cooldowns)

## Estimated Sessions: 8-10
```

**At session start during Phase 1:**
- Include `phases/phase_1_backend.md` in context

**At session start during Phase 2:**
- Include `phases/phase_2_frontend.md` in context
- Include summary of Phase 1 (not full details)

### Large Codebase Strategy

Once you have 20+ files, add a codebase map:

#### File: `docs/CODEBASE_MAP.md`

```markdown
# Codebase Structure

## Core Systems

### Character Management
- **character.py**: Character class, stat storage
- **stat_calculator.py**: Rating conversions, derived stats

### Combat Mechanics
- **spell.py**: Spell class, reads from JSON
- **spell_executor.py**: Execute spells, calculate damage
- **buff_manager.py**: Buff application, tracking, expiration
- **proc_manager.py**: Proc detection, buff triggers

### Simulation Engine
- **event_system.py**: EventQueue, Event classes
- **simulation.py**: Main simulation loop, orchestrates all systems

### Data
- **data/spells/**: JSON files with spell definitions
- **data/talents/**: JSON files with talent effects

## Dependency Graph

```
Character ──→ SpellExecutor ──→ Simulation
                    ↓               ↓
              DamageCalc      EventQueue
                                    ↓
                            ┌───────┴────────┐
                            ↓                ↓
                      BuffManager      ProcManager
                            ↑                ↓
                            └────────────────┘
```

## Quick Reference

When implementing new features, check:
1. Does it need character stats? → Use Character class
2. Does it affect spells? → Modify SpellExecutor
3. Does it create events? → Use EventQueue
4. Does it track buffs? → Use BuffManager
5. Does it detect procs? → Use ProcManager
```

**Include relevant sections in session context as needed.**

### Context Tiers System

Maintain three versions of each system's documentation:

#### Tier 1: Quick Reference (50-100 tokens)
```markdown
## Buff System (buff_manager.py)
- apply_buff(name, duration) - Apply buff
- has_buff(name) - Check if active  
- update(time) - Expire old buffs (call every tick!)
```

**Use when:** Just listing what exists in overview

#### Tier 2: Interface Signatures (200-400 tokens)
```markdown
## Buff System (buff_manager.py)

```python
def apply_buff(self, name: str, duration: float) -> None
def has_buff(self, name: str) -> bool
def get_buff_value(self, name: str) -> float
def update(self, time: float) -> List[str]
```
```

**Use when:** Implementing something that calls these methods

#### Tier 3: Full Integration Brief (500-800 tokens)
```markdown
## Buff System (Complete Reference)

[Full integration brief with usage examples, notes, etc.]
```

**Use when:** 
- First time integrating with a system
- Complex interactions
- Haven't worked on this area in a while

**Strategy:** Load the tier appropriate for your needs to minimize token usage.

### Progressive Context Loading

For very complex sessions, build context progressively:

**Initial message:**
```markdown
I'm implementing proc detection. Let me start by asking:
What systems will I need to integrate with?
```

**Claude responds:** "You'll need BuffManager and EventQueue"

**Your follow-up:**
```markdown
Here's the BuffManager interface:
[Paste integration brief]

Here's the EventQueue interface:
[Paste integration brief]

Now implement ProcManager that coordinates these...
```

**Benefit:** Only load what you actually need, based on Claude's guidance.

---

## Real-World Example: Complete Session Flow

### Session 5 - Complete Walkthrough

#### Morning: Preparation (5 minutes)

1. Open `docs/templates/SESSION_START.md`
2. Copy to `session_05_prep.md`
3. Fill in:

```markdown
# Session 5 - November 25, 2025

## Project Identity
WoW Combat Simulator - Cataclysm Fire Mage
Goal: Match SimC within 5%
Tech: Python + NumPy, data-driven

## Current State
✅ Stats, Spells, Rotation, Events, Buffs all working
🚧 Starting: Proc detection system

## Today's Goal
Implement ProcManager that detects Hot Streak procs.

Requirements:
- Check last 2 spell results for double-crit
- Trigger Hot Streak buff via BuffManager
- Handle proc cooldowns (can't proc again while active)

Success Criteria:
- Two crits → Hot Streak buff appears
- Buff expires or gets consumed
- Cannot proc while Hot Streak already active
- Unit tests pass

## Relevant Systems

[Paste from docs/sessions/session_04_integration.md:]

### BuffManager Interface
```python
def apply_buff(self, name: str, duration: float, effect: str)
def has_buff(self, name: str) -> bool
```

[Paste from docs/sessions/session_03_integration.md:]

### Simulation Event Tracking
```python
def get_recent_spell_results(self, count: int) -> List[dict]
# Returns: [{"spell": "fireball", "result": "crit"}, ...]
```

## Reference Data

Hot Streak proc rules:
- Trigger: Last 2 spells both crit
- Cannot proc: If Hot Streak already active
- Duration: 10 seconds
- Consumed by: Casting Pyroblast
```

4. Save and paste into Claude

#### During Session: Work

[Normal development conversation with Claude]

#### End of Day: Wrap-up (5 minutes)

**You:** 
```markdown
Session wrap-up time. Please provide:

1. What we built today
2. Integration brief with interfaces
3. Test status
4. What next session should do
```

**Claude provides integration brief**

**You:**
1. Save Claude's response to `docs/sessions/session_05_integration.md`
2. Update `docs/PROJECT_STATE.md`:

```markdown
## Completed ✅
...
- Session 5: Proc Detection (proc_manager.py)
  - Hot Streak proc working
  - Test coverage: 90%

## In Progress 🚧
- Session 6: Full simulation loop (next)

## Recent Decisions
- Procs checked after every spell result (not on a timer)
- Proc cooldowns stored per-proc (some have cooldowns, others don't)
```

---

## Troubleshooting

### Problem: Claude Forgets What Was Built

**Symptom:** Claude suggests re-implementing something that exists

**Solution:** You forgot to include the integration brief. Add:

```markdown
We already have a BuffManager system. Here's the interface:
[Paste from session_XX_integration.md]

Please use this existing system instead of creating a new one.
```

### Problem: Claude Makes Incompatible Changes

**Symptom:** New code doesn't match existing patterns/structures

**Solution:** Show the existing code structure:

```markdown
All our managers follow this pattern:

```python
class SomeManager:
    def __init__(self):
        self.data = {}
    
    def add_something(self, ...):
        # Add to self.data
    
    def get_something(self, ...):
        # Retrieve from self.data
```

Please make ProcManager follow the same pattern.
```

### Problem: Context Getting Too Large

**Symptom:** Approaching token limits, responses slow

**Solution:** Use more selective context:

**Bad:** Pasting 3 full integration briefs (2,400 tokens)
**Good:** Pasting just the method signatures you need (600 tokens)

```markdown
Today I only need to call two methods:

```python
buff_manager.apply_buff(name, duration)
simulation.get_recent_results(count)
```

That's all I need from existing systems.
```

### Problem: Lost Track of Project State

**Symptom:** You don't remember what's built or what's next

**Solution:** Ask Claude to reconstruct from your brief context:

```markdown
Quick reorientation. Based on what I've told you about this project:

1. What systems are complete?
2. What seems to be the next logical step?
3. Any obvious gaps I should address?

Then I'll specify what to work on.
```

---

## Key Principles Recap

1. **Incremental Context**: Load only what's needed for today
2. **Structured Documentation**: Consistent format aids LLM understanding
3. **LLM-Generated Briefs**: Claude documents what it built accurately
4. **Manual State Tracking**: You maintain PROJECT_STATE.md (2 min/session)
5. **Token Efficiency**: 1,000-1,300 tokens of context, 198,000+ for work

---

## Final Workflow Summary

### One-Time (30 min)
- Create PROJECT_VISION.md, TECH_STACK.md, PROJECT_STATE.md
- Create templates directory

### Every Session Start (5 min)
- Fill SESSION_START template
- Include: identity, state, today's goal, relevant systems
- Paste into Claude

### Every Session End (5 min)  
- Ask Claude for integration brief
- Save to sessions/session_XX_integration.md
- Update PROJECT_STATE.md

### Result
- Clear context for Claude
- Self-documenting codebase
- ~7 minutes overhead per session
- Maintains <1,500 tokens per session start

---

## Conclusion

This workflow enables you to:

- ✅ Work on complex projects across multiple sessions
- ✅ Provide Claude with exactly what it needs (no more, no less)
- ✅ Build a knowledge base as you go
- ✅ Return to projects after breaks without losing context
- ✅ Keep token usage efficient (~1,300 per session)
- ✅ Maintain consistency across sessions

**Total overhead: ~7 minutes per session for massive productivity gains.**

The key insight: **Structure your documentation for LLM consumption, not human reading.** Claude needs interfaces, examples, and current state - not narratives or explanations.

Happy building!
