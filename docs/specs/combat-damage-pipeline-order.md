# Damage pipeline order

**Description**

An attack passes through checks and damage calculations before it changes the target's life. Life stealing, cold, poison, item effects and reactions each occur at a particular point. Their order matters because later steps use the damage and random choices left by earlier steps.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: DPO-R, DPO-Q, DPO-V, DPO-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-damage-pipeline-order.md)
Open evidence gaps: [gap file](../plans/gaps/combat-damage-pipeline-order.md)

## 1. Scope


This spec defines the ordered stages that turn one connected (or attempted) melee attack into final resource changes on the defender, and the RNG draw order of that path: the hit-result hand-off (which result bits gate which later stages), the damage-record build (physical roll, crit/deadly doubling, elemental fields, leech inputs, status-length terms, conversion, the monster damage double), the difficulty/PvP scale placement, total and predicted-lethal flag, the apply step (gates, events, clamp, leech, absorb heal, life subtraction, mana/stamina burn, status appliers, death flag), the post-apply steps of the queued path (durability draw counts, events 7/3, reflection, reaction) and every dispatch point where other code may draw. The per-type mitigation arithmetic (flat reduction, resistance, pierce, absorb, bypass) belongs to [combat-damage-mitigation](combat-damage-mitigation.md); the to-hit roll and block/avoid to [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md); leech arithmetic to [combat-leech-and-status-effects](combat-leech-and-status-effects.md); the status appliers to [combat-status-application](combat-status-application.md); scale percentages and hostility to [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md); durability-loss arithmetic to [item-durability-loss](item-durability-loss.md); item-event callbacks to [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md); the reaction to [combat-hit-recovery](combat-hit-recovery.md). Out of scope: Crushing Blow, Open Wounds, thorns attribution, missile-unit damage rolls, the skill handlers reached through states (DPO-R33) and monster modifiers (DPO-R30), which are only placed in the order here.

## 2. Trigger and ownership

D2Game on the server owns the path. A player attack frame event runs skill do dispatch; `Skills.srvdofunc = 1` selects the basic-attack operation (SCD-R06, SCD-R09a, SCD-R14a). The path target is the defender. Missing target or self-target returns failure before this path. An early alternate branch is excluded from this supplied-path contract (DPO-Q19); monster frame-event roots and upstream hit-callback callers remain DPO-Q20. The companion records every original source link.

The basic attack zeroes a damage record, obtains hit result, sets already-built, runs the before-build state pass, builds with weapon mode 0 and ratio 128, runs the after-build state pass, then enters queued build completion. Build runs on hit and miss. An eligible hit runs scale/mitigation and predicted lethal; either hit or miss then enqueues. An intermediate route or the direct route reaches queued apply. Optional-route eligibility and side effects remain DPO-Q19.

Queued apply finds and copies the matching entry, tests reach, applies on hit, adds marks and durability on the pre-apply hit test, dispatches events 7/3, reflects on a surviving hit, reacts on hit or miss, then removes matching entries. These operations occur synchronously within the attack's server frame. Deferred operations include regeneration at frame+1, freeze event type 2 and timed-state expiration.

The separate hit callback returns before build on miss. Other queued callers build only when eligible and not already built. Missile contact supplies its own record and requests scale/mitigation inside apply. DPO-R08c/d distinguish the direct caller profiles and their number of scale runs. Client resource/reaction messages are outside this boundary; no message identity is established.

## 3. Inputs and data

Notation consumes NRF-R01/R02 for seeds and draws, NRF-R05 for `mulDiv`, NRF-R06a for signed truncation/wrapping, NRF-R08a for raw 8.8, NRF-R12 for stat/state reads and the stat-write dependency for writes. `Stat(unit,id)` is a layer-0 read; `SetStat` is a layer-0 write; `HasState` is a state test. `draw` and `drawAlt` give identical results; `drawInRange(m)` uses lower bound 0. `stepMod(unit,m)` steps once then takes unsigned low-word remainder by m. Ordinary arithmetic is signed 32-bit with wrapping products/sums and truncation toward zero unless an unsigned test is stated.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| units/game | type, class, identity, mode, path target, room, inventory, seed, difficulty, expansion flag, frame | integers, optional references, seed pair | gates, ownership and scheduling |
| damage record | physical, fire, lightning, magic, cold, lifeDrain, manaDrain, staminaDrain, absorbHeal, total | i32 raw 8.8 | mutable damage/resource terms |
| poison/burn | poisonMagnitude, poisonLength, burnMagnitude, burnLength | i32 raw stat units/frames | poison endpoints are not shifted by 8 |
| lengths | coldLength, freezeLength, stunLength | i32 frames | status arguments |
| controls | flags, result, physicalPercent, flatReductionScale, hitClass, markCondition, conversionType, conversionPercent, overlayId | flags u32, result u16, markCondition/conversionType u8; others i32 | rule controls and supplied record terms |
| supplied producer results | primary weapon, two-handed selector, weapon factors/mastery/crit, item type/qualification/throwable, hostility/town/reach, to-hit/avoidance, freeze eligibility and cold effect | references, integers and booleans | supplied-input vectors do not verify unread producers |
| unit conditions | invalid/dead condition, last attacker type/id/presence, special-stun condition, ordered queue entries, event registrations, pending events, modifier identities | semantic values and ordered lists | no original memory layout is required |

| Input | Source | Format | Meaning |
|---|---|---|---|
| attacker stats 21 / 22 (`mindamage` / `maxdamage`), 23 / 24 (`secondary_mindamage` / `secondary_maxdamage`) | stat id | i32, shifted `<< 8` by DPO-R02b | weapon damage pair; secondary pair when two-handed selector returns 2 |
| attacker stat 111 (`item_normaldamage`) | stat id | i32 `<< 8` | added to both ends |
| attacker stats 25* (`damagepercent`), 0 (`strength`), 2 (`dexterity`), 18* (`item_mindamage_percent`), 17* (`item_maxdamage_percent`) | stat id | i32 percent | physical percent terms (DPO-R02b) |
| attacker stat 0x145* (`progressive_tohit`) | stat id | i32 | passed to the hit result and the to-hit roll (DPO-R01b) |
| attacker stats 0x30 / 0x31 (`firemindam` / `firemaxdam`), 0x32 / 0x33 (`lightmindam` / `lightmaxdam`), 0x36 / 0x37 (`coldmindam` / `coldmaxdam`), 0x34 / 0x35 (`magicmindam` / `magicmaxdam`) | stat id | i32 `<< 8` | elemental pairs (DPO-R04) |
| attacker stats 0x149 (`passive_fire_mastery`), 0x14A (`passive_ltng_mastery`), 0x14B (`passive_cold_mastery`), 0x165 (`passive_mag_mastery`) | stat id | i32 percent | both percent arguments of the element roll |
| attacker stats 0x151 (`passive_critical_strike`), 0x8D (`item_deadlystrike`) | stat id | i32 percent | crit sources b and c (DPO-R03a) |
| attacker stats 0x3C / 0x3D* (`lifedrainmindam` / `lifedrainmaxdam`), 0x3E / 0x3F* (`manadrainmindam` / `manadrainmaxdam`), 0x40 / 0x41* (`stamdrainmindam` / `stamdrainmaxdam`) | stat id | i32 `<< 8` (monster roll) or raw (player add) | drain inputs (DPO-R05) |
| attacker stats 0x67 (`skill_bypass_undead`), 0x68 (`skill_bypass_demons`), 0x6A (`skill_bypass_beasts`) | stat id | i32, non-zero test | flags 0x100 / 0x200 / 0x400 (DPO-R05) |
| attacker stats 0x39* / 0x3A* (`poisonmindam` / `poisonmaxdam`), 0x14C* (`passive_pois_mastery`), 0x65* (`skill_poison_override_length`), 0x3B* (`poisonlength`), 0x146* (`poison_count`) | stat id | i32 raw | poison roll and length (DPO-R06) |
| attacker stats 0x38* (`coldlength`), 0x42* (`stunlength`), 0x13B* (`firelength`) | stat id | i32 | length terms (DPO-R06b, DPO-R21b) |
| attacker stats 0x78* (`item_damagetargetac`), 0x79 (`item_demondamage_percent`), 0x7A* (`item_undeaddamage_percent`), 0xB4* (`damage_vs_montype`, layered list) | stat id | i32 | pre-roll adjustments (DPO-R31) |
| attacker stat 0xAC (`alignment`) on the state 0x69 (`alignment`) list | alignment-state stat read | i32 | `== 2` extends DPO-R31 to non-player attackers |
| defender stats 6 (`hitpoints`), 7 (`maxhp`, through the max-life read), 8 (`mana`), 10 (`stamina`) | stat id | i32 8.8 | resources (DPO-R14, R16, R17, R18) |
| defender stat 0x1F* (`armorclass`) | stat id | i32 | written by DPO-R31 |
| defender stats 34 (`normal_damage_reduction`), 35 (`magic_damage_reduction`) | stat id | i32 `<< 8` | flat-reduction setup inside scale/mitigation operation ([combat-damage-mitigation](combat-damage-mitigation.md)) |
| defender stats 0x99 (`item_cannotbefrozen`), 0x76 (`item_halffreezeduration`) | stat id | i32, non-zero test | status-length strip (DPO-R10) |
| defender stat 0x4A (`hpregen`) | stat id | i32, non-zero test | regen event reschedule (DPO-R19c) |
| defender stat 0x83 (`thorns_percent`) | stat id | i32 | reflection (DPO-R25) |
| item stat 0x48* (`durability`) | stat id | i32 | durability loss (DPO-R24a; arithmetic DUR) |
| stat 0x15E* (`modifierlist_skill`) on state lists | stat id | i32 skill id | state-driven handlers (DPO-R33) |
| stat 0xB2* (`unit_dooverlay`) | stat id | i32 overlay id | written by overlay request (DPO-R22, DPO-R32) |

An asterisk marks a stat whose name/id cell has a narrow table record in the companion; the other names consume GDF-R11.


States consume GDF-R12: state 54 uninterruptable, 131 shrine_resist_fire, 133 shrine_resist_poison, 105 alignment; state 92 is the absorb-heal gate consumed from LSE-R12. Player Run=3, Death=0, Dead=17; monster Death=0, Dead=12 (GDF-R14/R15). The special-stun condition's producer is DPO-Q28.

MonStats inputs are inTown, lUndead, hUndead, demon, killable, Crit (u8), ColdEffect and the reach/monster-type selectors whose column identities remain DPO-Q23/Q26. DifficultyLevels.HireableBossDamagePercent is consumed by scale selection; its eligible class predicate remains DPO-Q01. The configured state list and Skills handler selector are DPO-Q25. Weapon strength/dexterity factors are 16-bit values that the physical roll sign-extends (DPO-R02b); they return 100 for a missing item record or stored value 1, and their column names remain DPO-Q27.

Original numeric flags are retained as logical masks for exact vectors: 1 physical supplied, 2 already built, 4/8/16 skip monster life/mana/stamina roll, 32 build marker, 128 suppress attacker event 6, 256/512/1024 bypass undead/demon/beast, 4096 hostility override. Result masks: 1 hit, 2 lethal/dead, 4 interruptible, 16/128/256/32768 avoidance results, 32 suppress specified item events, 8192 crit. Result mask 32's producer/wider meaning remains DPO-Q15. These controls impose no object layout.

| Mapping/constant | Origin | Binding |
|---|---|---|
| masks, conversion cases, ratio normalization, lethal threshold, ordered tables and durability weights | engine code | fixed by original code |
| stat/state names | effective ItemStatCost.Stat / States tables; GDF-R11/R12 and companion narrow records | code fixes numeric consumers; names are data mappings |
| types 45 Weapon/weap, 50 Any Armor/armo, 57 Blunt/blun | effective data/global/excel/itemtypes.txt; zero-based record index, ItemType/Code | code-fixed type consumers; equivalent matching is DPO-Q27 |
| skill selectors, MonStats and DifficultyLevels fields | game data; selector meaning engine code | resolved from data or data selects engine-defined handlers; uncertain fields remain explicit |

Ordered mitigation entries are physical, fire, lightning, cold, magic, coldLength, freezeLength, poisonLength, poisonMagnitude, lifeDrain, manaDrain, staminaDrain. All except the three lengths are scale-enabled; only the last three are monster-only. Durability's ordered (body location, weight) pairs are (1,3),(3,5),(4,4),(5,4),(8,2),(9,2),(10,2), count 7. Their names remain DPO-Q29. These are engine constants, not copied archive rows.

## 4. Rules (execution order)

Shared local operations use i32 inputs, wrapping additions/products and signed truncation. These names are illustrative:

```
rangeRoll(unit,base,mn,mx,minPct,maxPct):
    if mx <= 0: return max(0,base)
    low = mn + mulDiv(minPct,mn,100)
    high = mx + mulDiv(maxPct,mx,100)
    value = base + low
    if high > low: value += drawAlt(unit,high-low)
    return max(0,value)
elementRoll(unit,maxId,minId,masteryId,prior):
    mx = Stat(unit,maxId) << 8
    if mx < 8: return max(0,prior)
    mn = Stat(unit,minId) << 8
    mastery = 0 if masteryId absent else Stat(unit,masteryId)
    return rangeRoll(unit,prior,mn,mx,mastery,mastery)
scale(value,ratio):
    return value if ratio == 128 else truncatingDivide(wrapped32(value*ratio),128)
scaleAlways(value,ratio):
    return truncatingDivide(wrapped32(value*ratio),128)
```

Build ratio is u8. Zero normalizes to 128 at entry, before every use including burn. Only the elemental fields (DPO-R04a) skip the multiplication at ratio 128; the drain, poison, length and burn terms use `scaleAlways`, so at ratio 128 a term with magnitude at least 2^24 wraps.

**DPO-R01** **Hit result and avoidance hand-off.**

- **DPO-R01a** [verified in 1.13c code (static)]. Start result 0. Both units must be player/monster and supplied hostility must succeed, else return 0. A running player defender pre-sets hit. Reach must succeed; unless hit was pre-set, obtain the supplied to-hit result using attacker, defender, progressive-to-hit and zero extra term. With hit set, obtain avoidance value b: map b masks 2/4/16/1 to result masks 256/128/32768/16; nonzero b clears hit. Result mask 16 clears hit again. Surviving hit without defender state 54 sets result 4. Scale/mitigation and prediction require hit and no result mask in 0x8380. Queued apply and state passes require hit. Basic attack builds on hit and miss; the separate hit callback returns before building on miss. Queued completion skips building when already-built. This hand-off draws nothing itself; supplied hit/avoidance operations own their draws.
- **DPO-R01b** [asm-only]. Reach allowance k is 3 for monster attacker, else 1. Nonzero progressive-to-hit s subtracts weapon-style reach+s from k. Supplied reach tests special monster defender classes 258/261 using defender reach+8>distance; otherwise limit=attacker weapon-style reach+k+1, allowing distance<=0, rejecting distance>limit, then requiring collision test with mask 0x804 to return 0. Before queued apply k is plain 3/1 without the s subtraction. Distance/collision and player weapon-style producer integration remain DPO-Q23.

**DPO-R02** **Physical range.**

- **DPO-R02a** [verified in 1.13c code (static)]. Physical-supplied skips roll and crit, preserving physical. Otherwise with supplied endpoints/percentages, the first build draw is rangeRoll on attacker, one draw iff adjusted high>low. Store physical. Ratio 128 keeps it; otherwise use mulDiv(ratio, value,128), in that argument order, consuming NRF-R05. With value>0x100000 its applicable branch is (value/128)*ratio; ratio<=255 excludes the large-first-argument branch.
- **DPO-R02b** [asm-only]. With weapon present, selector 2 uses stats 23/24, otherwise 21/22; without weapon use the unarmed 21/22 pair, first raising a minimum below 1 to 1 and a maximum below 2 to 2. Add stat 111 to endpoints and shift left 8. Minimum<1 becomes 256; maximum<=minimum becomes minimum+256. Percent starts at physicalPercent+stat 25. With weapon add signed strength*weaponStrengthFactor/100, dexterity*weaponDexterityFactor/100 for nonzero factors (each factor sign-extended from 16 bits) and supplied mastery percent; without weapon add strength. Floor percent at -90. Call rangeRoll with prior physical as base, percent+stat 18 for minimum and percent+stat 17 for maximum, then DPO-R02a scaling. Weapon/factor/selector/mastery integration is DPO-Q27.

**DPO-R03** **Exclusive crit/deadly doubling.**

- **DPO-R03a** [verified in 1.13c code (static)]. After physical roll, try weapon chance only with weapon mode 0 and weapon present, then stat 337, then stat 141. Each positive chance costs one attacker drawInRange(100); signed roll<chance succeeds. First success shifts physical left once, sets result 8192 and stops. Nonpositive chances draw nothing. No elemental doubling here. Physical-supplied skips every crit source.
- **DPO-R03b** [asm-only]. Weapon identity and weapon chance producer remain DPO-Q27. The ordered supplied-chance operation above does not verify their production.

**DPO-R04** **Elemental fields.**

- **DPO-R04a** [verified in 1.13c code (static)]. In order fire, lightning, cold, magic: write scale(elementRoll(attacker, maxId, minId, mastery, prior), ratio). Each positive shifted maximum with adjusted high>low takes one attacker draw; otherwise keep max(0, prior), no draw, still ratio-scale the returned value. Prior field is the roll base.
- **DPO-R04b** [verified in 1.13c code (static)]. Maximum/minimum/mastery ids are fire 49/48/329, lightning 51/50/330, cold 55/54/331, magic 53/52/357. Read mastery only with positive maximum and use it for both endpoint percentages.

**DPO-R05** **Drain inputs** [verified in 1.13c code (static)]. Hireling classifier returns 0 for null/non-monster; monster class 271 gives 1,338 gives 2,359 gives 3,560/561 give 4, otherwise 0 (GDF-R17). Regular monster rolls life 61/60, mana 63/62, stamina 65/64 in order, skipping flags 4/8/16: field+=scaleAlways(elementRoll(attacker, maxId, minId, absent, field), ratio). Prior field counts twice. Player/hireling draws nothing: lifeDrain+=stat 60 unless flag 4; manaDrain+=stat 62 unless flag 8; no stamina addition. Only on this player/hireling path nonzero stats 103/104/106 set flags 256/512/1024. Other attacker types add nothing.

**DPO-R06** **Poison and status lengths.**

- **DPO-R06a** [verified in 1.13c code (static)]. After drains, poisonMagnitude=scaleAlways(rangeRoll(attacker, prior, stat 57, stat 58, m, m), ratio), with raw endpoints. m=stat 332 only when stat 58>0, otherwise 0. One attacker draw iff maximum>0 and adjusted high>low.
- **DPO-R06b** [verified in 1.13c code (static)]. Nonzero poisonMagnitude: positive stat 101 adds to poisonLength; otherwise poisonLength=poisonLength+stat 59 and, only in that else branch, when stat 326>1 the whole sum is divided by stat 326 with signed truncation. Poison length is not ratio-scaled. Positive cold adds scaleAlways(stat 56, ratio) to coldLength. Zero incoming stunLength adds scaleAlways(stat 66, ratio); nonzero stunLength stays. Burn follows DPO-R21.

**DPO-R07** **Conversion** [verified in 1.13c code (static)]. After build event 3, before monster double, nonzero conversionType computes converted=mulDiv(conversionPercent, physical,100), with that argument order, then physical=max(0, physical-converted). Type 10 takes attacker draw(5), selecting result+1. Types 1/2/3 add converted to fire/lightning/magic. Type 4 adds to cold and raises coldLength to at least 50. Type 5 adds signed converted/8 to poisonMagnitude and raises poisonLength to at least 50. Types 6..10 add nothing further. Type 11 adds to fire and raises burnLength to at least 50; type 12 adds to cold and raises freezeLength to at least 50. Other nonzero types only subtract physical. Type 0 does nothing.

**DPO-R08** **Scaling and entry profiles.**

- **DPO-R08a** [verified in 1.13c code (static)]. With supplied P!=100, visit section 3's twelve entries in order. Positive scale-enabled values become mulDiv(P, value,100), preserving NRF-R05 argument and branch order. Nonpositive fields and the three lengths stay unchanged. P=100 skips scaling. Numeric P selection remains DPO-R08b.
- **DPO-R08b** [asm-only]. P is the factor PVP-R02a selects (17, 25, the difficulty hireling/boss percent, 200/400 or 100 in that contract's order); this spec consumes that contract and does not restate its predicates. The production of PVP's scale condition remains PVP-R02b (DPO-Q01). Before selection, physical/magic flat reductions are stats 34/35 shifted 8; each positive reduction with positive flatReductionScale becomes mulDiv(flatReductionScale, reduction,1024) (DMI-R02, which owns use of these inputs).
- **DPO-R08c** [verified in 1.13c code (static)]. Eligible queued completion runs the scale/mitigation stage after build, before prediction/enqueue; later queued apply requests no second run. Apply with scale request true runs it after gates, before apply events. The identified direct-then-flagged profile runs twice; direct-then-unflagged once. Two unflagged local-record profiles run zero times. Original caller identities are in the companion; callers supply one of these entry profiles.
- **DPO-R08d** [verified in 1.13c code (static)]. Queued eligible hit runs once, miss zero; either still enqueues. Enqueue accepts only player/monster defenders, records game, attacker/defender identities and damage, and pushes at attacker queue head. Missile/flagged callbacks request one apply-time run and do not also enqueue. Direct-then-flagged runs twice, direct-then-unflagged once, unscaled local profiles A/B zero. Allocation dependency integration remains DPO-Q30.

**DPO-R09** **Pre-strip event 11** [verified in 1.13c code (static)]. After scaling, unconditionally including P=100, dispatch event 11 on defender list with other=attacker and the shared record. Callbacks see scaled, unstripped fields. Placement adds no draw and ignores dispatcher result; IEP-R07 owns traversal.

**DPO-R10** **Length strip** [verified in 1.13c code (static)]. If coldLength>0 or freezeLength>0, nonzero defender stat 153 zeros both; otherwise nonzero stat 118 halves both with signed truncation, including a negative partner. Positive poisonLength becomes 0 with state 133. Positive burnLength becomes 0 with state 131. No draw.

**DPO-R11** **Mitigation dispatch** [verified in 1.13c code (static)]. Start bypass 0. Monster defender: undead+flag 256 or demon+flag 512 sets 1; only neither undead nor demon tests flag 1024. Null/non-monster defender uses flag 1024. Visit the twelve entries in section 3 order. At a monster-only entry, end loop unless attacker is a regular monster. Each visited field consumes DMI's per-entry mitigation with bypass and supplied flat reductions. No local draw; event 11 callbacks own theirs.

**DPO-R12** **Total and predicted lethal.**

- **DPO-R12a** [verified in 1.13c code (static)]. Store wrapping total=physical+fire+lightning+magic+cold+poisonMagnitude, plus lifeDrain only for regular monster attacker. Eligible queued completion independently recomputes the six-field sum plus lifeDrain for any monster-type attacker, including hireling; this prediction neither reads nor writes stored total. Set result 2 if signed (prediction&~255)>(defender life&~255). Equality does not pre-set lethal. Then enqueue, even when hit was ineligible for scaling/prediction.
- **DPO-R12b** [unresolved]. Poison enters immediate total and also supplies the poison status operation. Combined gameplay interpretation remains DPO-Q02; DPO-R12a's supplied-field sum remains available.

**DPO-R13** **Queue consumption and apply gates.**

- **DPO-R13a** [verified in 1.13c code (static)]. Find first entry matching attacker/defender type/id; none returns without effect. Copy record. Supplied reach uses allowance 3 for monster attacker, else 1; failure removes all matching entries and returns without events/apply. Miss skips apply/marks/durability and goes to 7/3. On hit, player/monster attacker in mode 0 returns before apply, retaining queue. Repeat this mode test after 7/3 if hit still set; failure returns before reflection/reaction/removal. Replace copied flags with 32 and apply with scale request false.

  Apply gates in order: (1) failed supplied hostility without flag 4096 clears result 1/2 and returns; absent defender room does the same even with override. (2) In supplied town room only monster attacker with inTown continues; others clear only result 2 and return. (3) Monster defender without killable returns unchanged. (4) Null/invalid defender, non-player/non-monster, mode 0, monster mode 12 or player mode 17 returns unchanged. (5) Store last attacker type/id and presence; null attacker zeros both and clears presence. (6) Requested scale/mitigation runs here. (7) With result 32 clear and hit set, unflagged apply dispatches attacker 5 then defender 1; flagged apply requires flags 32, then attacker 6 unless flags 128, then defender 2. Each dispatch supplies other=opposite unit and shared record. Events precede clamp; ignore dispatcher return. Supplied-predicate ordering/effects are available; missing integration remains DPO-Q23/DPO-Q30.
- **DPO-R13b** [asm-only]. Hostility/town meanings consume PVP-R05a/PVP-R06; unread subpaths remain DPO-Q30. Override bypasses hostility alone.

**DPO-R14** **Clamp** [verified in 1.13c code (static)]. Set physical=min(physical, defender raw life), signed. Keep total unchanged. No draw.

**DPO-R15** **Leech hand-off** [verified in 1.13c code (static)]. Invoke LSE with clamped, mitigated physical. LSE owns the resource writes (LSE-R09/R10/R15), the drain rewrites (LSE-R04/R06/R15) and its draw(2) (LSE-R13). Zero lifeDrain and manaDrain return before rewrite/draw (LSE-R02); the draw occurs only on LSE's base path (not LSE-R15's monster branch), when both rewritten leech values are nonzero and the base is positive, even if both added amounts are 0. Placement adds none.

**DPO-R16** **Absorb heal** [verified in 1.13c code (static)]. Nonpositive absorbHeal, invalid/dead defender(DPO-R13a gate 4), or state 92 returns 0 with no write. Otherwise n=raw life+absorbHeal, cap n at maxhp, store raw life; return absorbHeal unless capped, then maxhp-oldLife. Apply ignores return. No draw.

**DPO-R17** **Life subtract** [verified in 1.13c code (static)]. Positive total subtracts from raw life; signed remainder<256 becomes 0. Store without whole-life rounding. Nonpositive total leaves life. No draw.

**DPO-R18** **Mana then stamina burn** [verified in 1.13c code (static)]. Positive manaDrain subtracts from raw mana, storing 0 when remainder<256; then positive staminaDrain does the same to stamina(negative or<256 becomes 0). Player/hireling current manaDrain is the LSE-R17 leech term left behind: hireling M<<6; player (M<<6)/ManaStealDivisor when divisor nonzero; unshifted M if leech aborted before shift. Monster leech rewrites lifeDrain/manaDrain back down by 64. This stage consumes current terms and adds no draw.

**DPO-R19** **Status calls and final death.**

- **DPO-R19a** [verified in 1.13c code (static)]. After burns call in order stun(attacker, defender, stunLength), cold(attacker, defender, coldLength), freeze(source=attacker, target=defender, length=freezeLength), poison(attacker, defender, poisonMagnitude, poisonLength), burn(attacker, defender, burnMagnitude, burnLength). CSA owns interiors; ordering adds no draw.
- **DPO-R19b** [verified in 1.13c code (static)]. After all appliers, life<=0 sets result 2;life>0 clears it, overwriting prediction. No draw.
- **DPO-R19c** [asm-only]. Surviving monster defender with total!=0: nonzero hpregen requests removing all pending type 3 events and replacing one type 3 at frame+1 with three zero payload values. With or without reschedule, invoke per-hit monster update. Scheduler/removal/update integration remains DPO-Q24.

**DPO-R20** **Post-apply summary** [asm-only]. Queued events 7/3, reflection, reaction and marks are placed by DPO-R23..R26/DPO-R32. This summary adds no callback-interior assertion.

**DPO-R21** **Burn update and mandatory draw.**

- **DPO-R21a** [verified in 1.13c code (static)]. Every build after poison/length terms takes attacker drawAlt(1), returning 0 and advancing seed, including hit-in and miss.
- **DPO-R21b** [verified in 1.13c code (static)]. With that draw, wrapped x=burnMagnitude*ratio+316+drawAlt(1); add signed max(0, x)/128 to burnMagnitude. Default ratio 128 and prior 0 produce 2. Only nonzero incoming burnLength adds scaleAlways(stat 315, ratio). CSA-R10 requires positive magnitude and length.

**DPO-R22** **Monster double** [verified in 1.13c code (static)]. Last in build after conversion: monster with valid class row and nonzero u8 Crit takes attacker stepMod(100). Unsigned byte roll<Crit doubles physical, fire, lightning, magic, cold, poisonMagnitude. With defender present after doubling, if hitClass low byte has none of mask 0xF0, set 0x10;otherwise request overlay 54. No doubling means neither marker nor overlay. Crit 0, missing row or other attacker type draws nothing. Missile paths also run this on owner. Overlay effect remains DPO-Q18; request is fixed.

**DPO-R23** **Item-event placement** [verified in 1.13c code (static)]. Consume IEP-R07 dispatch(game, eventId, owner, other, record), with registered values passed to callbacks; ignore final return. Build 3 runs on defender/other attacker only for player/hireling and result 32 clear, after burn before conversion. Event 11 runs on defender before strip. Apply 5/1 or 6/2 follows DPO-R13a. On lethal with result 32 clear, defender 10 then attacker 9. Queued post-apply attacker 7 then defender 3 run on hit and miss. DPO-Q11 gates callback-specific draws; supplied callbacks may log/mutate the shared record under IEP-R07.

**DPO-R24** **Durability draws.**

- **DPO-R24a** [verified in 1.13c code (static)]. After apply/marks, before 7/3, use hit test made before apply; clearing hit during apply does not skip durability. Player attacker with primary weapon uses supplied type/qualification: qualified armor threshold 10; weapon 4, or 10 if throwable in Expansion. Classic throwable returns before draw. Qualified item takes one owner's stepMod(100); unsigned roll<threshold requests DUR-R05 loss. Player defender with inventory: eligible armor in seven ordered body locations contributes weight W; W>0 takes defender draw(7) for starting table index, then draw(W) for remainder. Walk cyclically, skip ineligible entries, subtract weights until remainder<weight selects item; process selected item using defender seed with the same qualification/draw rule. Thus 0..1 attacker and 0..3 defender local draws. Item producer integration is DPO-Q27; break effects DPO-R24b.
- **DPO-R24b** [unresolved]. Additional break/inventory effects and draws remain DPO-Q12.

**DPO-R25** **Reflection** [asm-only]. After 7/3, hit still set and attacker passing DPO-R13a's surviving mode check invokes reflection. Positive defender thorns percent is adjusted by signed (percent+4)/8 for player/hireling attacker, then a blank record applies back to attacker with scale request true and reacts. Exact reflection integration/attribution remains DPO-Q04. The record supplies physical=mulDiv(adjustedPercent, incoming physical,100), result 0x4021 and hitClass 0x8D; no leech/stun/cold/freeze terms. The result event-suppression mask prevents apply events. Reaction owns its target seed draws.

**DPO-R26** **Reaction placement** [verified in 1.13c code (static)]. After reflection, before queue removal, invoke reaction on defender with attacker and record, on hit and miss unless an earlier queued return occurred. HRC owns veto/interruption/mode draws; placement adds none.

**DPO-R27** **Stun draw** [verified in 1.13c code (static)]. Positive stunLength and monster defender with monster data and supplied special-stun condition takes attacker stepMod(100). Roll<90 cancels stun;otherwise CSA continues. No draw otherwise. Producer meaning is DPO-Q28 and does not gate this supplied-condition result.

**DPO-R28** **Cold/freeze fallback.**

- **DPO-R28a** [verified in 1.13c code (static)]. Positive cold length with supplied E!=0 takes defender stepMod(100);E=0 returns before draw. E is MonStats.ColdEffect for monster,-50 otherwise(CSA-R05/R06). Freeze on player or monster failing CSA-R07 eligibility re-applies its freeze length through cold, adding one defender draw under the same conditions.
- **DPO-R28b** [asm-only]. Monster freeze schedules type 2;state 54 pending-death/reaction behavior and draws remain DPO-Q13.

**DPO-R29** **Draw order** [verified in 1.13c code (static)]. Basic path order: hit operations;before-build state pass on hit;build on hit/miss: physical, crit a/b/c, fire/lightning/cold/magic, monster life/mana/stamina, poison, mandatory drawAlt(1), build 3, conversion draw(5), monster double;after-build state pass on hit;eligible scale/mitigation with event 11;apply 5/1 or 6/2, leech draw(2), modifier hook, stun attacker draw, cold defender draw, freeze-as-cold defender draw, lethal 10/9;durability on pre-apply hit;7/3 on hit/miss;reflection on surviving hit;reaction on hit/miss. Handler/callback/dependency draws occur at these positions and belong to their contracts;DPO-Q11/Q22/Q25 gate complete integration counts. Burn uses the mandatory draw without another range draw.

**DPO-R30** **Modifier hook** [asm-only]. Between leech and absorb, monster attacker with hit set invokes ordered modifier callbacks for on-hit slot 3, with game, attacker, modifier identity and hook argument. Callback/argument meaning and draws remain DPO-Q22.

**DPO-R31** **Monster pre-roll adjustments** [asm-only]. Player attacker or alignment-state value 2 against monster: nonzero stat 120 adds to defender armorclass, floored 0;positive demon stat 121 adds physicalPercent against demon;against undead positive sum of stat 122 plus 50 for present Blunt weapon adds physicalPercent;matching layered damage_vs_montype entries add values. Armor write persists. Selector/matching is DPO-Q26;weapon/type producers DPO-Q27.

**DPO-R32** **Wrappers and marks** [asm-only]. Weapon-state wrapper runs before build and its partner after with attacker/weapon mode;effects remain DPO-Q21. After queued apply, use pre-apply hit test: positive overlayId requests it on defender;if markCondition 0 and low hitClass nibble 0, merge supplied attacker hit-class value. Overlay/producer integration remains DPO-Q18/Q21. No wrapper-local draw identified.

**DPO-R33** **State handlers** [asm-only]. Basic attack with hit set iterates configured states before/after build. Active state with state list and in-range modifierlist_skill selects skill-record handler:1 before build,2/3/4 after. Pass attacker, state list, shared record. Enumeration, selector-column identity and handler effects/draws remain DPO-Q25.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| DPO-R01a | Hit result bits; what the result gates; build not gated on the basic attack | verified in 1.13c code (static) | DPO-Q23: reach integration only | DPO-V03 |
| DPO-R01b | Reach check before the to-hit roll and before the apply | asm-only | DPO-Q23: reach | none |
| DPO-R02a | Physical roll: one draw iff hit-in clear and `mx' > mn'` | verified in 1.13c code (static) | DPO-Q27: endpoint integration only | DPO-V04 |
| DPO-R02b | Physical min/max and percent arithmetic | asm-only | DPO-Q27: producers | none |
| DPO-R03a | Crit/deadly: ordered sources, one draw each, one doubling | verified in 1.13c code (static) | DPO-Q27: weapon chance integration only | DPO-V02, DPO-V04, DPO-O02 |
| DPO-R03b | Weapon crit chance and weapon identity sources | asm-only | DPO-Q27: chance producer | none |
| DPO-R04a | Elemental order, draws, scaling | verified in 1.13c code (static) | none | DPO-V05 |
| DPO-R04b | Elemental min/max/mastery pairing | verified in 1.13c code (static) | none | DPO-V24 |
| DPO-R05 | Leech/drain inputs; monster drain draws; hireling predicate | verified in 1.13c code (static) | none | DPO-V06 |
| DPO-R06a | Poison roll draw | verified in 1.13c code (static) | none | DPO-V07 |
| DPO-R06b | Poison, cold and stun length terms | verified in 1.13c code (static) | none | DPO-V22 |
| DPO-R07 | Conversion | verified in 1.13c code (static) | none | DPO-V08 |
| DPO-R08a | Scale arithmetic and scaled fields | verified in 1.13c code (static) | DPO-Q01: scale selection only | DPO-V09 |
| DPO-R08b | Scale `P` values; flat-reduction setup | asm-only | DPO-Q01: selection | none |
| DPO-R08c | Paths that run the scale/mitigation entry | verified in 1.13c code (static) | DPO-Q19/DPO-Q20: caller integration only | DPO-V27 |
| DPO-R08d | Runs per path; unconditional enqueue | verified in 1.13c code (static) | DPO-Q30: allocator integration | DPO-V27 |
| DPO-R09 | Pre-mitigation event 0xB | verified in 1.13c code (static) | none | DPO-V26 |
| DPO-R10 | Status-length strip | verified in 1.13c code (static) | none | DPO-V21 |
| DPO-R11 | Bypass flag and the 12-type loop's monster-only stop | verified in 1.13c code (static) | none | DPO-V28, DPO-V09; arithmetic DMI-V01 to DMI-V05 ([combat-damage-mitigation](combat-damage-mitigation.md)) |
| DPO-R12a | Total and predicted-lethal flag; enqueue | verified in 1.13c code (static) | DPO-Q02: combined poison meaning; DPO-Q30: allocation | DPO-V01, DPO-V30, DPO-O01 |
| DPO-R12b | Poison in total and in state | unresolved | DPO-Q02: combined behavior | none |
| DPO-R13a | Queue match, reach gate, apply gates, last-attacker record, events | verified in 1.13c code (static) | DPO-Q23/DPO-Q30: predicates integration only | DPO-V10, DPO-V31 |
| DPO-R13b | Meaning of the hostility and town predicates | asm-only | DPO-Q30: predicate meaning | none |
| DPO-R14 | Clamp physical to life | verified in 1.13c code (static) | none | DPO-V11 |
| DPO-R15 | Leech on clamped mitigated damage | verified in 1.13c code (static) | none | DPO-V11 |
| DPO-R16 | Absorb heal capped at max life | verified in 1.13c code (static) | none | DPO-V12 |
| DPO-R17 | Life subtract, below 1.0 kills | verified in 1.13c code (static) | none | DPO-V01, DPO-V12, DPO-O01 |
| DPO-R18 | Mana/stamina burn | verified in 1.13c code (static) | none | DPO-V13, DPO-O03 |
| DPO-R19a | Applier order | verified in 1.13c code (static) | none | DPO-V25 |
| DPO-R19b | Death flag from final life | verified in 1.13c code (static) | none | DPO-V14 |
| DPO-R19c | Monster regen event reschedule | asm-only | DPO-Q24: scheduler/update | none |
| DPO-R20 | Pointers to the post-apply stages | asm-only | DPO-Q04: reflection | none |
| DPO-R21a | Unconditional `drawAlt(1)` in every build | verified in 1.13c code (static) | none | DPO-V15 |
| DPO-R21b | Burn magnitude and length update | verified in 1.13c code (static) | none | DPO-V23 |
| DPO-R22 | Monster damage double | verified in 1.13c code (static) | DPO-Q18: overlay effect only | DPO-V16 |
| DPO-R23 | Dispatch points consuming IEP-R07 | verified in 1.13c code (static) | DPO-Q11: callback draws only | DPO-V26, DPO-V27, DPO-V31; dispatcher IEP-V08 |
| DPO-R24a | Durability step draws and seeds | verified in 1.13c code (static) | DPO-Q27: item producers only | DPO-V17 |
| DPO-R24b | Draws from item breaking | unresolved | DPO-Q12: breaking | none |
| DPO-R25 | Reflection sub-hit | asm-only | DPO-Q04: reflection | none |
| DPO-R26 | Reaction stage last, hit and miss | verified in 1.13c code (static) | none | DPO-V20, DPO-V29 |
| DPO-R27 | Stun draw on the attacker's seed | verified in 1.13c code (static) | none | DPO-V18 |
| DPO-R28a | Cold draw; freeze fallback extra cold draw | verified in 1.13c code (static) | none | DPO-V19 |
| DPO-R28b | Freeze event type 2 | asm-only | DPO-Q13: event | none |
| DPO-R29 | Draw order of the whole path | verified in 1.13c code (static) | DPO-Q11/DPO-Q22/DPO-Q25: callback/handler integration only | DPO-V20 |
| DPO-R30 | Monster-modifier on-hit hook | asm-only | DPO-Q22: callbacks | none |
| DPO-R31 | Pre-roll adjustments against monsters | asm-only | DPO-Q26/DPO-Q27: adjustments | none |
| DPO-R32 | Wrappers on the build and the queued apply | asm-only | DPO-Q18/DPO-Q21: wrappers/marks | none |
| DPO-R33 | State-driven skill handlers around the build | asm-only | DPO-Q25: handlers | none |

## 5. Outputs and side effects

| Write/effect | Owner | Rule |
|---|---|---|
| hit, avoidance, interruptible, crit, predicted/final lethal flags | damage record | DPO-R01a, R03a, R12a, R13a, R19b |
| built/build-marker/bypass flags; queued apply replacement with 32 | damage record | section 2, DPO-R05, R13a |
| physical, physicalPercent, elements, drains, poison/burn and lengths | damage record | DPO-R02..R07, R10, R21, R22, R31 |
| positive scale-enabled values;stored total | damage record | DPO-R08a, R12a |
| queue head insertion and matching-entry removal | attacker queue | DPO-R08d, R12a, R13a |
| last attacker type/id/presence | defender | DPO-R13a |
| build 3, event 11, apply 5/1 or 6/2, lethal 10/9, post 7/3 dispatches | event owners and shared record | DPO-R09, R13a, R23 |
| leech resource increases and drain rewrites | attacker stats/record | DPO-R15;LSE |
| raised/capped life, then lowered life, mana, stamina at layer 0 | defender stats 6/8/10 | DPO-R16..R18 |
| ordered status calls;states 21/11/1/2/115, timed lists, shatter and timer effects belong to CSA | defender/scheduler | DPO-R19a;CSA-R03..R13 |
| replacement pending regeneration event 3 at frame+1 and monster update | defender scheduler | DPO-R19c (gated) |
| durability loss request, at most one item per side | item stat 72 | DPO-R24a;DUR-R05 |
| persistent armorclass adjustment | defender stat 31 | DPO-R31 (gated) |
| hitClass marker/nibble and overlay 54/overlayId requests | record/defender | DPO-R22, R32 (effect gated) |
| modifier/state callbacks, reflection, hit reaction | dependency owners | DPO-R30, R33, R25, R26 |
| ordered attacker/defender seed advances | owning units | DPO-R29 and individual draw rules |

No client message identity or persistence format is established within this boundary.

## 6. Branches and edge cases

- Attacker type: a regular monster (type 1, hireling predicate 0) rolls drain inputs (DPO-R05), adds `lifeDrain` to the total (DPO-R12a), is mitigated for all 12 table entries (DPO-R11) and may double all six damage fields (DPO-R22); a hireling (classes of GDF-R17) is treated like a player by DPO-R05 and DPO-R12a's stored total, but like a monster by the predicted-lethal recomputation; a player attacker adds leech stats unrolled (DPO-R05) and pays the mana-burn quirk (DPO-R18); other types add nothing and dispatch no build event 3.
- Defender type: only types 0 and 1 reach the hit result (DPO-R01a) and the queue (DPO-R08d); a monster defender needs `killable` (DPO-R13a step 3) and is the only one that receives the stun draw (DPO-R27), the regen reschedule (DPO-R19c) and the pre-roll adjustments (DPO-R31); a player defender in `Run` has bit 0 pre-set (DPO-R01a).
- Hit versus miss on the basic attack: the build and all its draws run either way; the scale, predicted-lethal bit, apply, durability, reflection and the state handlers run only on a hit; the enqueue, events 7/3, the reaction and the dequeue run either way (DPO-R01a, DPO-R08d, DPO-R23, DPO-R26). On the hit-callback path a miss ends before the build.
- Hit-in flag (`flags & 1`) skips the physical roll and all crit draws (DPO-R02a, DPO-R03a), not the other build draws (DPO-R21a); flag 0x1000 allows the apply when the hostility check fails (DPO-R13a), except on the queued path, where queued apply operation overwrites the flags with 0x20 first.
- Scale/mitigation placement (DPO-R08c): at build time on the queued path, inside the apply on the flag-1 paths, twice on direct-then-flagged profile, never on unscaled local profile A and the self-apply of unscaled local profile B (DPO-R08d).
- Crit applies to physical only (DPO-R03a); conversion runs on the doubled value (DPO-R07); the monster double applies after conversion to all six damage fields (DPO-R22).
- Zero and negative inputs: a max stat `<= 0` keeps `max(0, previous)` with no draw (DPO-R04a, DPO-R05); `rangeRoll` never returns a negative value; `physical` after conversion is floored at 0 (DPO-R07); the strip halves negatives toward zero (DPO-R10); a total `<= 0` leaves life unchanged (DPO-R17); burn amounts `<= 0` are ignored (DPO-R18); `burnMagnitude` gains at least 2 per build from 0 (DPO-R21b).
- Overflow: the scale and conversion use `mulDiv` (NRF-R05); the sums of DPO-R12a wrap in 32 bits; the crit doubling is a plain shift.
- Missile versus melee: weapon actions and the hit callback use the same build; missile-unit contact does not (it rolls on the missile's seed, missile roll operation; [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md) BSF-R07), then runs DPO-R22 on the owner before the apply with flag 1; durability, events 7/3, reflection and the reaction of the queued path are not run by the missile apply.
- Difficulty: enters only through `P` (DPO-R08b: the hireling-vs-monster case reads `DifficultyLevelsRecord.HireableBossDamagePercent` of the game's difficulty; every other listed `P` is a constant) and the appliers' divisors (CSA-R05, CSA-R08); classic versus expansion only through the throwable durability rule (DPO-R24a).
- Missing data: a monster class at or beyond the MonStats count makes DPO-R22 draw nothing, DPO-R31's class tests false and DPO-R13a step 3 fail; a null stat list makes every `Stat` read 0 (NRF-R12a) and max-life read return 0 (so the absorb cap is 0).
- Neighbors: [combat-knockback](combat-knockback.md) and [combat-hit-recovery](combat-hit-recovery.md) consume `result` and `hitClass` after DPO-R26.

## 7. Open questions

| Id | Status | Question and evidence that would settle it |
|---|---|---|
| **DPO-Q01** | open (integration) | Selection of P from actual units: PVP-R02a owns the selector order and predicates; the remaining unknown is the hit-time scale-condition producer (PVP-R02b, PVP-Q01/PVP-Q02). Supplied P arithmetic remains available. |
| **DPO-Q02** | open (behavior) | Combined immediate poison damage and poison status interpretation; compare the consuming paths and observe one controlled hit without fitting expected values to a capture. |
| **DPO-Q04** | open (behavior) | Reflection classification, rounding and attribution: complete its source chain. |
| **DPO-Q06** | open (integration) | Physical endpoint producer integration remains DPO-Q27; supplied endpoint roll is independently available. |
| **DPO-Q07** | open (integration) | Dispatch placement is verified; callback-specific draws remain DPO-Q11. |
| **DPO-Q11** | open (integration) | Per-registration callbacks, stat/event selection and their draws consume IEP; missing callbacks cannot be assigned guessed counts. |
| **DPO-Q12** | open (integration) | Additional inventory/break effects and draws after durability loss: finish DUR break contract. |
| **DPO-Q13** | open (integration) | Whether freeze event type 2 draws only with state 54 and can run pending death/reaction: read its handler. |
| **DPO-Q15** | open (outside-boundary) | Producer and wider meaning of result mask 32; local conditional use remains fixed. |
| **DPO-Q16** | open (outside-boundary) | Which effective Skills rows select srvdofunc 1; local supplied selector is fixed. |
| **DPO-Q18** | open (integration) | Overlay 54 and supplied overlayId validation/downstream update; requests themselves remain available. |
| **DPO-Q19** | open (outside-boundary) | Basic-attack alternate branch, route selector and optional intermediate operation; contract begins after alternate branch excluded. |
| **DPO-Q20** | open (outside-boundary) | Monster frame-event roots and hit-callback upstream roots; trace callers to dispatch. |
| **DPO-Q21** | open (behavior) | Build weapon-state wrappers and post-apply hit-class producer; resolve remaining imports and callees. |
| **DPO-Q22** | open (integration) | Modifier slot-3 callback identities, hook argument and draws; read exact callbacks. |
| **DPO-Q23** | open (integration) | Reach distance, collision, player weapon-style branch and selector name; finish callees and corroborate allowances. |
| **DPO-Q24** | open (integration) | Pending-event removal, scheduler and per-hit monster update; read owned contracts. |
| **DPO-Q25** | open (integration) | Configured state enumeration, skill selector column and four handlers; resolve list stat producer and read handlers. |
| **DPO-Q26** | open (behavior) | Monster-type selector identity and layered damage_vs_montype matching; resolve list/match helper and loader. |
| **DPO-Q27** | open (integration) | Primary weapon, two-handed selection, mastery/crit, type equivalence/throwable and weapon-factor columns; complete producer paths. |
| **DPO-Q28** | open (outside-boundary) | Producer/meaning of special-stun condition; local supplied condition-to-draw contract remains available. |
| **DPO-Q29** | open (outside-boundary) | BodyLocs names for numeric locations 1,3,4,5,8,9,10; verify narrow named cells. |
| **DPO-Q30** | open (integration) | Queue allocator and hostility/town dependency subpaths; supplied predicates/queue state remain usable. |
| **DPO-Q31** | open (observation) | Gameplay observations remain pending for all behavioral rules; gameplay confirmation is pending. |

The [gap file](../plans/gaps/combat-damage-pipeline-order.md) records attempted evidence and exact blocked boundaries. Observation gaps limit runtime-confirmation claims. Outside-boundary producers do not gate their supplied-input consumers.

## 8. Test vectors

Conventions: raw 8.8 units; `S1 = {lo 0x1, hi 0x29A}` and `S5 = {lo 0x5, hi 0x29A}` are the seeds of NRF-V01 and CSA-V05; one step of `S1` gives `lo 0x6AC6935F, hi 0` (step 1), then `lo 0x2F2ED81B, hi 0x2C890AFD` (step 2), `lo 0x280687C4, hi 0x13ADFE65` (step 3), `lo 0xB6AAB839, hi 0x10B1BFE6` (step 4), `lo 0xBFC793C3, hi 0x4C305F9E` (step 5); one step of `S5` gives `lo 0x15E0D673, hi 0x2`. Draw results quoted below are `lo mod m` or `lo & (m - 1)` of the new `lo` per NRF-R02. Unless stated: Normal difficulty, `flatReductionScale = 0`, no conversion (`conversionType = 0`), no bypass stats, ratio 0x80, flags 0, `P = 100`, every stat not named is 0, the hit-in flag is clear, the attacker is a player (type 0, id 7, mode 7) with a primary weapon present (primary-weapon lookup non-null; for any monster attacker in a vector `w` is null, so crit source a is not tried and DPO-R02b takes its bare-handed branch), not two-handed (two-handed selector returns 0), strength and dexterity 0 (so the weapon words of DPO-R02b add 0), mastery term `weapon mastery/crit producer(u, w, 0, 1) = 0` and crit chance `ca = 0`, and the defender is a monster (type 1, id 0x2A, class 0x50, alive in mode 1, `killable` set, not a hireling class, no special-stun condition, invalid condition clear). Every vector that enters queued apply operation or the apply also has: hostility check 1, reach (reach predicate) 1 at both calls, attacker mode non-zero, defender room non-null and not a town, no leech inputs (`lifeDrain = manaDrain = 0` before the leech, so DPO-R15 neither rewrites nor draws), no state 0x5C. `conversionType` and `conversionPercent` are 0 on the basic attack (basic-attack operation zeroes the record; only the asm-only handlers of DPO-R33 or another caller's record template could set them), so DPO-V08 and DPO-V20 supply them as record inputs. "Draws" count seed steps on the named unit. Every vector: `runtime_observed: no`.

### 8.1 Executable vectors

Each case restarts from shared defaults and its own overrides. Unnamed states are absent, event registrations are empty, supplied callbacks make no unstated writes/draws, Monster Crit is 0, defender armor locations are empty, and an item does not qualify for durability unless stated. No case inherits another case's outputs or stepped seed. Inputs, draws and exact outputs embedded in each hand trace are part of that case. Supplied resource writes persist within the case; supplied dependency operations log their named calls and have only stated effects. A placement-stage vector logs its dependencies without executing unknown interiors, so it makes no unlisted dependency draw. Full-path callback/handler integration remains gated.

- **DPO-V01** lethal threshold (DPO-R12a, DPO-R17, DPO-R19b). (boundary: supplied post-build stage through final death flag)
  - Inputs: Post-build record flags 34, result 5, physical T, all other damage/drain/status/absorb terms 0. Defender life 2560, hpregen 0; no mitigation reductions/resistance; supplied P=100 and gates succeed. Independent T cases: (a)2560,(b)2557,(c)2305,(d)2304,(e)2816. End at final death-flag processing, before deferred/post-apply work.
  - Seeds: no draw at this stage
  - Draws: none; build drawAlt(1) precedes this entry; zero leech/status terms cause no local draw
  - Trace and exact outputs:
    Path: queued basic attack (section 2). Inputs: attacker type 0 id 7, defender type 1 id 0x2A alive in mode 1, stat 6 = 2560 (10.0), stat 0x4A = 0, no resistances or reductions (so mitigation leaves every field unchanged), `P = 100`. The vector starts at queued build completion after the build, with the record as built: flags `0x22` (bit 0x2 from basic-attack operation, bit 0x20 from the build), result word 0x0005, `physical = T`, every other damage, leech and status field 0, `absorbHeal = 0`. Draws: none from this point (the build's `drawAlt(1)` of DPO-R21a precedes it; no leech stats, so DPO-R15 makes no change to life). Trace: DPO-R12a `total = T`; pre-flag iff `(T & ~0xFF) > 2560`; enqueue; queued apply operation overwrites flags with 0x20 and applies with flag argument 0; DPO-R14 clamps `physical` to `min(T, 2560)`; DPO-R17 `n = 2560 - T`, `n < 0x100 -> 0`; DPO-R19b sets or clears bit 0x2. Outputs: `T = 2560`: pre-flag clear, stat 6 = 0, result bit 0x2 set. `T = 2557`: pre-flag clear (2304 > 2560 false), stat 6 = 0 (3 < 0x100), bit 0x2 set. `T = 2305`: stat 6 = 0, bit 0x2 set. `T = 2304`: stat 6 = 256 (exactly 1.0), bit 0x2 clear. `T = 2816`: pre-flag set at build (2816 > 2560), stat 6 = 0, bit 0x2 set.
  - Outputs: (a) prediction clear, life 0, lethal set; (b) prediction clear, life 0, lethal set; (c) life 0, lethal set; (d) life 256, lethal clear; (e) prediction set, life 0, lethal set. Stored total is T; physical after clamp is min(T,2560).
  - Discriminates: The strict prediction comparison differs from the final below-256 death threshold.
  - runtime_observed: no

- **DPO-V02** crit exclusivity (DPO-R03a, DPO-R02a). (boundary: supplied physical/crit stage)
  - Inputs: Attacker weapon stats 21=5,22=6; prior physical=0, physicalPercent=0; weapon present, weapon mode 0. Independent (weapon chance, stat 337, stat 141) cases: A(100,100,100), B(0,0,50), C(0,0,30), D(20,20,50). Other percentages/mastery/endpoints use shared zero defaults.
  - Seeds: attacker S1 in each independent case
  - Draws: A/B/C: attacker drawAlt(256)=95, then drawInRange(100)=31. D adds attacker drawInRange(100)=12 after the chance-20 weapon source fails. In A/B first successful crit source stops later sources; C fails its only positive source
  - Trace and exact outputs:
    Attacker seed `S1`; weapon stats 21 = 5, 22 = 6; prior `physical = 0`, `physicalPercent = 0`. Physical: `mn = 0x500`, `mx = 0x600`, `pct = 0`, `mn' = 0x500`, `mx' = 0x600`; draw 1 `drawAlt(0x100)` = step 1 `lo & 0xFF` = 95 → `physical = 0x55F`. Then:
      - case A: `ca = 100` (fifth argument 0, weapon present), stat 0x151 = 100, stat 0x8D = 100: source a draws `drawInRange(100)` = step 2 `lo mod 100` = 31 < 100 → `physical = 0xABE`, result bit 0x2000; b and c not tried; 2 draws; seed `{0x2F2ED81B, 0x2C890AFD}`.
      - case B: `ca = 0`, 0x151 = 0, 0x8D = 50: no draw for a, b; c: step 2 → 31 < 50 → `physical = 0xABE`, bit 0x2000; 2 draws.
      - case C: `ca = 0`, 0x151 = 0, 0x8D = 30: c: 31 >= 30 → no doubling, `physical = 0x55F`, no bit; 2 draws.
      - case D: `ca = 20`, 0x151 = 20, 0x8D = 50: a: 31 >= 20 fail; b: step 3 → 12 < 20 → `physical = 0xABE`, bit 0x2000; c not tried; 3 draws; seed `{0x280687C4, 0x13ADFE65}`.
  - Outputs: Physical roll 0x55F in every case. A/B/D physical=0xABE, result crit mask 8192 set; C physical=0x55F, crit clear. A/B/C make 2 draws and end at {lo 0x2F2ED81B, hi 0x2C890AFD}; D makes 3 and ends at {lo 0x280687C4, hi 0x13ADFE65}.
  - Discriminates: First successful source stops later draws and permits one doubling.
  - runtime_observed: no

- **DPO-V03** hand-off (DPO-R01a). (boundary: supplied hit-result stage)
  - Inputs: Attacker player mode 7, defender monster mode 1; hostility=1, reach=1, progressive-to-hit=0, to-hit=1, state 54 absent. Independent avoidance b cases:2,1,16,0. Additional case: defender player Run 3, state 54 present, b 0; final case: hostility 0. Hit/avoidance operations are supplied draw-free results; downstream stages are logged placement operations.
  - Seeds: no local draw; supplied hit/avoidance results make none here
  - Draws: none in the hand-off; downstream build draws belong to their own vectors
  - Trace and exact outputs:
    Inputs to hit-result operation: attacker type 0 in mode 7 (the mode matters only downstream, at queued apply operation), defender type 1 in mode 1, hostility check returns 1, reach check returns 1, `s = 0`, to-hit roll returns 1, block/avoid returns `b`, `HasState(defender, 0x36) = 0`. No draw inside hit-result operation itself (the to-hit and block draws are SBA's). Outputs (result word, then the downstream effect on the basic attack):
      - `b = 2`: 0x0100; the build still runs with its draws; queued build completion runs no scale and sets no pre-flag (bit 0 clear) but enqueues; queued apply operation: no apply, no durability, events 7 and 3 dispatched, reaction run, dequeue.
      - `b = 1`: 0x0010; same downstream.
      - `b = 0x10`: 0x8000; same downstream (`0x8000 & 0x8380 != 0`).
      - `b = 0`: 0x0005; `0x5 & 0x8380 == 0` and bit 0 set: scale and pre-flag run, apply runs.
      - defender type 0 in mode 3, `b = 0`, `HasState(defender, 0x36) = 1`: 0x0001 with the to-hit roll never called.
      - hostility check returns 0: 0x0000; the build still runs; nothing is applied.
  - Outputs: b 2 -> result 0x0100; b 1 ->0x0010; b 16 ->0x8000: build/enqueue and post 7/3/reaction/removal log, without scale/apply/durability. b 0 ->0x0005 and eligible scale/prediction/apply. Running/state 54 case ->0x0001 without to-hit call. Hostility 0 ->0x0000, still basic build, but no apply.
  - Discriminates: Avoided basic attacks still build/enqueue and dispatch post events.
  - runtime_observed: no

- **DPO-V04** physical roll (DPO-R02a, DPO-R03a). (boundary: supplied physical/crit stage)
  - Inputs: Attacker stats 21=5,22=9, no crit chances, ratio 128, prior physical 0. Case 1 other percent inputs 0. Case 2 stat 17=-50. Case 3 physical-supplied flag set, physical 0x700, weapon/passive/deadly chances 100 each.
  - Seeds: attacker S1 independently; case 3 takes no draw
  - Draws: case 1 attacker drawAlt(1024)=863; cases 2/3 none
  - Trace and exact outputs:
    Seed `S1`, stats 21 = 5, 22 = 9 (`mn = 0x500`, `mx = 0x900`), no crit chances. Case 1: `mn' = 0x500`, `mx' = 0x900`: draw 1 `drawAlt(0x400)` = step 1 `lo & 0x3FF` = 863 → `physical = 0x500 + 863 = 0x85F`; 1 draw. Case 2: stat 17 (`item_maxdamage_percent`) = -50: `mx' = 0x900 + mulDiv(-50, 0x900, 100) = 0x900 - 1152 = 0x480 < mn' = 0x500` → no draw, `physical = 0x500`. Case 3: hit-in set with `physical = 0x700` supplied and `ca = cb = cc = 100`: 0 physical and 0 crit draws, `physical = 0x700`, no bit 0x2000.
  - Outputs: case 1 physical 0x85F, one draw; case 2 physical 0x500, no draw; case 3 physical 0x700, no crit mask, no physical/crit draw.
  - Discriminates: Adjusted maximum below minimum costs no draw; hit-in skips every crit.
  - runtime_observed: no

- **DPO-V05** elements (DPO-R04a). (boundary: isolated elemental stage)
  - Inputs: Physical-supplied flag set; fire min 2, max 8, mastery 0; lightning min 3, max 0;cold/magic maxima 0. Incoming fire/lightning/magic 0, cold=-5. Independent ratio 128 and 64 cases.
  - Seeds: attacker S1 at fire in each case
  - Draws: one attacker drawAlt(1536)=351; no other elemental draw
  - Trace and exact outputs:
    Hit-in set (no physical or crit draw), seed `S1` at the fire roll. Attacker stats 0x30 = 2, 0x31 = 8, 0x149 = 0; 0x33 = 0 (0x32 = 3), 0x37 = 0, 0x35 = 0; incoming `fire = 0`, `lightning = 0`, `cold = -5`, `magic = 0`. Fire: `mx = 0x800 >= 8`, `mn = 0x200`, `m = 0`; draw 1 `drawAlt(0x600)` = step 1 `lo mod 0x600` = 351 → `fire = 0x200 + 351 = 0x35F`. Lightning: max stat 0 → `max(0, 0) = 0`, no draw. Cold: `max(0, -5) = 0` → `cold = 0`. Magic 0. Exactly 1 draw. With ratio 0x40 instead: `fire = 0x35F * 0x40 / 128 = 431 = 0x1AF`; `cold = 0 * 0x40 / 128 = 0`.
  - Outputs: ratio 128:fire 0x35F, lightning/cold/magic 0. Ratio 64:fire 431(0x1AF), other elements 0. Exactly 1 draw each.
  - Discriminates: Negative prior clamps before ratio scaling.
  - runtime_observed: no

- **DPO-V06** drain inputs (DPO-R05). (boundary: isolated drain stage)
  - Inputs: Baseline: attacker monster class 80(hireling classifier 0), life stats 60=2,61=4, all other drain stats 0, flags 0, incoming drain fields 0, ratio 128. Case 2 flag 4;case 3 prior lifeDrain 256. Case 4 player with stats 60=6,62=3,64=7,103=1. Case 5 hireling monster class 271 with case 4 stats. Case 6 attacker type 2, classifier 0, stats 60=6,62=3,64=7,103=1, all other drain stats 0, flags 0 and incoming drain fields 0. Cases reset baseline except their explicit overrides; player/hireling incoming terms are 0.
  - Seeds: attacker S1 independently in every case; player/hireling/unsupported-type paths draw nothing
  - Draws: cases 1/3 attacker drawAlt(512)=351; cases 2/4/5/6 none
  - Trace and exact outputs:
    Case 1: attacker type 1, class 0x50 (hireling predicate 0), seed `S1`, flags 0, stats 0x3C = 2, 0x3D = 4, 0x3E = 0x3F = 0x40 = 0x41 = 0, incoming `lifeDrain = manaDrain = staminaDrain = 0`: life: `mx = 0x400`, `mn = 0x200`, draw 1 `drawAlt(0x200)` = step 1 `lo & 0x1FF` = 351 → `lifeDrain = 0 + (0 + 0x200 + 351) = 0x35F`; mana and stamina: max stat 0 → `max(0, 0) = 0` added, no draw; 1 draw. Case 2: as case 1 with flags bit 4 set: 0 draws, `lifeDrain = 0`. Case 3: as case 1 with incoming `lifeDrain = 0x100`: `lifeDrain = 0x100 + (0x100 + 0x200 + 351) = 0x55F` (the prior field is counted twice). Case 4: attacker type 0 with stats 0x3C = 6, 0x3E = 3, 0x40 = 7, 0x67 = 1: 0 draws, `lifeDrain = 6`, `manaDrain = 3`, `staminaDrain = 0`, flags `|= 0x100`. Case 5: attacker type 1 of class 0x10F (hireling predicate 1) with the stats of case 4: same outputs as case 4, 0 draws.
    Case 6 is neither player/hireling nor regular monster, so no drain or bypass addition runs: all three drain fields and flags stay 0; attacker seed stays S1.
  - Outputs: case 1 lifeDrain0x35F;case 2 lifeDrain 0;case 3 lifeDrain0x55F. Cases 4/5 lifeDrain 6, manaDrain 3, staminaDrain 0, bypass-undead flag 256 set. Case 6 lifeDrain/manaDrain/staminaDrain 0, flags 0, no draw, seed S1 unchanged. Other drain outputs 0.
  - Discriminates: Monster prior counts twice; player/hireling stamina is not added. Treating every non-monster as a player fails case 6's zero drain and flag assertions.
  - runtime_observed: no

- **DPO-V07** poison roll (DPO-R06a). (boundary: isolated poison stage)
  - Inputs: Physical-supplied, no elements/drains; incoming poisonMagnitude/poisonLength 0. Case 1 stats 57=10,58=20,332=0,101=0,59=50,326=0. Case 2 min=max 20 with other case 1 inputs. Case 3 maximum 0 with other case 1 inputs. Boundary ends before mandatory burn draw.
  - Seeds: attacker S1 at poison in each case
  - Draws: case 1 attacker drawAlt(10)=1; cases 2/3 none
  - Trace and exact outputs:
    Seed `S1` at the poison roll (hit-in set, no elements, no drains), incoming `poisonMagnitude = 0`, `poisonLength = 0`. Case 1: stats 0x39 = 10, 0x3A = 20, 0x14C = 0, 0x65 = 0, 0x3B = 50, 0x146 = 0: `mn' = 10`, `mx' = 20`, draw 1 `drawAlt(10)` = step 1 `lo mod 10` = 1 → `poisonMagnitude = 11`; then (DPO-R06b) `poisonLength = 50`; 1 draw. Case 2: 0x39 = 0x3A = 20: `mx' = mn'` → no draw, `poisonMagnitude = 20`. Case 3: 0x3A = 0: `max(0, 0) = 0` → `poisonMagnitude = 0`, `poisonLength` untouched. In every case the `drawAlt(1)` of DPO-R21a follows next.
  - Outputs: case 1 poisonMagnitude 11, length 50;case 2 magnitude 20, no roll draw;case 3 magnitude 0, length untouched 0. Mandatory drawAlt(1) is the next build operation outside this isolated stage.
  - Discriminates: Raw poison endpoints are not shifted by 8.
  - runtime_observed: no

- **DPO-V08** conversion (DPO-R07). (boundary: isolated conversion stage)
  - Inputs: Physical0x1000, conversionPercent 25, other damage 0. Independent conversion types 1,5,4,7,10,0. Type 5 poisonLength cases 20 and 70;type 4 coldLength 0. Type 10 cases use S1 and S5. The physical-only 100-percent check supplies a nonzero conversion type; it asserts the subtraction before that type handler. Additional independent type 11 case starts burnLength 20;type 12 starts freezeLength 20. Each has physical 0x1000, conversionPercent 25, and other damage/length fields 0.
  - Seeds: no draw for fixed types; attacker S1/S5 at random-type cases
  - Draws: type 10 with S1:attacker draw(5)=1 -> selected type 2;with S5:draw(5)=4 -> type 5. All other cases no draw
  - Trace and exact outputs:
    Additional fixed cases: converted 1024 and physical 0xC00. Type 11 adds fire 0x400 and raises burnLength 20 to 50. Type 12 adds cold 0x400 and raises freezeLength 20 to 50. Neither takes a draw.
    `physical = 0x1000` after crit, `conversionPercent = 25`, no draw before this step. `conv = mulDiv(25, 0x1000, 100) = 1024 = 0x400`, `physical = 0xC00`. Type 1: `fire += 0x400`. Type 5: `poisonMagnitude += 0x80`, `poisonLength = max(prior, 50)`: prior 20 → 50, prior 70 → 70. Type 4: `cold += 0x400`, `coldLength = max(prior, 50)`. Type 7: only `physical = 0xC00`. `conversionPercent = 100`: `physical = 0`. Type 0xA with seed `S1`: 1 draw `draw(5)` = step 1 `lo mod 5` = 1 → `t = 2`: `lightning += 0x400`; with seed `S5`: `lo mod 5` = 4 → `t = 5` (poison). Type 0: nothing, `physical = 0x1000`.
  - Outputs: At percent 25 and nonzero type, converted 1024, physical 0xC00. Type 1 adds fire 0x400;type 5 adds poison 0x80 and lengths 50/70;type 4 adds cold 0x400, length 50;type 7 adds nothing. Percent 100 physical subtraction gives 0. Random S1 adds lightning 0x400;S5 selects poison. Type 0 keeps physical 0x1000. Type 11 fire 0x400, burnLength 50;type 12 cold 0x400, freezeLength 50;physical 0xC00 in each, with no draw.
  - Discriminates: Unsupported nonzero type still subtracts physical; random type is selected after subtraction.
  - runtime_observed: no

- **DPO-V09** scale (DPO-R08a, DPO-R11 table). (boundary: supplied-percent scaling)
  - Inputs: Case 1 P 17, physical 0x1000, fire 0x200000, cold=-0x100, coldLength 40, lifeDrain0x100. Case 2 P0x10001, physical 0x1000, other fields 0. Case 3 P 100 with the case 1 incoming record. This is scaling only, not mitigation.
  - Seeds: no draw
  - Draws: none
  - Trace and exact outputs:
    No draw. `P = 17`; record `physical = 0x1000`, `fire = 0x200000`, `cold = -0x100`, `coldLength = 40`, `lifeDrain = 0x100`. Outputs: `physical = mulDiv(17, 0x1000, 100) = 696 (0x2B8)`; `fire`: `v > 0x100000` and `100 <= v >> 4` → `(0x200000 / 100) * 17 = 20971 * 17 = 356507`; `cold` unchanged (`v <= 0`); `coldLength = 40` (scale flag 0); `lifeDrain = 0x100 * 17 / 100 = 43`. `P = 0x10001`, `physical = 0x1000`: `P > 0x10000` and `100 <= P >> 4` → `(0x10001 / 100) * 0x1000 = 655 * 4096 = 2682880`. `P = 100`: nothing changes.
  - Outputs: Case 1 physical 696(0x2B8), fire 356507, cold-0x100, coldLength 40, lifeDrain 43. Case 2 physical 2682880. Case 3 entire incoming record unchanged.
  - Discriminates: Reversing mulDiv arguments changes the large-value branch.
  - runtime_observed: no

- **DPO-V10** apply gate (DPO-R13a). (boundary: supplied apply-gate prefix ending before clamp)
  - Inputs: Attacker player id 7;defender monster id 42, class 80, mode 1, killable, invalid condition clear, room present/non-town. All record fields 0 except F/R. (1)F 0, R=1, hostility 0, scale false;(2)F 4096, R=1, hostility 0, false;(3)F 4096, R=33, hostility 0, false;(4)F 4096, R=1, hostility 0, scale true;(5)F 0, R=3, hostility 1, town true;(6)F 0, R=1, hostility 1, defender mode 12. Separate queue reach-failure cases use attacker type 1 or 0 and reach 0. Stop before clamp;scale is a supplied logged operation. Where a case's Outputs say "no events", that means no apply-dispatch requests 5/1 or 6/2, not internal event 11. Other cases explicitly request 5/1. Empty event registrations mean no callbacks; they do not suppress the caller's dispatch request.
  - Seeds: no draw; gate predicates and logged scale/dispatch operations are supplied
  - Draws: none
  - Trace and exact outputs:
    No draw. Apply entered with flag argument 0 directly (not through queued apply operation), record flags `F`, result `R = 0x0001`, attacker type 0 id 7, defender type 1 id 0x2A in mode 1 with invalid condition clear and `killable` set, defender room non-null and not a town; `physical = 0`, every other field 0 (so the later stages change nothing).
      - case 1: hostility check returns 0, `F = 0`: result 0x0000; defender `lastAttackerType`, `lastAttackerId`, lastAttackerPresence untouched; no events.
      - case 2: hostility 0, `F = 0x1000`: proceeds; `lastAttackerType = 0`, `lastAttackerId = 7`, lastAttackerPresence = true; no scale run; event 5 on the attacker's list with the defender pushed, then event 1 on the defender's list with the attacker pushed.
      - case 3: as case 2 with `R = 0x0021`: proceeds, no events.
      - case 4: hostility 0, flag argument 1, `F = 0x1000`, `R = 0x0001`: proceeds, the scale/mitigation entry runs once (DPO-R08c), no events (flags bit 0x20 clear).
      - case 5: room is a town, hostility 1, `R = 0x0003`: result 0x0001 (bit 1 cleared only), return before step 3.
      - case 6: hostility 1, `F = 0`, defender in mode 0xC: return after step 4 with `R` untouched and `lastAttackerType` not written.
      - queued path: queued apply operation with attacker type 1 calls reach predicate with `k = 3`, with type 0 `k = 1`; a return of 0 dequeues only.
  - Outputs: (1)R=0, last-attacker fields untouched, no apply events. (2)last attacker=(type 0, id 7, present), no scale, apply event 5(attacker, other defender) then 1(defender, other attacker). (3)same identity write, no apply events. (4)one scale request, no apply events. (5)R=1, return before killable test. (6)R=1, return before last-attacker write. Failed queued reach logs allowance 3 for monster/1 for player, then removal only.
  - Discriminates: Override bypasses hostility alone and town failure clears only lethal.
  - runtime_observed: no


- **DPO-V11** clamp and leech input (DPO-R14, DPO-R15). (boundary: supplied clamp/leech hand-off)
  - Inputs: Defender life 0x1000. Case 1 post-mitigation physical=total=0x1800;case 2 physical=total=0xC00. Zero drain/status/absorb terms;leech hand-off is logged with no resource effect.
  - Seeds: no draw at clamp/hand-off/subtraction; leech interior is supplied
  - Draws: none locally
  - Trace and exact outputs:
    No draw of these rules. Defender stat 6 = 0x1000; after mitigation `physical = 0x1800`, `total = 0x1800`. DPO-R14: `physical = 0x1000`; the leech call receives `physical = 0x1000` (its own draw: LSE-V09); `total` stays 0x1800, so DPO-R17 gives `0x1000 - 0x1800 < 0x100` → stat 6 = 0. With `physical = 0xC00`, `total = 0xC00`: unchanged by the clamp, stat 6 = 0x400 after DPO-R17.
  - Outputs: Case 1 physical 0x1000 reaches leech, total stays0x1800, life subtraction gives 0. Case 2 physical 0xC00 reaches leech, total0xC00, final life 0x400.
  - Discriminates: Clamping total would change final life; leech receives clamped physical.
  - runtime_observed: no

- **DPO-V12** absorb then subtract (DPO-R16, DPO-R17). (boundary: isolated absorb/subtract stage)
  - Inputs: Defender life 0x800, maxhp0x1000, mode 1, state 92 absent. (a)absorb0x400, total0xB00;(b)absorb0x400, total0xB01;(c)absorb0xC00 with total 0;(d)total 0 leaves either stated heal result unchanged by subtraction. Direct absorb-only dead-monster case uses mode 12 and positive absorb, with the same initial life/max. Final flag assertions invoke DPO-R19b after subtraction.
  - Seeds: no draw
  - Draws: none
  - Trace and exact outputs:
    No draw. Defender stat 6 = 0x800, stat 7 = 0x1000 (max-life read), alive (mode 1), no state 0x5C. `absorbHeal = 0x400`: `n = 0xC00 < 0x1000` → stat 6 = 0xC00, return 0x400; then `total = 0xB00` → stat 6 = 0x100, bit 0x2 clear; `total = 0xB01` instead → 0xFF → stat 6 = 0, bit 0x2 set. `absorbHeal = 0xC00`: `n = 0x1400 >= 0x1000` → stat 6 = 0x1000, return 0x800. `total = 0`: stat 6 unchanged. A direct call of absorb-heal operation for a defender in mode 0xC (not reachable through the apply, whose step 4 returns first): no-op, return 0.
  - Outputs: (a)heal to0xC00, return 0x400, final life 0x100, lethal clear. (b)same heal/return, final life 0, lethal set. (c)heal cap0x1000, return 0x800. Zero total leaves the healed life. Direct mode 12 absorb returns 0 and makes no write.
  - Discriminates: Absorb occurs before subtraction; a 255 remainder dies but 256 survives.
  - runtime_observed: no

- **DPO-V13** burn (DPO-R18). (boundary: isolated resource-burn stage)
  - Inputs: Independent cases: mana 0x1000, manaDrain0x140;mana 0x180, manaDrain0x100;mana 0x1000, manaDrain 0;stamina 0x800, staminaDrain0x900;stamina 0x800, staminaDrain0x600. Other burn term 0 in each case.
  - Seeds: no draw
  - Draws: none
  - Trace and exact outputs:
    No draw. Mana 0x1000, `manaDrain = 0x140` → stat 8 = 0xEC0. Mana 0x180, `manaDrain = 0x100` → 0x80 < 0x100 → stat 8 = 0. `manaDrain = 0` → unchanged. Stamina 0x800, `staminaDrain = 0x900` → negative → stat 10 = 0. Stamina 0x800, `staminaDrain = 0x600` → stat 10 = 0x200.
  - Outputs: Mana outputs respectively0xEC0,0,0x1000;stamina outputs respectively 0,0x200.
  - Discriminates: Positive fractional remainder below 256 becomes zero.
  - runtime_observed: no

- **DPO-V14** death flag (DPO-R19b). (boundary: isolated death-flag stage)
  - Inputs: After status calls, all status terms 0, hpregen 0. Case 1 life 0, predicted lethal clear;case 2 life 0x100, predicted lethal set. Other supplied result bits remain unchanged.
  - Seeds: no draw
  - Draws: none
  - Trace and exact outputs:
    No draw (no status fields, defender stat 0x4A = 0). After the appliers, stat 6 = 0: result bit 0x2 set (even if the pre-flag was clear). Stat 6 = 0x100 with the pre-flag set: bit 0x2 cleared.
  - Outputs: Case 1 result lethal mask 2 set;case 2 mask 2 cleared. Other result bits unchanged.
  - Discriminates: Final life replaces prediction.
  - runtime_observed: no

- **DPO-V15** unconditional draw (DPO-R21a, DPO-R21b). (boundary: isolated build stage)
  - Inputs: Player build, physical-supplied, all roll maxima 0, conversionType 0, ratio 128, burnMagnitude 0, burnLength 0;no callbacks/states modifying record.
  - Seeds: attacker S1 at build entry
  - Draws: one attacker drawAlt(1)=0
  - Trace and exact outputs:
    Any build: hit-in set, every roll range empty (all max stats 0), no conversion, player attacker, `burnMagnitude = 0`, `burnLength = 0`, seed `S1`: exactly 1 attacker draw `drawAlt(1)` = 0, seed becomes `{0x6AC6935F, 0}`; `burnMagnitude = (0 * 0x80 + 316 + 0) / 128 = 2`; `burnLength` stays 0 (not updated because it was 0).
  - Outputs: Seed{lo 0x6AC6935F, hi 0};burnMagnitude 2, burnLength 0;exactly 1 draw.
  - Discriminates: Modulus 1 advances the seed even though its result is zero.
  - runtime_observed: no

- **DPO-V16** monster double (DPO-R22). (boundary: isolated monster-double stage)
  - Inputs: Attacker monster with valid class row and supplied Crit 60, defender present. Physical0x400, fire 0x200, poisonMagnitude 10, lifeDrain0x100, coldLength 40, hitClass 0, other damage 0. Independent variants:hitClass0x20;Crit 51;Crit 0;attacker player. Overlay request is logged without downstream effect.
  - Seeds: attacker S1 independently
  - Draws: one attacker stepMod(100)=51 for each valid monster/nonzero-Crit case;no draw for Crit 0 or player
  - Trace and exact outputs:
    Attacker type 1 with supplied `Crit = 60` (isolated-rule input, not a named-cell claim), seed `S1` at this step, record `physical = 0x400`, `fire = 0x200`, `poisonMagnitude = 10`, `lifeDrain = 0x100`, `coldLength = 40`, `hitClass = 0`, defender passed. 1 draw `stepMod(100)` = 51 < 60 → `physical = 0x800`, `fire = 0x400`, `poisonMagnitude = 20`, `lifeDrain = 0x100` and `coldLength = 40` unchanged, `hitClass = 0x10`. With `hitClass = 0x20` instead: `hitClass` stays 0x20 and overlay request is called once with `(defender, 0x36, 0)` (its effect: DPO-Q18). `Crit = 51`: 1 draw, 51 < 51 false → no change, `hitClass` untouched. `Crit = 0`, or attacker type 0: 0 draws.
  - Outputs: Crit 60:physical 0x800, fire 0x400, poison 20, lifeDrain0x100, coldLength 40, hitClass0x10. With prior hitClass0x20 it stays0x20 and requests overlay 54 once. Crit 51 fails equality, record unchanged. Crit 0/player no draw or marker.
  - Discriminates: Crit equality fails and drain/length fields do not double.
  - runtime_observed: no

- **DPO-V17** durability draws (DPO-R24a). (boundary: supplied durability stage)
  - Inputs: Attacker case(a):player with primary weapon matching weap, not armo, not throwable, qualified;seed{lo 57, hi 666}. Variants use S1;throwable in Classic;or monster attacker. Defender case(b):attacker monster, defender player with inventory, armor only in locations 1(weight 3) and 3(weight 5), both qualified, others empty. Variant chosen item unqualified;final variant monster/monster. Loss is a supplied logged DUR request.
  - Seeds: (a)specified attacker seed or S1 per independent variant;(b)defender S1;other seed unchanged
  - Draws: (a)one attacker stepMod(100)=3 for seed 57, or 51 for S1;Classic throwable/monster none. (b)defender draw(7)=2, draw(8)=3, then selected qualified item stepMod(100)=12;unqualified selected item stops after 2 draws
  - Trace and exact outputs:
    (a) Attacker type 0 with seed `{lo 57, hi 0x29A}` and a primary weapon of type `weap` (0x2D), not `armo`, not throwable, that qualifies (durability qualification non-zero): `t = 4`; 1 attacker draw `stepMod(100)`: step gives `lo 0xC6363E77`, `lo mod 100 = 3 < 4` → a loss (DUR-R05 arithmetic). With seed `S1`: 51 → no loss. A throwable weapon in a classic game (`Expansion flag = 0`): 0 draws. Monster attacker: 0 attacker draws. (b) Attacker type 1, defender type 0 with seed `S1` whose `inventory` holds `armo` items only in slots 1 (weight 3) and 3 (weight 5), both qualifying: `W = 8`; draw 1 `Rand(7)` = step 1 `lo mod 7` = 2 → start index 2; draw 2 `Rand(8)` = step 2 `lo & 7` = 3 → `r = 3`; walk: index 2 (slot 4) empty, 3 (slot 5) empty, 4 (slot 8) empty, 5 (slot 9) empty, 6 (slot 10) empty, 0 (slot 1, weight 3): `3 < 3` false → `r = 0`; index 1 (slot 3, weight 5): `0 < 5` → chosen; draw 3 in item durability operation with the defender's seed, `t = 10`: step 3 `lo mod 100` = 12 >= 10 → no loss; 3 defender draws, seed `{0x280687C4, 0x13ADFE65}`. If the chosen item does not qualify: 2 draws. Monster vs monster: 0.
  - Outputs: (a)roll 3<threshold 4 requests loss;S1 roll 51 no loss. Classic throwable/monster no attacker draw. (b)location 3 selected, roll 12>=10 so no loss;3 draws, end{lo 0x280687C4, hi 0x13ADFE65}. Unqualified choice 2 draws. Monster/monster 0 draws.
  - Discriminates: Starting modulus is seven table entries, not eligible count; qualification controls third draw.
  - runtime_observed: no

- **DPO-V18** stun draw (DPO-R27). (boundary: isolated stun-draw stage)
  - Inputs: StunLength 25, monster defender with monster data and special-stun condition true. Independent attacker S1/S5 cases. Other variants:player defender;stunLength 0;condition false. This stage asserts draw/cancellation only.
  - Seeds: attacker S1 or S5 per case;defender seed supplied unchanged
  - Draws: S1 attacker stepMod(100)=51;S5=99;other variants no draw
  - Trace and exact outputs:
    `stunLength = 25`, defender type 1 with special-stun condition set. Attacker seed `S1`: 1 attacker draw, 51 < 90 → stun cancelled; defender seed unchanged. Attacker seed `S5`: 99 >= 90 → the applier continues. Defender type 0, or `stunLength = 0`, or bit 3 clear: 0 draws.
  - Outputs: 51<90 cancels;99>=90 continues CSA. Defender seed unchanged;player/zero-length/condition-false variants 0 draws.
  - Discriminates: Attacker seed advances while defender seed stays.
  - runtime_observed: no

- **DPO-V19** freeze fallback (DPO-R28a). (boundary: supplied cold/fallback stage)
  - Inputs: Defender player, cold effect E=-50, coldLength 25, freezeLength 50;independent variant coldLength 0, freezeLength 50. Monster variant E 0, coldLength 25, freezeLength 0. Freeze player eligibility always falls back. This stage asserts cold/fallback draw contract.
  - Seeds: defender S1 in each independent case
  - Draws: Player 25/50:defender stepMod(100)=51 then 31. Player 0/50:one 51. Monster E 0:none
  - Trace and exact outputs:
    Defender type 0 with seed `S1`; its cold effect `E = -50` (cold-effect operation for a non-monster), `coldLength = 25`, `freezeLength = 50`: cold draw (step 1, roll 51) then the freeze applier re-applies 50 as cold (a player always falls back; step 2, roll 31): 2 defender draws, seed `{0x2F2ED81B, 0x2C890AFD}`. `coldLength = 0`, `freezeLength = 50`: 1 defender draw. A monster defender whose `MonStatsRecord.ColdEffect` is 0 (`E = 0`) with `coldLength = 25`, `freezeLength = 0`: 0 draws.
  - Outputs: Two-draw case seed{lo 0x2F2ED81B, hi 0x2C890AFD};0/50 case 1 draw;E 0 case 0 draws.
  - Discriminates: Player freeze fallback adds another defender draw.
  - runtime_observed: no

- **DPO-V20** draw order (DPO-R29, DPO-R26). (boundary: supplied build stage)
  - Inputs: Attacker player, hit result 5, physical-supplied clear;stats 21=5,22=9,337=50, other crit chances 0;fire 48=2,49=8, mastery 0;other elements/drains/poison 0, burnMagnitude/Length 0, conversionType 10, conversionPercent 25. No callbacks/state handlers, stun/cold/freeze/leech effects. Independent variant stat 337=0. Build-only boundary.
  - Seeds: attacker S1 independently
  - Draws: Base:attacker drawAlt(1024)=863, drawInRange(100)=31, drawAlt(1536)=452, drawAlt(1)=0, draw(5)=2. No-crit variant:863, fire drawAlt(1536)=27, drawAlt(1)=0, draw(5)=3
  - Trace and exact outputs:
    Attacker type 0 with seed `S1`, hit (result 0x0005), hit-in clear, weapon stats 21 = 5, 22 = 9, stat 0x151 = 50 (other crit sources 0), fire stats 0x30 = 2, 0x31 = 8, other elements 0, no drains, no poison, `burnMagnitude = 0`, `burnLength = 0`, `conversionType = 0xA`, `conversionPercent = 25`, defender without the stun flag, no cold/freeze, no leech, no event callbacks, no DPO-R33 states. Build draws on the attacker, in order: 1 physical `drawAlt(0x400)` = 863 → `physical = 0x85F`; 2 crit b `drawInRange(100)` = 31 < 50 → `physical = 0x10BE`, bit 0x2000; 3 fire `drawAlt(0x600)` = step 3 `lo mod 0x600` = 452 → `fire = 0x3C4`; 4 `drawAlt(1)` = 0 → `burnMagnitude = 2`; 5 conversion `draw(5)` = step 5 `lo mod 5` = 2 → `t = 3`: `conv = mulDiv(25, 0x10BE, 100) = 1071 (0x42F)`, `physical = 0xC8F`, `magic = 0x42F`. 5 draws; seed `{0xBFC793C3, 0x4C305F9E}`. Durability (DPO-R24a) and then the reaction draws ([combat-hit-recovery](combat-hit-recovery.md)) come after all of these. With stat 0x151 = 0: 4 draws (physical 863; fire = step 2 `lo mod 0x600` = 27 → `fire = 0x21B`; `drawAlt(1)`; `draw(5)` = step 4 `lo mod 5` = 3 → `t = 4`, cold).
  - Outputs: Base physical 0xC8F, fire 0x3C4, magic 0x42F, burn 2, crit set;5 draws, seed{lo 0xBFC793C3, hi 0x4C305F9E}. No-crit:4 draws, fire 0x21B, conversion selects cold type 4. Durability/reaction placement is outside this build boundary;V29 supplies a reaction case.
  - Discriminates: Removing crit shifts later elemental/conversion draws.
  - runtime_observed: no

- **DPO-V21** status-length strip (DPO-R10). (boundary: isolated strip stage)
  - Inputs: Independent cases:coldLength 40, freezeLength 20, stat 153=1;(40,20), stat 153=0, stat 118=1;(-3,20), stat 118=1;poisonLength 50 with state 133;burnLength 30 with state 131;poison 50/burn 30 without those states. Other strip controls 0/absent.
  - Seeds: no draw
  - Draws: none
  - Trace and exact outputs:
    No draw. `coldLength = 40`, `freezeLength = 20`: defender stat 0x99 = 1 → both 0. Stat 0x99 = 0, 0x76 = 1 → 20 and 10. `coldLength = -3`, `freezeLength = 20`, 0x76 = 1 → -1 and 10 (truncation toward zero). `poisonLength = 50` with state 0x85 → 0; `burnLength = 30` with state 0x83 → 0; without those states both unchanged.
  - Outputs: Cold/freeze respectively(0,0),(20,10),(-1,10). Poison/burn with matching shrine state become 0;without them remain 50/30.
  - Discriminates: A negative partner halves toward zero.
  - runtime_observed: no

- **DPO-V22** length terms (DPO-R06b). (boundary: isolated length stage)
  - Inputs: Poison case:poisonMagnitude 11, prior length 0, stat 101=0, stat 59=50, stat 326=3. Variant stat 101=75;variant poisonMagnitude 0. Cold cases:cold 0x300, prior coldLength 0, stat 56=25, ratio 128 or 64;variant cold 0. Stun cases:incoming 0 or 5, stat 66=40. Cases are independent.
  - Seeds: no local draw
  - Draws: none in length stage
  - Trace and exact outputs:
    No draw of this rule. Poison: after DPO-V07 case 1 (`poisonMagnitude = 11`), stats 0x65 = 0, 0x3B = 50, 0x146 = 3: `poisonLength = (0 + 50) / 3 = 16`; with 0x65 = 75: `poisonLength = 75` and 0x3B, 0x146 unused; with `poisonMagnitude = 0`: `poisonLength` untouched. Cold: `cold = 0x300`, stat 0x38 = 25 → `coldLength += 25`; with ratio 0x40 → `+= 12`; with `cold = 0` → unchanged. Stun: `stunLength = 0`, stat 0x42 = 40 → `stunLength = 40`; `stunLength = 5` → stays 5.
  - Outputs: Poison length 16 with ordinary length/count,75 with override, unchanged 0 when poison 0. Cold length additions 25 or 12, none when cold 0. Stun 0 becomes 40;incoming 5 stays 5.
  - Wrap case (independent): cold 0x300, prior coldLength 0, stat 56=0x1000000, ratio 128. scaleAlways multiplies even at ratio 128: `wrapped32(0x1000000*128)` = -2147483648, truncated /128 = -16777216. Output coldLength -16777216; no draw.
  - Discriminates: Poison override bypasses division and nonzero incoming stun stays. Skipping the multiplication at ratio 128 would give coldLength 16777216 in the wrap case.
  - runtime_observed: no

- **DPO-V23** burn update (DPO-R21b). (boundary: isolated burn-update stage)
  - Inputs: BurnMagnitude0x100, burnLength 25, stat 315=10, ratio 128. Independent variants:incoming burnLength 0;ratio 64 with length 25. Mandatory burn draw belongs to this stage. Additional independent nonpositive-x case: burnMagnitude=-4, burnLength 0, ratio 128, stat 315=10, attacker S1.
  - Seeds: attacker S1 in each case
  - Draws: one attacker drawAlt(1)=0 per case
  - Trace and exact outputs:
    Nonpositive-x case: wrapped x=(-4*128)+316+0=-196. max(0, x)=0, so the added quotient is 0 and burnMagnitude remains-4. Zero incoming burnLength stays 0. The one drawAlt(1) returns 0 and steps S1 once.
    Seed `S1` (the `drawAlt(1)` returns 0). `burnMagnitude = 0x100`, `burnLength = 25`, stat 0x13B = 10, ratio 0x80: `x = 0x100 * 0x80 + 316 + 0 = 33084`, `33084 / 128 = 258` → `burnMagnitude = 514`; `burnLength = 25 + 10 = 35`. With `burnLength = 0`: `burnMagnitude = 514`, `burnLength = 0`. Ratio 0x40: `x = 16384 + 316 = 16700`, `/128 = 130` → `burnMagnitude = 386`; `burnLength = 25 + 10 * 0x40 / 128 = 30`.
  - Outputs: Base magnitude 514, length 35;zero-length variant magnitude 514, length 0;ratio 64 magnitude 386, length 30. Nonpositive-x case burnMagnitude=-4, burnLength 0, one attacker draw, seed{lo 0x6AC6935F, hi 0}.
  - Discriminates: Prior burn magnitude is added to the scaled term.
  - runtime_observed: no

- **DPO-V24** element pairing and mastery (DPO-R04b). (boundary: isolated element-pair stage)
  - Inputs: Physical-supplied;fire stats 48=2,49=8, mastery 329=50;cold 54=1,55=4, mastery 331=0;lightning/magic maxima 0;incoming elements 0, ratio 128. At this supplied-input boundary each layer-0 Stat read logs (unit, stat id, returned value), in request order, before returning the supplied value; reads have no other effect.
  - Seeds: attacker S1 at fire
  - Draws: attacker drawAlt(2304)=1887, then cold drawAlt(768)=27;no lightning/magic draw
  - Trace and exact outputs:
    Hit-in set, seed `S1` at the fire roll. Fire 0x30 = 2, 0x31 = 8, 0x149 = 50: `mn' = 0x200 + mulDiv(50, 0x200, 100) = 0x300`, `mx' = 0x800 + mulDiv(50, 0x800, 100) = 0xC00`; draw 1 `drawAlt(0x900)` = step 1 `lo mod 0x900` = 1887 → `fire = 0x300 + 1887 = 0xA5F`. Cold 0x36 = 1, 0x37 = 4, 0x14B = 0: `mn' = 0x100`, `mx' = 0x400`; draw 2 `drawAlt(0x300)` = step 2 `lo mod 0x300` = 27 → `cold = 0x11B`. Lightning and magic max stats 0: no draw, stats 0x14A and 0x165 not read. 2 draws.
  - Outputs: Fire0xA5F, cold 0x11B;2 draws. Lightning/magic 0. Exact stat-read log: (attacker,49,8), (attacker,48,2), (attacker,329,50), (attacker,51,0), (attacker,55,4), (attacker,54,1), (attacker,331,0), (attacker,53,0). There is no read of mastery 330 or 357, nor minimum 50 or 52.
  - Discriminates: Mastery changes both endpoints. Reading a skipped element's mastery even when its default value is 0 adds a log entry and fails the exact stat-read assertion.
  - runtime_observed: no

- **DPO-V25** applier order (DPO-R19a). (boundary: supplied application-order stage)
  - Inputs: After life subtraction:manaDrain0x100, staminaDrain0x100, stunLength 25, coldLength 40, freezeLength 50, poisonMagnitude 10, poisonLength 50, burnMagnitude 5, burnLength 25;defender mana/stamina 0x1000. Appliers are supplied logged operations without internal effects/draws. Independent zero-term case: all listed drain, magnitude and length inputs 0, defender mana/stamina 0x1000, with the same logged appliers.
  - Seeds: no local draw; supplied appliers draw nothing here
  - Draws: none; separate applier-draw cases are V18/V19
  - Trace and exact outputs:
    Record after DPO-R17 with `manaDrain = 0x100`, `staminaDrain = 0x100`, `stunLength = 25`, `coldLength = 40`, `freezeLength = 50`, `poisonMagnitude = 10`, `poisonLength = 50`, `burnMagnitude = 5`, `burnLength = 25`, defender mana 0x1000 and stamina 0x1000. Output: the calls in order `manaBurn(0x100)` (stat 8 = 0xF00), `staminaBurn(0x100)` (stat 10 = 0xF00), `stun(attacker, defender, 25)`, `cold(attacker, defender, 40)`, `freeze(attacker=attacker, defender=defender, length=50)`, `poison(attacker, defender, 10, 50)`, `burn(attacker, defender, 5, 25)`; the resource writes happen before the first applier. Draws inside the appliers: DPO-V18, DPO-V19.
    In the zero-term case burns make no resource write, but all five applier calls still run in the same order with zero arguments; the supplied appliers log and return without effects.
  - Outputs: Base case: mana0xF00 then stamina 0xF00 before first applier. Call log:stun(attacker, defender,25), cold(attacker, defender,40), freeze(source attacker, target defender,50), poison(attacker, defender,10,50), burn(attacker, defender,5,25). Zero-term case: mana/stamina stay 0x1000, no resource writes; exact call log stun(attacker, defender,0), cold(attacker, defender,0), freeze(source attacker, target defender,0), poison(attacker, defender,0,0), burn(attacker, defender,0,0); no draw.
  - Discriminates: Mana/stamina writes precede stun; freeze target is defender. Skipping zero-term applier calls fails the zero-term call log.
  - runtime_observed: no

- **DPO-V26** event 11 (DPO-R09, DPO-R23; IEP-R07). (boundary: supplied shared-entry integration)
  - Inputs: P 17, physical 0x1000, coldLength 40, defender stat 153=1. One persistent defender event 11 registration:skill kind, executing false, removeAfterCall false, values(0,0). Callback logs arguments/observed fields, makes no write/draw, returns 0.
  - Seeds: no draw
  - Draws: none including supplied callback
  - Trace and exact outputs:
    Execution: shared-entry integration
      with supplied scale `P = 17`. No draw, including the supplied callback. Inputs: physical
      `0x1000`, coldLength 40, defender stat 153 `item_cannotbefrozen` = 1; one persistent event-11
      entry on the defender's list, skill kind, executing false, removeAfterCall false, callback
      values `(0,0)`. Its callback records arguments and field values, writes nothing, and returns 0.
      Expected: scaling gives physical `0x2B8`; dispatch occurs before the strip with owner
      defender and other attacker; callback arguments are `(game,11,defender,attacker,record,0,0)`
      using the same record. The callback observes physical `0x2B8` and coldLength 40; the entry
      remains registered, running false; DPO-R10 then sets coldLength 0. The callback's observed
      duration distinguishes dispatch before the strip from dispatch after it.
  - Outputs: Scale physical 0x2B8;callback(game,11, defender, attacker, shared record,0,0) observes physical 0x2B8, coldLength 40. Registration remains, running false. Subsequent strip gives coldLength 0.
  - Discriminates: Callback sees cold length before strip.
  - runtime_observed: no

- **DPO-V27** dispatch sequence of one basic attack (DPO-R23, DPO-R08c, DPO-R08d). (boundary: supplied dispatch-placement integration)
  - Inputs: Basic-attack stage with supplied physical endpoints low 256, high 512 and endpoint percentages 0, weapon/mastery bonuses 0;physical-supplied clear, ratio 128, all elemental/drain/poison maxima 0, conversion 0, burn prior 0/length 0, attacker player. Gate results succeed, P 100, callbacks/state handlers absent, Crit 0, qualification false, no armor. Independent cases:hit result 5/defender life 2560;hit result 5/defender life 256;miss result 256/defender life 2560. Status lengths 0, leech 0, thorns 0. Unread wrappers, post-apply mark/overlay producers, queue storage, monster update and reaction are supplied draw-free operations without record/resource effects. Build arithmetic and dispatches execute;only the named dependency interiors are supplied.
  - Seeds: attacker S1 in every independent case;defender S5, unchanged
  - Draws: attacker physical drawAlt(256)=95, then mandatory drawAlt(1)=0, in every hit/lethal/miss case;no other draw at this supplied dependency boundary
  - Trace and exact outputs:
    Supplied low 256 plus roll 95 gives physical 351. The mandatory modulus 1 draw returns 0 and burnMagnitude becomes 2 with length 0. The first hit survives raw subtraction;the second reaches final life 0;the miss skips scale/apply.
    Attacker type 0, defender type 1, no lethal, no state 0x36 test needed. Hit (result 0x0005): dispatcher calls in order `(defender list, 3, attacker)` during the build (after DPO-R21, before DPO-R07), `(defender, 0xB, attacker)` inside the one scale/mitigation run, `(attacker, 5, defender)`, `(defender, 1, attacker)` in the apply, `(attacker, 7, defender)`, `(defender, 3, attacker)` after the durability step: 6 dispatches, 1 scale run. Lethal hit: `(defender, 0xA, attacker)`, `(attacker, 9, defender)` are inserted after the appliers and before event 7. Miss (result 0x0100): `(defender, 3, attacker)` during the build, then `(attacker, 7, defender)`, `(defender, 3, attacker)`; no 0xB, 5, 1; 0 scale runs.
  - Outputs: Surviving hit dispatches(defender 3, defender 11, attacker 5, defender 1, attacker 7, defender 3):6 dispatches,1 scale run. Lethal inserts defender 10 then attacker 9 before 7:8 dispatches,1 run. Miss dispatches defender 3, attacker 7, defender 3:3 dispatches,0 runs. Each case has 2 attacker draws, ending{lo 0x2F2ED81B, hi 0x2C890AFD};defender seed unchanged. Physical roll is 351 before any lethal clamp.
  - Discriminates: Miss has build 3 and post 7/3, without 11/5/1; lethal inserts 10/9.
  - runtime_observed: no

- **DPO-V28** bypass flag and loop stop (DPO-R11). (boundary: supplied mitigation-dispatch stage)
  - Inputs: Mitigation operations log(entry index, field, bypass), make no write/draw. Independent cases:(1)attacker player, defender monster class 80, lUndead true, demon false, flags 256;(2)same classification, flags 512;(3)defender neither undead/demon, flags 1024;(4)undead, flags 1024;(5)attacker regular monster class 80, defender player, flags 1024;(6)attacker hireling class 271, defender neutral monster, flags 0. Use the section 3 ordered table.
  - Seeds: no draw
  - Draws: none including supplied mitigation
  - Trace and exact outputs:
    No draw. Defender type 1 (class 0x50), table of section 3; the output is the sequence of per-entry mitigation calls (entry index, field, `by`), whose arithmetic is [combat-damage-mitigation](combat-damage-mitigation.md)'s.
      - case 1: attacker type 0, defender MonStats lUndead set, demon clear, flags 0x100: `by = 1`; 9 calls, entries 0 to 8 in order (`physical`, `fire`, `lightning`, `cold`, `magic`, `coldLength`, `freezeLength`, `poisonLength`, `poisonMagnitude`); entry 9 (`lifeDrain`, marker 1) ends the loop, so `lifeDrain`, `manaDrain`, `staminaDrain` are not mitigated.
      - case 2: as case 1 with flags 0x200: `by = 0`; same 9 calls.
      - case 3: defender neither undead nor demon, flags 0x400: `by = 1`; 9 calls.
      - case 4: defender undead, flags 0x400 only: `by = 0` (the 0x400 test is skipped for an undead or demon defender); 9 calls.
      - case 5: attacker type 1 class 0x50 (hireling predicate = 0), defender type 0, flags 0x400: `by = 1`; 12 calls, entries 0 to 11.
      - case 6: attacker type 1 class 0x10F (hireling predicate = 1, hireling), defender type 1 neither undead nor demon, flags 0: `by = 0`; 9 calls.
  - Outputs: Cases 1/2/3/4/6 log entries 0..8 in order, stop before lifeDrain entry 9;bypass respectively 1/0/1/0/0. Case 5 logs 0..11, bypass 1. Fields of unvisited entries unchanged.
  - Discriminates: Monster-only marker ends the loop rather than skipping one entry.
  - runtime_observed: no

- **DPO-V29** Reaction on a queued miss (DPO-R13a, DPO-R23, DPO-R26; boundary: supplied queue-consumption stage)
  - Inputs: game frame 100; attacker player type 0, id 7, mode 7, seed S1; defender monster type 1, id 42, class 80, mode 1, life 2560, seed S5. One matching queued record has flags 34, result 256, all damage/drain/status terms 0. Supplied reach returns 1. Event registrations are empty. Dispatch and reaction are supplied operations that log calls, return 0 and make no writes or draws. The entry profile starts at queue consumption after build; no build or hit operation executes.
  - Seeds: attacker S1 and defender S5 as defined above; no draw.
  - Draws: none; both seeds remain their entry pairs.
  - Trace: copy the matching record; reach succeeds. Result lacks hit, so skip apply, marks and durability. Dispatch attacker event 7 then defender event 3. Hit is still clear, so skip reflection. Invoke reaction with target defender, source attacker and the copied record. Remove matching queued entries.
  - Outputs: exact call log is reach(attacker, defender, allowance=1), dispatch(game,7, attacker, defender, record), dispatch(game,3, defender, attacker, record), reaction(game, target=defender, source=attacker, record), removeMatching(attacker, defender). Queue is empty. Result remains 256, life remains 2560, no resource/status/durability write, attacker seed remains {lo 1, hi 666}, defender seed remains {lo 5, hi 666}.
  - (a) Inputs: isolated queued tail immediately after events 7/3 have completed; their execution and all earlier stages are outside this case. Attacker player type 0/id 7 is now in mode 0; defender monster type 1/id 42 is in mode 1 with life 2560. One matching queue entry remains. Copied record has hit result 1, flags 32 and all damage/drain/status terms 0. Attacker seed S1, defender seed S5. Reflection, reaction and queue removal log requests and otherwise make no effects; the initial request log is empty. Draws: none. Trace: surviving-hit path repeats the attacker mode test; player mode 0 returns before reflection, reaction or removal. Outputs: empty request log, one matching queue entry retained unchanged, result 1, defender life 2560, both seeds unchanged. No earlier apply/event assertion is made.
  - Discriminates: hit-only reaction omits the base case's fourth logged call; reaction before events or after removal changes its exact position. Omitting the repeated mode test fails (a)'s retained queue and empty tail-request log.
  - runtime_observed: no

- **DPO-V30** Regular-monster and hireling totals (DPO-R12a; boundary: supplied post-build scale/total/prediction stage ending after enqueue)
  - Inputs: attacker cases (a)monster type 1, class 80, hireling classifier 0;(b)monster type 1, class 271, classifier 1;(c)player type 0, classifier 0. Attacker id 7, defender monster id 42, class 80, mode 1, life 512. Prepared record flags 34, result 5, physical 512, lifeDrain 256, all other fields 0. Already-built skips build. Supplied P 100, identity mitigation for each visited entry, no callbacks. Initial attacker queue is empty;queue storage is supplied without extra effects. Stop before queued apply.
  - Seeds: no draw;unit seeds unchanged.
  - Draws: none;scale, identity mitigation and supplied queue storage make no draw.
  - Trace: the six-field sum is 512. Stored total adds 256 only for case(a), giving 768;cases(b)/(c)store 512. Prediction adds 256 for both monster-type cases(a)/(b), giving 768;player(c)predicts 512. Compare masked prediction against masked life 512:768>512 true;512>512 false. Enqueue the resulting record at the empty head.
  - Outputs: (a)stored total 768, prediction 768, result 7;(b)stored total 512, prediction 768, result 7;(c)stored total 512, prediction 512, result 5. Each has one queued record with those exact fields, physical 512, lifeDrain 256, flags 34. Defender life stays 512 and both seeds stay unchanged.
  - Discriminates: treating hirelings as regular monsters in stored total makes(b)768;omitting lifeDrain from monster-type prediction leaves(b)result 5;using non-strict prediction sets(c)lethal. Each contradicts an exact assertion.
  - runtime_observed: no

- **DPO-V31** Flagged apply attacker-event suppression (DPO-R13a, DPO-R23; boundary: supplied apply-gate prefix ending before clamp)
  - Inputs: attacker player type 0, id 7, mode 7;defender monster type 1, id 42, class 80, mode 1, killable, invalid condition clear, life 2560. Hostility true, room present/non-town. Direct apply requests scale. Record result 1, all damage/resource/status terms 0;independent flags cases(a)32 and(b)160. Scale is a supplied logged operation with no record effect;its interior event 11 is outside this prefix vector. Dispatch operations log their arguments, return 0, and make no mutation/draw. No queued flags replacement occurs.
  - Seeds: no draw;unit seeds unchanged.
  - Draws: none in the prefix or its supplied operations.
  - Trace: gates succeed;write last attacker type 0/id 7/presence true;request scale once. Result suppress-events mask 32 is clear and hit is set. Flags 32 enables flagged events. In(a)flags 128 is clear, so dispatch attacker 6 then defender 2. In(b)flags 128 is set, so skip attacker 6 and dispatch defender 2 alone.
  - Outputs: (a)one scale request, then dispatch(game,6, attacker, defender, shared record), dispatch(game,2, defender, attacker, shared record);(b)one scale request, then dispatch(game,2, defender, attacker, shared record). Both write last attacker(0,7, present), keep result 1, and preserve their input flags 32/160, life 2560 and both seeds. No clamp/resource/status operation has run.
  - Discriminates: always dispatching 6 fails(b);always suppressing 6 fails(a);reversing 6/2 fails(a). Callback return 0 does not suppress either documented dispatch.
  - runtime_observed: no

### 8.2 Observation vectors

- **DPO-O01** Lethal boundary: measure raw life, total, predicted/final lethal. Expect raw subtraction with surviving remainder at least 256, otherwise 0; prediction uses strict whole-part comparison. [Procedure](../research/spec-evidence/combat-damage-pipeline-order.md#observation-procedures).
- **DPO-O02** Crit order/exclusivity: measure attacker seed and physical at build entry/return. Physical draws first;positive crit sources draw until first success;physical doubles at most once. Count other build draws too. [Procedure](../research/spec-evidence/combat-damage-pipeline-order.md#observation-procedures).
- **DPO-O03** Mana-burn quirk: with supplied Normal ManaStealDivisor 1, M 5 and completed leech, expect 320 raw defender mana loss, floored 0 below 256. This supplied divisor is not a new archive-cell claim. [Procedure](../research/spec-evidence/combat-damage-pipeline-order.md#observation-procedures).

## 9. Acceptance and dependencies

The verified subset is the statically verified Id-map rows and their named DPO-V01..V31 assertions, executed at the stated supplied-input stage. Passing this subset is partial acceptance. Supplied dependency/producer results do not verify their production; callback/handler and other explicit integration gates remain necessary for full reproduction. Full acceptance requires closing every in-scope behavioral/integration gate, vectors for those assertions, and matching DPO-O01 through DPO-O03 plus the remaining behavioral observations. Pending observation alone limits runtime-confirmation claims. No passing partial subset establishes complete parity.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) (NRF-R01, NRF-R02 for seeds/draws; NRF-R05, NRF-R06a for arithmetic; NRF-R08a for units; NRF-R12a, NRF-R12b for state/stat reads); [core-game-data-fields](core-game-data-fields.md) (GDF-R02, R05, R09, R11, R12, R14 to R18); [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md) (to-hit roll and block/avoid return bits, DPO-R01a); [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md) (PVP-R01a, R05a, R06 for DPO-R08b, R13); [combat-damage-mitigation](combat-damage-mitigation.md) (DMI-R05 and the 12-type arithmetic for DPO-R11); [combat-leech-and-status-effects](combat-leech-and-status-effects.md) (LSE-R10, R12, R17, V09 for DPO-R15 to R18); [combat-status-application](combat-status-application.md) (CSA-R06, CSA-R07, CSA-R10 to CSA-R12 for DPO-R19, R27, R28); [item-durability-loss](item-durability-loss.md) (DUR-R05 for DPO-R24a); [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md) (IEP-R01/IEP-R02 registration and IEP-R07 dispatcher for DPO-R23); [combat-hit-recovery](combat-hit-recovery.md) (DPO-R26); [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md) (SCD-R06, R09, R14a for section 2); [state-timed-state-lifecycle](state-timed-state-lifecycle.md) (expiry of the states set by DPO-R19a).

**Depended on by:** [combat-damage-mitigation](combat-damage-mitigation.md), [combat-leech-and-status-effects](combat-leech-and-status-effects.md), [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md), [combat-hit-recovery](combat-hit-recovery.md), [combat-knockback](combat-knockback.md), [combat-crushing-blow](combat-crushing-blow.md), [combat-open-wounds](combat-open-wounds.md), [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md), [item-durability-loss](item-durability-loss.md), [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md) (apply entry and DPO-R22), [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md), [skill-assassin-death-sentry-corpse-damage](skill-assassin-death-sentry-corpse-damage.md) and every skill spec whose do-function reaches queued build completion or direct damage application.

## 10. Related specs

- [combat-damage-mitigation](combat-damage-mitigation.md): the 12-type loop called by DPO-R11, the flat-reduction values of DPO-R08b, `mulDiv`.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md): the to-hit roll and block/avoid whose return bits DPO-R01a maps.
- [combat-leech-and-status-effects](combat-leech-and-status-effects.md): leech and drain effectiveness of DPO-R15 to DPO-R18.
- [combat-status-application](combat-status-application.md): the five appliers of DPO-R19, DPO-R27, DPO-R28.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): `P` selection, hostility and town gates of DPO-R08b and DPO-R13.
- [combat-hit-recovery](combat-hit-recovery.md) and [combat-knockback](combat-knockback.md): the reaction stage of DPO-R26 and the hit-class nibble of DPO-R32.
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md), [combat-crushing-blow](combat-crushing-blow.md), [combat-open-wounds](combat-open-wounds.md): callbacks at the dispatch points of DPO-R23.
- [item-durability-loss](item-durability-loss.md): the loss arithmetic behind the draws of DPO-R24a.
- [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md) and [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): the missile entries into the apply (DPO-R08c, DPO-R22) and the missile-unit damage roll (section 6).
- [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md): the frame event, do dispatch and do table at the top of section 2.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): the lists created by the appliers of DPO-R19a.
- [core-game-data-fields](core-game-data-fields.md): names of stats, states, modes and record columns.

## Appendix A. Provenance

Lord of Destruction Expansion 1.13c, Windows x86. Original D2Game/D2Common/Fog identities and data-profile limits are in the [source manifest](../research/source-manifest.md). The [companion](../research/spec-evidence/combat-damage-pipeline-order.md) records dispatch/source correspondences, compatible readings dated 2026-10-12/13, new targeted evidence, corrections and observation procedures. No DPO gameplay vector is runtime-confirmed.
