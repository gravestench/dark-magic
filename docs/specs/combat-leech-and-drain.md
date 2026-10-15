# Leech and Drain

**Description**

A hit with life or mana leech can restore a share of its physical damage to the attacker's life
or mana, up to the corresponding maximum. The target's drain effectiveness changes that share,
and player leech is reduced on Nightmare and Hell; the ordinary life-restoration path refuses a
dead or dying attacker. Monsters can instead use a separate life-restoration calculation based
on the hit and the target's available mana and stamina, and a mana-leeching hit can also drain the
target's mana. Leech requests a visual effect, choosing between the life and mana effects at
random when both kinds apply.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: LSE-R, LSE-Q, LSE-V, LSE-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-leech-and-drain.md)
Open evidence gaps: [gap file](../plans/gaps/combat-leech-and-drain.md)

## 1. Scope

This spec defines life and mana leech applied by the hit-application operation after the
defender's life clamp: the leech precondition, drain effectiveness, the leech scale, the
attacker routing, the player difficulty divisors, the base, the mana-then-life amounts, the capped
life and mana adds with their caps and life-add gates, the overlay choice draw and the overlay
write requests, the monster-attacker branch (apparent attacker form, defender mana and stamina
terms, room cap and additive life write), and the defender mana drain that consumes the value
leech leaves in the record. Out of scope: building, scaling and mitigating the record, the life
clamp, the absorb heal, the current-life write and the order of these stages within the hit are
[combat-damage-pipeline-order](combat-damage-pipeline-order.md) and
[combat-damage-mitigation](combat-damage-mitigation.md) (DPO-R12a computes the total, including its
leech-related terms, DPO-R14 the clamp, DPO-R15 places leech, DPO-R16 the absorb heal, DPO-R17 the
life write); the defender stamina drain arithmetic and its placement after the mana drain are
DPO-R18; the stun, cold, freeze, poison and burn appliers, their draws and the status hand-off
order are [combat-status-application](combat-status-application.md) (CSA-R01 to CSA-R13);
timed-list construction, attachment and list-stat writes are
[state-timed-state-lifecycle](state-timed-state-lifecycle.md); reflection, hit reaction and the
post-apply events are the pipeline and [combat-hit-recovery](combat-hit-recovery.md).

## 2. Trigger and ownership

- Module: D2Game (server). The leech operation and its D2Game helpers carry no source-file assert
  string; ownership follows the call chain: they are called only from the hit-application
  operation, which the pipeline owns under `combat`. The monster-form, stat and
  overlay helpers are D2Common operations consumed here.
- Trigger: once per applied hit, synchronously inside the hit-application operation reached from
  the queued-hit dequeue of melee and skill attacks and from the missile contact and callback
  routes, as [combat-damage-pipeline-order](combat-damage-pipeline-order.md) defines.
- Placement within the hit: owned by the pipeline. Leech runs after the life clamp (DPO-R14) as
  the leech hand-off (DPO-R15), before the absorb heal (DPO-R16) and the current-life write
  (DPO-R17); the defender mana drain (LSE-R15) is the mana burn of DPO-R18 and the first request
  of the status hand-off (CSA-R01), followed by the stamina drain. The clamp, absorb heal and life
  write make no draw, so the leech overlay draw (LSE-R12) is the first draw of this stage; every
  later draw of the status stage is listed by CSA-R01.
- Frequency: the leech operation and the mana drain each run exactly once per hit-application
  call; the overlay choice draw happens at most once.
- Client view: this stage requests server update marking. Its queue effects remain LSE-Q01 and
  RLI-Q02; which messages carry stat and overlay changes, and their delivery to clients, remain
  LSE-Q04. STL-R14's candidate state-message selection does not establish that sender contract.
- The full call chain with original locations is in the
  [companion](../research/spec-evidence/combat-leech-and-drain.md#dispatch-chain).

## 3. Inputs and data

Integers use the foundations notation (NRF section 1). Life, mana and stamina are raw 8.8
(NRF-R08a). `Stat(u, s, 0)` is the unit read of RLI-R02a and RLI-R02b.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| attacker `A`, defender `D` | units passed by the hit-application operation | association; either may be null | Roles are named; no positional order is required. The apply gate (DPO-R13) never passes a null `D`; LSE-R02 reads one as `P = 100` |
| absorb heal and life write | `AbsorbHeal(D, R.absorbHeal)` (DPO-R16) and `LifeWrite(D, R.total)` (DPO-R17) | consumed operations | Called by the hit-application operation after leech whatever their amounts; each applies its own amount guard inside (DPO-R16 returns for a nonpositive amount, DPO-R17 writes only for a positive total) |
| hit game `G` | the game the hit-application operation passes to leech | association | `G.difficulty` 0, 1 or 2 selects the defender's drain column; this contract covers 0..2 only (CSA-Q06) |
| unit identity | `unit.type`, `unit.classId`, `unit.id`, `unit.mode`, `unit.game`, `unit.pool` | i32 type/class/mode, u32 id, associations | Types 0 player, 1 monster (NRF section 3.4); modes GDF-R05/GDF-R06; `unit.pool` is the unit's allocation context for new lists |
| seed | `A.seed` | pair `{lo,hi}` of u32 | One step is NRF-R01; `Rand` is NRF-R02 |
| current stats and definitions | existing unit stat arrays and ItemStatCost definitions of NRF section 3.8; stats 6, 7, 8, 9 and 10 at layer 0 | i32 raw values; getter selection and minimum metadata | `Stat` consumes RLI-R02a/RLI-R02b; supplied fixtures define their own stat view below, without verifying array construction or stored-write integration |
| damage record `R` | NRF section 3.2 fields `physical`, `fire`, `lightning`, `magic`, `cold`, `poison`, `lifeLeech`, `manaBurn`, `staminaBurn`, `absorbHeal`, `total` | i32; magnitudes 8.8 | On entry `lifeLeech` and `manaBurn` are the leech percents built by the pipeline; leech rewrites both in place; `manaBurn` as left by leech is the defender mana drain amount (LSE-R15). The pipeline names `lifeLeech`, `manaBurn`, `staminaBurn` and `poison` as `lifeDrain`, `manaDrain`, `staminaDrain` and `poisonMagnitude` (same quantities) |
| monster data | `D.monsterData` present or absent; its MonStats record present or absent | association | Missing data aborts leech (LSE-R02) |
| `MonStats.Drain`, `Drain(N)`, `Drain(H)` | per-class column for difficulty 0, 1, 2 | u8, read unsigned | Leech effectiveness `P` of a monster defender |
| difficulty divisors | `DifficultyLevels.LifeStealDivisor`, `ManaStealDivisor` | i32 | Row Normal, Nightmare or Hell for `A.game.difficulty` 0, 1 or 2 (GDA-R05); this contract covers 0..2 only (CSA-Q06) |
| class group | `unitGate(A)` of CSA-R02 | i32 0..4 | Nonzero for the five hireling classes |
| alive exclusion | `unit.aliveExclusionCondition` | boolean | The unit exclusion condition CSA-R13 also tests; wider meaning CSA-Q03 |
| disguise condition | `A.disguiseCondition` | boolean | The condition the lifecycle's disguise bookkeeping maintains (STL-R07a); producer LSE-Q02 |
| form-state list | `formStateList` | ordered loaded list of state ids; count signed 16-bit | Contents and loader LSE-Q02 |
| state form kind | `States.gfxtype` of a state | u8 | GDA-R13 field; values 1 and 2 select a form (LSE-R14) |
| state membership | `HasState(unit, s)` | RLI-R03 | States 92 `death_delay` and the form states |
| overlay count | `overlayCount` | i32 | Upper bound for an overlay id (LSE-R13); table identity LSE-Q03 |
| attached lists | `A.statList`, its `hasStateStorage` condition, its primary attached-list chain, each list's `overlayListCondition`; the base list's `overlayPendingCondition` | NRF section 3 list fields; the two overlay conditions are neutral names for engine-defined list conditions (LSE-Q03) | Searched and set by LSE-R13 |
| dependency operations | unit replacing `SetStat`, unit additive `AddStat`, `NewList`, `Attach`, list `SetStat`, `MarkForUpdate`; supplied `AbsorbHeal`, `LifeWrite` and later status requests in the stage fixture | named operations and supplied effects | The fixture preamble defines persistence and logging; requested writes do not establish their stored integration (LSE-Q01) |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Stats 6 `hitpoints`, 7 `maxhp`, 8 `mana`, 10 `stamina` (GDF-R02); 9 `maxmana`, 178 `unit_dooverlay` ([table records](../research/spec-evidence/combat-leech-and-drain.md#table-records)) | Names from `data/global/excel/itemstatcost.txt` | Fixed by original code |
| State 92 `death_delay` ([table record](../research/spec-evidence/combat-leech-and-drain.md#table-records)) | Name from `data/global/excel/states.txt` | Fixed by original code |
| Leech scale 64, percent divisor 100, full effectiveness 100, mana drain floor 256, overlay ids 151 and 152, a new overlay list carrying only `overlayListCondition` (no expiry condition), attach mode 1, apparent type 6 for a null attacker, form kinds 1 and 2 | Engine code | Fixed by original code |
| Divisors (LSE-R06) | `data/global/excel/difficultylevels.txt`, rows `Normal`, `Nightmare`, `Hell`, columns `LifeStealDivisor`, `ManaStealDivisor` | Loaded data; code fixes the consumer and the zero-divisor skip |
| `Drain` per class | MonStats per-class cells | Loaded data; vectors supply them as named inputs |
| Class groups | CSA-R02 | Fixed numeric bindings |

Dependency operations, each with its owning contract: `Stat` (RLI-R02a, RLI-R02b); `HasState`
(RLI-R03); `mulDiv` (NRF-R05); `Rand` (NRF-R02); `unitGate`
(CSA-R02); `NewList` (STL-R06a); `Attach` (STL-R06b); list `SetStat` (STL-R06c). The unit replacing
write `SetStat(u, s, v, 0)` and additive write `AddStat(u, s, v, 0)` have supplied effects in these
fixtures; their original stored effects remain LSE-Q01. `MarkForUpdate(u)` likewise consumes the
pending update-queue boundary of RLI-Q02, retained locally by LSE-Q01. NRF-R08a defines numeric
formats and does not establish either unit-write operation.

`DifficultyRecord` names LSE-R05's local selection of the prepared difficulty row for the supplied
index 0..2. Its getter evidence is in the [local function table](../research/spec-evidence/combat-leech-and-drain.md#function-references).
GDA-R05 owns row preparation and field correspondence; it does not supply the getter's contract.
No selection behavior outside 0..2 is required here.

## 4. Rules (execution order)

Pseudocode names are illustrative. "Return" leaves the leech operation with no further write.
Divisions written `/` are signed and truncate toward zero; `min` is signed. Every comparison in
this section (`<=`, `<`, `>=`, `==`, `!=` and the minimums) is on signed i32 values.

The immediate total `R.total`, including its poison term and the unshifted `lifeLeech` term for a
regular monster attacker, is computed before this stage by DPO-R12a; the leech operation consumes
neither it nor its terms (LSE-R03).

**LSE-R01** **Precondition** [verified in 1.13c code (static)].

```
Leech(R, D, G, A):
    if R.lifeLeech == 0 and R.manaBurn == 0: return      // record unchanged, no draw, no overlay
```

**LSE-R02** **Effectiveness** [verified in 1.13c code (static)].

```
    if D == null or D.type != 1: P = 100
    else:
        if D.monsterData absent or its MonStats record absent: return   // record unchanged
        P = Drain column of that record for G.difficulty               // u8, 0..255; the hit game's difficulty
        if P <= 0: return                                              // only P == 0 occurs
```

**LSE-R03** **Scale and store** [verified in 1.13c code (static)].

```
    R.lifeLeech = low32(R.lifeLeech << 6)
    R.manaBurn  = low32(R.manaBurn << 6)
```

Both are written into the record before any later abort. No path of the leech operation reads or
writes `R.total` (DPO-R12a's total); it stays as mitigation left it for the life write (DPO-R17).

**LSE-R04** **Attacker routing** [verified in 1.13c code (static)].

```
    if A == null:                         continue at LSE-R14
    if A.type == 0:                       continue at LSE-R05, then LSE-R07
    if A.type == 1 and unitGate(A) != 0:  continue at LSE-R07            // hireling: no divisors
    continue at LSE-R14                                                   // any other attacker
```

**LSE-R05** **Player divisors** [verified in 1.13c code (static)].

```
    d = DifficultyRecord(A.game.difficulty)                   // row 0, 1 or 2, GDA-R05
    if d.LifeStealDivisor != 0: R.lifeLeech = R.lifeLeech / d.LifeStealDivisor
    if d.ManaStealDivisor != 0: R.manaBurn  = R.manaBurn  / d.ManaStealDivisor
```

Each quotient is stored back. A zero divisor skips its division. The divisor row follows the
attacker's game difficulty, while the drain column of LSE-R02 follows the hit game `G`; the two are
separate inputs.

**LSE-R06** **Divisor values** [verified in 1.13c data]. `LifeStealDivisor` is 1, 2, 3 and
`ManaStealDivisor` is 1, 2, 3 for Normal, Nightmare and Hell
([table record](../research/spec-evidence/combat-leech-and-drain.md#table-records)).

**LSE-R07** **Base** [verified in 1.13c code (static)].

```
    base = R.physical                     // already mitigated and clamped (DPO-R14)
    if base <= 0: return                  // no add, no draw, no overlay; the record keeps its rewrite
```

**LSE-R08** **Mana first, then life** [verified in 1.13c code (static)].

```
    manaFlag = false
    if R.manaBurn != 0:
        m = mulDiv(R.manaBurn, base, 100)
        if P != 100: m = mulDiv(P, m, 100)
        ManaAdd(A, m / 64)                // LSE-R09
        manaFlag = true                   // also when m / 64 is 0
    if R.lifeLeech != 0:
        l = mulDiv(R.lifeLeech, base, 100)
        if P != 100: l = mulDiv(P, l, 100)
        LifeAdd(A, l / 64)                // LSE-R09, LSE-R11
    continue at LSE-R12
```

Amounts are raw 8.8 like `base`. The tests use the rewritten record values, not the amounts.

**LSE-R09** **Capped add** [verified in 1.13c code (static)].

```
ManaAdd(u, amount):                       LifeAdd(u, amount):
    if amount <= 0: return 0                  if amount <= 0: return 0
                                              if LSE-R11 refuses u: return 0
    cap = Stat(u, 9, 0)                       cap = Stat(u, 7, 0)            // LSE-R10
    cur = Stat(u, 8, 0)                       cur = Stat(u, 6, 0)
    new = low32(cur + amount)                 new = low32(cur + amount)
    if new >= cap: new = cap                  if new >= cap: new = cap
    SetStat(u, 8, new, 0)                     SetStat(u, 6, new, 0)
    return amount if new != cap               return amount if new != cap
           else low32(cap - cur)                     else low32(cap - cur)
```

A unit above its cap is lowered to the cap. The leech operation ignores both return values; the
absorb heal (DPO-R16) uses `LifeAdd` on the defender and also ignores the return value.

**LSE-R10** **Cap sources** [verified in 1.13c code (static)]. The mana cap is the unit read
`Stat(u, 9, 0)` and the life cap is `Stat(u, 7, 0)`, with RLI-R02a's guards (0 for a null unit, a
null stat list or an unavailable definition) and RLI-R02b's minimum adjustment. No other
contribution is added.

**LSE-R11** **Life-add gates** [verified in 1.13c code (static)]. `LifeAdd` refuses `u`, after its
amount test and before any read, when `u` is not alive or `HasState(u, 92) != 0`:

```
alive(u) = u != null and not u.aliveExclusionCondition and
           ((u.type == 0 and u.mode not in {0, 17}) or (u.type == 1 and u.mode not in {0, 12}))
```

Units of other types are not alive. `ManaAdd` has no such gate, so a dead or dying attacker
regains mana but not life.

**LSE-R12** **Overlay choice and draw** [verified in 1.13c code (static)].

```
    if R.lifeLeech != 0:
        if manaFlag:
            r = Rand(A.seed, 2)           // exactly one step of the attacker's seed
            id = 151 if r == 0 else 152
        else: id = 151
    else if manaFlag: id = 152
    else: return                          // both rewritten values are 0: no draw, no overlay
    SetOverlay(A, id)                     // LSE-R13
```

The draw depends on the rewritten values, not on the amounts: a hit whose amounts are 0 still
draws when both values are nonzero. Both rewritten values are 0 only when the shift of LSE-R03
wraps to 0 (for example an input `lifeLeech` of 0x04000000 with `manaBurn` 0).

**LSE-R13** **Overlay write requests** [verified in 1.13c code (static)].

```
SetOverlay(u, id):
    if id < 0 or id >= overlayCount: return
    L = null
    if u.statList != null and u.statList.hasStateStorage:
        L = first list on u's primary attached-list chain whose overlayListCondition is set
                                      // starts at the chain head; the base list itself is not a candidate
    if L == null:
        L = NewList(u.pool, conditions {overlayListCondition}, expiry 0, owner u.type, u.id)   // STL-R06a
        Attach(u, L, 1)
    if u.statList != null: set u.statList.overlayPendingCondition       // u.statList re-read after Attach
    SetStat(L, 178, id AND 0xFFFF, 0)                                     // list write, STL-R06c
    MarkForUpdate(u)
```

No draw occurs. The new list carries no expiry condition, so the expiry walk (STL-R04) never
removes it. The wider meanings of `overlayListCondition` and `overlayPendingCondition` are LSE-Q03.

**LSE-R14** **Monster-attacker branch** [verified in 1.13c code (static)]. Reached for a null
attacker and for every attacker that is neither a player nor a hireling:

```
    t = A.type if A != null else 6
    t = ApparentType(A, t)
    if t == 0: continue at LSE-R07                        // no divisors
    defMana = Stat(D, 8, 0); defStam = Stat(D, 10, 0)
    R.lifeLeech = R.lifeLeech / 64                         // stored back
    R.manaBurn  = R.manaBurn / 64                          // stored back
    total = low32(min(R.lifeLeech, R.physical) + min(R.manaBurn, defMana)
                  + min(R.staminaBurn, defStam))           // staminaBurn is not divided
    if total == 0: return                                  // a negative total continues
    if P != 100: total = mulDiv(P, total, 100)
    room = low32(Stat(A, 7, 0) - Stat(A, 6, 0))            // LSE-R10 cap; 0 for a null attacker
    if total >= room: total = room                         // a scaled total of 0 or below also reaches here
    if total <= 0: return
    AddStat(A, 6, total, 0)
    SetOverlay(A, 151)                                     // LSE-R13; no draw

ApparentType(A, t):
    if A == null or not A.disguiseCondition: return t
    for i in 0 .. count(formStateList) - 1:              // none when the signed count is <= 0
        s = formStateList[i]
        if A.type not in {0, 1, 3}: continue
        if HasState(A, s) == 0: continue
        if States(s).gfxtype == 1: return 1
        if States(s).gfxtype == 2: return 0
    return t
```

There is no `base <= 0` test on this path, and no alive, exclusion or state 92 gate on `AddStat`:
a dead monster attacker still gains life here. A nonpositive list count leaves `t` unchanged. The apparent-form operation also yields a class and a
mode that leech does not use; it makes no draw and no write that leech observes.

**LSE-R15** **Defender mana drain** [verified in 1.13c code (static)]. The pipeline places this
call as the mana burn of DPO-R18, the first request of the status hand-off (CSA-R01). `v` is `R.manaBurn` as the
leech operation left it: `M << 6` for a hireling or an apparent player form, `(M << 6) /
ManaStealDivisor` for a player when that divisor is nonzero, `(M << 6) / 64` on the
monster-attacker path, and the unshifted input `M` when leech returned at LSE-R01 or LSE-R02.

```
ManaDrain(D, v):
    if v <= 0: return 0
    new = low32(Stat(D, 8, 0) - v)
    if new < 256: new = 0
    SetStat(D, 8, new, 0)
    return v if new != 0 else 0
```

The caller ignores the return value. No draw occurs.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| LSE-R01 | No leech when both inputs are 0 | verified in 1.13c code (static) | none | LSE-V02 |
| LSE-R02 | Effectiveness 100 or the unsigned drain byte; 0 or missing data returns | verified in 1.13c code (static) | none | LSE-V03 |
| LSE-R03 | Both inputs shifted by 6 and stored; the total is neither read nor written | verified in 1.13c code (static) | none | LSE-V01, LSE-V04, LSE-V06, LSE-V10 |
| LSE-R04 | Player, hireling and other-attacker routing | verified in 1.13c code (static) | none | LSE-V05, LSE-V12 |
| LSE-R05 | Player divisors; zero divisor skips | verified in 1.13c code (static) | none | LSE-V05, LSE-V14 |
| LSE-R06 | Divisor values per difficulty | verified in 1.13c data | none | [table record](../research/spec-evidence/combat-leech-and-drain.md#table-records) |
| LSE-R07 | Base is the clamped physical damage; nonpositive returns | verified in 1.13c code (static) | none | LSE-V06 |
| LSE-R08 | Mana then life; mulDiv chain, effectiveness scaling, signed division by 64 | verified in 1.13c code (static) | none | LSE-V03, LSE-V04, LSE-V05, LSE-V09 |
| LSE-R09 | Capped mana and life adds and their returns | verified in 1.13c code (static) | LSE-Q01: stored effect of the replacing unit write; requests, caps and returns eligible | LSE-V07 |
| LSE-R10 | Caps are the unit reads of stats 9 and 7 | verified in 1.13c code (static) | none | LSE-V07 |
| LSE-R11 | Life-add alive and `death_delay` gates | verified in 1.13c code (static) | none | LSE-V08 |
| LSE-R12 | One attacker draw only when both rewritten values are nonzero; overlay 151 or 152 | verified in 1.13c code (static) | none | LSE-V09, LSE-V10 |
| LSE-R13 | Overlay list search or creation and stat 178 request | verified in 1.13c code (static) | LSE-Q01: stored attach, list-stat and update effects; ordered requests eligible | LSE-V11 |
| LSE-R14 | Monster-attacker branch: apparent form, division by 64, three minimums, room cap, additive write | verified in 1.13c code (static) | LSE-Q01: stored effect of the additive unit write; arithmetic and requests eligible | LSE-V12 |
| LSE-R15 | Defender mana drain from the value leech leaves | verified in 1.13c code (static) | LSE-Q01: stored effect of the replacing unit write; arithmetic eligible | LSE-V13, LSE-V14 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| `R.total` (written by DPO-R12a before this stage) left unchanged by the leech operation | LSE-R03 |
| `R.lifeLeech`, `R.manaBurn` rewritten: shifted by 6; divided by the player divisors; divided by 64 on the monster-attacker path | LSE-R03, LSE-R05, LSE-R14 |
| Attacker stat 8 `mana` replaced by `min(cur + amount, cap)` | LSE-R08, LSE-R09 |
| Attacker stat 6 `hitpoints` replaced by `min(cur + amount, cap)`, unless refused | LSE-R08, LSE-R09, LSE-R11 |
| Attacker stat 6 `hitpoints` increased by `total` through the additive write | LSE-R14 |
| Attacker seed stepped once | LSE-R12 |
| Overlay request: optional list allocation and attach, `overlayPendingCondition`, list stat 178 `unit_dooverlay` = 151 or 152, update mark | LSE-R12, LSE-R13, LSE-R14 |
| Defender stat 8 `mana` replaced by `cur - v`, or 0 below 256 | LSE-R15 |
| Defender stat 6 raised by the absorb heal (placed by DPO-R16) through the same capped add and gates | LSE-R09, LSE-R11 |
| Return values: each capped add returns 0 when refused, otherwise the requested amount when uncapped or `low32(cap - cur)` when capped (including a negative result when initially above cap); mana drain returns `v` or 0; callers ignore these returns | LSE-R09, LSE-R15 |
| Messages: none sent here; server update marking is requested, while message selection and delivery remain LSE-Q04 | LSE-R13 |

Requested dependency writes become stored effects only under LSE-Q01.

## 6. Branches and edge cases

- Attacker kinds: a player takes the divisors; a hireling (class group nonzero) skips them; a monster
  disguised in a player form (apparent type 0) skips them; every other monster, unit of another type
  and a null attacker take LSE-R14. A null attacker ends LSE-R14 with a room of 0, so it writes
  nothing, but the record values are still divided by 64.
- Defender kinds: a monster uses its drain byte; a drain of 0 or missing monster data returns before
  the shift, and the mana drain then removes the unshifted `manaBurn`. A player defender uses 100.
- Both inputs nonzero but a tiny base: the amounts are 0 and the draw still happens (LSE-V10).
  Negative inputs divide toward zero; a nonpositive amount writes nothing.
- An attacker above its cap is lowered to the cap by a positive add (LSE-V07).
- On the `LifeAdd` paths, a dead or dying attacker, an excluded unit or one holding state 92 gets no
  life; mana, the draw and
  the overlay still happen. The monster-attacker branch has no such gate (LSE-R14).
- Shift overflow: an input whose shift wraps to 0 counts as 0 for the later tests.
- Difficulty enters through the drain column, the two divisors and nothing else; there is no
  classic/expansion, melee/missile or PvP branch here (the record is already scaled, DPO-R08).
- Missing data: a difficulty outside 0..2 is outside this contract (CSA-Q06); a missing MonStats
  record returns at LSE-R02;
  an overlay id outside the overlay count is not written.
- Overflow: additions wrap with `low32`; the monster-attacker total is not otherwise clamped.

## 7. Open questions

| Question | Kind / status | Evidence that would settle it |
|---|---|---|
| **LSE-Q01** | integration / open | Stored effects of the requested writes: the unit replacing write, the unit additive write, `Attach` (STL-R06b), list `SetStat` (STL-R06c) and `MarkForUpdate` (RLI-Q02). Complete unit-write contracts and accept the owners' contracts with integration vectors. Requests, arithmetic and returns here are verified. |
| **LSE-Q02** | outside-boundary / open | Producer of `disguiseCondition` (STL-R07a) and the loader and contents of `formStateList`, with the `gfxtype` values of its states. The apparent-form decision for supplied values is verified. |
| **LSE-Q03** | outside-boundary / open | Which table supplies `overlayCount`, the names of overlays 151 and 152, and the wider roles of `overlayListCondition` and `overlayPendingCondition`. Trace the loader that publishes the count. The bound test for a supplied count is verified. |
| **LSE-Q04** | outside-boundary / open | Which server-to-client messages carry the stat 6, stat 8 and overlay changes: trace the update sender in a `net-` spec. |
| **LSE-Q05** | observation / pending | Matching gameplay captures of LSE-O01, LSE-O02 and the verified rules' vectors, including the poison term of the total (DPO-Q02). No gameplay observation has been made. |

The [gap file](../plans/gaps/combat-leech-and-drain.md) records the attempts and exact
blocked assertions.

## 8. Test vectors

### 8.1 Executable vectors

Shared defaults, overridden only where a vector or case says so. All vectors are supplied-input
operation fixtures (boundary: supplied-input stage) unless a vector states another boundary.

Case structure: every lettered case starts from its vector's Inputs and Seeds with only the inputs it
names overridden. A case never takes another case's state or outputs.

- Units: `A` is a player, type 0, id 7, mode 1, `aliveExclusionCondition` false, no states,
  `disguiseCondition` false, seed `{1, 0}`; stats 6 = 0x1000, 7 = 0x6400, 8 = 0, 9 = 0x1000.
  `A.game` and `G` have difficulty 0 (Normal). `D` is a player, mode 1 (so `P = 100`), stat 8 =
  0x1000, stat 10 = 0. A monster unit is class 0 with monster data and its MonStats record present.
  A monster `A` has the same stats as the default `A`.
- Record: every field 0 unless stated.
- Data: divisors are the LSE-R06 values unless a vector supplies a fixture row; `unitGate` follows
  CSA-R02 (class 0 gives 0); `overlayCount` 1000 (a fixture input); `formStateList` empty.
- Operations: `Stat(u, s, 0)` returns `u`'s current value of `s` (0 for an unstated stat or a null
  unit) and is not logged. The supplied unit `SetStat` replaces the value and `AddStat` adds a
  nonzero value; both persist within a vector and are not logged. `HasState` answers from the
  states a case lists. `R` fields written by leech persist within a vector. Outside LSE-V11, "overlay `n`" means that LSE-R13 runs with id `n`; its
  list requests are asserted only in LSE-V11. In LSE-V11 the supplied `NewList`, `Attach`, list
  `SetStat` and `MarkForUpdate` log their arguments in order, and setting
  `overlayPendingCondition` is logged as an event in that order; `NewList` returns a fresh list
  `L1`; `Attach` does not create a stat list unless a case says so. `owner t/i` means owner type `t`
  and owner id `i`. In LSE-V13 the fixture logs, in order, the attacker
  draw, the supplied `AbsorbHeal` and `LifeWrite` calls (each call is logged whatever its amount;
  their internal guards belong to DPO-R16 and DPO-R17),
  the real `ManaDrain` call (LSE-R15) with its argument, and the supplied CSA-R01 requests that follow
  it; these log entries are named intermediates of consumed contracts, not outputs of this behavior.
- Seeds are full `{lo,hi}` pairs; "no draw" means `A.seed` equals its input. A draw is written
  (seed owner, modulus, result): one NRF-R02 draw on that unit's seed and its returned value.
- Every vector has `runtime_observed: no`.
- Local operations execute the rules themselves, including the leech routing, `ManaAdd`,
  `LifeAdd`, `ManaDrain`, apparent-form selection and overlay request sequence; supplied
  dependencies do not replace these local operations. In LSE-V07, observers record unit-field,
  stat and state reads and unit-stat writes made during the local capped-add operation; an empty
  read/write log asserts that none occurred. In LSE-V09, observers separately record ordered
  unit replacing-write requests as `(unit, stat, value, layer)` while retaining the supplied
  persistence above. These extra observers do not change another vector's stated log.

- **LSE-V01** Total consumed and left unchanged (DPO-R12a, LSE-R03; boundary: isolated rule)
  - Inputs: `R.physical` 1000, `R.poison` 500, `R.lifeLeech` 10, other damage fields 0; player `A`.
  - Seeds: `no draw`
  - Draws: none
  - Trace: DPO-R12a: 1000 + 0 + 0 + 0 + 0 + 500 = 1500; a player adds no leech term.
  - Outputs: `R.total` 1500 entering the leech stage.
    (a) `R.total` 1500 supplied, then the leech operation runs with player `A` and `D`, `R.manaBurn`
    0: `R.lifeLeech` 640, `mulDiv(640, 1000, 100)` = 6400, 6400 / 64 = 100, `A` stat 6 = 0x1064,
    overlay 151; `R.total` still 1500.
    (b) `A` monster class 0, `R.total` 1510 supplied (DPO-R12a's regular-monster leech term),
    `R.manaBurn` 0, `D` player with stats 8 = 0x1000, 10 = 0: the leech operation takes LSE-R14; `R.lifeLeech` 640 then 10; total min(10, 1000) + 0 + 0 =
    10; `A` stat 6 = 0x100A; overlay 151; `R.total` still 1510.
  - Discriminates: 1500 rejects a total without the poison term that this stage's life write
    consumes (DPO-V30 covers the regular-monster and hireling leech terms of DPO-R12a); (a) and (b)
    reject a leech that rewrites the total on either path.
  - runtime_observed: no

- **LSE-V02** Precondition (LSE-R01)
  - Inputs: `R.lifeLeech` 0, `R.manaBurn` 0, `R.physical` 25600.
  - Seeds: `no draw`
  - Draws: none
  - Trace: both inputs 0: return.
  - Outputs: record unchanged; `A` stats unchanged; no write, no overlay.
    (a) `A` monster class 0, `R.staminaBurn` 0x100, `D` stat 10 = 0x100: return; `A` stat 6 stays
    0x1000.
  - Discriminates: (a) rejects testing the precondition after routing, which would add `min(0x100,
    0x100)` = 256 to the monster's life through LSE-R14.
  - runtime_observed: no

- **LSE-V03** Effectiveness (LSE-R02, LSE-R08)
  - Inputs: `D` monster with `Drain` 50 at Normal; `R.lifeLeech` 10, `R.physical` 25600; `A` stat 6 = 0.
  - Seeds: `no draw`
  - Draws: none
  - Trace: `P` 50; 640; `mulDiv(640, 25600, 100)` = 163840; `mulDiv(50, 163840, 100)` = 81920;
    81920 / 64 = 1280.
  - Outputs: `A` stat 6 = 1280; `R.lifeLeech` 640; overlay 151.
    (a) `Drain` 0: return; `R.lifeLeech` 10; `A` stat 6 = 0; no overlay.
    (b) `Drain` 200: `mulDiv(200, 163840, 100)` = 327680; `A` stat 6 = 5120.
    (c) `D` monster with monster data absent: return; `R.lifeLeech` 10; no write.
    (d) `G` and `A.game` difficulty 1 (Nightmare, `LifeStealDivisor` 2), `Drain(N)` 50, `Drain` 0 at
    Normal: 640 / 2 = 320; `mulDiv(320, 25600, 100)` = 81920; `mulDiv(50, 81920, 100)` = 40960;
    `A` stat 6 = 640; `R.lifeLeech` 320.
    (e) `R.lifeLeech` 0, `R.manaBurn` 5, `A` stat 8 = 0: `m` = `mulDiv(320, 25600, 100)` = 81920,
    then `mulDiv(50, 81920, 100)` = 40960; 40960 / 64 = 640; `A` stat 8 = 640; overlay 152; no draw.
    (f) `G` difficulty 1, `A.game` difficulty 0, `Drain(N)` 50, `Drain` 0 at Normal: `P` 50 from `G`;
    divisor 1 from `A.game`; `R.lifeLeech` 640; `A` stat 6 = 1280.
    (g) `Drain` 150, `R.physical` 1008: `l` = `mulDiv(640, 1008, 100)` = 6451; `mulDiv(150, 6451,
    100)` = 9676; 9676 / 64 = 151; `A` stat 6 = 151.
  - Discriminates: (b) rejects a signed byte (200 read as -56 would return); (d) rejects reading the
    Normal column on Nightmare; (a) rejects scaling by 0 instead of returning; (e) rejects skipping
    effectiveness on the mana branch (1280); (f) rejects taking the drain column from `A.game` (return)
    or the divisor from `G` (stat 6 = 640); (g) rejects applying effectiveness after the division by
    64 (150).
  - runtime_observed: no

- **LSE-V04** Scale and amount (LSE-R03, LSE-R08)
  - Inputs: `R.lifeLeech` 10, `R.physical` 25600; `A` stat 6 = 0.
  - Seeds: `no draw`
  - Draws: none
  - Trace: 640; `mulDiv(640, 25600, 100)` = 163840; 163840 / 64 = 2560.
  - Outputs: `A` stat 6 = 2560; `R.lifeLeech` 640; overlay 151.
    (a) `R.physical` 1008: `mulDiv(640, 1008, 100)` = 6451; 6451 / 64 = 100; `A` stat 6 = 100.
    (b) `R.physical` 0x20000: `l` = `mulDiv(640, 0x20000, 100)` = 838860; 838860 / 64 = 13107; `A` stat
    6 = 0x3333.
    (c) `R.lifeLeech` 0, `R.manaBurn` 5, `R.physical` 0x20000, `A` stats 8 = 0, 9 = 0x4000: `m` =
    `mulDiv(320, 0x20000, 100)` = 419430; 419430 / 64 = 6553; `A` stat 8 = 0x1999; overlay 152; no
    draw.
  - Discriminates: 2560 rejects omitting the shift (40); (a) rejects rounding to nearest (101); (b)
    rejects the swapped operand order `mulDiv(0x20000, 640, 100)` = 838400 (stat 6 = 13100); (c)
    rejects it on the mana branch (419200, stat 8 = 6550).
  - runtime_observed: no

- **LSE-V05** Routing, divisors and zero-divisor skip (LSE-R04, LSE-R05, LSE-R08, LSE-R12)
  - Inputs: `G` difficulty 0; `A.game` difficulty 2 with a supplied fixture row `LifeStealDivisor` 3, `ManaStealDivisor`
    0 (not table data); `R.lifeLeech` 10, `R.manaBurn` 5, `R.physical` 25600; `A` stats 6 = 0, 8 = 0.
  - Seeds: `A.seed = {1, 0}`
  - Draws: (A, 2, 1)
  - Trace: 640, 320; 640 / 3 = 213; mana division skipped; mana `mulDiv(320, 25600, 100)` = 81920,
    / 64 = 1280; life `mulDiv(213, 25600, 100)` = 54528, / 64 = 852; both values nonzero: draw.
  - Outputs: `R.lifeLeech` 213, `R.manaBurn` 320; `A` stat 8 = 1280, stat 6 = 852;
    `A.seed = {0x6AC690C5, 0}`; overlay 152.
    (a) `G` difficulty 0, `A.game` difficulty 2 with the Hell table row (divisors 3 and 3): `R.manaBurn` 106; `mulDiv(106, 25600, 100)` = 27136,
    / 64 = 424; `A` stat 8 = 424, stat 6 = 852; draw (A, 2, 1); overlay 152.
    (b) `G` difficulty 0, `A.game` difficulty 2 with the fixture row, `A` monster class 338 (`unitGate`
    2): no division; `R.lifeLeech` 640, `R.manaBurn` 320; `A`
    stat 8 = 1280, stat 6 = 2560; draw (A, 2, 1); overlay 152.
    (c) `G` difficulty 0, `A.game` difficulty 2 with the Hell table row, `R.lifeLeech` -1, `R.manaBurn` 0: -1 << 6 = -64; -64 / 3 = -21;
    `R.lifeLeech` -21; `l` = `mulDiv(-21, 25600, 100)` = -5376; -5376 / 64 = -84; no write; overlay 151;
    no draw; `A.seed = {1, 0}`.
    (d) `G` difficulty 0, `A.game` difficulty 2 with a fixture row `LifeStealDivisor` 0,
    `ManaStealDivisor` 3 (not table data): `R.lifeLeech` 640,
    `R.manaBurn` 106; `A` stat 8 = 424, stat 6 = 2560; draw (A, 2, 1); overlay 152.
  - Discriminates: 320 rejects dividing by a zero divisor (fault) or by another default; (d) shows the
    same skip for the life divisor; (b) rejects dividing a hireling's leech or sending it to LSE-R14;
    (c) rejects floor division (-22).
  - runtime_observed: no

- **LSE-V06** Base returns after the shift (LSE-R03, LSE-R07)
  - Inputs: `R.lifeLeech` 10, `R.physical` 0.
  - Seeds: `no draw`
  - Draws: none
  - Trace: 640; divisor 1; base 0: return.
  - Outputs: `R.lifeLeech` 640; no write, no overlay.
    (a) `R.physical` -5: `R.lifeLeech` 640; no write, no overlay.
    (b) `G` and `A.game` difficulty 1 (`LifeStealDivisor` 2), `R.physical` 0: `R.lifeLeech` 320; no
    write, no overlay.
    (c) `R.manaBurn` 5, `R.physical` 0, `A.seed = {1, 0}`: record 640 and 320; no write, no overlay;
    `A.seed = {1, 0}`.
  - Discriminates: 640 rejects testing the base before the shift; (b) rejects testing it before the
    divisors (640); (c) rejects drawing before the base return.
  - runtime_observed: no

- **LSE-V07** Capped adds (LSE-R09, LSE-R10; boundary: isolated rule)
  - Inputs: `A` the preamble player (mode 1, alive, no state 92); each case gives the stats it
    reads, and any stat a case does not give keeps the preamble value.
  - Seeds: `no draw`
  - Draws: none
  - Trace: per case.
  - Outputs:
    (a) `LifeAdd(A, 2560)` with stat 6 = 0x6000, stat 7 = 0x6400: 0x6A00 >= 0x6400; stat 6 = 0x6400;
    returns 0x400.
    (b) `ManaAdd(A, 0x500)` with stat 8 = 0, stat 9 = 0x1000: stat 8 = 0x500; returns 0x500.
    (c) `LifeAdd(A, 0)`: no read or write; returns 0; read/write log empty.
    (g) `LifeAdd(A, 0)` with stat 6 = 0x6500, stat 7 = 0x6400: stat 6 stays 0x6500; returns 0;
    read/write log empty.
    (d) `ManaAdd(A, -1)`: no write; returns 0.
    (e) `LifeAdd(A, 0x100)` with stat 6 = 0x6500, stat 7 = 0x6400: stat 6 = 0x6400; returns -0x100.
    (f) `ManaAdd(A, 0x500)` with stat 8 = 0xF00, stat 9 = 0x1000, stat 7 = 0x6400: stat 8 = 0x1000;
    returns 0x100.
    (h) `LifeAdd(A, 0x7FFFFFFF)` with stat 6 = 0x100, stat 7 = 0x6400: `low32(0x100 +
    0x7FFFFFFF)` = -2147483393, below the signed cap; stat 6 = -2147483393; returns 2147483647.
    (i) `ManaAdd(A, 0)` with stat 8 = 0x1100, stat 9 = 0x1000: stat 8 stays 0x1100; returns 0;
    read/write log empty.
  - Discriminates: (a) rejects an uncapped add; (e) rejects capping only when the current value is
    below the cap; (f) rejects capping mana by stat 7 (0x1400); (g) rejects an amount test of `< 0`
    (which would lower stat 6 to 0x6400). Empty logs in (c) and (g) reject reading the unit or
    rewriting its unchanged value before returning; (h) rejects widening or saturating the sum
    before comparing with the cap (which would write 0x6400 and return 0x6300). Case (i) independently
    rejects a mana amount test of `< 0` (which would lower stat 8 to 0x1000 and return -0x100),
    and any mana reads or writes before its nonpositive-amount return.
  - runtime_observed: no

- **LSE-V08** Life-add gates (LSE-R11; boundary: isolated rule)
  - Inputs: `LifeAdd(A, 2560)` with stat 6 = 0x1000, stat 7 = 0x6400, stat 8 = 0, stat 9 = 0x1000;
    each case gives the unit's type, mode, states and exclusion condition, otherwise the preamble player.
  - Seeds: `no draw`
  - Draws: none
  - Trace: per case.
  - Outputs: player mode 1: stat 6 = 0x1A00, returns 2560.
    (a) Player mode 17: no write; returns 0. (b) Monster mode 12: no write.
    (c) Player mode 1 with `aliveExclusionCondition` true: no write.
    (d) Player mode 1 holding state 92: no write. (e) Monster mode 17: stat 6 = 0x1A00.
    (f) Unit type 2, mode 1, `aliveExclusionCondition` false: no write. (g) Player mode 17, `ManaAdd(A, 0x500)` with stat 8 = 0: stat 8 = 0x500.
    (h) Player mode 1 holding state 92 with `aliveExclusionCondition` true, `ManaAdd(A, 0x500)` with
    stat 8 = 0: stat 8 = 0x500.
  - Discriminates: (e) rejects one not-alive mode set shared by players and monsters; (g) and (h)
    reject gating mana by mode, state 92 or the exclusion condition.
  - runtime_observed: no

- **LSE-V09** Overlay choice and draw (LSE-R12, LSE-R08)
  - Inputs: `R.lifeLeech` 10, `R.manaBurn` 5, `R.physical` 25600.
  - Seeds: `A.seed = {1, 0}`
  - Draws: (A, 2, 1)
  - Trace: mana 1280, life 2560; both values nonzero: step gives `{0x6AC690C5, 0}`, result 1.
  - Outputs: `A` stat 8 = 0x500, stat 6 = 0x1A00; overlay 152; `A.seed = {0x6AC690C5, 0}`; record
    640 and 320. For the main case, the ordered unit-write log is exactly `(A, 8, 0x500, 0)`, then
    `(A, 6, 0x1A00, 0)`.
    (a) `R.manaBurn` 0: no draw; `A.seed = {1, 0}`; overlay 151; stat 8 = 0; stat 6 = 0x1A00.
    (b) `R.lifeLeech` 0: no draw; stat 8 = 0x500; overlay 152.
    (c) `A.seed = {0, 2}`: draw (A, 2, 0), seed `{2, 0}`; overlay 151; `A` stat 8 = 0x500, stat 6 =
    0x1A00.
    (d) `A.seed = {1, 1}`: step gives lo 0x6AC690C6, hi 0; draw (A, 2, 0); overlay 151; `A.seed =
    {0x6AC690C6, 0}`; `A` stat 8 = 0x500, stat 6 = 0x1A00.
  - Discriminates: (a) and (b) reject drawing for one leech kind; (c) fixes result 0 to overlay 151;
    (d) rejects using the old low word (odd, overlay 152). The main case's ordered write log
    rejects invoking the local life add before the local mana add even if final stats agree.
  - runtime_observed: no

- **LSE-V10** Draw with zero amounts and shift wrap (LSE-R12, LSE-R08, LSE-R03)
  - Inputs: `A` monster class 271 (`unitGate` 1), stats 6 = 0x100, 8 = 0x100; `R.lifeLeech` 10,
    `R.manaBurn` 5, `R.physical` 1.
  - Seeds: `A.seed = {1, 0}`
  - Draws: (A, 2, 1)
  - Trace: 640, 320; mana `mulDiv(320, 1, 100)` = 3, / 64 = 0, no write, `manaFlag` true; life
    `mulDiv(640, 1, 100)` = 6, / 64 = 0, no write; draw.
  - Outputs: stats unchanged; `A.seed = {0x6AC690C5, 0}`; overlay 152.
    (a) Player `A`, `R.lifeLeech` 0x04000000, `R.manaBurn` 0, `R.physical` 25600: the shift wraps to
    0; `R.lifeLeech` 0; no add, no draw, no overlay.
  - Discriminates: the main case rejects drawing only for nonzero amounts; (a) rejects testing the
    unshifted input.
  - runtime_observed: no

- **LSE-V11** Overlay write requests (LSE-R13)
  - Inputs: `SetOverlay(A, 152)`; `A.statList` present with `hasStateStorage` true and primary chain
    `Lx` (`overlayListCondition` false); the base list's own `overlayListCondition` is false.
  - Seeds: `no draw`
  - Draws: none
  - Trace: no overlay list found on the chain; new list.
  - Outputs: log `NewList(A.pool, conditions {overlayListCondition}, expiry 0, owner 0/7)`, `Attach(A, L1, 1)`,
    `overlayPendingCondition` set on `A.statList`, `SetStat(L1, 178, 152, 0)`, `MarkForUpdate(A)`.
    (a) Chain `Lx` (false), `Ly` (true), `Lz` (true): no allocation; `overlayPendingCondition` set;
    `SetStat(Ly, 178, 152, 0)`; `MarkForUpdate(A)`.
    (b) `hasStateStorage` false, chain `Ly` (true): `NewList(A.pool, conditions {overlayListCondition}, expiry 0, owner 0/7)`,
    `Attach(A, L1, 1)`, `overlayPendingCondition` set, `SetStat(L1, 178, 152, 0)`, `MarkForUpdate(A)`.
    (c) `A.statList` absent, supplied `Attach` creates no stat list: `NewList(A.pool, conditions {overlayListCondition}, expiry 0,
    owner 0/7)`, `Attach(A, L1, 1)`, no `overlayPendingCondition`, `SetStat(L1, 178, 152, 0)`,
    `MarkForUpdate(A)`.
    (d) Id 1000: no request. (e) Id -1: no request.
    (f) `overlayCount` 70000, `SetOverlay(A, 65687)`, chain `Lx` (false): `NewList(A.pool, conditions {overlayListCondition},
    expiry 0, owner 0/7)`, `Attach(A, L1, 1)`, `overlayPendingCondition` set, `SetStat(L1, 178, 151,
    0)`, `MarkForUpdate(A)`.
    (g) `A.statList` absent; this case's supplied `Attach` gives `A` a stat list `S1`: `NewList(A.pool,
    conditions {overlayListCondition}, expiry 0, owner 0/7)`, `Attach(A, L1, 1)`, `overlayPendingCondition` set on `S1`,
    `SetStat(L1, 178, 152, 0)`, `MarkForUpdate(A)`.
    (h) `A.statList` present with `hasStateStorage` true, its own `overlayListCondition` true, chain
    `Lx` (false): no list found; `NewList(A.pool, conditions {overlayListCondition}, expiry 0, owner 0/7)`, `Attach(A, L1, 1)`,
    `overlayPendingCondition` set, `SetStat(L1, 178, 152, 0)`, `MarkForUpdate(A)`.
    (i) `SetOverlay(A, 0)`, `overlayCount` 1000, primary chain `Ly` (`overlayListCondition` true):
    no allocation or attach; `overlayPendingCondition` set, `SetStat(Ly, 178, 0, 0)`,
    `MarkForUpdate(A)`.
  - Discriminates: (a) rejects taking the first list rather than the first flagged one; (b) rejects
    searching without state storage; (d) fixes the bound at `id >= overlayCount`; (f) rejects writing
    the full 32-bit id; (g) rejects testing the stat list before `Attach` without reading it again;
    (h) rejects treating the base list as a candidate; (i) rejects refusing overlay id 0 when
    the supplied count is positive.
  - runtime_observed: no

- **LSE-V12** Monster-attacker branch (LSE-R04, LSE-R14)
  - Inputs: `A` monster class 0, id 9, stats 6 = 0x1000, 7 = 0x2000; `D` player, stats 8 = 0x300,
    10 = 0x100; `R.lifeLeech` 10, `R.manaBurn` 5, `R.staminaBurn` 0x200, `R.physical` 0x800.
  - Seeds: `A.seed = {1, 0}`
  - Draws: none in every case except (g), (p) and (q), each of which draws (A, 2, 1)
  - Trace: 640, 320; apparent type 1; `R.lifeLeech` 10, `R.manaBurn` 5; total = min(10, 2048) +
    min(5, 768) + min(512, 256) = 271; room 0x1000; `AddStat(A, 6, 271, 0)`.
  - Outputs: `A` stat 6 = 0x110F; record 10 and 5; overlay 151; `A.seed = {1, 0}`; `D` unchanged.
    (a) `D` monster with `Drain` 50, stats 8 = 0x300, 10 = 0x100: total 271; `mulDiv(50, 271, 100)` =
    135; `A` stat 6 = 0x1087.
    (b) `A` stat 6 = 0x1F80: room 128; `A` stat 6 = 0x2000.
    (c) `A` stat 6 = 0x2000: room 0; no write, no overlay; record 10 and 5.
    (d) `R.lifeLeech` 0, `R.staminaBurn` 0, `D` stat 8 = 0: total 0; no write, no overlay;
    `R.manaBurn` 5.
    (e) `R.physical` 4: total 4 + 5 + 256 = 265; `A` stat 6 = 0x1109.
    (n) `R.physical` 0: total min(10, 0) + 5 + 256 = 261; `A` stat 6 = 0x1105; overlay 151.
    (f) `A` null: type 6; record 10 and 5; total 271; room 0; no write, no overlay.
    (g) `A.statList` present with `hasStateStorage` true; `A.disguiseCondition` true; `formStateList` = [140, 141]; `A` holds states 140 and 141;
    fixture `gfxtype` 0 for state 140 and 2 for state 141: apparent type 0; base path with record
    640 and 320; mana `mulDiv(320, 2048, 100)` = 6553, / 64 = 102; life `mulDiv(640, 2048, 100)` =
    13107, / 64 = 204; `A` stat 8 = 102, stat 6 = 0x10CC; draw (A, 2, 1), `A.seed = {0x6AC690C5,
    0}`; overlay 152.
    (h) `A.statList` present with `hasStateStorage` true; `A.disguiseCondition` true; `formStateList` = [140, 141]; `A` holds states 140 and 141;
    fixture `gfxtype` 1 for state 140 and 2 for state 141: apparent type 1; record 10 and 5; total
    271; `A` stat 6 = 0x110F; overlay 151; no draw.
    (i) `A.statList` present with `hasStateStorage` true; `A.disguiseCondition` true; `formStateList` = [140, 141]; `A` holds state 140 only; fixture
    `gfxtype` 0 for state 140 and 2 for state 141: apparent type 1; `A` stat 6 = 0x110F; overlay 151;
    no draw.
    (j) `D` monster with `Drain` 50, stats 8 = 0x300, 10 = 0xFF; `R.lifeLeech` 11: `R.lifeLeech` 704
    then 11; total 11 + 5 + 255 = 271; `mulDiv(50, 271, 100)` = 135; `A` stat 6 = 0x1087.
    (k) `D` monster with `Drain` 50, stats 8 = 0x300, 10 = 0x100; `A` stat 6 = 0x1F80: total 271;
    scaled 135; room 128; total 128; `A` stat 6 = 0x2000.
    (l) `A.statList` present with `hasStateStorage` true; `A.disguiseCondition` false; `formStateList` = [140, 141]; `A` holds states 140
    and 141; fixture `gfxtype` 0 for state 140 and 2 for state 141: apparent type 1; `A` stat 6 =
    0x110F; overlay 151; no draw.
    (m) `A` monster class 0 in mode 12 (`dead`), `aliveExclusionCondition` true, holding state 92:
    total 271; `A` stat 6 = 0x110F; overlay 151.
    (o) `D` monster with `Drain` 50, stats 8 = 0x300, 10 = 0x100; `R.lifeLeech` 100099, `R.manaBurn`
    0, `R.staminaBurn` 0, `R.physical` 0x20000; `A` stat 7 = 0x40000: `R.lifeLeech` 6406336 then
    100099; total 100099; `mulDiv(50, 100099, 100)` = 50049; room 0x3F000; `A` stat 6 = 0xD381.
    (p) Monster `A`, `A.game` difficulty 2 with the Hell divisors 3 and 3, `G` difficulty 0;
    `A.statList` present with `hasStateStorage` true; `A.disguiseCondition` true;
    `formStateList` = [140]; `A` holds state 140, whose fixture `gfxtype` is 2: apparent type 0;
    no player divisors; record 640 and 320; mana 102, life add 204; `A` stat 8 = 102, stat 6 =
    0x10CC; draw (A, 2, 1); `A.seed = {0x6AC690C5, 0}`; overlay 152.
    (q) `A.type` 3, `A.statList` present with `hasStateStorage` true; `A.disguiseCondition` true;
    `formStateList` = [140]; `A` holds state 140 with fixture `gfxtype` 2: apparent type 0;
    record 640 and 320; mana add 102; life add amount 204 is refused because the actual unit is
    type 3 (LSE-R11); `A` stat 8 = 102, stat 6 = 0x1000; draw (A, 2, 1);
    `A.seed = {0x6AC690C5, 0}`; overlay 152.
    (r) `A` stat 6 = 0x2100, stat 7 = 0x2000: total 271; room -256; total becomes -256, then
    returns; no additive write or overlay; `A` stat 6 stays 0x2100; record 10 and 5; no draw;
    `A.seed = {1, 0}`.
  - Discriminates: 271 rejects dividing `staminaBurn` by 64 (23); (e) shows the physical minimum; (n)
    rejects a base test on this path; (b) the room cap; (f) a null attacker writing; (g) and (h) the first form
    state of kind 1 or 2 in list order deciding the path; (j) rejects applying effectiveness per term
    (134, stat 6 = 0x1086); (k) rejects capping before scaling (64, stat 6 = 0x1FC0); (l) rejects
    ignoring `disguiseCondition` (base path with a draw); (m) rejects applying the life-add gates to
    this branch; (o) rejects the swapped operand order `mulDiv(100099, 50, 100)` = 50000. Case (p)
    rejects applying player divisors after selecting an apparent player form (which would give
    record 213 and 106, mana 33, stat 6 = 0x1044). Case (q) rejects excluding type 3 from form
    selection and using the apparent type for life-add eligibility. Case (r) rejects treating
    negative room as available life or omitting the nonpositive-total return after the cap.
  - runtime_observed: no

- **LSE-V13** Stage order and the drain value (LSE-R15, LSE-R08; placement DPO-R14 to DPO-R18, CSA-R01)
  - Inputs: `R.lifeLeech` 10, `R.manaBurn` 5, `R.physical` 0x30000, `R.absorbHeal` 0, `R.total`
    0x30000, `R.staminaBurn` 0; `D` monster with `Drain` 100 at Normal, stat 6 = 0x40000, stat 8 = 0;
    the life clamp (DPO-R14) already ran and left `R.physical` 0x30000 unchanged. `AbsorbHeal(D, a)`,
    `LifeWrite(D, t)` and the CSA-R01 requests after `ManaDrain` are supplied and log only; `ManaDrain`
    is LSE-R15.
  - Seeds: `A.seed = {1, 0}`
  - Draws: (A, 2, 1)
  - Trace: record 640 and 320; `P` 100 so no scaling; mana `mulDiv(320, 0x30000, 100)` = 629145, / 64
    = 9830, capped at 0x1000; life `mulDiv(640, 0x30000, 100)` = 1258291, / 64 = 19660; draw; then
    `AbsorbHeal(D, 0)`, `LifeWrite(D, 0x30000)`, `ManaDrain(D, 320)`: 0 - 320 < 256, stat 8 = 0; then
    `StaminaDrain(D, 0)`.
  - Outputs: log prefix in order: the leech draw, `AbsorbHeal(D, 0)`, `LifeWrite(D, 0x30000)`,
    `ManaDrain(D, 320)`, `StaminaDrain(D, 0)`; the later CSA-R01 requests are those CSA-V01 lists and
    are not asserted here. `A` stat 8 = 0x1000, stat 6 = 0x5CCC; overlay 152;
    `A.seed = {0x6AC690C5, 0}`; `D` stat 8 = 0.
  - Discriminates: `ManaDrain(D, 320)` rejects draining before leech (which would pass 5); the draw
    preceding `AbsorbHeal` rejects healing before leech; stat 6 = 0x5CCC rejects scaling by an
    effectiveness of 100 anyway (`mulDiv(100, 1258291, 100)` = 1258200, stat 6 = 0x5CCB).
  - runtime_observed: no

- **LSE-V14** Defender mana drain (LSE-R15, LSE-R05)
  - Inputs: `G` and `A.game` difficulty 1 (`ManaStealDivisor` 2); `R.manaBurn` 5, `R.lifeLeech` 0,
    `R.physical` 25600; `D` player, stat 8 = 0x1000.
  - Seeds: `no draw`
  - Draws: none
  - Trace: 320 / 2 = 160; mana `mulDiv(160, 25600, 100)` = 40960, / 64 = 640; overlay 152;
    `ManaDrain(D, 160)`: 0x1000 - 160 = 0xF60.
  - Outputs: `A` stat 8 = 640; `D` stat 8 = 0xF60; drain returns 160.
    (a) `D` stat 8 = 0x150: 176 < 256, stat 8 = 0; returns 0.
    (b) `D` stat 8 = 0x1A0: 256 is kept; returns 160.
    (c) `D` monster with `Drain(N)` 0, stats 8 = 0x1000, 10 = 0: leech returns at LSE-R02; `R.manaBurn` 5; `D` stat 8 = 0xFFB.
    (d) `A` monster class 338: `v` = 320; `A` stat 8 = 1280; `D` stat 8 = 0xEC0.
    (e) `A` monster class 0, `D` stat 10 = 0: LSE-R14 total 5; `A` stat 6 = 0x1005; `v` = 5; `D` stat
    8 = 0xFFB.
    (f) `ManaDrain(D, 0)` alone: no write; returns 0.
    (g) `A` monster class 0, `G` and `A.game` difficulty 0, `R.lifeLeech` 0, `R.manaBurn` 0x02000000,
    `D` player with stats 8 = 0x1000, 10 = 0: the shift wraps to -2147483648; LSE-R14 divides to
    -33554432; total -33554432 is nonzero, scaled unchanged, below the room, then `<= 0`: return;
    `v` = -33554432; no drain write; `D` stat 8 = 0x1000; drain returns 0.
    (h) `ManaDrain(D, 1)` alone with `D` stat 8 = -2147483648: `low32(-2147483648 - 1)` =
    2147483647, not below 256; `D` stat 8 = 2147483647; drain returns 1.
  - Discriminates: (b) fixes the floor test at `< 256`; (c), (d), (e) and (g) reject draining the
    unrewritten input on the paths that rewrite it ((g) would drain to 0). Case (h) rejects
    widening or saturating the subtraction before the floor test (which would write 0 and return 0).
  - runtime_observed: no

### 8.2 Observation vectors

Both are pending; procedures are in the
[companion](../research/spec-evidence/combat-leech-and-drain.md#observation-procedures).

- **LSE-O01** Leech draw count: across two otherwise identical hits, one with both life and mana leech
  and one with life leech only, the first steps the attacker seed exactly once more.
- **LSE-O02** Defender mana drain: a player with mana leech `M` on Normal lowers the defender's mana
  by `M << 6` raw units, or to 0 when the result is below 256.

## 9. Acceptance and dependencies

Partial acceptance: every assertion of LSE-V01 to LSE-V14 for LSE-R01 to LSE-R05 and LSE-R07 to
LSE-R15 within the stated supplied-input boundary, plus the LSE-R06 table record. Assertions of
LSE-V01 and LSE-V13 that the pipeline owns (the total of DPO-R12a and the stage order of DPO-R14 to
DPO-R18 and CSA-R01) are accepted under those owners. Stored effects of
the replacing and additive unit writes, the overlay list attach, the list stat write and the update
mark remain gated by LSE-Q01; passing the vectors with supplied operations does not accept them.
Full reproduction additionally requires accepting those dependency contracts and matching
observations for LSE-O01, LSE-O02 and the verified rules (LSE-Q05).

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) NRF-R01, Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).
NRF-R02 (draws), NRF-R05 (mulDiv), NRF-R08a (raw formats), RLI-R03 (state test), RLI-R02a and
RLI-R02b (stat reads), section 3.2 (damage record);
[core-game-data-fields](core-game-data-fields.md) GDA-R05, GDA-R08a, GDF-R02, GDF-R03, GDA-R13,
GDF-R05a, GDF-R06a, GDF-R08 (fields and names);
[combat-status-application](combat-status-application.md) CSA-R02 (`unitGate`);
[state-timed-state-lifecycle](state-timed-state-lifecycle.md) STL-R04 (expiry condition), STL-R06a,
STL-R06b and STL-R06c (overlay list requests). Stored attach and list-write effects remain gated;
unit replacing/additive effects have no complete consumed contract (LSE-Q01), and update-queue
effects remain RLI-Q02. Unknown disguise and form-list producers (LSE-Q02, STL-R07a) do not gate
the supplied apparent-form inputs.

**Placement and supporting contracts:**
[combat-damage-pipeline-order](combat-damage-pipeline-order.md) DPO-R12a (the total this stage leaves
unchanged), DPO-R13 (apply gate), DPO-R14 (already-clamped physical input), DPO-R15 (leech
placement), DPO-R16 (absorb heal
placement calling this spec's `LifeAdd`), DPO-R17 (current-life write placement) and DPO-R18
(stamina drain following mana drain);
[combat-status-application](combat-status-application.md) CSA-R01 (status hand-off order, whose
first request calls LSE-R15). These references establish surrounding stages; the supplied-input
vectors do not require reproducing their stored effects.

**Depended on by:** [combat-damage-pipeline-order](combat-damage-pipeline-order.md) DPO-R15, DPO-R16,
DPO-R18 (leech, the capped add used by the absorb heal, the drain value);
[combat-status-application](combat-status-application.md) CSA-R01 (`ManaDrain`, LSE-R15).

## 10. Related specs

- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): the hit path around this stage,
  the total, the clamp, the absorb heal, the life write and the stamina drain.
- [combat-status-application](combat-status-application.md): the status hand-off that follows leech
  and begins with the mana drain, and the class-group helper.
- [combat-damage-mitigation](combat-damage-mitigation.md): produces `R.physical` and the total.
- [combat-crushing-blow](combat-crushing-blow.md): uses the same overlay write with another id.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): list construction and attachment
  consumed by the overlay write, and the disguise bookkeeping.
- [combat-hit-recovery](combat-hit-recovery.md): stages after the status stage.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): player-versus-player
  scaling of the record before this stage.
- [core-game-data-fields](core-game-data-fields.md): column, stat, state, mode and class names.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. The original `D2Game.dll` and
`D2Common.dll` specimens are identified in the [source manifest](../research/source-manifest.md).
Static readings of the leech operation, the capped adds, the alive test and the mana drain were
recorded on 2026-10-12 and 2026-10-13 and reread on 2026-10-14; on 2026-10-14 the cap getters, the
overlay helper and its list search and update mark, the additive write and the apparent-form helper
were read, using original instruction listings and, where a byte run had no listing, a hand decode
of the original bytes. Table cells for the divisors, stat 9 and state 92 were read on 2026-10-14
through the running original game's loader. No gameplay observation has been made. The
[public evidence companion](../research/spec-evidence/combat-leech-and-drain.md) gives
module locations, rule-linked rationale, corrections, table records and observation procedures.
Public verification requires matching original game files only.
