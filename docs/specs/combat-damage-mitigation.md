# Damage mitigation

**Description**

When an attack hits, damage reduction and resistances change how much damage gets through.
Absorb can reduce some types of damage further and contribute healing. The result depends on
the damage type, the target, and the difficulty, so the same attack can affect different targets
in different ways.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: DMI-R, DMI-Q, DMI-V, DMI-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-damage-mitigation.md)
Open evidence gaps: [gap file](../plans/gaps/combat-damage-mitigation.md)

## 1. Scope

This spec defines the per-type mitigation stage of a damage record: the regular-monster
classification it uses, flat damage reduction, the pre-mitigation event 11 dispatch point,
the bypass of positive resistance, percent resistance with pierce, difficulty penalty, caps and
floor, and percent and flat absorb with the heal they accumulate. It is stage 5 of
[combat-damage-pipeline-order](combat-damage-pipeline-order.md) (DPO-R11). It starts with a
built damage record and ends with the per-type values rewritten and absorbed amounts added to
`record.absorbHeal`. Out of scope: the scale selection and scaling loop that run inside the same
entry (DPO-R08a; selection in
[combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md) PVP-R02a); the
event dispatcher's own walk and entry lifecycle (DPO-R23, IEP-R01, IEP-R05); the status-length
strip (DPO-R10); the final total (DPO-R12a); application of healing to life (DPO-R16); block
and avoidance ([combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md));
knockback and hit recovery; and the behavior of item-event callbacks
([combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md)).
Mitigation itself draws no random number (DMI-R10); callbacks are excluded from that claim.

## 2. Trigger and ownership

**Owner.** Server combat logic. The original module is D2Game. Its computing functions do not
identify a source folder through an assertion string; the neighboring pipeline functions place
the behavior in the `combat` category, as recorded in the companion.

**Trigger and timing.** For basic attacks and queued hits, mitigation runs during the attacker's
skill action, after building damage and before predicted-lethal calculation and queueing
(DPO-R02 to DPO-R07, DPO-R21, DPO-R22, DPO-R12a). The queued apply path does not mitigate again.
Apply paths that request scaling and mitigation run it between the apply gates and apply
events; missile contact is one such path. Which paths run the shared entry, and how many times
per record, belongs to DPO-R08c/DPO-R08d and PVP-R01a: one direct path runs it twice on the
same record, every other running path once.

Within the shared entry, the order is: regular-monster flags (DMI-R01), flat-reduction setup
(DMI-R02), scale selection and scaling (PVP-R02a, DPO-R08a), event 11 dispatch (DMI-R03),
status-length strip (DPO-R10), bypass flag (DMI-R05), per-type mitigation (DMI-R06), then the
total (DPO-R12a). The skill action's placement within a server frame belongs to
[skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md).

**Client.** This stage sends no message. The later apply stage produces the life change
(DPO-R17); the pipeline and network specs own its transmission.

The original dispatch roots, alternate caller identities, output-write anchors, and calling
conventions are in the [companion](../research/spec-evidence/combat-damage-mitigation.md#dispatch-chain).

## 3. Inputs and data

Operation names below are illustrative. They impose no function signatures or storage layout.
`Stat(unit, id)` means the layer-0 read supplied by NRF-R12b; a null unit or a unit without a
stat list reads 0 there. Its stored-value selection is NRF-R12b and conditional minimum
adjustment is NRF-R12c; both are verified at the getter boundary with existing stat arrays
and loaded definitions. Constructing and updating those arrays is outside this stage.
`HasState(unit, id)` is NRF-R12a's state test, false for a null unit.
Every vector except DMI-V27 supplies the stat values for non-null attacker and
defender units; DMI-V17 separately tests null input to the classification predicate.
Attacker and defender are optional logical unit associations at the isolated mitigation
boundary; the game and record are supplied. Entry eligibility and the surrounding apply
operation's null-unit guards belong to the pipeline, rather than being added here.

| Input | Semantic source | Width and format | Meaning |
|-------|-----------------|------------------|---------|
| `record.values[field]` | built damage record, NRF section 3.2/NRF-R08a and CSA-R01 | i32; damage magnitudes 8.8, lengths frames, drain inputs raw quantities | twelve values in the table below; status/drain interpretation belongs to the cited contracts |
| `record.absorbHeal` | damage record | i32, 8.8 | incoming accumulator, increased or decreased by DMI-R08 |
| `record.reductionScale` | damage record field supplied by the path that built the record; its producer is not owned by any spec yet (DMI-Q05) | i32; 0 = unscaled, 0x400 = 1.0 | scale used only when both reduction and scale are positive; vectors supply it |
| `record.bypassUndead`, `record.bypassDemons`, `record.bypassOther` | record flags set during the build by DPO-R05 (from attacker stats 103 `skill_bypass_undead`, 104 `skill_bypass_demons`, 106 `skill_bypass_beasts`, player or hireling attackers only) | logical flags | consumed by DMI-R05; vectors supply them directly |
| defender stat 34 `normal_damage_reduction` | `Stat(defender, 34)` | i32 whole points | flat physical reduction |
| defender stat 35 `magic_damage_reduction` | `Stat(defender, 35)` | i32 whole points | flat fire/lightning/cold/magic reduction |
| defender resistance, maximum resistance, and absorb stats | type bindings below | i32 percent; flat absorb is whole points | DMI-R07/DMI-R08 |
| attacker pierce stats 333-336 | type bindings below | i32 percent | DMI-R07 step 2 |
| attacker state 47 `sanctuary` | `HasState(attacker, 47)` | logical state | physical resistance exception, DMI-R07 step 6 |
| attacker and defender type/class | unit identity, GDF-R17 | type u32; class i32 | regular-monster and class tests; "monster" is the unit type whose class is a MonStats row; every other type, including a player, is a non-monster |
| `monsterClassCount`, class `lUndead`, `hUndead`, `demon` | MonStats data, GDF-R05/GDF-R18 | count i32; logical flags | bounds check and category tests, DMI-R04 |
| `game.difficulty` | game | u8: 0 Normal, 1 Nightmare, 2 Hell; remaining u8 values follow DMI-R07 | penalty selection; expansion clamps the row index, classic applies no penalty above 2 |
| `game.expansion` | game | logical (original non-zero flag) | expansion data penalty versus classic constants |
| `difficulty.ResistPenalty` | selected difficulty row, GDF-R02 | i32 | expansion penalty: Normal 0, Nightmare -40, Hell -100 (section 3.1) |
| `difficultyRowCount` and prepared difficulty rows | GDF-R02 field preparation; selection in DMI-R07 | positive i32 count; original-profile count 3 | selected row is `min(game.difficulty, difficultyRowCount - 1)`; an unavailable/empty prepared collection is outside the supplied-table boundary |
| defender event list | entries created by IEP-R01 (event id, callback, two callback values, entry type) | logical ordered entries | walked for event 11 by DPO-R23; DMI-R03 |
| attacker, defender, game RNG seeds | NRF-R01 | each `{lo, hi}`, two u32 words | unchanged by mitigation itself |
| stat/state-read context | NRF section 3.8 and NRF-R12a/NRF-R12b/NRF-R12c | optional lists/storage/owners; i32 state and definition counts; ordered stat identities and i32 values; signed 16-bit active array counts; boolean `usesResolvedStats`/`fMin`, i32 `MinAccr`, u8 `ValShift` | supplied getter inputs in DMI-V27; other numeric vectors supply returned values, not array production |
| whole-entry scale-selector conditions | PVP-R02a/PVP-R02b | logical conditions and selected i32 percent `P` | common whole-entry setup supplies the attacker's PVP-R02b flag clear and `P=100`; isolated mitigation does not select `P` |
| supplied event callbacks | IEP-R07 | logical callback behavior; i32 results and two registered i32 values | callback effects, draws and record/list mutations are explicit fixture inputs; DMI-V19 supplies logging callbacks with no writes or draws |

**Derived context.** `context` denotes the game, selected difficulty data, attacker, defender,
record, `regularAttacker`, `regularDefender`, and `reduction[0..2]`. These are logical inputs
and intermediate results. DMI-R01 computes the two regular-monster flags once per entry;
DMI-R02 computes the reductions. Their original temporary layout is evidence only.

**Type bindings and execution order.** `none` means this entry has no stat binding (the
original sentinel is recorded in the companion). These bindings are an engine-defined table,
not an MPQ damage-policy table. Each named stat/id maps to GDF-R11. Field names map to the
foundations' damage-record fields (section 3.2, CSA-R01): physical; fire, lightning, cold,
magic and poison damage; poisonLength = poison length; coldLength and freezeLength = cold and
freeze length; lifeDrain, manaDrain and staminaDrain = the foundations' `lifeLeech`,
`manaBurn` and `staminaBurn` raw inputs. DPO calls `poison` `poisonMagnitude` and
`record.reductionScale` `flatReductionScale`; these are aliases for the same semantic
quantities, with no unit conversion. The unvisited `burnLength` is also a frame count.

| Entry | Field | Resist | Max | Pierce | % absorb | Flat absorb | Reduction slot | Regular attacker required |
|------:|-------|--------|-----|--------|----------|-------------|---------------:|---------------------------|
| 0 | physical | 36 `damageresist` | none | none | none | none | 1 | false |
| 1 | fire | 39 `fireresist` | 40 `maxfireresist` | 333 `passive_fire_pierce` | 142 `item_absorbfire_percent` | 143 `item_absorbfire` | 2 | false |
| 2 | lightning | 41 `lightresist` | 42 `maxlightresist` | 334 `passive_ltng_pierce` | 144 `item_absorblight_percent` | 145 `item_absorblight` | 2 | false |
| 3 | cold | 43 `coldresist` | 44 `maxcoldresist` | 335 `passive_cold_pierce` | 148 `item_absorbcold_percent` | 149 `item_absorbcold` | 2 | false |
| 4 | magic | 37 `magicresist` | 38 `maxmagicresist` | none | 146 `item_absorbmagic_percent` | 147 `item_absorbmagic` | 2 | false |
| 5 | coldLength | 43 `coldresist` | 44 `maxcoldresist` | 335 `passive_cold_pierce` | none | none | 0 | false |
| 6 | freezeLength | 43 `coldresist` | 44 `maxcoldresist` | 335 `passive_cold_pierce` | none | none | 0 | false |
| 7 | poisonLength | 110 `item_poisonlengthresist` | none | 336 `passive_pois_pierce` | none | none | 0 | false |
| 8 | poison | 45 `poisonresist` | 46 `maxpoisonresist` | 336 `passive_pois_pierce` | none | none | 0 | false |
| 9 | lifeDrain | none | none | none | none | none | 0 | true |
| 10 | manaDrain | none | none | none | none | none | 0 | true |
| 11 | staminaDrain | none | none | none | none | none | 0 | true |

### 3.1 Origin and binding

Each table claim below has a table-data record (TD-1 to TD-5) in the
[companion](../research/spec-evidence/combat-damage-mitigation.md#table-data-records), read
through the vanilla 1.13c loader. The records establish archive contents; the code-defined
uses are cited separately.

| Mapping or constant | Origin | Binding in original 1.13c | Table identity / evidence |
|---------------------|--------|--------------------------|---------------------------|
| reduction/flat-absorb minimum metadata | table data | fMin is clear and the getter retains the stored value before DMI converts whole points to 8.8 | ItemStatCost rows 34,35,143,145,147,149, columns fMin/MinAccr/ValShift; [text and compiled evidence](../research/spec-evidence/core-numeric-and-rng-foundations.md#stat-table-verification-2026-10-14) |
| stat names and numeric identities in this spec, including common-vector stats 118 and 153 | table data | named identity supplied by data; mitigation type bindings and build-stat selectors are fixed by code | `data/global/excel/itemstatcost.txt`, row by `Stat` name, column `ID` (GDF-R11); TD-1 |
| state 47 `sanctuary` | table data | identity supplied by data; selector fixed in code | `data/global/excel/states.txt`, row `state=sanctuary`, column `id` (GDF-R12); TD-2 |
| type bindings, entry order, reduction slots, regular-attacker marker | engine code | fixed by original code's internal table | image read and live read-only memory corroboration in the companion; not MPQ policy |
| expansion `ResistPenalty` | table data | resolved from selected difficulty row | `data/global/excel/difficultylevels.txt`, rows `Name` Normal/Nightmare/Hell = zero-based data rows 0/1/2, column `ResistPenalty`: 0, -40, -100 (GDF-R02); compiled table agrees; TD-3 |
| monster class identity and classification | table data | `hcIdx` selects class; code tests data flags | `data/global/excel/monstats.txt`, row by `Id`, column `hcIdx`; category columns `lUndead`, `hUndead`, `demon` (GDF-R05/GDF-R17/GDF-R18); vector class: `sandleaper1` `hcIdx` 78 (0x4E); TD-4 |
| hireling exceptions: 0x10F, 0x152, 0x167, 0x230, 0x231 | engine code for selection; table data for names | original code compares fixed class IDs; it does not select an arbitrary data-defined hireling category | MonStats `Id` `roguehire` 271, `act2hire` 338, `act3hire` 359, `act5hire1` 560, `act5hire2` 561, column `hcIdx`; class 0x232 (562) is `baaltentacle1`; TD-4 |
| event 11 dispatch | engine code | fixed event selector; callback registrations supplied elsewhere | `data/global/excel/events.txt` zero-based data row 11, column `event` = `absorbdamage`; TD-5 |
| classic penalties 20/50; resistance floor -100; caps 50/75/95; absorb cap 40; scale unit 0x400 | engine code | fixed arithmetic constants | DMI-R02, DMI-R07, DMI-R08; source citations in companion |

These are original-profile bindings. They do not require the same internal representation in
another language or establish custom remapping support. Any custom mapping capability is a
separate implementation decision with its own acceptance boundary.

## 4. Rules (execution order)

Notation follows [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md):
i32 two's complement; arithmetic wraps as specified there; division truncates toward zero;
`>>` is arithmetic. All arithmetic values are i32 and comparisons signed unless stated.
`none` is an absent binding, not a numeric stat. `mulDiv(a, b, c)` consumes NRF-R05 exactly;
its argument order matters to precision loss.

**DMI-R01** Regular-monster predicate [verified in 1.13c code (static)].
A unit is a regular monster when it is a non-null monster and its hireling-class category
is 0. For a monster, unsigned class comparisons give category 1 for 0x10F `roguehire`,
2 for 0x152 `act2hire`, 3 for 0x167 `act3hire`, 4 for 0x230 or 0x231 (`act5hire1`/`act5hire2`),
otherwise 0. The category of a null or non-monster unit is 0, but the regular-monster type
test still fails. Thus only those five monster classes are excluded; this predicate does
not exclude champions or bosses. At entry, compute once per shared entry:

```
context.regularAttacker = isRegularMonster(attacker)
context.regularDefender = isRegularMonster(defender)
```

No random draw.

**DMI-R02** Flat-reduction setup [verified in 1.13c code (static)].
Preconditions: the shared entry receives the game, attacker, defender, and record.

```
reduction[0] = 0
red = Stat(defender, 34) << 8
if red > 0 and record.reductionScale > 0:
    red = mulDiv(record.reductionScale, red, 0x400)
reduction[1] = red
red = Stat(defender, 35) << 8
if red > 0 and record.reductionScale > 0:
    red = mulDiv(record.reductionScale, red, 0x400)
reduction[2] = red
```

The shift is 32-bit with no overflow guard. Zero or negative reduction, or zero or negative
scale, leaves the shifted reduction unscaled; a negative reduction is kept, so DMI-R06's
subtraction then raises the value (DMI-V20). Slot 1 serves physical; slot 2 serves fire,
lightning, cold, and magic; slot 0 serves the other entries. No random draw.

**DMI-R03** Pre-mitigation event 11 dispatch [verified in 1.13c code (static)].
After scale selection and scaling, and before the status-length strip, the shared entry
dispatches event 11 (`absorbdamage`) on the defender's event list with the attacker as the
other unit and the record (placement DPO-R09; dispatcher DPO-R23):

```
dispatch(game, event 11, owner = defender, other = attacker, record)   # DPO-R23
# the dispatcher's result is not used; the entry continues with the strip
```

The step runs unconditionally, including when scaling changed nothing. Callbacks therefore
see the scaled record before the strip and before any per-type mitigation. A null defender or
a list without event-11 entries runs no callback. Which entries run, their order, their
arguments and whether an entry remains registered afterwards are the dispatcher contract
(DPO-R23 consumes IEP-R07), with entries created by IEP-R01. Matching callbacks receive the
event ID and traversal reloads the successor after each callback; IEP-V08 exercises the
registration-mutation case in the owning dispatcher contract.
Callbacks may draw and write; those effects belong to their own rules (NRF-R04c, DPO-R23,
IEP). The entry itself draws nothing here.

**DMI-R04** Undead and demon class tests [verified in 1.13c code (static)].
Both tests return false unless the unit is non-null, a monster, and its class id is in
`[0, monsterClassCount)`. For a valid monster class:

```
checkA = class.lUndead or class.hUndead
checkB = class.demon
```

DMI-R05 and DMI-R07 step 6 consume these tests. No random draw.

**DMI-R05** Bypass flag [verified in 1.13c code (static)]. After the strip and before the
loop, compute once per record:

```
b = 0
if defender != null and defender.type == monster:
    A = checkA(defender); B = checkB(defender)     # DMI-R04
    if A != 0 and record.bypassUndead: b = 1
    if B != 0 and record.bypassDemons: b = 1
    if A == 0 and B == 0 and record.bypassOther: b = 1
else:
    if record.bypassOther: b = 1
```

Thus `bypassUndead` bypasses undead, `bypassDemons` bypasses demons, and `bypassOther`
bypasses defenders that are neither, including a null or non-monster defender such as a
player. When `b=1`, DMI-R06 treats positive resistance as 0, preserves negative resistance,
and skips flat reduction and absorb. No random draw.

**DMI-R06** Per-type loop and order [verified in 1.13c code (static)].
Walk the twelve entries in section 3 order. Entries 0-8 always run; entries 9-11 run only for
a regular-monster attacker:

```
for entry in entries:                              # section 3 order
    if entry.regularAttackerRequired and context.regularAttacker == 0:
        break
    v = record.values[entry.field]
    if v <= 0:
        record.values[entry.field] = 0
        continue
    r = effectiveResistance(entry, context)        # DMI-R07; before the bypass test
    if b != 0:
        if r > 0: r = 0
    else:
        v = v - reduction[entry.slot]              # no clamp
    if v > 0 and r != 0:
        if r >= 100: r = 100
        v = mulDiv(100 - r, v, 100)
    if b == 0: v = absorb(v, entry, context)       # DMI-R08, even when v <= 0
    record.values[entry.field] = v
```

Order inside an entry: flat reduction, resistance, absorb. DMI-R07's floor normally limits the
resistance multiplier to 200; the negative-cap edge case in section 6 can exceed that limit.
No random draw.

**DMI-R07** Effective resistance [verified in 1.13c code (static)].
Use the entry's `resistId`, `maxId`, `pierceId`, and `regularDef=context.regularDefender`.

1. `r=Stat(defender,resistId)` when there is a resistance binding; otherwise `r=0`.
   Entries 9-11 start at 0 but still run steps 3-5.
2. When a pierce binding exists and (`r<100` or `regularDef=0`), read
   `p=Stat(attacker,pierceId)` and subtract it when non-zero, with no clamp: a negative
   pierce stat raises `r`. A regular monster at `r>=100` is not pierced; any other defender is
   pierced at any resistance.
3. When `regularDef=0` and `resistId` is neither 36 `damageresist` nor 37 `magicresist`,
   expansion adds the selected row's `difficulty.ResistPenalty` value unchanged. With the
   supplied positive `difficultyRowCount`, select row
   `min(game.difficulty, difficultyRowCount - 1)`; the game's difficulty is u8, so its index
   is nonnegative. GDF-R02 supplies field preparation; the clamping getter's existing
   [source evidence](../research/spec-evidence/combat-damage-mitigation.md#rule-linked-rationale)
   supports this consumed selection within DMI-R07. Classic
   subtracts 20 at difficulty 1, 50 at difficulty 2, and nothing for any other difficulty
   value (0, or a value above 2). This also applies when the
   resistance binding is absent (entries 9-11) and to stat 110 (entry 7), increasing drain
   inputs and poison length under a negative penalty (DMI-V15/DMI-V16). With the 1.13c
   penalties (all at most 0) an absent binding never reaches step 5; if `r` were positive
   there, step 5 uses `cap = 75` (the binding is not stat 36) and step 6 does not apply.
4. If `r<=0`, return `r` when `r>-100`, otherwise -100. Steps 5 and 6 do not run.
5. Only for `regularDef=0`: if maximum binding is absent, `cap=50` for resistance stat 36,
   otherwise 75. If present, `cap=Stat(defender,maxId)+75`, replaced by 95 when the sum is
   `>=95`; no lower bound. Replace `r` by `cap` when `r>=cap`. Regular monsters are not capped;
   DMI-R06's clamp at 100 supplies immunity.
6. Only for resistance stat 36, for any defender, and only when step 4 did not return:
   if `HasState(attacker,47 sanctuary)` and `checkA(defender)`, set `r=0`.
   Negative physical resistance has already returned and is not zeroed.

Return `r`. No random draw.

**DMI-R08** Absorb [verified in 1.13c code (static)].
`pctId` and `flatId` are the type bindings; `heal` denotes `record.absorbHeal`.

```
if pctId == none: return v                       # skip whole step: entries 0, 5-11
pct = Stat(defender, pctId)
if pct >= 40: pct = 40
if pct > 0:
    a = mulDiv(pct, v, 100)
    heal += a; v -= a
f = Stat(defender, flatId) << 8
if f > 0:
    if f >= v: f = v                             # v may be negative
    heal += f; v -= f
return v
```

Percent absorb precedes flat absorb on the post-resistance value. Percent is capped at 40;
flat absorb is converted from whole points by the wrapping i32 shift. Each part is skipped
when its shifted/read stat is non-positive. Healing is only accumulated here; DPO-R16 owns
application to life. No random draw.

**DMI-R09** `mulDiv` argument order [verified in 1.13c code (static)].
Use NRF-R05's exact branches and rounding, consumed here in three orders:
DMI-R02 `mulDiv(a=record.reductionScale,b=red,c=0x400)`;
DMI-R06 `mulDiv(a=100-r,b=v,c=100)`;
DMI-R08 `mulDiv(a=pct,b=v,c=100)`.
Divide-first precision loss is required behavior (DMI-V13), including zero-divisor handling
from NRF-R05. No random draw. The shared arithmetic is not redefined locally.

**DMI-R10** No draws in the stage [verified in 1.13c code (static)].
The shared entry, per-type mitigation, resistance, absorb, scale selector, status strip,
and their read/arithmetic helpers draw nothing from any seed for any input. The sequence
is DMI-R01, DMI-R02, scale selection/scaling (PVP-R02a, DPO-R08a), event 11 dispatch
(DMI-R03), status strip (DPO-R10), DMI-R05, DMI-R06, then total (DPO-R12a). Event callbacks
are excluded from this no-draw claim.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|----|---------|--------------|-------|---------|
| DMI-R01 | Regular monster excludes five hireling classes; flags computed at entry | verified in 1.13c code (static) | none | DMI-V17, DMI-V18, DMI-V23 |
| DMI-R02 | Flat-reduction setup and scale | verified in 1.13c code (static) | none at getter boundary; existing arrays supplied | DMI-V01, DMI-V06, DMI-V20, DMI-V22, DMI-V26, DMI-V27 |
| DMI-R03 | Event 11 dispatch point on the defender's list; result ignored | verified in 1.13c code (static) | none; consumes IEP-R07 | DMI-V19 |
| DMI-R04 | Undead and demon class tests | verified in 1.13c code (static) | none | DMI-V07, DMI-V11, DMI-V18 |
| DMI-R05 | Bypass flag from class tests and record flags | verified in 1.13c code (static) | none | DMI-V07, DMI-V18, DMI-V24 |
| DMI-R06 | Ordered loop: flat, resistance, absorb, early end | verified in 1.13c code (static) | none at getter boundary; existing arrays supplied | DMI-V01, DMI-V05, DMI-V07, DMI-V08, DMI-V15, DMI-V20, DMI-V21, DMI-V22, DMI-V24, DMI-V26, DMI-V27 |
| DMI-R07 | Pierce, penalty, floor, caps, sanctuary | verified in 1.13c code (static) | none at getter boundary; existing arrays supplied | DMI-V02, DMI-V03, DMI-V04, DMI-V09, DMI-V10, DMI-V11, DMI-V15, DMI-V16, DMI-V21, DMI-V23, DMI-V25, DMI-V26, DMI-V27 |
| DMI-R08 | Percent then flat absorb | verified in 1.13c code (static) | none at getter boundary; existing arrays supplied | DMI-V05, DMI-V12, DMI-V22, DMI-V26, DMI-V27 |
| DMI-R09 | Three mulDiv argument orders, NRF-R05 | verified in 1.13c code (static) | none | DMI-V13, DMI-V26 |
| DMI-R10 | No draws; callbacks excluded | verified in 1.13c code (static) | none | DMI-V14 |

## 5. Outputs and side effects

| Write or effect | Rule | Detail |
|-----------------|------|--------|
| `context.regularAttacker`, `context.regularDefender` | DMI-R01 | computed classification flags; temporary logical values |
| `reduction[0..2]` | DMI-R02 | derived flat reductions; temporary logical values |
| event 11 dispatched on the defender's event list | DMI-R03 | synchronous before the strip and loop; callback effects and entry lifecycle belong to DPO-R23 and the callbacks' rules |
| physical, fire, lightning, cold, magic, coldLength, freezeLength, poisonLength, poison | DMI-R06 | rewritten per entry; input at or below 0 becomes 0 |
| lifeDrain, manaDrain, staminaDrain | DMI-R06 | rewritten only for a regular attacker; otherwise untouched |
| `record.absorbHeal` | DMI-R08 | percent and flat absorb contributions added across entries 1-4 |
| attacker, defender, game RNG seeds | DMI-R10 | unchanged by this stage; callbacks excluded |
| event 11 callback calls and event-list retention | DPO-R23, IEP-R07 (consumed) | call order, callback arguments and whether entries stay registered; asserted in DMI-V19 as integration assertions |
| `record.total` | DPO-R12a (consumed, not redefined here) | `record.total = physical + fire + lightning + magic + cold + poison` (i32 adds), plus `lifeDrain` when `context.regularAttacker = 1` |
| stat writes, states, scheduled events, units created/destroyed, messages | none directly | mitigation itself performs none; callbacks own their effects; DPO-R16 applies healing later |

**Rule-level return values.** DMI-R01 returns the hireling category and regular-monster
predicate; DMI-R04 returns the two category tests; DMI-R07 returns effective resistance;
DMI-R08 returns the post-absorb value; DMI-R06 stores the resulting field. DMI-R09 consumes
NRF-R05's arithmetic return. These are semantic intermediate results available to vectors,
not required APIs. No shared-entry return value is part of this spec's acceptance.

## 6. Branches and edge cases

- **Unit types.** Regular-monster defenders are not pierced at resistance at least 100,
  penalized, or capped; resistance at least 100 yields immunity. Players, hirelings, and other
  defenders are pierced at any resistance, penalized except physical/magic, and capped.
  Only regular-monster attackers have their drain inputs mitigated (DMI-R01, DMI-R06).
- **Null units.** A null attacker is not regular and has no sanctuary state; its pierce reads
  follow NRF-R12b. A null defender is not regular; bypass depends only on
  `bypassOther`; its stat reads follow NRF-R12b; the event dispatch runs no callback.
- **Difficulty.** Expansion adds the selected data penalty (0/-40/-100) except physical/magic;
  classic subtracts 20/50. Poison length and absent-resistance drain entries can therefore
  increase up to twice their input on Hell (DMI-V15/DMI-V16).
- **Melee/missile.** Same mitigation; path timing and count are DPO-R08c/DPO-R08d.
- **PvP.** Scaling before this loop belongs to PVP-R02a/DPO-R08a; mitigation does not test PvP.
- **Negative resistance.** Survives bypass; floor -100. Sanctuary does not zero negative
  physical resistance (DMI-V11).
- **Negative values after flat reduction.** No post-subtraction clamp. A negative physical
  field can lower the total (DMI-V08; runtime cross-check DMI-Q03). With absorb bindings,
  positive percent absorb can contribute negative healing; positive flat absorb is then
  clamped to the negative remaining value. Example (DMI-V22): `v=-512`, percent 20, flat 3 gives
  `a=-102`, heal -102, `v=-410`, then `f=-410`, heal -512, field 0. Without absorb bindings,
  the negative value remains.
- **Negative reduction stats.** A negative defender stat 34 or 35 gives a negative reduction,
  never scaled; subtracting it raises the field (DMI-V20).
- **Zero inputs.** A non-positive field is stored as 0 before other processing. Zero
  resistance skips percent mitigation, zero pierce subtracts nothing, zero max-resist gives
  cap 75 when a maximum binding exists.
- **Overflow.** Stat shifts and `100-r` are i32. NRF-R05 defines mulDiv's product handling.
  A sufficiently negative max-resist can produce a negative cap in step 5; it is applied
  after step 4, so the resistance floor is not reapplied to that negative cap (DMI-V21).
- **Missing data.** Difficulty selection clamps to the supplied positive row range under DMI-R07;
  with three rows an index above 2 selects the last row. Classic subtracts nothing for a
  difficulty above 2. Out-of-range monster class IDs make checks A/B false; the hireling
  category still compares IDs directly (DMI-V18 (e)).
- **Runs per record.** One direct path runs the entry twice; queued apply never repeats
  mitigation. DPO-R08c/DPO-R08d and PVP-R01a own these distinctions.
- **RNG.** No draws except callbacks outside DMI-R10's claim.

## 7. Open questions

| Id | Status | Question and evidence that would settle it |
|----|--------|-------------------------------------------|
| **DMI-Q01** | settled | NRF-R12b/NRF-R12c are verified for reads from existing stat arrays and loaded definitions. Original-profile reduction and flat-absorb definitions have `fMin` clear and return stored values unchanged before DMI's whole-point-to-8.8 conversion. DMI-V27 exercises getter-to-mitigation integration; it does not exercise production of the arrays. |
| **DMI-Q05** | open (outside this boundary) | Which rule produces `record.reductionScale` on each path that builds a damage record. The build steps DPO-R02 to DPO-R07 do not mention it and no sibling spec owns it. DMI-R02 consumes it as a supplied input, so it gates no rule here. Settled when the spec that owns each record builder names the writer. |
| **DMI-Q03** | unresolved (runtime cross-check) | Does a negative physical field survive into the total as the static reading says (DMI-V08)? DMI-O03 measures field and total after mitigation when flat reduction exceeds damage. |
| **DMI-Q04** | unresolved (runtime cross-check) | Confirm penalty-induced growth of absent-resistance drain entries and poison length (DMI-V15/DMI-V16). DMI-O02 measures both sides of the stage on Hell. |

No rule carries not found in d2-decomp. Gameplay observations are possible and pending. The
gaps that remain after static and table-data attempts are listed in the
[gap file](../plans/gaps/combat-damage-mitigation.md).

## 8. Test vectors

Common setup unless overridden: raw 8.8 quantities; `record.reductionScale=0`,
all three bypass flags false, `record.absorbHeal=0`, all unnamed fields 0;
`game.expansion=0` (classic), `game.difficulty=0` (Normal), `difficultyRowCount=3` with
the named Normal/Nightmare/Hell rows in section 3.1; all unnamed stats 0, including
defender 118 `item_halffreezeduration` and 153 `item_cannotbefrozen`. No states except an
explicitly named attacker sanctuary state, so DPO-R10 leaves cold/freeze/poison lengths and
the burn length (CSA-R01; not read by this stage) unchanged. The defender's event list is empty.

Default attacker: regular monster class 0x4E `sandleaper1` (GDF-R17); defender: player.
"Monster defender" means a monster whose class id `c` satisfies `0 <= c < monsterClassCount`
and is not one of DMI-R01's five hireling classes (so it is a regular monster), with lUndead,
hUndead, and demon clear unless a vector supplies other class flags; the attacker is a player.
`monsterClassCount` and the class flags are supplied inputs. Unless a vector says otherwise,
the monster defender's class id is 0x4E with `monsterClassCount = 0x200`, and its class flags
are the vector's stated flags; these are mitigation-boundary test inputs, not
`sandleaper1`'s data or the 1.13c MonStats count, which this spec does not assert. Every
vector names its pairing.

Whole-entry scale selection is settled by PVP-R02a for exactly two pairings:
- default pairing (regular-monster attacker, player defender): `P=100` when the attacker is not
  a hireling class and its PVP-R02b flag is clear; this setup supplies that flag clear;
- monster-defender pairing (player attacker, regular-monster defender): `P=100`, because a
  player is neither a hireling class nor passes the MonStats bit-7 test.
In both, the scaling loop changes no field. Any other pairing in a vector (DMI-V15's
player-attacker variant) is evaluated at the mitigation boundary only; through the whole entry
PVP-R02a would select another `P` and DPO-R08a would scale first.

Numeric vectors can be executed at the mitigation boundary: use the stated stat reads,
class flags, and difficulty data as supplied inputs, compute DMI-R01/DMI-R02 intermediates,
then DMI-R05/DMI-R06. A whole-entry execution under the setup additionally consumes PVP-R02a,
DPO-R08a, DMI-R03 with an empty list, DPO-R10, and the stat reads of NRF-R12b/NRF-R12c.
Vectors V01-V26 supply the values those reads return; V27 supplies the arrays and definitions.
Each vector uses **no draw** (DMI-R10);
DMI-V14 explicitly supplies unchanged seeds.
A vector that does not list `record.total` or `record.absorbHeal` among its outputs does not
assert them. Vectors may also assert named intermediate results (`mulDiv` returns, the hireling
category, `b`, `r`, `cap`) as rule-level checks; section 5 lists only writes.
In separate lettered cases, start a fresh record using the common setup, the vector's shared
inputs, and that case's overrides. "Same with" retains the named inputs of the referenced
case with only the stated change. Do not carry mitigation outputs into the next case.
In vectors, `physical`, `fire`, and the other type-field names abbreviate
`record.values[field]`; `heal` abbreviates `record.absorbHeal`; single-argument `Stat(id)`
uses the attacker or defender named in that clause. All class flags are supplied logical
values for the isolated numeric checks. IDs and names map through section 3's bindings.

### 8.1 Executable vectors

- **DMI-V01** (DMI-R02, DMI-R06, DMI-R07) physical order. Inputs: `physical = 5120`;
  defender `Stat(34) = 5`, `Stat(36) = 25`. Trace: `reduction[1] = 5 << 8 = 1280` (unscaled,
  `record.reductionScale = 0`); `b = 0` (player defender, bypassOther false). Entry 0: `v = 5120 > 0`; `r = 25`; no
  pierce id; stat 36 so no penalty; `r > 0`, not regular: `maxId = none`, stat 36 so `cap = 50`,
  `r = 25`; sanctuary step: no state 47, `r = 25`. `v = 5120 - 1280 = 3840`; `r != 0`, `< 100`:
  `v = mulDiv(75, 3840, 100) = 2880`; absorb: `pctId = none`, skipped. Entries 1-8: `v = 0`,
  stored 0. Entries 9-11 run (regular-monster attacker): `v = 0`, stored 0. Outputs:
  `physical = 2880` (11.25), `record.absorbHeal = 0`, `record.total = 2880` (DPO-R12a). `runtime_observed: no`.
- **DMI-V02** (DMI-R07 step 4) negative resistance. Monster defender; `fire =
  2560`; defender `Stat(39) = -50`; attacker `Stat(333) = 0`. Trace: `b = 0` (A = B = 0,
  bypassOther false). Entry 1: `r = -50`; pierce: `r < 100`, `p = 0`, nothing subtracted; regular defender:
  no penalty; `r <= 0` and `> -100`: return -50. `v = 2560 - reduction[2] (0) = 2560`;
  `v = mulDiv(150, 2560, 100) = 3840`; absorb: `Stat(142) = 0` (`<= 0`, skipped), `Stat(143) = 0`
  (skipped). Output `fire = 3840`. Same with `Stat(39) = -150`: step 4 returns -100;
  `v = mulDiv(200, 2560, 100) = 5120`. `runtime_observed: no`.
- **DMI-V03** (DMI-R07 step 5) caps. Player defender; `fire = 2560`. (a)
  `Stat(39) = 95`, `Stat(40) = 0`: `r = 95`; no pierce (`Stat(333) = 0`); classic Normal: no
  penalty; `cap = 0 + 75 = 75 < 95`, `r = 75`; `v = mulDiv(25, 2560, 100) = 640`. (b)
  `Stat(39) = 95`, `Stat(40) = 20`: `cap = 95` (sum 95 is `>= 95`), `r = 95`, `v = mulDiv(5, 2560, 100) = 128`.
  (c) `physical = 5120`, `Stat(36) = 90`, `Stat(34) = 0`: `cap = 50`, `r = 50`,
  `v = mulDiv(50, 5120, 100) = 2560`. Outputs: `fire = 640` / `128`; `physical = 2560`; heal 0.
  (d) Fresh common setup, player defender; `fire=2560`, `Stat(39)=120`, `Stat(40)=40`:
  `75+40=115` is capped to 95; `r=95`, so `fire=mulDiv(5,2560,100)=128`, heal 0.
  Omitting the 95 ceiling would give `r=115`, then DMI-R06's 100 clamp and fire 0.
  `runtime_observed: no`.
- **DMI-V04** (DMI-R07 step 2) pierce against a regular monster. Monster defender;
  `fire = 2560`; attacker `Stat(333) = 10`. (a) `Stat(39) = 100`: `r = 100`; `r >= 100` and
  regular defender: pierce skipped; no penalty; `r > 0`, regular: no cap; not stat 36: return
  100. DMI-R06: `v = 2560`; `r >= 100` so `r = 100`; `v = mulDiv(0, 2560, 100) = 0`; absorb:
  both stats 0, skipped. Output `fire = 0` (immune). (b) `Stat(39) = 99`: `r < 100`, pierced:
  `r = 89`; return 89; `v = mulDiv(11, 2560, 100) = 281` (28160 / 100 truncated). Output
  `fire = 281`. `runtime_observed: no`.
- **DMI-V05** (DMI-R06, DMI-R08) absorb order. Player defender; `cold = 2560`;
  `Stat(43) = 0`, `Stat(44) = 0`, `Stat(35) = 0`, `Stat(148) = 60`, `Stat(149) = 3`. Trace:
  entry 3: `r = 0`; pierce 0; classic Normal: no penalty; `r <= 0`, `> -100`: return 0.
  `v = 2560 - 0`; `r == 0`: no percent step. Absorb: `pct = 60 >= 40`, so `pct = 40`;
  `a = mulDiv(40, 2560, 100) = 1024`; heal 1024; `v = 1536`; `f = 3 << 8 = 768 > 0`;
  `768 < 1536` so `f = 768`; heal 1792; `v = 768`. Outputs: `cold = 768` (3.0),
  `record.absorbHeal = 1792` (7.0).
  Fresh cases use the common setup and player defender, `cold=2560`, with unnamed stats 0.
  (b) `Stat(148)=-20`, `Stat(149)=3`: negative percent is skipped; flat is 768;
  outputs cold 1792, heal 768. Applying the negative percent would instead give cold 2304,
  heal 256. (c) `Stat(148)=40`, `Stat(149)=-3`: percent contributes 1024, cold becomes 1536;
  shifted flat is -768 and skipped. Outputs cold 1536, heal 1024. Applying the negative
  flat would instead give cold 2304, heal 256. `runtime_observed: no`.
- **DMI-V06** (DMI-R02) scaled flat reduction. Player defender; `physical = 5120`; `Stat(36) = 0`.
  (a) `Stat(34) = 5`, `record.reductionScale = 0x200`: `red = 1280 > 0`, scale `> 0`:
  `reduction[1] = mulDiv(0x200, 1280, 0x400)`: `b = 1280 <= 0x100000`, `a = 512 <= 0x10000`,
  so `(512 * 1280) / 1024 = 640`. Entry 0: `r = 0`; `v = 5120 - 640 = 4480`; `r == 0`: no
  percent; no absorb id. Output `physical = 4480`. (b) `Stat(34) = 3`, `record.reductionScale = 0x155`:
  `reduction[1] = (341 * 768) / 1024 = 261888 / 1024 = 255` (255.75 truncated); output
  `physical = 4865`. (c) `Stat(34) = 3`, `record.reductionScale = 0`: `reduction[1] = 768`; output `4352`. (d)
  `Stat(35) = 2`, `record.reductionScale = 0x200`: `reduction[2] = (512 * 512) / 1024 = 256`; with `Stat(34) = 0`,
  `physical = 0`, `fire = 2560` and `Stat(39) = 0`: entry 1 gives `r = 0`,
  `v = 2560 - 256 = 2304`, no percent step, absorb stats 0 skipped; output `fire = 2304`
  (slot 2 serves entries 1-4 only).
  (e) Fresh common setup, player defender; `physical=5120`, `Stat(34)=5`,
  `record.reductionScale=-512`: shifted reduction 1280 is positive but scale is negative,
  so reduction stays 1280. Outputs physical 3840, heal 0. Scaling whenever the scale is
  merely nonzero would give reduction -640 and physical 5760. `runtime_observed: no`.
- **DMI-V07** (DMI-R04, DMI-R05, DMI-R06) bypass. Defender: monster defender (preamble) whose class has
  `lUndead` set and `demon` clear (check A = 1, check B = 0); attacker player with
  `Stat(333) = 0`; `fire = 2560`; defender `Stat(39) = 50`, `Stat(35) = 2` (so
  `reduction[2] = 512`), `Stat(142) = 20`, `Stat(143) = 0`. (a) `bypassUndead=true`: A and bypassUndead: `b =
  1`. Entry 1: `r = 50` (no pierce, no penalty, no cap: regular); DMI-R06 with `b = 1`: `r > 0`
  so `r = 0`; no flat; `r == 0`: no percent; no absorb. Output `fire = 2560`, heal 0. (b)
  `bypassDemons=true`: B = 0 so the second test fails; A != 0 so the "neither" test is not reached: `b = 0`.
  `v = 2560 - 512 = 2048`; `v = mulDiv(50, 2048, 100) = 1024`; absorb: `pct = 20`,
  `a = mulDiv(20, 1024, 100) = 204`, heal 204, `v = 820`; flat 0 skipped. Output `fire = 820`,
  `record.absorbHeal = 204`. (c) `bypassUndead=true`, `Stat(39) = -25`: `b = 1`; `r = -25` kept (not `> 0`);
  `v = mulDiv(125, 2560, 100) = 3200`; absorb skipped. Output `fire = 3200`, heal 0. (d)
  `bypassOther=true` on this defender: A != 0, so `b = 0`: same outputs as (b). `runtime_observed: no`.
- **DMI-V08** (DMI-R06) unclamped flat reduction. Player defender; `physical = 256`; `Stat(34) = 3`,
  `Stat(36) = 0`. Trace: `reduction[1] = 768`; entry 0: `v = 256 > 0`; `r = 0`; `b = 0`:
  `v = 256 - 768 = -512`; `v > 0` false: no percent step; absorb: `pctId = none`, returns -512.
  Output `physical = -512`, `record.absorbHeal = 0`, `record.total = -512` (so DPO-R17 subtracts nothing). Static
  reading; runtime check DMI-Q03.
  (b) Fresh common setup, player defender; `physical=-256`, `Stat(34)=3`, `Stat(36)=0`:
  the field is initially nonpositive, so DMI-R06 stores 0 before reduction or resistance.
  Outputs physical 0, heal 0, total 0 (DPO-R12a). Preserving the initial negative field
  would fail the field and total assertions. `runtime_observed: no`.
- **DMI-V09** (DMI-R07 step 3, classic) difficulty penalty. Player defender; `fire = 2560`;
  `Stat(39) = 30`, `Stat(40) = 0`; `game.expansion = 0`. (a) `game.difficulty = 1`: `r = 30 - 20 = 10`;
  `cap = 75`, `r = 10`; `v = mulDiv(90, 2560, 100) = 2304`. (b) `game.difficulty = 2`:
  `r = 30 - 50 = -20`; step 4 returns -20; `v = mulDiv(120, 2560, 100) = 3072`. (c)
  `game.difficulty = 2`, `physical = 5120`, `Stat(36) = 30`: stat 36, no penalty; `cap = 50`,
  `r = 30`; `v = mulDiv(70, 5120, 100) = 3584`. Outputs: `fire = 2304` / `3072`;
  `physical = 3584`. `runtime_observed: no`.
- **DMI-V10** (DMI-R07 steps 2 and 3, expansion) penalty and non-monster pierce. (a)
  `game.expansion = 1`, `game.difficulty = 2`, `DifficultyLevels.ResistPenalty` (Hell, row 2)
  = -100 (section 3.1); player defender `Stat(39) = 30`; `fire = 2560`: `r = 30 + (-100) = -70`;
  step 4 returns -70; `v = mulDiv(170, 2560, 100) = 4352`. (b) classic Normal; `fire = 2560`; player defender
  `Stat(39) = 120`, `Stat(40) = 0`; attacker `Stat(333) = 10`: `r = 120`; `r >= 100` but the
  defender is not regular: pierced, `r = 110`; no penalty; `cap = 75`, `r = 75`;
  `v = mulDiv(25, 2560, 100) = 640`. Outputs: `fire = 4352` / `640`.
  Fresh cases start from the common setup, player defender, `fire=2560`, `Stat(39)=30`,
  `Stat(40)=0`, with unnamed stats 0. (c) `Stat(attacker,333)=-10`: subtracting negative
  pierce raises `r` to 40; output fire `mulDiv(60,2560,100)=1536`, heal 0. Ignoring or
  clamping negative pierce would leave `r=30` and fire 1792.
  (d) `game.expansion=1`, `game.difficulty=3`, supplied three named difficulty rows:
  selection clamps to Hell row 2, penalty -100; `r=-70`, output fire 4352, heal 0.
  (e) `game.expansion=0`, `game.difficulty=3`: classic applies no penalty above 2;
  `r=30`, output fire `mulDiv(70,2560,100)=1792`, heal 0. Treating this classic case as
  Hell would instead produce 3072. `runtime_observed: no`.
- **DMI-V11** (DMI-R04, DMI-R07 step 6) sanctuary. Defender: monster defender (preamble) whose class has `lUndead`
  set; attacker player; `physical = 5120`; defender `Stat(36) = 40`, `Stat(34) = 0`. (a) attacker
  has state 47: `r = 40`; no pierce id; regular: no penalty, no cap; stat 36, `HasState = 1`,
  check A = 1: `r = 0`. `v = 5120`; `r == 0`: no percent. Output `physical = 5120`. (b) no state
  47: `r = 40`; `v = mulDiv(60, 5120, 100) = 3072`. (c) state 47 and `Stat(36) = -20`: step 4
  returns -20 before step 6; `v = mulDiv(120, 5120, 100) = 6144`. (d) attacker has state 47,
  `Stat(36) = 40`, but the defender's class has `lUndead` and `hUndead` clear: `r = 40`; step
  6: `HasState = 1` but check A = 0, so `r` stays 40; `v = mulDiv(60, 5120, 100) = 3072`.
  Outputs: `physical = 5120` / `3072` / `6144` / `3072`. `runtime_observed: no`.
- **DMI-V12** (DMI-R08) flat absorb larger than the value. Player defender; `cold = 512`;
  `Stat(43) = 0`, `Stat(148) = 0`, `Stat(149) = 5`. Trace: `r = 0`; `v = 512`; absorb: `pct =
  0`, `<= 0`, skipped; `f = 1280 > 0`; `1280 >= 512` so `f = 512`; heal 512; `v = 0`. Outputs:
  `cold = 0`, `record.absorbHeal = 512`. `runtime_observed: no`.
- **DMI-V13** (DMI-R09) `mulDiv` branches. `mulDiv(75, 5120, 100) = 3840` (i32 branch).
  `mulDiv(50, 0x200000, 100)`: `b > 0x100000`, `100 > 0x20000` false, divide first:
  `(2097152 / 100) * 50 = 20971 * 50 = 1048550` (exact 1048576). `mulDiv(0x20000, 300, 7)`:
  `b <= 0x100000`, `a > 0x10000`, `7 > 0x2000` false: `(131072 / 7) * 300 = 18724 * 300 =
  5617200`. `mulDiv(5, 9, 0) = 0`. `runtime_observed: no`.
- **DMI-V14** (DMI-R10) draw count. Attacker seed `{lo = 1, hi = 0x29A}`, defender seed
  `{lo = 5, hi = 0x29A}`, game seed `{lo = 7, hi = 0x29A}`, defender event list empty; run
  DMI-V01 through the shared entry under the common setup (the scaling loop, the event
  dispatch and the strip then change no field, so the record outputs are those of DMI-V01).
  Output: all three seeds unchanged (0 steps), checked by comparing each seed before and after
  the entry or by a draw counter that reads 0. The same check applies to DMI-V02 to DMI-V12,
  DMI-V15, DMI-V16 and DMI-V18 to DMI-V25 run through the entry (DMI-V15's player-attacker
  variant at the mitigation boundary only); DMI-V13 and DMI-V17 call no stage.
  `runtime_observed: no`.
- **DMI-V15** (DMI-R06, DMI-R07 step 3) drain inputs under the penalty. Attacker regular monster;
  player defender; `game.expansion = 1`, `game.difficulty = 2`, `ResistPenalty` row 2 = -100;
  `lifeDrain = 1000`, all damage fields 0. Trace: entries 0-8: `v = 0`, stored 0. Entry 9:
  regular attacker required and `context.regularAttacker = 1`: runs. `v = 1000`; `r`: binding absent so 0; no pierce;
  defender not regular and resistance binding neither 36 nor 37: `r = 0 + (-100) = -100`; step 4 returns -100.
  `b = 0`; slot 0: `v -= 0`; `v = mulDiv(200, 1000, 100) = 2000`; absorb: `pctId = none`,
  skipped. Entries 10, 11: `v = 0`. Outputs: `lifeDrain = 2000`, `record.total = 0 + 2000 = 2000`
  (DPO-R12a adds `lifeDrain` for a regular-monster attacker). Variants: `game.expansion = 1`,
  `game.difficulty = 1` (`ResistPenalty` -40): `lifeDrain = mulDiv(140, 1000, 100) = 1400`,
  `record.total = 1400`; classic (`game.expansion = 0`) Hell: `lifeDrain = mulDiv(150, 1000,
  100) = 1500`, `record.total = 1500`; Normal: `lifeDrain = 1000`, `record.total = 1000`. Same
  record with a **player** attacker (`context.regularAttacker = 0`), mitigation boundary only
  (preamble): the loop ends at entry 9, `lifeDrain = 1000` untouched and not in the total,
  `record.total = 0`. The supplied
  `ResistPenalty` value is added as given. `runtime_observed: no`.
- **DMI-V16** (DMI-R07 steps 1, 3, 5) poison length under the penalty. Player defender;
  `poisonLength = 500`; attacker `Stat(336) = 0`. (a) `game.expansion = 1`, `game.difficulty = 2`,
  `ResistPenalty` -100, `Stat(110) = 0`: entry 7: `r = 0 - 100 = -100`;
  `v = mulDiv(200, 500, 100) = 1000`; no absorb id. Output `poisonLength = 1000`. (b) same with
  `Stat(110) = 75`: `r = -25`; `v = mulDiv(125, 500, 100) = 625`. (c) classic Normal,
  `Stat(110) = 75`: `r = 75`; `maxId = none`, stat 110 so `cap = 75`; `r = 75`;
  `v = mulDiv(25, 500, 100) = 125`. The supplied `ResistPenalty` value is added as given.
  `record.total` is not asserted (poisonLength is not a total term). `runtime_observed: no`.
- **DMI-V17** (DMI-R01) regular-monster predicate. Monster unit, class 0x152: hirelingCategory
  returns 2, not regular (regular flag 0). Monster, class 0x4E: returns 0, regular (flag 1).
  Monster, class 0x231: returns 4, not regular. Monster, class 0x232: returns 0, regular. Player (any
  class field): the type test fails, flag 0. Null unit: flag 0. The shared category operation's
  complete five-class mapping is exercised by CSA-V04(c), including classes 359 and 560;
  DMI-R01 combines its returned category with the monster-type test. `runtime_observed: no`.
- **DMI-V18** (DMI-R01, DMI-R04, DMI-R05) bypass by class flags. Monster defenders (preamble; regular) with
  class flags (a) `hUndead` only, (b) `demon` only, (c) `lUndead` and `demon`, (d)
  none; attacker player; `fire = 2560`, defender `Stat(39) = 50`, nothing else. `bypassUndead=true`:
  `b` = 1, 0, 1, 0; `fire` = 2560, 1280, 2560, 1280. `bypassDemons=true`: `b` = 0, 1, 1, 0;
  `fire` = 1280, 2560, 2560, 1280. `bypassOther=true`: `b` = 0, 0, 0, 1; `fire` = 1280, 1280,
  1280, 2560. All three bypass flags true: `b = 1` for all four; `fire = 2560`. (Unbypassed:
  `r = 50`, `v = mulDiv(50, 2560, 100) = 1280`.) (e) a monster defender with class id 0x200 and
  `monsterClassCount = 0x200` (out of range; 0x200 is not a hireling class), whose class
  flags are ignored (the range test fails before any flag is read): checks A and B are false and the defender is still regular (DMI-R01 compares ids
  directly). `bypassUndead=true`: `b = 0`, `fire = 1280`; `bypassOther=true`: `b = 1`,
  `fire = 2560`.
  (f) Fresh inputs as in (e), except class id -1. Both category tests return false before
  reading any class flags. This non-hireling monster is still regular. With only
  bypassUndead true: b 0, fire 1280; with only bypassOther true: b 1, fire 2560.
  `runtime_observed: no`.
- **DMI-V19** (DMI-R03; dispatcher behavior consumed from DPO-R23 and IEP-R07) event 11
  dispatch point. Execution: shared-entry integration, with dispatcher-specific assertions
  supplied by the verified IEP-R07 contract. Inputs: DMI-V01's record and stats under the common setup (default pairing),
  except `coldLength = 40` and defender `Stat(153 item_cannotbefrozen) = 1`, and the
  defender's event list holds, head to tail, three entries, each of
  entry type 2 (not one-shot: IEP-R07 removes only type-0 entries) with no removal mark and an arbitrary
  owner key, unused because nothing removes entries in this vector: X with event 11,
  callback `cbX`, callback values `(7, 0)`; Y with event 5 `domeleedamage`, callback `cbY`,
  callback values `(3, 0)`; Z with event 11, callback `cbZ`, callback values `(9, 0)`. Each
  callback records its arguments and the record's `physical` and `coldLength` values at
  call time, writes nothing, changes no event list, draws nothing, and returns 5 (`cbX`)
  or 7 (`cbZ`). No draw.
  Trace: DMI-R01, DMI-R02 as DMI-V01; `P=100`, so scaling changes nothing; DMI-R03 dispatches
  event 11 with owner = defender, other = attacker: `cbX` runs, then `cbZ` (head-to-tail
  order); `cbY` does not run. Both observe `physical = 5120` and `coldLength = 40`.
  The dispatcher's final result is 7 (the last matching callback's result), and is ignored.
  DPO-R10 then sets `coldLength = 0`; DMI-R05 and DMI-R06 continue as in DMI-V01, with the
  cold-length entry remaining 0. Outputs: call log `[cbX(game, event 11, owner defender, other attacker, record,
  7, 0), cbZ(game, event 11, owner defender, other attacker, record, 9, 0)]`, where `record`
  is the same record object the entry is mitigating (passed by reference, not a copy);
  each callback observation `(physical = 5120, coldLength = 40)`;
  `physical = 2880`, `coldLength = 0`, `record.absorbHeal = 0`, `record.total = 2880`; list unchanged (X, Y, Z;
  no entry is one-shot or marked); all seeds unchanged. The log form is this spec's own
  illustrative notation. Assertions on call order, arguments and entry retention belong to
  IEP-R07, consumed by DPO-R23, and are static integration assertions. The placement before
  the strip and loop, the use of the defender's list with the attacker as other unit, and the
  unused result belong to DMI-R03. No callback changes the list in this vector; the
  successor-reload case is exercised by IEP-V08 in the owning contract.
  Discriminating assertions: dispatch after the strip would observe `coldLength = 0`;
  dispatch after mitigation would observe `physical = 2880`; stopping on the nonzero final
  dispatch result would leave `physical = 5120` and `coldLength = 40`. Local placement and
  continuation may also be checked with a supplied dispatcher that records the same call-time
  fields and returns 7; that boundary check does not verify the real dispatcher walk or arguments.
  `runtime_observed: no`.
- **DMI-V20** (DMI-R02, DMI-R06) negative flat reduction. Player defender; `physical = 5120`;
  `Stat(34) = -2`, `Stat(36) = 0`; `record.reductionScale = 0x200`. Trace: `red = -2 << 8 =
  -512`; `red > 0` false, so unscaled: `reduction[1] = -512`; `b = 0`. Entry 0: `v = 5120 > 0`;
  `r = 0`; `v = 5120 - (-512) = 5632`; `r == 0`: no percent step; absorb: `pctId = none`.
  Output `physical = 5632`, `record.absorbHeal = 0`. `runtime_observed: no`.
- **DMI-V21** (DMI-R06, DMI-R07 step 5) negative cap. Player defender; `fire = 2560`;
  `Stat(39) = 50`; attacker `Stat(333) = 0`; classic Normal. (a) `Stat(40) = -100`: `r = 50`;
  no pierce; no penalty; `r > 0`; `cap = -100 + 75 = -25` (below 95, kept); `50 >= -25` so
  `r = -25`; not stat 36: return -25. DMI-R06: `v = 2560 - 0`; `r != 0`, `< 100`:
  `v = mulDiv(125, 2560, 100) = 3200`. (b) `Stat(40) = -200`: `cap = -125`, `r = -125` (no
  floor after step 4); `v = mulDiv(225, 2560, 100) = 5760`. Outputs: `fire = 3200` / `5760`,
  heal 0. `runtime_observed: no`.
- **DMI-V22** (DMI-R02, DMI-R06, DMI-R08) negative value through absorb. Player defender;
  `fire = 256`; `Stat(35) = 3`, `Stat(39) = 0`, `Stat(142) = 20`, `Stat(143) = 3`. Trace:
  `reduction[2] = 768`; `b = 0`. Entry 1: `v = 256 > 0`; `r = 0`; `v = 256 - 768 = -512`;
  `v > 0` false: no percent step. Absorb: `pct = 20 > 0`; `a = mulDiv(20, -512, 100)`:
  `b = -512 > 0x100000` false and `a = 20 > 0x10000` false (signed compares), so the i32
  branch `(20 * -512) / 100 = -10240 / 100 = -102` (truncated toward zero); `heal = -102`;
  `v = -512 - (-102) = -410`; `f = 3 << 8 = 768 > 0`; `768 >= -410` so `f = -410`;
  `heal = -102 + (-410) = -512`; `v = -410 - (-410) = 0`. Outputs: `fire = 0`,
  `record.absorbHeal = -512`. `runtime_observed: no`.
- **DMI-V23** (DMI-R01, DMI-R07 steps 2, 3, 5) hireling-class defender. Attacker player with
  `Stat(333) = 10` unless stated; defender a monster of class 0x10F `roguehire`, so DMI-R01
  gives category 1 and `context.regularDefender = 0`; its class flags do not matter here (no
  bypass flag, no sanctuary). `fire = 2560`, defender `Stat(40) = 0`. (a) `game.expansion = 1`,
  `game.difficulty = 2`, `ResistPenalty` -100, `Stat(39) = 120`: `r = 120`; `r >= 100` but the
  defender is not regular, so pierced: `r = 110`; penalty: `r = 110 + (-100) = 10`; `r > 0`;
  `cap = 0 + 75 = 75`, `10 < 75`; return 10. DMI-R06: `v = mulDiv(90, 2560, 100) = 2304`. (b)
  classic Normal, `Stat(39) = 90`, attacker `Stat(333) = 0`: `r = 90`; no penalty;
  `cap = 75`, `r = 75`; `v = mulDiv(25, 2560, 100) = 640`. Outputs: `fire = 2304` / `640`.
  Contrast with a regular monster defender (monster-defender pairing, same inputs): (a) not
  pierced, not penalized, not capped, `r = 120`, clamped to 100 in DMI-R06, `fire = 0`; (b)
  `r = 90`, `fire = mulDiv(10, 2560, 100) = 256`. Whole-entry `P = 100` for both (player
  attacker, non-player defender; PVP-R02a). `runtime_observed: no`.
- **DMI-V24** (DMI-R05 non-monster branch, DMI-R06) bypass against a player. Default pairing
  (regular-monster attacker with `Stat(333) = 0`, player defender); `fire = 2560`; defender
  `Stat(39) = 50`, `Stat(40) = 0`, `Stat(35) = 2` (so `reduction[2] = 512`). (a)
  `bypassOther=true`: the defender is not a monster, so DMI-R05's else branch gives `b = 1`.
  Entry 1: `r = 50` (no pierce, classic Normal no penalty, `cap = 75`, `r = 50`); DMI-R06 with
  `b = 1`: `r > 0` so `r = 0`; flat reduction skipped; `r == 0`: no percent step; absorb
  skipped. Output `fire = 2560`. (b) `bypassUndead=true` only: else branch, `bypassOther` false,
  so `b = 0`. `v = 2560 - 512 = 2048`; `r = 50`; `v = mulDiv(50, 2048, 100) = 1024`; absorb
  stats 0, skipped. Output `fire = 1024`. `runtime_observed: no`.
- **DMI-V25** (DMI-R07 step 3) magic resistance exempt from the penalty. Default pairing; player
  defender; `game.expansion = 1`, `game.difficulty = 2`, `ResistPenalty` -100. (a)
  `magic = 2560`; defender `Stat(37) = 40`, `Stat(38) = 0`, `Stat(35) = 0`: entry 4: `r = 40`;
  no pierce binding; resistance binding 37, so no penalty; `r > 0`; maximum binding present:
  `cap = 0 + 75 = 75`, `r = 40`; not stat 36. DMI-R06: `v = 2560 - 0`;
  `v = mulDiv(60, 2560, 100) = 1536`; absorb: stats 146 and 147 are 0, skipped. Output
  `magic = 1536`. (b) contrast, `fire = 2560`, `Stat(39) = 40`, `Stat(40) = 0`, attacker
  `Stat(333) = 0`: `r = 40 + (-100) = -60`; step 4 returns -60;
  `v = mulDiv(160, 2560, 100) = 4096`. Output `fire = 4096`. `runtime_observed: no`.
- **DMI-V26** (DMI-R02, DMI-R06, DMI-R07, DMI-R08, DMI-R09) mitigation callers' mulDiv
  argument orders. Execution: supplied-input mitigation boundary, default pairing and common
  setup, each case with a fresh record. (a) Inputs: `physical = 100002`, defender
  `Stat(36 damageresist) = 25`, all reduction and
  absorb stats 0. No draw. Trace: reduction slot 1 is 0; `b = 0`; effective resistance
  is 25 (no physical penalty, cap 50, no sanctuary). DMI-R06 calls
  `mulDiv(a = 75, b = 100002, c = 100)`. Since `b <= 0x100000` and `a <= 0x10000`, NRF-R05
  uses the i32 product: `7500150 / 100 = 75001` (truncated). Outputs: `physical = 75001`,
  `record.absorbHeal = 0`. Discriminating assertion: swapping the caller's first two
  arguments would make `a = 100002 > 0x10000`; with `100 <= (100002 >> 4) = 6250`,
  the divide-first branch gives `(100002 / 100) * 75 = 1000 * 75 = 75000`, which fails
  this case.
  (b) Inputs: `physical = 200000`, `record.reductionScale = 100002`, defender
  `Stat(34 normal_damage_reduction) = 5`, resistance 0. Reduction is initially `5 << 8 = 1280`.
  DMI-R02 calls `mulDiv(100002, 1280, 1024)`: the a-large divide-first branch gives
  `(100002 / 1024) * 1280 = 97 * 1280 = 124160`. DMI-R06 subtracts it with no resistance
  or absorb. Outputs: `physical = 75840`, `record.absorbHeal = 0`. Swapping the arguments
  selects the i32 product branch: `(1280 * 100002) / 1024 = 125002`, giving physical 74998,
  which fails the case.
  (c) Inputs: `fire = 100003`, defender `Stat(142 item_absorbfire_percent) = 40`, all reduction,
  resistance and flat absorb stats 0. DMI-R08 calls `mulDiv(40, 100003, 100)`, giving
  `(40 * 100003) / 100 = 40001`; remaining fire is 60002. Outputs: `fire = 60002`,
  `record.absorbHeal = 40001`. Swapping the arguments selects the a-large divide-first branch:
  `(100003 / 100) * 40 = 40000`, giving fire 60003 and heal 40000, which fail the case.
  These cases check all three mitigation call-site orders, separately from V13's helper checks.
  `runtime_observed: no`.

- **DMI-V27** (NRF-R12b, NRF-R12c; DMI-R02, DMI-R06, DMI-R07, DMI-R08)
  **getter integration from supplied arrays**. Use the common setup and default pairing;
  both units have non-null lists with `usesResolvedStats=true` and associated owners equal
  to themselves. Definition count is 359; all reads have available definitions with
  `fMin=false`, `MinAccr=0`, `ValShift=0`. For each case below, start fresh.
  The attacker arrays are empty with active counts 0. Tuples are `(statId,layer,value)`;
  unlisted defender identities are absent and therefore read 0. Class/state inputs remain
  those of the common setup; array construction and state-bit production are not exercised.
  (a) `physical=5120`; defender ownStats `[(34,0,1),(36,0,0)]`, resolvedStats
  `[(34,0,5),(36,0,25)]`, counts 2. Getter results are 5 and 25 unchanged. Reduction is
  `5 << 8=1280`; physical becomes `mulDiv(75,5120-1280,100)=2880`, heal 0.
  (b) `fire=2560`; defender ownStats `[(143,0,1)]`, resolvedStats `[(143,0,3)]`, counts 1.
  Flat absorb reads 3 unchanged, converts to `3 << 8=768`; fire becomes 1792, heal 768.
  (c) Fresh inputs: `physical=5120`; defender ownStats `[(34,0,1),(36,0,0)]`, count 2,
  resolvedStats `[(36,0,25)]`, active count 1; use this vector's common lists, owners and definitions:
  stat 34 reads 0 despite ownStats containing 1; physical becomes
  `mulDiv(75,5120,100)=3840`, heal 0.
  All other type fields remain 0, and no list, array, definition or seed changes. **No draw**.
  `runtime_observed: no`. These are getter-to-stage assertions, not whole-game parity.

### 8.2 Observation vectors

Priorities: caps/floor, penalty on poison length/drain inputs, negative physical fields.
These are proposed observations, not reproduced results.

- **DMI-O01** Per-type inputs and outputs. Measure defender stats 34/36/39/40/
  142/143, attacker stat 333, physical/fire before and after per-type mitigation, and absorbHeal
  at the shared entry's exit. Expect DMI-R02 and DMI-R05 to DMI-R08 applied to the recorded
  inputs (DMI-V01, DMI-V03, DMI-V05). [Reproduction procedure](../research/spec-evidence/combat-damage-mitigation.md#dmi-o01-reproduction).
- **DMI-O02** Poison-length/drain penalty. Measure poisonLength and lifeDrain before
  and after the shared entry on expansion Hell, with stat 110=0 and regular-monster attacker.
  Expected each is `mulDiv(200,input,100)` under DMI-V15/DMI-V16; control: Normal leaves
  poisonLength unchanged. Settles DMI-Q04. [Reproduction procedure](../research/spec-evidence/combat-damage-mitigation.md#dmi-o02-reproduction).
- **DMI-O03** Negative physical field. Measure physical on per-type exit and total on
  shared-entry exit. Expect `input-(Stat(defender,34)<<8)` when negative, the same negative
  total, and unchanged life under DPO-R17 (DMI-V08). Settles DMI-Q03.
  [Reproduction procedure](../research/spec-evidence/combat-damage-mitigation.md#dmi-o03-reproduction).

## 9. Acceptance and dependencies

**Verified subset.** DMI-R01 to DMI-R10 carry static labels. DMI-V01 to DMI-V18 and DMI-V20 to DMI-V26 exercise
DMI-R01, DMI-R02 and DMI-R04 to DMI-R10 at the mitigation boundary using supplied stat values
and classifications; DMI-V19 exercises DMI-R03's placement. Passing their numeric, placement
and no-draw assertions is partial acceptance. DMI-V27 additionally exercises verified
NRF-R12b/NRF-R12c with existing arrays and loaded definitions; it does not confirm array
production or gameplay observations. Dispatcher assertions consume the reconciled IEP-R07;
this does not verify the production registration path or each item's callback behavior.

**Assertion ownership.** Per-type outputs, bypass/classification, reduction, and absorbHeal
belong to the cited DMI rules on each vector. `record.total` assertions in DMI-V01/V08/V15/V19
belong to DPO-R12a; later life effects belong to DPO-R17. Whole-entry assertions in DMI-V14 and
DMI-V19 also depend on PVP-R02a's `P=100` selection, DPO-R08a, and DPO-R10 (including V19's
post-dispatch cold-length removal). Callback order and
arguments in DMI-V19 belong to DPO-R23 and IEP-R07. These integration assertions keep their dependency
labels and are reported separately from isolated DMI numeric results.

**Gates and full reproduction.** Stat-getter behavior is supplied by verified NRF-R12b/NRF-R12c;
dispatcher behavior is supplied by IEP-R07 through DPO-R23. Full reproduction within this stage's
boundary requires DMI-V01 to DMI-V27 to pass and DMI-O01 to DMI-O03 to match on an owned 1.13c
game. Supplying existing stat arrays does not verify their production by other systems.
A table read cannot satisfy gameplay observation acceptance.

**Gap boundaries.** DMI-Q03 and
DMI-Q04 are pending observation cross-checks: they block runtime-confirmation claims, not
the existing static arithmetic. DMI-Q05 concerns an outside-boundary producer and blocks no
local supplied-input rule. A gap-file reference to a rule does not expand these gates.

**Depends on:**

- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): NRF-R05 exact
  mulDiv, NRF-R08a units, NRF-R12a/NRF-R12b/NRF-R12c stat/state reads,
  NRF-R01/NRF-R02 seed contracts for DMI-V14, section 3.2 damage-field meanings.
- [core-game-data-fields](core-game-data-fields.md): GDF-R02 difficulty data,
  GDF-R05/GDF-R18 monster category flags, GDF-R11 stat identities, GDF-R12 sanctuary state,
  GDF-R17 class identities. These are data/meaning contracts, not original layout requirements.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): DPO-R05/DPO-R06 built
  inputs; DPO-R08a scaling loop, DPO-R09 event 11 placement, DPO-R23 dispatcher, DPO-R10 strip,
  DPO-R08c/DPO-R08d timing/count; DPO-R12a totals and DPO-R16/DPO-R17 later healing/life
  effects for integration assertions.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): PVP-R02a scale
  selection for the whole-entry setup; PVP-R01a runs per path.
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): IEP-R01
  entry creation and IEP-R07 callback arguments for the event list read by DMI-R03. Callback behavior is separate.

**Depended on by:**

- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): DPO-R11 consumes this
  loop; DPO-R14 to DPO-R18 consume its results.
- [combat-status-application](combat-status-application.md): consumes mitigated
  poison and lengths; [combat-leech-and-status-effects](combat-leech-and-status-effects.md): regular-monster drain inputs.

**Supporting relationships.** GDF-R08/SCD-R06 identify the original dispatch root; they are
evidence references here. CSA-R01 in
[combat-status-application](combat-status-application.md) consumes the record's status and
drain fields; it does not execute during mitigation. Reverse consumers and supporting
references do not automatically set implementation-plan prerequisites.

## 10. Related specs

- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): surrounding stages,
  scaling/strip, totals, healing/life application, and event dispatch points.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): PvP/hireling
  scale selected before this loop; an adjacent stage rather than a consumer of this loop.
- [combat-status-application](combat-status-application.md): later use of length fields; [combat-leech-and-status-effects](combat-leech-and-status-effects.md): later use of drain fields.
- [combat-item-event-triggers-and-procs](combat-item-event-triggers-and-procs.md): event
  entries and callbacks eligible for the event 11 dispatch, separate from mitigation arithmetic.
- [skill-cast-start-and-do-dispatch](skill-cast-start-and-do-dispatch.md): SCD-R06 original
  dispatch-root evidence, together with GDF-R08's srvdofunc identity.
- [combat-shield-block-and-avoidance](combat-shield-block-and-avoidance.md),
  [combat-knockback](combat-knockback.md), [combat-hit-recovery](combat-hit-recovery.md):
  adjacent hit outcomes outside this boundary.
- [core-game-data-fields](core-game-data-fields.md): stat/state/class/column identities.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. Original `D2Game.dll`,
`D2Common.dll` and `Fog.dll` analysis specimens are identified by exact SHA-256 fingerprints
in the [source manifest](../research/source-manifest.md).

The [public evidence companion](../research/spec-evidence/combat-damage-mitigation.md) gives
module-relative function locations, export ordinals where recorded, dispatch/callee chains,
rule-linked rationale, named table cells and observation procedures. Source readings recorded
on 2026-10-12/13 used disassembly, reconstructions checked against original bytes and independent
instruction decoding where recorded. The targeted 2026-10-14 stat-reader records are linked
through the numeric foundations evidence. Function descriptions are analyst labels.

Original-profile table reads dated 2026-10-13 (TD-1 to TD-5) establish their named cells and
selected compiled correspondence. Effective-file hashes and supplying MPQs were not captured;
the manifest states the runtime identity limits. A read-only memory comparison on 2026-10-13 corroborated the
type table; no executable vector was runtime-observed. Gameplay observations remain pending
and the spec remains draft. Public verification requires matching original artifacts and
recorded data scope, without access to the private reconstruction project.
