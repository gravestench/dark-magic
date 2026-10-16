# Combat Status Application

**Description**

When a hit lands, the game decides whether the target is stunned, chilled, frozen, poisoned or set on
fire. A new chill or freeze only extends an existing effect; stun, poison and burn can replace a longer
effect with a shorter one, while weaker poison and burn are ignored. Some monsters resist stun or
receive chill instead of freeze, and chilling a monster can mark it to shatter if it dies.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: CSA-R, CSA-Q, CSA-V, CSA-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/combat-status-application.md)
Open evidence gaps: [gap file](../plans/gaps/combat-status-application.md)

## 1. Scope

This spec defines how a resolving hit applies timed status effects to its defender: the order in
which the hit-application operation hands its record to the status appliers after the defender's
current-life write, and the final lethal-flag decision that follows them; the stun, cold, freeze,
poison and burn appliers with their gates, length arithmetic, caps, refresh policies, state, list,
stat and timer requests; the three status gate helpers; the attacker stun-cancel draw, the defender
cold-shatter draw and the freeze-to-cold fallback; the freeze tail and the poison/burn monster
pre-step timer-node requests; and the cold expiry callback. Out of scope: everything before the
life write (damage build, mitigation and the life clamp) is
[combat-damage-pipeline-order](combat-damage-pipeline-order.md) and
[combat-damage-mitigation](combat-damage-mitigation.md); leech, the mana and stamina drain arithmetic
and the poison term of the immediate damage total stay in
[combat-leech-and-drain](combat-leech-and-drain.md) and the pipeline; random
stepping, draw forms and the seed-ownership table stay in
[core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), whose NRF-R04a rows for the
stun-cancel, cold-shatter and freeze-fallback sites refer to the rules here; stat reads, state
lookup/test/set routing and the damage record stay in
[core-record-layouts-and-ids](core-record-layouts-and-ids.md); the timer queue, list
construction, attachment, list-stat propagation, expiry, removal, the generic expiry callback and
stay-death stay in [state-timed-state-lifecycle](state-timed-state-lifecycle.md); what the cold and
freeze states do once applied, including the freeze expiry callback (MMF-R09b), is
[state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md); lengths reaching the
record (resistance, cannot-be-frozen) are produced upstream.

## 2. Trigger and ownership

- Module: D2Game (server); category `combat` by call chain (basis in the companion): every applier
  here is called only from the hit-application operation (cold also from the freeze applier), which the
  pipeline owns under `combat`.
- Trigger: once per applied hit, synchronously inside the hit-application operation reached from the
  queued-hit dequeue of melee and skill attacks and from missile contact and callback routes, as
  [combat-damage-pipeline-order](combat-damage-pipeline-order.md) defines. Within that operation the
  status stage runs after the life clamp, leech, absorb heal and the current-life write (DPO-R14,
  DPO-R16, DPO-R17), and before the surviving-monster regeneration restart (DPO-R19c, OWD-R12), the
  durability step, post-apply events, reflection and hit reaction.
- The cold expiry callback is requested later through the lifecycle's expired-list removal path
  (STL-R03, STL-R05). Actual contribution removal and callback invocation remain CSA-Q02; CSA-R13
  defines the callback at supplied entry after the list's stats have been removed.
- Frequency: each applier is requested exactly once per hit-application call, in CSA-R01 order; each
  returns at its own gates.
- Client view: state-change and unit-update message integration remains unresolved under STL-R14;
  the sender and message content belong to a `net-` owner (CSA-Q04).
- The full call chain with original locations is in the
  [companion](../research/spec-evidence/combat-status-application.md#dispatch-chain).

## 3. Inputs and data

Integers use the foundations notation (NRF section 1). Durations and frames are i32 frame counts;
expiry frames are absolute game frames.

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| attacker `A`, defender `D` | units passed by the hit-application operation | association, may be null | Roles are named; no positional order is required |
| unit identity | `unit.type`, `unit.classId`, `unit.id`, `unit.mode` | i32 type/class/mode, u32 id | Types 0 player, 1 monster (RLI section 3.3); a null defender is type 6 in the freeze routing test; modes by GDF-R05a/GDF-R06a |
| seeds | `A.seed`, `D.seed` | pair `{lo,hi}` of u32 | One step is NRF-R01; `r = lo mod 100` is unsigned |
| attacker game | `A.game.frame`, `A.game.difficulty`, `A.game.pool` | i32 frame, u8 difficulty 0..2, allocation context | Every applier takes frame, difficulty and pool from the attacker's game |
| damage record | RLI section 3.1 fields `manaBurn`, `staminaBurn`, `stunLength`, `coldLength`, `freezeLength`, `poison`, `poisonLength`, `burnMagnitude`, `burnLength`, `resultFlags.lethal` | i32 (magnitudes 8.8 per frame, lengths frames); `lethal` is the 0x0002 bit of the u16 `resultFlags` | `manaBurn` holds the value left by the leech stage (LSE-R15); CSA-R01 modifies only the 0x0002 bit |
| monster data | `D.monsterData` present or absent; its 16-bit condition word | association; u16 | `monsterFlag8`: monster data present and condition word AND 8 nonzero; wider meaning CSA-Q03 |
| alive exclusion | `unit.aliveExclusionCondition` | boolean | When true the unit is not alive for CSA-R13 (the same condition STL-R11 tests); wider meaning CSA-Q03 |
| pending timer nodes | `D.eventChain`: ordered nodes, each with `kind` (u8) and successor | ordered list | Walked by CSA-R09 and CSA-R11 |
| MonStats | `MonStats.Velocity`, `MonStats.boss`, `MonStats.ColdEffect` / `ColdEffect(N)` / `ColdEffect(H)`, `MonStats.MonStatsEx`, record count | i16, boolean, i8 per difficulty, i16 index, i32 | Per-class loaded data (GDA-R08a; column names compare case-insensitively, GDA-R03c; the signed i8 reading of `coldeffect` is this consumer's) |
| MonStats2 | `MonStats2.deadCol`, record count | boolean, i32 | Shatter suppression (GDA-R09a fields; GDA-R09b index bounds) |
| difficulty divisors | `DifficultyLevels.MonsterColdDivisor`, `MonsterFreezeDivisor` | i32 | Row Normal, Nightmare or Hell for the attacker game's difficulty 0, 1 or 2 (GDA-R05); this contract covers difficulty 0..2 only (CSA-Q06) |
| state set | `HasState`, `FindStateList` results | RLI-R03 | States 1 `freeze`, 2 `poison`, 11 `cold`, 21 `stunned`, 54 `uninterruptable`, 107 `shatter`, 115 `burning` (GDF-R03) |
| list stat | own stat 74 `hpregen` of an existing poison/burn list | i32 | Read through the list read of RLI-R02a/RLI-R02b |
| defender life | `Stat(D, 6, 0)` | i32 8.8 | Unit read of RLI-R02a/RLI-R02b, after all appliers |
| helper arguments | optional supplied monster record, `mask`, callback state `s` | record association; u32 mask; i32 state | Supplied record takes precedence in `recordGate`; `monsterGate` intersects the low 16 mask bits; callback state comes from its removed list |
| state and stat-read context | base list/state storage, state count; selected stat array, stat-definition count and availability, `fMin`, `MinAccr`, `ValShift` | optional associations; i32 counts, stat values and `MinAccr`; u8 `ValShift`; boolean conditions | Supplied state lookups/tests and list/unit stat reads consume RLI-R03/RLI-R02a/RLI-R02b; these inputs are distinct from list writes |

| Mapping / constant | Origin | Binding |
|---|---|---|
| States 1, 2, 11, 21, 54, 107, 115; stats 6, 67 `velocitypercent`, 68 `attackrate`, 69 `other_animrate`, 74 `hpregen` | Names from original-profile tables (GDF-R02, GDF-R03) | Fixed by original code; custom remapping is not established |
| Timer kinds 2, 3, 12 | Engine code | Fixed by original code; scheduling is STL-R01 and the kind-12 handler STL-R03; the kind-2 and kind-3 handlers are not specified by the lifecycle (kind 2: DPO-R28b, DPO-Q13; kind 3 is also requested by DPO-R19c) and their execution stays gated by CSA-Q02 |
| Stun cancel threshold 90, shatter threshold 20, modulus 100, stun caps 13 and 250, cold minimum 1, cold default effect -50, list flags 2, attach mode 1, recordGate result 64 | Engine code | Fixed by original code. List flags 2 marks the list as timed, so the lifecycle's expiry walk considers it (STL-R04, STL-R06a). Attach mode 1 is the mode that applies all of the list's stats to the unit (STL-R06b, gated: CSA-Q01). recordGate's 64 is the boss bit's value; callers test only nonzero |
| Class groups 271 `roguehire`, 338 `act2hire`, 359 `act3hire`, 560 `act5hire1`, 561 `act5hire2` | Class numbers fixed in code; names from MonStats (GDF-R08) | Fixed numeric bindings, not resolved from editable class names |
| Cold and freeze divisors: Normal 1 and 1, Nightmare 2 and 2, Hell 4 and 4 | `data/global/excel/difficultylevels.txt`, rows `Normal`, `Nightmare`, `Hell`, columns `MonsterColdDivisor`, `MonsterFreezeDivisor` ([table records](../research/spec-evidence/core-game-data-table-records.md#named-cell-records)) | Loaded data; code fixes which consumer divides and the zero-divisor policy |
| `ColdEffect`, `Velocity`, `boss`, `deadCol` per class | MonStats / MonStats2 per-class cells | Loaded data; vectors supply them as named inputs |

Callback identities: `coldExpiryCallback` (CSA-R13), `freezeExpiryCallback` (MMF-R09b) and
`genericStatusCallback` (STL-R11). Dependency operations, each with its owning contract:
`FindStateList`, `HasState`, `SetState` request (RLI-R03; bookkeeping STL-R07, STL-R07a);
`NewList`, `SetListState`, `SetListCallback`, `SetListExpiry`, `GetListExpiry` (STL-R06a);
`Attach` (STL-R06b); replacing list `SetStat` (STL-R06c); `GetStat` / `Stat` (RLI-R02a, RLI-R02b);
`Schedule(game, unit, kind, frame)` (STL-R01); `RemoveTimerNode(game, node)` and
`AnimationRefresh(unit)` (CSA-Q02); `StayDeath(unit, s)` (STL-R12a). The mana and stamina drains
are LSE-R15 and DPO-R18.

## 4. Rules (execution order)

Pseudocode names are illustrative. "Fatal" is a fatal assertion: execution ends with no defined
later state. `step(S)` advances seed `S` once by NRF-R01; `low32` keeps the low 32 bits.

**CSA-R01** **Status hand-off and final lethal decision** [verified in 1.13c code (static)]. After the
defender's current-life write, the hit-application operation requests, in this order, with record `R`:

```
ManaDrain(D, R.manaBurn)                                   // LSE-R15
StaminaDrain(D, R.staminaBurn)                             // DPO-R18
Stun(A, D, R.stunLength)                                   // CSA-R03, CSA-R04
Cold(A, D, R.coldLength)                                   // CSA-R05, CSA-R06
Freeze(attacker = A, defender = D, length = R.freezeLength) // CSA-R07 to CSA-R09
Poison(A, D, magnitude = R.poison, length = R.poisonLength)               // CSA-R10 to CSA-R12
Burn(A, D, magnitude = R.burnMagnitude, length = R.burnLength)            // CSA-R10 to CSA-R12
life = Stat(D, 6, 0)                                       // RLI-R02a, RLI-R02b
if life > 0 (signed): R.resultFlags.lethal = false
else:                 R.resultFlags.lethal = true
```

No applier returns a value the caller uses. The stage makes no direct draw; its only draws are the
attacker stun-cancel step (CSA-R03), the defender shatter step of every cold application that reaches
it (CSA-R06), including a freeze that falls back to cold (CSA-R07). Poison, burn and the two drains
draw nothing.

**CSA-R02** **Status gate helpers** [verified in 1.13c code (static)]. None draws or mutates.

- `recordGate(record, unit)`: 0 unless `unit` is a monster. Use `record` when non-null, else the
  loaded MonStats record of `unit.classId`; an invalid or unavailable record gives 0. Return 64 when
  that record's `boss` is set, else 0.
- `monsterGate(unit, mask)`: 0 for a null unit, a nonmonster or absent monster data; otherwise 1
  when the 16-bit condition word AND the low 16 bits of `mask` is nonzero, else 0. `monsterFlag8`
  (section 3) is `monsterGate(D, 8) != 0`.
- `unitGate(unit)`: 0 for a null unit or a nonmonster; monster class 271 returns 1, 338 returns 2,
  359 returns 3, 560 or 561 returns 4, every other class 0.

**CSA-R03** **Stun entry, monster checks and attacker cancel draw** [verified in 1.13c code (static)].

```
Stun(A, D, len):
    if len <= 0: return                                    // no draw, no request
    if D != null and D.type == 1:
        if monsterFlag8(D):
            step(A.seed); r = A.seed.lo mod 100            // attacker's seed
            if r < 90: return                              // stun cancelled
        if recordGate(null, D) != 0: return
        if MonStats(D.classId).Velocity == 0: return
        if unitGate(D) != 0 and len >= 13: len = 13; continue at CSA-R04
    if len > 250: len = 250
    continue at CSA-R04
```

The checks run in that order; each later check runs only if the earlier ones did not return. A
nonmonster defender skips every monster check and uses only the 250 cap. A monster class with no
loaded MonStats record has no guard at the Velocity read (a fault, not a successful path). With a
null attacker the cancel step faults when `monsterFlag8(D)` holds; otherwise the null attacker reaches
CSA-R04's assertion.

**CSA-R04** **Stun list requests** [verified in 1.13c code (static)].

```
    L = FindStateList(D, 21)                               // a null D faults here (RLI-R03)
    if A == null: fatal
    g = A.game; expiry = low32(g.frame + len)
    if L == null:
        L = NewList(g.pool, flags 2, expiry, owner A.type, A.id)
        Schedule(g, D, 12, expiry)
        SetListState(L, 21); SetListCallback(L, genericStatusCallback)
        Attach(D, L, 1)
        SetState(D, 21, on)                                // after attach
    else:
        SetListExpiry(L, expiry)                           // unconditional: may shorten
        Schedule(g, D, 12, expiry)
```

No list stat is written. No draw occurs in this rule.

**CSA-R05** **Cold length, expiry and list requests** [verified in 1.13c code (static)].

```
Cold(A, D, len):
    if len <= 0: return                                    // no draw
    L = FindStateList(D, 11)                               // a null D faults here
    if A == null: fatal
    g = A.game
    E = ColdEffect(D, g)
    if E == 0: return                                      // no request, no draw
    if D.type == 1 and E < 0:
        div = DifficultyRecord(g.difficulty).MonsterColdDivisor
        if div != 0: len = len / div                       // signed, truncating
    if len <= 1: len = 1
    expiry = low32(g.frame + len)
    if L == null:
        SetState(D, 11, on)
        L = NewList(g.pool, flags 2, expiry, owner A.type, A.id)
        Schedule(g, D, 12, expiry)
        SetListState(L, 11); SetListCallback(L, coldExpiryCallback)
        Attach(D, L, 1)
        SetStat(L, 67, E, 0); SetStat(L, 68, E, 0); SetStat(L, 69, E, 0)
        AnimationRefresh(D)
    else if GetListExpiry(L) < expiry (signed):
        SetListExpiry(L, expiry); Schedule(g, D, 12, expiry)
    continue at CSA-R06                                    // on every path reaching here
```

`ColdEffect(unit, g)` is -50 when the unit is null, not a monster, has a negative class or a class at
or beyond the MonStats count, or has no loaded record; otherwise it is the signed MonStats cold effect
column for `g.difficulty` (`ColdEffect`, `ColdEffect(N)`, `ColdEffect(H)` for 0, 1, 2). An existing
list's stats are never rewritten; an equal or earlier candidate expiry makes no expiry write and no
timer request. Cold has no maximum length.

**CSA-R06** **Cold shatter draw** [verified in 1.13c code (static)]. Every cold application that passed
`len > 0` and `E != 0` (new list, extension or unchanged expiry) ends with:

```
    step(D.seed); r = D.seed.lo mod 100                    // defender's seed
    if r < 20 and D.type == 1 and not deadCol(D.classId):
        SetState(D, 107, on)
    else:
        SetState(D, 107, off)
```

`deadCol(class)` is true only when the class has a loaded MonStats record whose MonStatsEx index
selects a valid MonStats2 record with `deadCol` set; every unavailable lookup gives false, so the
shatter request can still be on.

**CSA-R07** **Freeze routing and cold fallback** [verified in 1.13c code (static)].

```
Freeze(attacker A, defender D, length len):
    if len <= 0: return
    t = D.type if D != null else 6
    if t == 0: Cold(A, D, len); return                     // player: cold with undivided length
    if t == 1:
        if HasState(D, 54) != 0: return                    // nothing applied, no draw
        if recordGate(MonStats(D.classId) or null, D) != 0
           or monsterGate(D, 8) != 0
           or unitGate(D) != 0:
            Cold(A, D, len); return                        // first nonzero ends the tests
    continue at CSA-R08                                    // monsters past the gates and other types
```

The fallback is the full cold applier (CSA-R05, CSA-R06), so it takes one defender shatter step when
its `E != 0`.

**CSA-R08** **Freeze length, expiry and list requests** [verified in 1.13c code (static)].

```
    L = FindStateList(D, 1)                                // a null D faults here
    if A == null: fatal
    g = A.game
    if ColdEffect(D, g) >= 0: return
    if D.type == 1:
        div = DifficultyRecord(g.difficulty).MonsterFreezeDivisor
        if div == 0: fatal
        len = len / div                                    // signed, truncating; no minimum, may be 0
    expiry = low32(g.frame + len)
    if L == null:
        SetState(D, 1, on)
        L = NewList(g.pool, flags 2, expiry, owner A.type, A.id)
        SetListState(L, 1); SetListCallback(L, freezeExpiryCallback)
        Attach(D, L, 1)
    else:
        old = GetListExpiry(L)
        if old < expiry (signed): SetListExpiry(L, expiry)
        else: expiry = old
    Schedule(g, D, 12, expiry)                             // also when the expiry did not change
    continue at CSA-R09
```

No list stat is written. A requested frame at or before the current frame is queued for the next frame
by STL-R01, so a zero adjusted length gives a list expiring at `g.frame` with its timer queued at
`g.frame + 1`.

**CSA-R09** **Freeze tail: type assertion and kind-2 timer requests** [verified in 1.13c code (static)].

```
    if D == null or D.type != 1: fatal                     // after CSA-R08's requests
    node = D.eventChain.head
    while node != null:
        next = node.successor                              // saved before any request
        if node.kind == 2: RemoveTimerNode(g, node)
        node = next
    Schedule(g, D, 2, low32(g.frame + len + 1))            // len after division
```

No draw occurs.

**CSA-R10** **Poison and burn entry gates and field roles** [verified in 1.13c code (static)]. Poison uses
state `s = 2` with `mag = R.poison`, `len = R.poisonLength`; burn uses `s = 115` with
`mag = R.burnMagnitude`, `len = R.burnLength`. The two appliers are otherwise identical:

```
    if len <= 0: return                                    // length tested first
    if mag <= 0: return
    L = FindStateList(D, s)                                // a null D faults here
    if A == null: fatal
    g = A.game
    continue at CSA-R11, then CSA-R12
```

**CSA-R11** **Poison and burn monster pre-step** [verified in 1.13c code (static)]. Before any list
decision:

```
    if D != null and D.type == 1:
        node = D.eventChain.head
        while node != null:
            next = node.successor
            if node.kind == 3: RemoveTimerNode(g, node)
            node = next
        Schedule(g, D, 3, low32(g.frame + 1))
```

Nonmonster defenders skip the pre-step. The pre-step runs on every monster application that passed
CSA-R10, including one CSA-R12 then rejects. Every scheduled node is linked at the head of the
unit's chain (scheduling request STL-R01; the head-first link is verified at the scheduler's node
allocation, [companion record](../research/spec-evidence/combat-status-application.md#correction-nodes-scheduled-earlier-in-the-same-hit-csa-r11)), so within one hit the burn pre-step finds and requests removal of the
kind-3 node the poison pre-step scheduled.

**CSA-R12** **Poison and burn list requests** [verified in 1.13c code (static)].

```
    expiry = low32(g.frame + len)
    if L == null:
        SetState(D, s, on)
        L = NewList(g.pool, flags 2, expiry, owner A.type, A.id)
        Schedule(g, D, 12, expiry)
        SetStat(L, 74, low32(-mag), 0)
        SetListState(L, s); SetListCallback(L, genericStatusCallback)
        Attach(D, L, 1)
    else:
        cur = GetStat(L, 74, 0)                            // own list stat, RLI-R02a/RLI-R02b
        if i32(low32(-cur)) > mag: return                  // a stronger existing effect wins
        SetListExpiry(L, expiry)                           // unconditional: may shorten
        Schedule(g, D, 12, expiry)
        SetStat(L, 74, low32(-mag), 0)
```

Equal magnitude replaces. No draw occurs.

**CSA-R13** **Cold expiry callback** [verified in 1.13c code (static)]. At supplied callback entry, use
the unit and removed list's state `s` (11 for every list tagged with this callback), with that list's
stats already removed. Actual removal and callback invocation through STL-R05 remain CSA-Q02:

```
ColdExpiry(unit, s):
    alive = unit != null and not unit.aliveExclusionCondition and
            ((unit.type == 0 and unit.mode not in {0, 17}) or
             (unit.type == 1 and unit.mode not in {0, 12}))
    if not alive and StayDeath(unit, s) != 0: return       // STL-R12a
    SetState(unit, s, off)
    if alive: SetState(unit, 107, off)
    AnimationRefresh(unit)
```

`StayDeath` is consulted only when the unit is not alive. The callback does not remove the list and
makes no draw. The lifecycle calls a list's callback only when the list has a unit (STL-R05), so the
callback is not reached with a null unit on that path; a direct call with a null unit would consult
`StayDeath(null, s)` and, when it returns 0, request a state change on a null unit, which
RLI-R03 leaves outside the successful boundary.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| CSA-R01 | Hand-off order, arguments and final lethal decision | verified in 1.13c code (static) | CSA-Q01, CSA-Q02: stored applier effects only; supplied hand-off and lethal decision eligible | CSA-V01, CSA-V02, CSA-V03 |
| CSA-R02 | Boss, monster-flag and class-group gate helpers | verified in 1.13c code (static) | none | CSA-V04 |
| CSA-R03 | Stun monster checks, caps and attacker cancel draw | verified in 1.13c code (static) | none | CSA-V05, CSA-V06, CSA-V07 |
| CSA-R04 | Stun list requests, expiry replacement | verified in 1.13c code (static) | CSA-Q01: stored attach/state effects; ordered requests eligible | CSA-V05, CSA-V07 |
| CSA-R05 | Cold effect, monster divisor, minimum 1, extend-only requests | verified in 1.13c code (static) | CSA-Q01, CSA-Q02: stored attach/stat/state and refresh effects; ordered requests eligible | CSA-V08, CSA-V09 |
| CSA-R06 | Defender shatter draw and shatter on/off request | verified in 1.13c code (static) | CSA-Q01: stored state bookkeeping; draw and request eligible | CSA-V08, CSA-V09, CSA-V03 |
| CSA-R07 | Freeze player/state-54/gate routing to cold | verified in 1.13c code (static) | none | CSA-V10, CSA-V03 |
| CSA-R08 | Freeze divisor without minimum, extend-only list, timer always requested | verified in 1.13c code (static) | CSA-Q01: stored attach/state effects; ordered requests eligible | CSA-V11 |
| CSA-R09 | Freeze tail assertion, kind-2 removal and kind-2 timer requests | verified in 1.13c code (static) | CSA-Q02: node removal and handler effects; ordered requests eligible | CSA-V11 |
| CSA-R10 | Poison/burn gates and field roles | verified in 1.13c code (static) | none | CSA-V13, CSA-V14 |
| CSA-R11 | Poison/burn monster kind-3 pre-step | verified in 1.13c code (static) | CSA-Q02: node removal and handler effects; ordered requests eligible | CSA-V12, CSA-V02 |
| CSA-R12 | Poison/burn magnitude comparison and replacement | verified in 1.13c code (static) | CSA-Q01: stored attach/stat effects; ordered requests and comparison eligible | CSA-V12, CSA-V13, CSA-V14 |
| CSA-R13 | Cold expiry callback | verified in 1.13c code (static) | CSA-Q01: disguise/update effects; CSA-Q02: actual free/callback and refresh effects; supplied-entry requests eligible | CSA-V15 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| Ordered drain and applier requests; `resultFlags.lethal` cleared when final life is positive, set otherwise (the same decision DPO-R19b describes) | CSA-R01 |
| Gate helper results: recordGate 0 or 64, monsterGate 0 or 1, unitGate 0 to 4; no mutation | CSA-R02 |
| Attacker seed stepped once (flagged monster defender, positive stun length) | CSA-R03 |
| Stun list allocation, kind-12 timer at `frame + len`, tag 21, generic callback, attach mode 1, state 21 on; or expiry replacement and kind-12 timer | CSA-R04 |
| Cold state 11 on, list allocation, kind-12 timer, tag 11, cold callback, attach mode 1, list stats 67, 68, 69 = `E`, animation refresh; or expiry extension and kind-12 timer | CSA-R05 |
| Defender seed stepped once per cold application reaching the draw; state 107 on or off request | CSA-R06 |
| Cold applier invoked with the undivided freeze length | CSA-R07 |
| Freeze state 1 on, list allocation, tag 1, freeze callback, attach mode 1; expiry extension; kind-12 timer at the resulting expiry | CSA-R08 |
| Kind-2 node removal requests; kind-2 timer at `frame + len + 1` | CSA-R09 |
| Kind-3 node removal requests; kind-3 timer at `frame + 1` (monster defenders) | CSA-R11 |
| Poison/burn state on, list allocation, kind-12 timer, list stat 74 = `-mag`, tag, generic callback, attach mode 1; or expiry replacement, kind-12 timer and stat 74 replacement | CSA-R12 |
| Cold expiry: state off, state 107 off when alive, animation refresh | CSA-R13 |
| Fatal assertion: null attacker reaching a list lookup's assertion (stun, cold, freeze, poison, burn), zero freeze divisor for a monster, a freeze reaching CSA-R09 on a non-monster; faults: null defender at a state-list lookup, stun Velocity read for a class without a record, cancel step on a null attacker | CSA-R03 to CSA-R05, CSA-R08 to CSA-R10 |
| Messages: no direct send; lifecycle/update-message integration is unresolved (CSA-Q04) | none |

Requested dependency effects become stored or queued effects only under the owning dependency's
acceptance (CSA-Q01, CSA-Q02).
The gate helpers return the values listed above; no applier return value is consumed by CSA-R01.
The remaining listed list, stat, state, refresh and removal effects are ordered requests at this
boundary, with stored membership established only by the verified portion of STL-R07 and supplied
fixture operations. Disguise and update-queue integration remain gated.

## 6. Branches and edge cases

- Defender kinds: players never take the stun monster checks, never divide cold, and always turn a
  freeze into cold. Monsters with `monsterFlag8` lose 90% of stuns before the other checks and never
  freeze (cold instead). Boss-record monsters are never stunned and never frozen (cold instead).
  Class-group monsters cap stuns at 13 and are never frozen. State 54 blocks freeze entirely but not
  stun or cold. Units of other types take the generic paths and a freeze on them ends at CSA-R09's
  fatal assertion after its list requests.
- Refresh policies: cold and freeze extend only (strictly later candidate); freeze still requests its
  kind-12 timer at the retained expiry. Stun replaces the expiry even when shorter. Poison and burn
  ignore a weaker application and replace the expiry on an equal or stronger one, even when shorter.
- Zero and negative inputs: a nonpositive length returns before any lookup or draw; poison and burn
  also require a positive magnitude, tested after the length. `E == 0` disables cold and freeze; a
  positive `E` disables freeze but still applies cold (no monster division, since the division needs
  `E < 0`).
- Divisors: a zero cold divisor skips the division; a zero freeze divisor is fatal for a monster.
  Cold's minimum 1 applies after division; freeze has no minimum.
- Shatter: requested on only for a monster on a roll below 20 without `deadCol`; every other cold
  application that reaches the draw requests it off, even for a player.
- Difficulty enters only through `ColdEffect`, the two divisors and the attacker's game; there is no
  classic/expansion, melee/missile or PvP branch in these rules (the record is already scaled).
- Missing data: an invalid or unavailable monster record gives cold effect -50 and `deadCol` false,
  and recordGate 0; the stun Velocity read has no guard.
- Overflow: expiry and timer frames wrap with `low32`; magnitudes and lengths are not otherwise
  clamped.

## 7. Open questions

| Question | Kind / status | Evidence that would settle it |
|---|---|---|
| **CSA-Q01** | integration / open | Stored effects of `Attach`, replacing list `SetStat`, disguise and state-update bookkeeping where STL-R06b, STL-R06c, STL-R07a and RLI-Q02 are gated: accept those contracts with integration vectors. Ordered requests, arithmetic and the verified membership subset of STL-R07 remain available. |
| **CSA-Q02** | integration / open | Animation refresh effects; active/deferred frame-event node removal; kind-2, kind-3 and kind-12 handler execution; actual list contribution removal and callback invocation through unresolved STL-R05; and the scheduler's own handling of a kind-2 request for a monster holding state 54, which the freeze path can reach only if a dependency sets that state after CSA-R07's test. Accept the lifecycle and animation owners' contracts. Direct supplied-entry callback requests remain available. |
| **CSA-Q03** | outside-boundary / open | Producer and wider meaning of the monster condition word's bit 8 (`monsterFlag8`) and of `aliveExclusionCondition`: trace their writers in monster spawning and unit lifecycle owners. The tested branches here are verified. |
| **CSA-Q04** | outside-boundary / open | Which server-to-client messages carry these state and stat changes: trace the update sender in a `net-` spec. |
| **CSA-Q05** | observation / pending | Matching gameplay captures of CSA-O01, CSA-O02 and the verified rules' vectors; no gameplay observation has been made. |
| **CSA-Q06** | outside-boundary / open | Whether any caller can supply a game difficulty outside 0..2. The divisor lookup's handling of such a value and the unguarded per-difficulty cold-effect read are not part of this contract; establish the difficulty producer's range in its owner. |

The [gap file](../plans/gaps/combat-status-application.md) records the attempts and exact blocked
assertions.

## 8. Test vectors

### 8.1 Executable vectors

Shared defaults, overridden only where a vector says so. All vectors are supplied-input operation
fixtures (boundary: supplied-input stage) unless a vector states another boundary.

Case structure: every vector and every lettered case starts from the vector's Inputs and Seeds, with
only the inputs the case names overridden. A case never takes another case's state or outputs; any
existing list, expiry, state or stepped seed a case needs is listed among that case's own inputs, and
every case states its own outputs.

- Units: `A` is a non-null player, type 0, id 7. `A.game` has frame 1000, difficulty 0 (Normal) and
  pool `P`. `D` is a distinct non-null unit with id 9. A monster `D` is class 0, mode 1, monster data present
  with `monsterFlag8` false, `aliveExclusionCondition` false; a player `D` is mode 1. No states, lists
  or pending timer nodes unless stated; state count 185.
- Data: MonStats count 1; the class 0 record has `boss` false, `Velocity` 1, `ColdEffect` -50 at every
  difficulty unless the vector states another value, and `MonStatsEx` 0; MonStats2 count 1, and its
  record 0 has `deadCol` false unless the vector states true. Event chains are listed head first.
  Divisors are the Normal, Nightmare and Hell values of section 3. Stat definitions count 359;
  definition 74 is available with `fMin` false, `MinAccr` 0, `ValShift` 0. Every stated existing
  list has `usesResolvedStats` false and holds its stated stats in its own array, so `GetStat(L, 74,
  0)` returns the stated value unchanged.
- Operations: these supplied operations persist their stored effects within one vector and log
  their semantic arguments in call order. `FindStateList(D, s)` logs and returns the list attached to
  `D` for state `s` (a list stated in the case's inputs, or one attached earlier in the same case), else null.
  `NewList` succeeds and returns a new unattached list holding its owner and expiry. `SetListState`
  and `SetListCallback` store the tag and callback; `SetListExpiry` stores the expiry; `GetListExpiry`
  returns the stored expiry; `SetStat` stores the list stat (replacing); `Attach(D, L, 1)` attaches
  `L` so later lookups on `D` find it; `SetState` sets or clears the state bit; `AnimationRefresh`
  logs only. `RemoveTimerNode` unlinks the node from `D`'s chain and clears the removed node's
  successor. `Schedule` logs `(kind, requested frame)` and its queue frame by STL-R01 (requested frame
  at or before the current frame queues at current + 1) and, as the original scheduler does (CSA-R11),
  links a new node of that kind at the head of the scheduled unit's event chain. New nodes are named by kind and requested
  frame, for example `K3@1001`.
  `GetListExpiry`, `HasState`, the gate helpers and data reads (MonStats, MonStats2, difficulty record,
  `ColdEffect`) are not logged; their order is therefore not asserted by these vectors. None of these
  operations draws. Supplied drains and, where a vector says so, supplied appliers only log.
- Unit stat reads: `Stat(D, 6, 0)` reads `D`'s unit stat list (`D.statList`), as RLI-R02a defines;
  definition 6 is available with `fMin` false.
- Entry: CSA-V01 to CSA-V03 run the CSA-R01 hand-off with a damage record. CSA-V04 to CSA-V15 call
  the named helper, applier or callback directly with the stated arguments; record fields play no
  part there.
- Supplied difficulty rows: a vector may supply a difficulty row with divisors that differ from the
  table values of section 3. Such a row is a declared fixture input selected by the stated
  `A.game.difficulty`, not table data.
- Games: every applier reads frame, difficulty and pool from `A.game` only. A defender associated
  with a different game changes no output; CSA-V05 (k) and CSA-V09 (n) show it for stun and cold.
- Seeds are full `{lo,hi}` pairs; "no draw" means every seed equals its input. Request lists are in
  order; "new-list requests" means the complete ordered sequence of the cited rule's no-list branch.
  Request lists omit each applier's one `FindStateList` call, which always precedes that applier's
  other logged requests; a vector mentions lookups only to assert that none occurred.
- Every vector has `runtime_observed: no`.

- **CSA-V01** Hand-off argument roles, order and lethal decision (CSA-R01; boundary: supplied-input stage)
  - Inputs: record `manaBurn` 1, `staminaBurn` 2, `stunLength` 3, `coldLength` 4, `freezeLength` 5,
    `poison` 6, `poisonLength` 7, `burnMagnitude` 8, `burnLength` 9; `resultFlags` = 0x000A (`lethal`
    is 0x0002; 0x0008 is an unrelated result bit); every other field 0. Drains and appliers are supplied operations that
    only log. `D.statList` has `usesResolvedStats` false and `ownStats = [(6, 0, 256)]`, count 1;
    definition 6 available, `fMin` false.
  - Seeds: `no draw`
  - Draws: none
  - Trace: log the seven requests in order, then `Stat(D, 6, 0)` = 256 > 0 clears `lethal`.
  - Outputs: log `ManaDrain(D, 1)`, `StaminaDrain(D, 2)`, `Stun(A, D, 3)`, `Cold(A, D, 4)`,
    `Freeze(attacker A, defender D, length 5)`, `Poison(A, D, 6, 7)`, `Burn(A, D, 8, 9)`, `Stat(D, 6, 0)`;
    `resultFlags` = 0x0008; all other fields unchanged.
    (a) Own stat 6 value 0: `resultFlags` stays 0x000A. (b) Own stat 6 value -1: stays 0x000A.
    (c) `resultFlags` 0x0008, own stat 6 value 256: stays 0x0008.
    (d) `resultFlags` 0x0008, own stat 6 value 0: becomes 0x000A.
    (e) `resultFlags` 0x0008, own stat 6 value -1: becomes 0x000A.
    (f) Own stat 6 value 128: `resultFlags` becomes 0x0008.
  - Discriminates: distinct values reject any swapped field, swapped magnitude/length or swapped
    applier order; case (b) rejects an unsigned positive-life test; case (a) rejects `>= 0`; (c)
    rejects toggling the bit; (d) and (e) reject leaving the bit unchanged for nonpositive life;
    (f) rejects a whole-point (256) test.
  - runtime_observed: no

- **CSA-V02** Hand-off over real appliers on a monster, Nightmare (CSA-R01, CSA-R03 to CSA-R06, CSA-R11, CSA-R12; boundary: supplied-input stage)
  - Inputs: `A.game.difficulty` 1 (Nightmare, cold divisor 2); `D` monster with `monsterFlag8` true,
    `ColdEffect(N)` -30, empty event chain; record `manaBurn` 0, `staminaBurn` 0, `stunLength` 10,
    `coldLength` 50, `freezeLength` 0, `poison` 0x100, `poisonLength` 25, `burnMagnitude` 0x100,
    `burnLength` 25; supplied drains log only; `D.statList` has `usesResolvedStats` false and
    `ownStats = [(6, 0, 0)]` (no applier writes it); `lethal` set.
  - Seeds: `A.seed = {1, 0x29A}`, `D.seed = {5, 0x29A}`
  - Draws: (A, 100, 51); (D, 100, 99)
  - Trace: stun: `monsterFlag8`, A steps to `{0x6AC6935F, 0}`, 51 < 90 cancels. Cold: E = -30,
    50 / 2 = 25, expiry 1025, new-list requests of CSA-R05 with stats 67, 68, 69 = -30; D steps to
    `{0x15E0D673, 2}`, 99 >= 20 requests state 107 off. Freeze length 0 returns. Poison: pre-step
    walk finds only the cold node `K12@1025`, no removal; timer 3 at 1001; new list expiry 1025, stat
    74 = -0x100, state 2. Burn: pre-step walk over `K12@1025` (poison), `K3@1001`, `K12@1025` (cold)
    requests removal of `K3@1001` only; timer 3 at 1001; new list expiry 1025, stat 74 = -0x100,
    state 115. Life 0 keeps `lethal`.
  - Outputs: states 11, 2, 115 on; state 107 off requested once; kind-12 requests at 1025 three times;
    kind-3 requests at 1001 twice; exactly one removal request, by burn, of poison's kind-3 node; final
    chain `K12@1025` (burn), `K3@1001` (burn), `K12@1025` (poison), `K12@1025` (cold); no state 21 or
    state 1; `A.seed = {0x6AC6935F, 0}` (1 step), `D.seed = {0x15E0D673, 2}` (1 step); `lethal` still
    set.
    (a) `ColdEffect(N)` 0: draw (A, 100, 51) cancels the stun, `A.seed = {0x6AC6935F, 0}`; cold
    returns with no list, request or draw, `D.seed = {5, 0x29A}`; state 107 not requested; freeze
    returns; poison requests kind-3 at 1001 and its new list (state 2, expiry 1025, stat 74 = -0x100,
    kind-12 at 1025); burn requests removal of poison's kind-3 node, kind-3 at 1001 and its new list
    (state 115, expiry 1025, stat 74 = -0x100, kind-12 at 1025); `lethal` still set.
    (b) Freeze-only, from the Inputs with `monsterFlag8` false and every record length 0 except
    `freezeLength` 40 (`ColdEffect(N)` stays -30, Nightmare freeze divisor 2): 40 / 2 = 20; requests state 1 on, `NewList(P, 2, 1020, owner 0/7)`, tag 1,
    freeze callback, attach 1, kind-12 at 1020, kind-2 at 1021; no draw; `Stat(D, 6, 0)` = 0, so
    `lethal` stays set.
  - Discriminates: the attacker step proves the stun draw owner (a defender-owned draw would leave A
    unchanged); one kind-3 request per applier rejects sharing one pre-step; burn's removal of
    poison's kind-3 node rejects a pre-step that ignores nodes scheduled earlier in the same hit;
    state 107 off rejects omitting the off request. Applier order is rejected by CSA-V01's log, not by this vector.
  - runtime_observed: no

- **CSA-V03** Status draw order with a cancelled stun and a freeze-to-cold fallback (CSA-R01, CSA-R03, CSA-R05 to CSA-R07, CSA-R10 to CSA-R12; boundary: supplied-input stage)
  - Inputs: Normal; `D` monster with `monsterFlag8` true, class 0 record `Velocity` 6, `ColdEffect` -50;
    record `stunLength` 25, `coldLength` 25, `freezeLength` 25, `poison` 100, `poisonLength` 50,
    `burnMagnitude` 10, `burnLength` 50, `manaBurn` 320, `staminaBurn` 0; supplied drains log only;
    `D.statList` has `usesResolvedStats` false and `ownStats = [(6, 0, 0x19B9C)]`; `lethal` set.
  - Seeds: `A.seed = {0x6AC690C5, 0}`, `D.seed = {0, 5}`
  - Draws: (A, 100, 9); (D, 100, 5); (D, 100, 33)
  - Trace: stun: A steps to `{0x669A3799, 747178471}`, 9 < 90 cancels. Cold 25: 25 / 1 = 25, expiry
    1025, new list with stats -50; D steps to `{5, 0}`, 5 < 20, monster, no `deadCol`: state 107 on.
    Freeze 25: no state 54; recordGate 0; `monsterGate(D, 8)` = 1, so Cold(A, D, 25): the lookup finds the state 11 list attached by
    the first cold, with stored expiry 1025; candidate 1025 is not later than 1025, no expiry write or timer; D steps to `{0x15E0D3D9, 2}`,
    33 >= 20: state 107 off. Poison: pre-step removes nothing, timer 3 at 1001, new list expiry 1050,
    stat 74 = -100. Burn: pre-step requests removal of poison's `K3@1001`, timer 3 at 1001, new list
    expiry 1050, stat 74 = -10. Life 0x19B9C > 0 clears `lethal`.
  - Outputs: `A.seed = {0x669A3799, 747178471}` (1 step), `D.seed = {0x15E0D3D9, 2}` (2 steps); state
    11 on with expiry 1025 and one kind-12 request at 1025; states 2 and 115 on with expiry 1050,
    list stat 74 = -100 (state 2) and -10 (state 115);
    state 107 requested on then off; one removal request (burn removes poison's kind-3 node); no state
    21, no state 1, no kind-2 request; `lethal` cleared.
  - Discriminates: the second defender step rejects skipping the shatter draw on a non-extending cold;
    a freeze that ignored `monsterFlag8` would create state 1 and a kind-2 request and leave
    `D.seed = {5, 0}`. A swapped applier order is rejected by CSA-V01, not here.
  - runtime_observed: no

- **CSA-V04** Gate helpers (CSA-R02; boundary: isolated rule)
  - Inputs: MonStats count 1, class 0 record `boss` false unless stated; monster flag word 0x0008 where
    stated.
  - Seeds: `no draw`
  - Draws: none
  - Trace: evaluate each helper directly.
  - Outputs: (a) recordGate: monster class 0 with supplied record `boss` true returns 64; supplied
    `boss` false returns 0; null record uses loaded class 0 (`boss` false) and returns 0; loaded
    `boss` true returns 64; class -1 or 1 with a null record returns 0; a player with a supplied
    `boss` true record returns 0. (b) monsterGate: flag word 0x0008 with mask 8 returns 1, mask 2
    returns 0, mask 0x10008 returns 1; null unit, player or absent monster data returns 0.
    (c) unitGate: monster classes 271, 338, 359, 560, 561 return 1, 2, 3, 4, 4; classes 0, 270, 339,
    559, 562 return 0; a null unit or a player with class 271 returns 0. All inputs unchanged.
  - Discriminates: mask 2 rejects ignoring the mask; the player-with-boss-record case rejects
    skipping the monster test; neighbor classes reject range tests. The 0x10008 case only shows that
    mask bits above the 16-bit word cannot change the result; it rejects no alternative.
  - runtime_observed: no

- **CSA-V05** Stun cancel draw, caps and new-list request order (CSA-R03, CSA-R04; boundary: supplied-input stage)
  - Inputs: monster `D` with `monsterFlag8` true, no state 21 list; `len` 10. `D.seed = {9, 0x29A}`.
  - Seeds: per case
  - Draws: per case
  - Trace: per case, by CSA-R03 then CSA-R04.
  - Outputs:
    (a) `A.seed = {1, 0x29A}`: draw (A, 100, 51); A becomes `{0x6AC6935F, 0}`; 51 < 90 cancels; no
    requests; D unchanged.
    (b) `A.seed = {5, 0x29A}`: draw (A, 100, 99); A becomes `{0x15E0D673, 2}`; expiry 1010; requests
    `NewList(P, 2, 1010, owner 0/7)`, kind-12 at 1010, tag 21, generic callback, attach mode 1,
    state 21 on; D unchanged.
    (c) Player `D`, `A.seed = {1, 0x29A}`: no draw; expiry 1010; new-list requests at 1010.
    (d) Monster `D` with `monsterFlag8` false, `A.seed = {1, 0x29A}`: no draw; expiry 1010; new-list
    requests at 1010.
    (e) `len` 0, `A.seed = {1, 0x29A}`: no draw, A stays `{1, 0x29A}`; no requests.
    (f) Monster `D` with `monsterFlag8` true, class 0 `Velocity` 6, `len` 300, `A.seed = {0, 95}`: draw (A, 100, 95); A becomes `{95, 0}`;
    len 250, expiry 1250; new-list requests at 1250.
    (g) Monster `D` with `monsterFlag8` true, class 0 `Velocity` 6, `len` 300, `A.seed = {1, 0}`: draw (A, 100, 85); A becomes `{0x6AC690C5, 0}`; cancelled; no
    requests.
    (h) Player `D`, `len` 300, `A.seed = {1, 0}`: no draw; len 250; new-list requests at 1250.
    (i) Monster `D` with `monsterFlag8` true, `len` 10, `A.seed = {0, 89}`: draw (A, 100, 89); A becomes `{89, 0}`; cancelled; no requests.
    (j) Monster `D` with `monsterFlag8` true, `len` 10, `A.seed = {0, 90}`: draw (A, 100, 90); A becomes `{90, 0}`; not cancelled; expiry 1010;
    new-list requests at 1010.
    (k) Player `D` associated with a different game (frame 5000, difficulty 2, pool `P2`; fixture
    input), `A.seed = {1, 0x29A}`, `len` 10: no draw; expiry 1010 from `A.game`; requests
    `NewList(P, 2, 1010, owner 0/7)`, kind-12 at 1010 on `A.game`, tag 21, generic callback, attach
    mode 1, state 21 on.
  - Discriminates: (b) rejects placing `SetState` before `Attach`; (c) and (d) reject drawing for
    players or unflagged monsters; (f) rejects comparing the old low word (old low word 0 would
    cancel); (i) and (j) fix the threshold at `r < 90`; (k) rejects using the defender's game.
  - runtime_observed: no

- **CSA-V06** Stun early monster returns (CSA-R03; boundary: supplied-input stage)
  - Inputs: `len` 10 unless stated; `D.seed = {9, 0x29A}`; `A.seed = {1, 0x29A}`.
  - Seeds: as stated
  - Draws: per case: none in (a) to (d); (A, 100, 51) in (e); (A, 100, 99) in (f) and (g)
  - Trace: per case.
  - Outputs:
    (a) Monster `D`, `monsterFlag8` false, class 0 `boss` true: recordGate 64 returns; no logged
    lookup and no request; seeds unchanged.
    (b) Monster `D`, `monsterFlag8` false, `boss` false, class 0 `Velocity` 0: returns; no logged
    lookup and no request.
    (c) Player `D` class 0 (the class 0 record has `Velocity` 0): no monster check runs; requests
    `NewList(P, 2, 1010, owner 0/7)`, kind-12 at 1010, tag 21, generic callback, attach mode 1, state
    21 on; no list stat write.
    (d) Monster `D`, `boss` true, `Velocity` 6, `len` 100, `A.seed = {0, 95}`: no stun, no list, no
    draw; A stays `{0, 95}`.
    (e) Monster `D`, `monsterFlag8` true, `boss` true, `A.seed = {1, 0x29A}`: draw (A, 100, 51), A
    becomes `{0x6AC6935F, 0}`; cancelled; no logged lookup and no request.
    (f) Monster `D`, `monsterFlag8` true, `boss` true, `A.seed = {5, 0x29A}`: draw (A, 100, 99), A
    becomes `{0x15E0D673, 2}`; not cancelled; the boss check then returns; no logged lookup and no
    request.
    (g) Monster `D`, `monsterFlag8` true, `boss` false, class 0 `Velocity` 0, `A.seed = {5, 0x29A}`:
    draw (A, 100, 99), A becomes `{0x15E0D673, 2}`; not cancelled; the Velocity check then returns; no
    logged lookup and no request.
  - Discriminates: (c) rejects applying the zero-Velocity rejection to players; (a) and (b) reject
    looking up the list before these checks; (e), (f) and (g) reject running the boss or Velocity check before the
    cancel draw (A would stay unchanged). The order among the boss, Velocity and class-group checks
    is not observable here (data reads and gate helpers are not logged, and each rejecting check
    returns with no effect); CSA-R03 states it, and no vector claims it.
  - runtime_observed: no

- **CSA-V07** Stun caps and shortening (CSA-R03, CSA-R04; boundary: supplied-input stage)
  - Inputs: monster `D`, `monsterFlag8` false; MonStats count 562 with records for classes 0, 271,
    338, 359, 560 and 561, each `boss` false; the class record each case names has `Velocity` 1 in cases (a)
    to (c) and (i), and `Velocity` 6 in cases (d) to (h) and (j) to (l).
  - Seeds: `no draw`; `A.seed = {0, 95}`
  - Draws: none
  - Trace: per case.
  - Outputs: with an existing state 21 list of expiry 1200:
    (a) class 0, `len` 50: `SetListExpiry(1050)`, kind-12 at 1050 (shortened).
    (b) class 0, `len` 300: capped 250; `SetListExpiry(1250)`, kind-12 at 1250.
    (c) class 271, `len` 50: capped 13; `SetListExpiry(1013)`, kind-12 at 1013.
    With no state 21 list:
    (d) class 338, `len` 100: capped 13; new-list requests at expiry 1013.
    (e) class 338, `len` 12: stays 12; new-list requests at 1012.
    (f) class 0, `len` 300: capped 250; new-list requests at 1250.
    (g) class 0, `len` 250: stays 250; new-list requests at 1250.
    (h) class 0, `len` 251: capped 250; new-list requests at 1250.
    (i) With the existing list of expiry 1200, class 271, `len` 13: stays 13; `SetListExpiry(1013)`,
    kind-12 at 1013.
    (j) No state 21 list, class 560, `len` 100: capped 13; new-list requests at 1013.
    (k) No state 21 list, class 561, `len` 100: capped 13; new-list requests at 1013.
    (l) No state 21 list, class 359, `len` 100: capped 13; new-list requests at 1013.
  - Discriminates: (a) rejects extend-only stun; (e) rejects capping class-group stuns below 13 to 13;
    (b) and (f) reject an uncapped stun; (g) and (h) fix the cap test at `len > 250`; (c) (class 271,
    value 1), (d) (338, value 2), (l) (359, value 3), (j) and (k) (560 and 561, value 4) exercise every
    class-group return value through the stun applier.
  - runtime_observed: no

- **CSA-V08** Cold extends only and draws on every application (CSA-R05, CSA-R06; boundary: supplied-input stage)
  - Inputs: player `D`, no state 11 list; `A.seed = {5, 0x29A}`.
  - Seeds: `D.seed = {1, 0x29A}`
  - Draws: per case
  - Trace: player E = -50, no division.
  - Outputs:
    (a) `len` 50: expiry 1050; requests state 11 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12 at 1050,
    tag 11, cold callback, attach mode 1, stats 67, 68, 69 = -50, animation refresh; draw (D, 100, 51),
    D becomes `{0x6AC6935F, 0}`; state 107 off.
    (b) Inputs: player `D` with an attached state 11 list (owner 0/7, cold callback, stats 67, 68, 69 =
    -50, expiry 1050), state 11 on, `D.seed = {0x6AC6935F, 0}`, frame 1010, `len` 20: candidate 1030 < 1050; no expiry write,
    timer or stat write; draw (D, 100, 31), D becomes `{0x2F2ED81B, 0x2C890AFD}`; state 107 off.
    (c) Inputs: player `D` with an attached state 11 list (owner 0/7, cold callback, stats 67, 68, 69 =
    -50, expiry 1050), state 11 on, `D.seed = {0x6AC6935F, 0}`, frame 1010, `len` 60: candidate 1070; `SetListExpiry(1070)`,
    kind-12 at 1070; draw (D, 100, 31), D becomes `{0x2F2ED81B, 0x2C890AFD}`; state 107 off.
    (d) Inputs: player `D` with an attached state 11 list (owner 0/7, cold callback, stats 67, 68, 69 =
    -50, expiry 1050), state 11 on, `D.seed = {0x6AC6935F, 0}`, frame 1010, `len` 40: candidate 1050 equals 1050; no expiry
    write, timer or stat write; draw (D, 100, 31), D becomes `{0x2F2ED81B, 0x2C890AFD}`; state 107 off.
    (e) Inputs: an attached state 11 list of expiry 1200 (stats -50), state 11 on,
    `D.seed = {1, 0x29A}`, frame 1000, `len` 50: retained 1200, no write or timer; draw (D, 100, 51),
    D becomes `{0x6AC6935F, 0}`; state 107 off.
    (f) Inputs: an attached state 11 list of expiry 1200 (stats -50), state 11 on,
    `D.seed = {1, 0x29A}`, frame 1000, `len` 300: `SetListExpiry(1300)`, kind-12 at 1300 (no length cap); draw
    (D, 100, 51), D becomes `{0x6AC6935F, 0}`; state 107 off.
    (g) Player `D` of class 0 whose class 0 record has `ColdEffect` -30 at every difficulty, no list,
    `D.seed = {1, 0x29A}`, `len` 50: E = -50 (no record lookup for a player); expiry 1050; new-list
    requests with stats 67, 68, 69 = -50; draw (D, 100, 51), D becomes `{0x6AC6935F, 0}`; state 107 off.
    (h) Player `D`, no list, `D.seed = {1, 0x29A}`, frame 1000, `len` 300: expiry 1300; new-list
    requests at 1300 with stats -50; draw (D, 100, 51), D becomes `{0x6AC6935F, 0}`; state 107 off.
    A stays `{5, 0x29A}` in every case.
  - Discriminates: (d) rejects an equal-candidate write; (b) rejects skipping the draw when the
    expiry does not change; (f) rejects a stun-style 250 cap on cold; (g) rejects reading a monster
    record for a player; (h) shows the absent cap on the no-list branch.
  - runtime_observed: no

- **CSA-V09** Cold monster divisor, minimum and shatter conditions (CSA-R05, CSA-R06; boundary: supplied-input stage)
  - Inputs: `A.game.difficulty` 1 (Nightmare, cold divisor 2); monster `D` with `ColdEffect(N)` -50;
    `len` 25; no state 11 list; `A.seed = {1, 0x29A}`.
  - Seeds: `D.seed = {0, 5}` unless stated
  - Draws: per case
  - Trace: per case.
  - Outputs:
    (a) 25 / 2 = 12, expiry 1012; new-list requests with stats -50; draw (D, 100, 5), D becomes
    `{5, 0}`; 5 < 20, monster, no `deadCol`: state 107 on.
    (b) `D.seed = {0, 50}`: expiry 1012; new-list requests with stats -50; draw (D, 100, 50), D
    becomes `{50, 0}`; state 107 off.
    (c) `deadCol` true: expiry 1012; new-list requests with stats -50; draw (D, 100, 5), D becomes
    `{5, 0}`; state 107 off.
    (d) `len` 1: 1 / 2 = 0, raised to 1, expiry 1001; new-list requests; draw (D, 100, 5); state 107 on.
    (e) Player `D`: no division, expiry 1025; new-list requests; draw (D, 100, 5); state 107 off.
    (f) `ColdEffect(N)` 0: no request, no draw.
    (g) Existing state 11 list with expiry 1020, `len` 12: 12 / 2 = 6, candidate 1006, retained 1020,
    no write or timer; draw (D, 100, 5); state 107 on.
    (h) `D.seed = {0, 20}`: expiry 1012; new-list requests with stats -50; draw (D, 100, 20), D
    becomes `{20, 0}`; state 107 off.
    (i) `D.seed = {0, 19}`: expiry 1012; new-list requests with stats -50; draw (D, 100, 19), D
    becomes `{19, 0}`; state 107 on.
    (j) `ColdEffect(N)` +10: E > 0, so no division; expiry 1025; new-list requests with stats 67, 68,
    69 = 10; draw (D, 100, 5); state 107 on.
    (k) `A.game.difficulty` 1 selects a supplied difficulty row (a fixture input, not table data)
    with `MonsterColdDivisor` 3 and
    `MonsterFreezeDivisor` 5, `len` 30: 30 / 3 = 10, expiry 1010; new-list requests; draw (D, 100, 5);
    state 107 on.
    (l) `A.game.difficulty` 0 (Normal, cold divisor 1), class 0 `ColdEffect` -40, `ColdEffect(N)` -50,
    `ColdEffect(H)` -60: E = -40, 25 / 1 = 25, expiry 1025; new-list requests with stats -40; draw
    (D, 100, 5); state 107 on.
    (m) `A.game.difficulty` 2 (Hell, cold divisor 4), class 0 `ColdEffect` -40, `ColdEffect(N)` -50,
    `ColdEffect(H)` -60: E = -60, 25 / 4
    = 6, expiry 1006; new-list requests with stats -60; draw (D, 100, 5); state 107 on.
    (n) `D` associated with a different game (frame 5000, difficulty 0, pool `P2`; a fixture input):
    the attacker's game is used; 25 / 2 = 12, expiry 1012; new-list requests with `NewList(P, ...)`,
    stats -50; kind-12 at 1012 on `A.game`; draw (D, 100, 5); state 107 on.
    (o) `A.game.difficulty` 1 selects a supplied difficulty row with `MonsterColdDivisor` 0
    (fixture input, not table data): skip division, `len` stays 25, expiry 1025; new-list requests
    with stats 67, 68, 69 = -50 and kind-12 at 1025; draw (D, 100, 5), D becomes `{5, 0}`;
    state 107 on. No fatal assertion occurs.
    (p) Monster `D` of class 1; MonStats count 2, class 1 record as the class 0 defaults (including
    `ColdEffect(N)` -50) with `MonStatsEx` 0; MonStats2 count 2, record 0 `deadCol` true, record 1
    `deadCol` false: expiry 1012; new-list requests with stats -50; draw (D, 100, 5), D becomes
    `{5, 0}`; `deadCol` is read through `MonStatsEx` 0 (true), so state 107 off.
    (q) Class 0 record `MonStatsEx` 1 with MonStats2 count 1 and record 0 `deadCol` true (index 1 selects
    no valid MonStats2 record): expiry 1012; new-list requests with stats -50; draw (D, 100, 5), D
    becomes `{5, 0}`; the unavailable lookup gives `deadCol` false, so state 107 on.
    (r) As (q) but `MonStatsEx` -1: the negative index selects no record; same requests and draw;
    state 107 on.
    A unchanged in every case.
  - Discriminates: (l) and (m) reject reading one fixed difficulty column; (n) rejects using the
    defender's game; (e) rejects dividing players; (d) rejects a minimum applied before division;
    (c) rejects ignoring `deadCol`; (e) also rejects allowing players to shatter; (h) and (i) fix the
    threshold at `r < 20`; (j) rejects dividing for positive E; (k), with CSA-V11 (e), rejects using
    the freeze divisor for cold; (o) rejects a fatal zero-divisor policy for cold; (p) rejects indexing MonStats2 by class id
    instead of through `MonStatsEx`; (q) and (r) reject treating an out-of-range or negative
    MonStats2 index as `deadCol` true, as a fault, or as a fallback, clamp or wrap to record 0 (whose
    `deadCol` is true there). An unsigned reading of `MonStatsEx` in (r) is equivalent unless MonStats2
    has more than 32767 records; GDA-R08a fixes the signed i16 reading.
  - runtime_observed: no

- **CSA-V10** Freeze routing to cold (CSA-R07, CSA-R05, CSA-R06; boundary: supplied-input stage)
  - Inputs: `freezeLength` per case; no lists.
  - Seeds: per case; each case states both seeds
  - Draws: per case
  - Trace: per case.
  - Outputs:
    (a) Player `D`, `D.seed = {1, 0x29A}`, `A.seed = {5, 0x29A}`, length 10: Cold(A, D, 10); requests
    state 11 on, `NewList(P, 2, 1010, owner 0/7)`, kind-12 at 1010, tag 11, cold callback, attach 1,
    stats 67, 68, 69 = -50, refresh; draw (D, 100, 51), D becomes `{0x6AC6935F, 0}`; state 107 off; no
    state 1 lookup or request; A unchanged.
    (b) Monster `D` whose class 0 record has `boss` true, `D.seed = {1, 0x29A}`, `A.seed = {5, 0x29A}`,
    length 10: Cold(A, D, 10) with 10 / 1 = 10; requests state 11 on, `NewList(P, 2, 1010, owner
    0/7)`, kind-12 at 1010, tag 11, cold callback, attach 1, stats 67, 68, 69 = -50, refresh; draw
    (D, 100, 51), D becomes `{0x6AC6935F, 0}`; state 107 off; A unchanged.
    (c) Monster `D` with state 54 on, length 10, `D.seed = {1, 0x29A}`, `A.seed = {5, 0x29A}`: return;
    no request, no draw; both seeds unchanged.
    (d) Nightmare (freeze and cold divisors 2), monster `D`, length 25, `D.seed = {0, 50}`,
    `A.seed = {5, 0x29A}`: with
    `boss` true, or `monsterFlag8` true, or class 359 (MonStats count 360, class 359 record `boss`
    false, `ColdEffect(N)` -50): cold with 25, 25 / 2 = 12, expiry 1012, state 11 new list, draw
    (D, 100, 50), D becomes `{50, 0}`; state 107 off, no state 1; A unchanged.
    (e) Nightmare, monster `D` with `boss` false, `monsterFlag8` false, class 0, length 25,
    `D.seed = {0, 50}`, `A.seed = {5, 0x29A}`: freeze path; 25 / 2 = 12; requests state 1 on,
    `NewList(P, 2, 1012, owner 0/7)`, tag 1, freeze callback, attach 1, kind-12 at 1012, kind-2 at
    1013; no draw; both seeds unchanged.
    (f) Nightmare, player `D`, length 25, `D.seed = {0, 50}`, `A.seed = {5, 0x29A}`: cold with 25, no
    division, expiry 1025; draw (D, 100, 50), D becomes `{50, 0}`; state 107 off; A unchanged.
    (g) Player `D` with state 54 on, `D.seed = {1, 0x29A}`, `A.seed = {5, 0x29A}`, length 10:
    Cold(A, D, 10); requests state 11 on, `NewList(P, 2, 1010, owner 0/7)`, kind-12 at 1010, tag 11,
    cold callback, attach 1, stats 67, 68, 69 = -50, refresh; draw (D, 100, 51), D becomes
    `{0x6AC6935F, 0}`; state 107 off; A unchanged.
    (h) Monster `D` with `boss` true and state 54 on, length 10, `D.seed = {1, 0x29A}`,
    `A.seed = {5, 0x29A}`: return; no request, no draw; both seeds unchanged.
    (i) Monster `D` of class 271 (MonStats count 272, class 271 record `boss` false, `ColdEffect` -50),
    Normal, length 10, `D.seed = {1, 0x29A}`, `A.seed = {5, 0x29A}`: unitGate 1, so Cold(A, D, 10)
    with 10 / 1 = 10; cold new-list requests at expiry 1010 with stats -50; draw (D, 100, 51), D
    becomes `{0x6AC6935F, 0}`; state 107 off; no state 1; A unchanged.
    (j) `A.game.difficulty` 1 with a supplied difficulty row (`MonsterColdDivisor` 3,
    `MonsterFreezeDivisor` 5), monster `D` with `boss` true, length 30, `D.seed = {1, 0x29A}`,
    `A.seed = {5, 0x29A}`: Cold(A, D, 30) with 30 / 3 = 10; cold new-list requests at expiry 1010
    with stats -50; draw (D, 100, 51), D becomes `{0x6AC6935F, 0}`; state 107 off; no state 1; A
    unchanged.
  - Discriminates: (c) rejects a cold fallback for state 54; (h) rejects testing state 54 after the
    gates; (j) rejects dividing the fallback length by the freeze divisor; (i) exercises the class-group gate through the applier; (g) rejects testing state 54 before the
    player routing; (b) and (d) reject freezing gated
    monsters; (f) rejects dividing the fallback length by the freeze divisor.
  - runtime_observed: no

- **CSA-V11** Freeze length, expiry and tail requests (CSA-R08, CSA-R09; boundary: supplied-input stage)
  - Inputs: Nightmare (freeze divisor 2); monster `D`, `ColdEffect(N)` -30, all gates 0, state 54 off,
    no state 1 list; event chain Q (kind 2), R (kind 3), S (kind 2) in that order.
  - Seeds: `A.seed = {1, 0x29A}`, `D.seed = {1, 0x29A}`; no draw
  - Draws: none
  - Trace: per case.
  - Outputs:
    (a) length 1: 1 / 2 = 0, expiry 1000; requests state 1 on, `NewList(P, 2, 1000, owner 0/7)`, tag 1,
    freeze callback, attach 1, kind-12 requested 1000 queued 1001; remove Q, remove S, kind-2
    requested 1001 queued 1001. The walk passes `K12@1000` (scheduled first, kind 12, kept), Q, R, S.
    Final chain `K2@1001`, `K12@1000`, R; no list stat write.
    (b) `ColdEffect(N)` 0, length 1: return after the effect test; no request; chain Q, R, S.
    (c) Existing state 1 list `L`, state 1 on, length 20 (adjusted 10, candidate 1010), empty chain:
    old expiry 1020: no write, kind-12 at 1020, kind-2 at 1011. Old 1010: no write, kind-12 at 1010,
    kind-2 at 1011. Old 1005: `SetListExpiry(1010)`, kind-12 at 1010, kind-2 at 1011. No allocation or
    removal; state 1 stays on.
    (d) `ColdEffect(N)` -50, length 25, empty chain: 12, new-list requests at expiry 1012, kind-12 at
    1012, kind-2 at 1013.
    (e) `A.game.difficulty` 1 selects a supplied difficulty row (a fixture input, not table data)
    with `MonsterColdDivisor` 3 and
    `MonsterFreezeDivisor` 5, length 30, empty chain: 30 / 5 = 6; new-list requests at expiry 1006,
    kind-12 at 1006, kind-2 at 1007.
    (f) `ColdEffect(N)` +10, length 1: return after the effect test; no request; chain Q, R, S.
    (g) `D` associated with a different game (frame 5000, difficulty 0, pool `P2`; fixture
    input), length 20, empty chain: 20 / 2 = 10 (attacker's Nightmare divisor); requests state 1 on,
    `NewList(P, 2, 1010, owner 0/7)`, tag 1, freeze callback, attach 1, kind-12 at 1010 and kind-2 at
    1011, both on `A.game`; no draw.
  - Discriminates: (a) rejects a minimum length of 1 and removing kind-3 nodes, and, because removal
    clears the removed node's successor, rejects reading the successor after removing Q (the walk
    would stop and S would remain); (c) rejects skipping the kind-12 request when the expiry is
    retained and rejects an equality write; (e) rejects using the cold divisor; (f) rejects testing
    only `E == 0`; (g) rejects using the defender's game.
  - runtime_observed: no

- **CSA-V12** Poison monster pre-step and new list (CSA-R10, CSA-R11, CSA-R12; boundary: supplied-input stage)
  - Inputs: monster `D`, event chain Q (kind 2), R (kind 3), S (kind 2); `poison` 100,
    `poisonLength` 50; no state 2 list.
  - Seeds: `no draw`
  - Draws: none
  - Trace: pre-step, then the no-list branch with expiry 1050.
  - Outputs: remove R; kind-3 at 1001; then state 2 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12 at
    1050, stat 74 = -100 layer 0, tag 2, generic callback, attach 1. Final chain `K12@1050`,
    `K3@1001`, Q, S.
    (a) Player `D` with chain Q (kind 2), R (kind 3), S (kind 2): no removal and no kind-3 request;
    state 2 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12 at 1050, stat 74 = -100 layer 0, tag 2,
    generic callback, attach 1; final chain `K12@1050`, Q, R, S.
    (b) Monster `D` with chain Q (kind 2), R (kind 3), S (kind 2) and an existing state 2 list (expiry 1080, own stat 74 = -200):
    remove R, kind-3 at 1001; then 200 > 100 rejects; no list request; final chain `K3@1001`, Q, S.
    (c) Monster `D` with chain Q (kind 3), R (kind 2), S (kind 3), no state 2 list: remove Q, remove S,
    kind-3 at 1001, then state 2 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12 at 1050, stat 74 = -100
    layer 0, tag 2, generic callback, attach 1; final chain `K12@1050`, `K3@1001`, R.
  - Discriminates: removal of R only rejects removing kind-2 nodes; (a) rejects a pre-step for players;
    (b) rejects placing the pre-step after the magnitude acceptance; (c) rejects removing only the
    first kind-3 node.
  - runtime_observed: no

- **CSA-V13** Poison magnitude comparison, shortening and guards (CSA-R10, CSA-R12; boundary: supplied-input stage)
  - Inputs: player `D`; existing state 2 list with expiry 1080, `usesResolvedStats` false,
    `ownStats = [(74, 0, -100)]`, count 1, unless stated.
  - Seeds: `no draw`
  - Draws: none
  - Trace: per case.
  - Outputs:
    (a) magnitude 150, length 50: 100 > 150 false; `SetListExpiry(1050)`, kind-12 at 1050, stat 74 =
    -150.
    (b) magnitude 100, length 200: equal passes; `SetListExpiry(1200)`, kind-12 at 1200, stat 74 = -100.
    (c) magnitude 50, length 200: 100 > 50; no request; expiry 1080 and stat -100 unchanged.
    (d) Existing stat 74 = -200, magnitude 100, length 50: no request.
    (e) No list: length 0 with magnitude 100, or length 25 with magnitude 0: return with no logged
    `FindStateList` and no request.
    (f) No list, magnitude 100, length 50: state 2 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12 at
    1050, stat 74 = -100, tag 2, generic callback, attach 1.
    (g) Monster `D`, no list, chain K (kind 3): length 0 with magnitude 100, or length 25 with
    magnitude 0: no logged lookup, no removal, no kind-3 request; chain K.
    (h) Player `D` associated with a different game (frame 5000, difficulty 2, pool `P2`; fixture
    input), no list, magnitude 100, length 50: state 2 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12
    at 1050 on `A.game`, stat 74 = -100, tag 2, generic callback, attach 1.
    (i) Existing list with own stat 74 = -2147483648, magnitude 100, length 50:
    `i32(low32(-(-2147483648)))` = -2147483648, which is not greater than 100; accept,
    `SetListExpiry(1050)`, kind-12 at 1050, stat 74 = -100. No draw; both seeds unchanged.
  - Discriminates: (a) rejects extend-only poison; (b) rejects a strict comparison; (e) rejects a
    lookup before the gates; (g) rejects a pre-step before the gates; (h) rejects using the
    defender's game; (i) rejects widening the existing-magnitude negation, which would produce
    2147483648 and incorrectly reject the application.
  - runtime_observed: no

- **CSA-V14** Burn gates, field roles and refresh (CSA-R10, CSA-R12; boundary: supplied-input stage)
  - Inputs: player `D`, no state 115 list unless stated.
  - Seeds: `no draw`
  - Draws: none
  - Trace: per case.
  - Outputs:
    (a) `burnMagnitude` 0x500, `burnLength` 0: no logged lookup, no request. (b) magnitude 0, length
    25: no logged lookup, no request.
    (c) magnitude 0x500, length 25: state 115 on, `NewList(P, 2, 1025, owner 0/7)`, kind-12 at 1025,
    stat 74 = -0x500 layer 0, tag 115, generic callback, attach 1.
    (d) magnitude 10, length 50: state 115 on, `NewList(P, 2, 1050, owner 0/7)`, kind-12 at 1050,
    stat 74 = -10 layer 0, tag 115, generic callback, attach 1.
    With an existing state 115 list, expiry 1080, own stat 74 = -100:
    (e) magnitude 50, length 200: no request. (f) magnitude 100, length 50: `SetListExpiry(1050)`,
    kind-12 at 1050, stat 74 = -100.
    (g) Monster `D`, no list, chain K (kind 3): `burnMagnitude` 0x500 with `burnLength` 0, or
    magnitude 0 with length 25: no logged lookup, no removal, no kind-3 request; chain K.
  - Discriminates: (d) rejects swapping burn magnitude and length (that would give expiry 1010 and
    stat -50); (f) rejects extend-only burn; (a) and (b) reject a lookup before the gates; (g)
    rejects a pre-step before the gates.
  - runtime_observed: no

- **CSA-V15** Cold expiry callback (CSA-R13; boundary: supplied-input stage)
  - Inputs: states 11 and 107 on; the lifecycle already removed the cold list and its stats; `s` = 11;
    supplied `StayDeath` returns the stated value and logs its arguments.
  - Seeds: `no draw`
  - Draws: none
  - Trace: evaluate `alive`, then the requests.
  - Outputs:
    (a) Monster mode 1: state 11 off, state 107 off, refresh.
    (b) Monster mode 12, `StayDeath` nonzero: log `StayDeath(D, 11)`; no request; both states stay on.
    (c) Monster mode 12, `StayDeath` 0: log `StayDeath(D, 11)`; state 11 off, refresh; state 107 stays
    on.
    (d) Player mode 1: state 11 off, state 107 off, refresh.
    (e) Player mode 17, `StayDeath` 0: log `StayDeath(D, 11)`; state 11 off, refresh; state 107 stays
    on.
    (f) Monster mode 1, `aliveExclusionCondition` true, `StayDeath` 0: log `StayDeath(D, 11)`; state 11
    off, refresh; 107 on.
    (g) Monster mode 1, `StayDeath` nonzero: no `StayDeath` call logged; state 11 off, state 107 off,
    refresh.
    (h) Monster mode 17, `StayDeath` nonzero: alive; no `StayDeath` call logged; state 11 off, state
    107 off, refresh.
    (i) Player mode 12, `StayDeath` nonzero: alive; no `StayDeath` call logged; state 11 off, state
    107 off, refresh.
    Cases (a) and (d) log no `StayDeath` call.
  - Discriminates: (g) rejects consulting `StayDeath` for alive units; (c) and (f) reject clearing
    shatter for units that are not alive; the logged `StayDeath(D, 11)` arguments reject passing another unit or state; (h) and (i) reject one not-alive mode set shared by players
    and monsters.
  - runtime_observed: no

### 8.2 Observation vectors

All are pending; procedures are in the
[companion](../research/spec-evidence/combat-status-application.md#observation-procedures).

- **CSA-O01** Cold extends only: on a monster surviving three cold hits, the first sets the state 11
  list expiry to `frame + adjusted length`; a second whose candidate is not later leaves it; a third
  later candidate raises it; each application steps the defender seed once.
- **CSA-O02** Poison is replaced, not extended: after a second poison hit with a larger magnitude and a
  shorter length, the state 2 list holds `hpregen = -magnitude` of the second hit and the shorter
  expiry.

## 9. Acceptance and dependencies

Partial acceptance: every assertion of CSA-V01 to CSA-V15 for CSA-R01 to CSA-R13 within the stated
supplied-input boundary. Stored list, stat and state effects of the requests, animation refresh effects,
frame-event node removal, handler execution and actual lifecycle free/callback invocation remain
gated by CSA-Q01 and CSA-Q02; passing the vectors
with supplied operations does not accept those integrations. Full reproduction additionally requires
accepting those dependency contracts and matching observations for CSA-O01, CSA-O02 and the verified
rules (CSA-Q05).

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) NRF-R01
(seed step); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) RLI-R03 (state lookup,
test and set routing), RLI-R02a and RLI-R02b (stat reads), section 3.1 (damage record), section 3.3
(unit types); RLI-Q02 retains update-queue integration;
[state-timed-state-lifecycle](state-timed-state-lifecycle.md) STL-R01
(scheduling and frame clamp), STL-R03 (expiry routing) and STL-R05 (actual free/callback integration,
unresolved), STL-R06a
(list primitives), STL-R06b and STL-R06c (attach and list-stat writes, gated), STL-R07 and STL-R07a
(verified state membership and unresolved disguise), STL-R11 (generic callback identity),
STL-R12a (stay-death consumer with supplied masks);
[core-game-data-fields](core-game-data-fields.md) GDF-R02, GDF-R03, GDF-R05a, GDF-R06a, GDF-R08
(stat, state, mode and class names);
[core-game-data-table-records](core-game-data-table-records.md) GDA-R05 (difficulty divisors),
GDA-R03c (case-insensitive column names), GDA-R08a, GDA-R09a and GDA-R09b (MonStats and MonStats2
fields and index bounds);
[combat-leech-and-drain](combat-leech-and-drain.md) LSE-R15 (mana drain) and
[combat-damage-pipeline-order](combat-damage-pipeline-order.md) DPO-R18 (stamina drain), DPO-R17
(the preceding life write).

**Depended on by:** [combat-damage-pipeline-order](combat-damage-pipeline-order.md) (the status stage
of the hit path and its draw order); [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md)
NRF-R04a (seed-ownership rows for the stun-cancel, shatter and fallback sites);
[state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md) (the cold and freeze
lists it describes); [combat-hit-recovery](combat-hit-recovery.md) (reads states 21 and 1);
[combat-open-wounds](combat-open-wounds.md) OWD-R12 (consumes the CSA-R01 lethal decision).

## 10. Related specs

- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): the hit path around this stage;
  DPO-R19a, DPO-R19b, DPO-R27 and DPO-R28a describe the applier order, final life decision and draws
  this spec owns.
- [combat-leech-and-drain](combat-leech-and-drain.md): leech, drain and the poison
  term of the immediate total, which run before or beside this stage.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): what cold and freeze
  do once applied; it consumes this spec's cold-effect, routing and length decisions (CSA-R01,
  CSA-R02, CSA-R05 to CSA-R09); MMF-R09b defines the freeze expiry callback that CSA-R08 installs.
- [combat-open-wounds](combat-open-wounds.md): OWD-R12 consumes CSA-R01's final lethal decision for
  the surviving-monster regeneration restart that follows this stage.
- [state-timed-state-lifecycle](state-timed-state-lifecycle.md): list, timer, expiry and callback
  machinery consumed here.
- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): random steps and the
  seed-ownership table.
- [core-record-layouts-and-ids](core-record-layouts-and-ids.md): stat and state primitives, the damage
  record and unit types.
- [combat-hit-recovery](combat-hit-recovery.md): consumes the stun and freeze states.
- [core-game-data-fields](core-game-data-fields.md): state, stat, mode and class names.
- [core-game-data-table-records](core-game-data-table-records.md): MonStats, MonStats2 and difficulty
  fields.

## Appendix A. Provenance

Evidence target: Lord of Destruction Expansion 1.13c, Windows x86. The original `D2Game.dll` and
`D2Common.dll` specimens are identified in the [source manifest](../research/source-manifest.md).
Static readings of the appliers, helpers, callback and hand-off were recorded on 2026-10-12 and
2026-10-13 and reused; targeted rereadings on 2026-10-14 settled the shatter request, the null-defender
fault point, the monster-flag identity and the hand-off fields. Readings used original instruction
listings and, for the cold callback, a reconstruction checked against the original bytes. Table values
reuse the recorded DifficultyLevels cells. No gameplay observation has been made. The
[public evidence companion](../research/spec-evidence/combat-status-application.md) gives module
locations, rule-linked rationale, contradiction records, the identifier mapping and observation
procedures. Public verification requires matching original game files only.
