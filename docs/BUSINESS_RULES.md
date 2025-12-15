# Business Rules (Spells & Mechanics)

Source of truth for the mechanical assumptions in the simulator. Covers spell data, talents, runes/Mystic Enchants, and core constants currently implemented.

## Scope & Constants
- Character level 60, WotLK talents with custom server tuning.
- Stat conversions: 14 crit rating = 1%, 10 haste rating = 1%.
- Hit: 17% cap vs boss (+3 levels), 4% base miss vs equal level.
- GCD: base 1.5s, minimum 1.0s. Haste applies to casts/GCD; DoT tick haste is gated behind Agent of Chaos.
- PvE Power: temporary fixed 1.25 multiplier in spell damage (pending config-ification).

## Spells
- **Immolate**: 404 direct + 770 DoT over 15s (5 ticks). 1.5s cast. SP coeff: 0.20 direct / 1.00 DoT. DoT snapshots multipliers at cast.
- **Incinerate**: 416–490 base plus 104–123 bonus when Immolate is active. 2.25s cast. SP coeff: 0.714.
- **Chaos Bolt**: 1000–1206 base, 12s cooldown (10s with Glyph of Chaos Bolt). 2.0s cast. SP coeff: 0.821.
- **Soul Fire**: 808–1014 base, 4.0s cast, SP coeff: 1.0.
- **Conflagrate**: Instant, 10s cooldown. Deals 60% of Immolate’s DoT as direct damage and applies a DoT equal to 40% of that hit. SP coeff: 0.60. Triggers Backdraft/pyro procs.
- **Life Tap**: Instant (GCD only). Health cost: `827 + spirit * 1.5`. Mana gain: `827 + spellpower * 0.5`. Improved Life Tap talent not present; glyph may add Spirit → SP buff.
- **Curse of Agony**: 24s DoT ticking every 2s (12 ticks). Base ramps 50% → 100% → 150% in 4-tick blocks; SP coefficient 1.2 splits evenly per tick; snapshots multipliers.
- **Pet (Imp)**: Firebolt casting with talent/rune hooks; shares core hit/crit/damage math.

## Talents
- **Emberstorm**: +15% fire/shadow damage.
- **Improved Immolate**: +30% all Immolate damage.
- **Aftermath**: +6% DoT damage (Immolate DoT).
- **Fire and Brimstone**: +10% damage to Incinerate/Chaos Bolt when Immolate is up; +25% Conflagrate crit chance.
- **Ruin**: Crits deal 200% (vs 150%).
- **Shadow and Flame**: +20% of bonus SP added to damage calculations.
- **Devastation**: 1% crit per point (5 points default).
- **Backlash**: 1% crit per point (1 point default).
- **Pyroclasm**: Conflagrate crit can grant +6% fire/shadow damage for 10s (duration extended by Endless Flames ME).
- **Backdraft**: Conflagrate grants 3 charges for 15s; each charge reduces next Destruction spell cast time and GCD by 30%; charges consumed by any Destruction spell (including instants).
- **Improved Soul Leech**: 30% proc on fire spells; returns 2% max mana instantly and applies a HoT for 15s ticking every 5s for 1% max mana.
- **Demonic Power**: Imp Firebolt cast time reduced by 0.25s per point (2 points).
- **Empowered Imp**: 10% damage per point and 33% proc chance per point for crit buff (8s duration).

## Mystic Enchants / Runes (implemented hooks)
- **Destruction Mastery**: Damage multiplier to core Destruction spells.
- **Cataclysmic Burst**: Interaction with Immolate ticks (extended uptime) and other Destruction spells.
- **Heating Up**: Haste/timer aura helper.
- **Gul'dan's Chosen**: Periodic buff with damage bonus (toggleable via config).
- **Agent of Chaos**: Allows Immolate DoT ticks to haste-scale; Chaos Bolt cooldown reduction with a direct damage penalty.
- **Chaos Manifesting**: Additional buff/debuff handling via aura system.
- **Decisive Decimation**: Buff applied by Conflagrate for Soul Fire.
- **Inner Flame**: Additional damage modifier hook.
- **Endless Flames**: Extends Pyroclasm duration.
- **Glyph of Life Tap**: Spirit → spell power buff after Life Tap.
- **Glyph of Conflagrate**: Immolate is not consumed on Conflagrate.
- **Glyph of Chaos Bolt**: -2s Chaos Bolt cooldown.
- **Glyph of Incinerate**: +5% Incinerate damage.
- **Glyph of Immolate**: Immolate enhancement (handled via spell modifiers).
- **Demonic Aegis**: +0.09 Spirit → spell power conversion.
- **Suppression**: +3% spell hit.
- **Improved Imp**: Pet bonus hooks.
- **Curse of the Elements**: 10% damage taken multiplier on target.
- **Pure Shadow**: Shadow spell casts grant 15s stacking buff (up to 6) giving +1% Shadow Bolt damage and +10% Shadowfury damage per stack (Shadow Crash reserved for later).
- **Dusk till Dawn**: Casting Shadow Bolt/Incinerate/Soul Fire/Chaos Bolt grants a 15s stack (up to 3); next Shadowburn gains +10% damage per stack and at 3 stacks also applies Corruption automatically.
- **Pyroclasmic Shadows**: While Pyroclasm is active, Shadow Bolt gains +10% crit chance.
- **Unstable Void**: Shadowfury triggers Backdraft (Shadow Crash to be added later); respects existing Backdraft/Gul'dan’s Chosen rules.
- **Nightfall**: Corruption ticks start at 2% to grant Shadow Trance; each failed tick adds +2% until it procs. Stacks drop when Corruption ends. Shadow Trance lasts 10s and makes the next Shadow Bolt instant.
- **Twilight Reaper**: When Shadow Trance procs (from Nightfall talent or ME), the Shadow Bolt it empowers is free and leeches 50% of its damage as healing.
- **Cursed Shadows**: Curse of Agony ticks have 30% chance to grant a 12s buff making the next Shadow Bolt cost 20% less mana and deal 20% more damage (consumed on cast).

## Planned Mystic Enchants (non-pet focus)
- **Shadow Siphon** (Epic): Shadowburn deals +25% damage to targets below 35% HP. Hooks: target health predicate in damage calc. *Requires Shadowburn implementation.*
- **Unstable Void – Shadow Crash**: Add Shadow Crash hook later to also trigger Backdraft.

## Reference baselines (from wotlk sim; re-verify base numbers on our server)
- **Curse of Agony**: 24s DoT, 12 ticks (14 with Glyph), 2s tick interval. Base tick dmg 145 ramping (start 0.5*base + 0.1*SP; every 4th tick add +0.5*base). Multipliers: Shadow Mastery +3%/pt, Contagion +1%/pt, Improved CoA +5%/pt. Mana 10% base reduced by Suppression. Cancels CoD on apply.
- **Corruption**: 6 ticks every 3s (hasteable with Glyph of Quick Decay). Base dmg 1080/6 + spellCoeff*SP with spellCoeff = 0.2 + Empowered Corruption contribution + 0.01*Everlasting Affliction. Crits if Pandemic; snapshot crit/multiplier on apply (not on rollover). Multipliers: Shadow Mastery, Contagion, Improved Corruption, Siphon Life, Grand Spellstone, set bonuses.
- **Shadowburn**: 15s CD, 20% base mana (reduced by Cataclysm). Base hit 775–865 + SP*0.429*(1+4% per Shadow and Flame). Damage mult includes Grand Firestone, Shadow Mastery; crit mult uses Ruin. Glyph: +20 crit rating during execute (35% HP) via reset callback. Backdraft does not shorten Shadowburn GCD per existing implementation comment.
- **Shadow Bolt**: 3.0s base cast minus 0.1s per Bane point; mana 17% base reduced by Cataclysm (4/7/10%) and Glyph of Shadow Bolt (-10%); base hit 694–775 + SP*0.857*(1+4% per Shadow and Flame). Damage mult includes Firestone bonus, Shadow Mastery (+3%/pt), Improved Shadow Bolt (+2%/pt), Malefic 4p (6%); crit bonuses from Devastation (+5% crit rating), Deathbringer/Dark Coven 4p, Ruin crit multiplier. Improved Shadow Bolt talent procs Shadow Mastery aura (20% per stack) on hit.

## Rotation / APL Rules of Thumb
- Maintain Immolate; gate Incinerate/Chaos Bolt damage bonuses on active Immolate.
- Prioritize Conflagrate on cooldown (Backdraft + Pyroclasm source).
- Life Tap when needed (or to maintain glyph buff).
- YAML APL lives in `configs/rotations/`; validate with `go run ./cmd/aplvalidate -rotation <file>`.
