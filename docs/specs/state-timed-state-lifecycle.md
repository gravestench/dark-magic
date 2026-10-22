# Timed State Lifecycle

**Description**

Temporary effects attach to a character or monster until their duration ends. The game tracks the effect's bonuses separately from whether its state is active, so removing bonuses and clearing the state are separate actions. Reapplying an effect can extend or replace its duration, depending on the effect, and death can preserve selected states.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: STL-R, STL-Q, STL-V, STL-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/state-timed-state-lifecycle.md)
Open evidence gaps: [gap file](../plans/gaps/state-timed-state-lifecycle.md)

## 1. Scope

This contract covers timed-state list creation, attachment, state membership, expiry scheduling and traversal, removal callbacks, refresh policies by reference, death cleanup, and client state-callback dispatch. Effect magnitudes belong to [combat-status-application](combat-status-application.md) and [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md); expression evaluation belongs to [skill-calc-expressions](skill-calc-expressions.md). Curse overrides, aura ticks, network message layouts and client visual callback bodies are separate behaviors. STL-R16 records the unresolved transition boundary. STL-R17 bounds a separate auxiliary timed effect without state identity. No local rule draws randomness.

## 2. Trigger and ownership

D2Game owns server applications, timer dispatch, removal callbacks and death cleanup. D2Common owns effect lists and state membership. D2Client owns set/remove callback dispatch. Server callbacks identify original source folders `UNIT\SUnitDmg`, `SKILLS\Skills`, `SKILLS\SkillBar`, `SKILLS\SkillSor` and `SKILLS\SkilPal`; shared list operations identify `Stats\StatsEx`.

A damage or skill application creates or refreshes a timed list and schedules state-expiry event type 12. Per-unit frame-event dispatch reaches the expiry walker for players, monsters and items. An expiry event concerns the whole unit: it examines every due timed primary list. Player and monster death cleanup also reaches list removal. Binary dispatch/output anchors are in the [companion](../research/spec-evidence/state-timed-state-lifecycle.md#dispatch-and-output-anchors).

Dispatch matches exact stored frames once per server frame (STL-R02). Cross-group/immediate/equal-slot ordering is STL-R02a. Changed-state membership feeds state-on, state-with-stats and state-off message selection (STL-R14). Synchronization eligibility and changed-state clearing are STL-Q01; packet-to-client-callback integration is STL-Q02.

## 3. Inputs and data

Semantic operation names are illustrative and impose no API or storage layout.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| Current/requested frame, list expiry | Game frame; effect expiry | i32 frames | Signed comparisons, except unsigned auxiliary expiry |
| Target and owner identity | Unit type/id ([core-record-layouts-and-ids](core-record-layouts-and-ids.md) section 3.3; absent sentinel 6/-1 is engine code) | u32 type; i32 id | Player 0, monster 1, object 2, missile 3, item 4; absent owner 6/-1 |
| Unit mode | GDF-R05a/GDF-R06a | u32 | Player death/dead 0/17; monster death/dead 0/12 |
| Additional death condition | `unit.aliveExclusionCondition` | boolean | Treated as dead by removal callbacks for every unit type; shared name from [core-game-data-fields](core-game-data-fields.md) section 3; wider meaning STL-Q06 |
| Base list and chains | RLI-R03 | Optional base, ordered primary/secondary lists | State storage eligibility and effect-list traversal |
| Timed list | Owner, state, expiry, stats, callback | i32 state/expiry/stat values; callable | Effect metadata independent of active membership |
| List conditions | `timed`, `hasStateStorage`, `secondary`, `removeOnModeChange` (RLI name), `filteredApply`, `deathRemovalExempt` | booleans | Code-tested conditions; `timed` is the condition RLI calls `expiryFlag` and `hasStateStorage` the one RLI calls `usesResolvedStats` (RLI section 3); unidentified category meanings STL-Q06 |
| State membership | Active and changed state sets | Bounded state identities | State count determines valid identities |
| Frame event | Type, target, frame, override, two parameters | Signed type; i32 frame/parameters | Timed-state callers use type 12/no override/parameters 0,0 |
| State definitions | GDF-R03/GDA-R13 | Named fields | `nosend`, `disguise`, `notondead`, player/monster/boss stay-death flags |
| Stay-death masks | Supplied player/monster/boss state sets | Membership bits | Consumer STL-R12a; producer STL-R12b |
| Skill fields | `Skills.aurastate`, `auralencalc`, `aurastat1..6`, `aurastatcalc1..6` | i16 selectors; expression references | State, duration and six ordered stat slots |
| Application inputs | Skill count, skill identity/level, skill record (present or missing), caster, target | i32 identity/level; optional record; optional caster | Expression inputs and effect owner |
| Client selectors | `States.setfunc`, `remfunc`, `cltevent`, `clteventfunc` | i16 | Select dispatch behavior, not callback bodies |
| Boss condition | GDA-R08a `MonStats.boss`, through CSA-R02 `recordGate` | boolean field; i32 selector 0 or 64 | Selects boss death partition |
| Unit timer entries | Kind, key, busy, pending removal | i32 kind/key; booleans | Kind 1/key state cancellation |
| Auxiliary effect | Optional effect and expiry | u32 expiry | Independent type 6 lifetime |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Event types 12/6, type bound 15, 64 frame slots | Engine code | Fixed by original code |
| Unit/dead-mode tests above | Engine code; names through GDF | Fixed by original code |
| State/stat identities in policy table and vectors | GDF-R02/GDF-R03 data identity | Appliers use fixed identities; skill selectors resolve from data |
| Definition fields and stay-death membership | Table schema, GDA-R13 | Resolved from data; execution meaning established separately |
| Six stat slots; attackrate 68 and paired stat 69 | Engine code | Fixed by original code; paired operation STL-Q07 |

Vectors supply masks and selectors explicitly; they do not claim archive values for unnamed rows. Actual named cells require their own table-data evidence.

## 4. Rules (execution order)

**STL-R01** Schedule [verified in 1.13c code (static)]. Signed type >=15 produces no event. Frame -1 uses the immediate path. Otherwise signed `requestedFrame <= now` stores `now+1`; a future frame is unchanged. Addition wraps at 32 bits. For nonnegative stored frames, slot=`frame mod 64`. Unit-type groups are player, monster, missile, object and item. Store target, type, frame, override and parameters and link the event at the head of its target unit's event chain, so that unit's newest event comes first (a null target is recorded as type 6/id -1). A type 2 request for a monster holding state 54 first makes one further request owned by [combat-status-application](combat-status-application.md) (CSA-Q02); timed-state callers never request type 2. Timed-state callers supply type 12, no override and parameters 0,0. Negative event types and wrapped-negative frame indexing remain STL-Q10.

**STL-R02** Exact-frame dispatch [verified in 1.13c code (static)]. In the current slot, fire only events whose stored frame equals current frame. Retain other events. In every group, including missiles, invoke a supplied override when present, otherwise the group's event-type handler; remove the fired event afterward. Without an override, missiles call their single handler regardless of event type.

**STL-R02a** Queue ordering [unresolved]. Candidate group order is missiles, players, monsters, objects, items, immediate queue before ordinary slot in each group, oldest-first within a slot. Checked decoding and insertion/dispatch mutation behavior are required for acceptance (STL-Q07).

**STL-R03** State-expiry routing [verified in 1.13c code (static)]. Player, monster and item type 12 handlers call `ExpireLists(unit,now)`. Monster dispatch, without an override, consults a monster gate that can drop events of other types; for types 3,4,5,8,12 it does not consult that gate and always calls the type handler. The gate's own condition is outside this contract. The event names no particular list/state. The object dispatch table has no type 12 handler; its failure behavior remains STL-Q10.

**STL-R04** Expiry traversal [verified in 1.13c code (static)]. Return for null unit, missing base or a base without state storage. Walk primary lists only, saving each successor before testing. When now!=0, timed lists with signed expiry<=now are due; untimed lists are never due. When now==0, decrement each visited timed expiry with 32-bit wrap, store the result whether or not it is due, and treat signed result<1 as due. For a due list whose own hasStateStorage condition is false request FreeList, then restart at the current primary head. A due list with hasStateStorage true is retained but still restarts; this can loop indefinitely. Actual free effects are STL-R05; supplied fixtures specify chain mutation. Countdown producers are STL-Q09.

**STL-R05** Free effects [unresolved]. Candidate observable behavior removes the list's applied contributions before invoking its callback with associated unit, state and still-readable list, then releases it. Secondary lists do not invoke that callback. Complete propagation/indirect effects have not been read through all output writes (STL-Q07); a callback-order reading alone does not establish stored totals.

**STL-R06a** Creation/expiry primitives [verified in 1.13c code (static)]. Successful allocation initializes supplied owner type/id, list conditions and expiry, with other semantic fields zero/empty. State, callback and expiry setters ignore null lists. Setting expiry on a non-null list always stores it; a positive value sets timed, and a nonpositive value never clears timed. Null-list expiry read returns 0. Lookup consumes RLI-R03: state 0 returns null; otherwise the first primary match precedes secondary matches, with no state-range test.

**STL-R06b** Attach [unresolved]. Candidate observable behavior attaches secondary lists without applying stats, and primary lists newest-first with full or filtered stat application. Cycle refusal, reattachment callbacks, extended-list processing and stored totals require complete verification (STL-Q07). Timed states request full application.

**STL-R06c** Replacing list-stat write [unresolved]. Candidate behavior replaces the stat/layer value, does nothing for equality, removes a zero entry, and propagates new-minus-old. Direct replacement control flow is read; entry management, propagation and special-list effects remain STL-Q07.

**STL-R07** State membership [verified in 1.13c code (static)]. Signed s<0 or s>=stateCount makes no request. Otherwise request the active-bit operation carrying on unchanged, then update marking. With state storage present, nonzero on sets active membership and zero clears it; only a membership change sets changed membership. No list is allocated or found. A state can be active without a list and a list can exist with its state off. Disguise is STL-R07a; update-queue integration is RLI-Q02.

**STL-R07a** Disguise [unresolved]. A changed disguise state is intended to update the unit's disguise condition; clearing must account for another active disguise state. Byte decoding and the full state-test helper remain STL-Q07; no guessed default is established.

**STL-R08** Forced changed membership [verified in 1.13c code (static)]. For non-null unit with state storage and valid signed state, set/clear changed membership according to nonzero/zero on, leaving active membership unchanged; request update even when the changed bit already matched. Invalid state or missing unit/list/storage makes no request. Update-queue effects remain RLI-Q02.

**STL-R09** Per-applier refresh [verified in 1.13c code (static)]. There is no shared policy. Consume the contracts below, including when each applier requests another type 12 event (freeze requests one even when its expiry is unchanged; STL-R10a requests one on every successful application). A queued event for an earlier scheduled expiry can fire with no due list.

| Applier | State | Existing expiry | Values | Contract |
|---|---|---|---|---|
| Cold |11 cold|Extend only|Unchanged|CSA-R05|
| Freeze |1 freeze|Extend only|No magnitude|CSA-R08|
| Stun |21 stunned|Replace, can shorten|No magnitude|CSA-R04|
| Poison/burn |2 poison/115 burning|Replace when new magnitude>=old|Replace hpregen|CSA-R12|
| Timed skill |Supplied aurastate|Replace, can shorten|Stat-slot operation|STL-R10a/STL-R10b|

**STL-R10a** Timed skill-state request order [verified in 1.13c code (static)]. A skill identity outside 0..skillCount-1 (signed), a missing skill record, or aurastate outside 0..stateCount-1 (signed) returns 0 with no request. Only after these checks is the duration evaluated: the skill's `auralencalc` through SCX-R11's skill front end with context (caster, skill identity, level), the caster being null when absent; expiry is that supplied result+now with 32-bit wrap. Look up the target state list. If absent, request allocation with expiry, the timed condition as its only list condition, and caster type/id, or 6/-1 without caster; allocation failure returns 0. For a new list, set state, install generic removal callback, request full attachment (its result is ignored), request active state on. Both paths then request forced changed on, store expiry, schedule type 12, invoke stat-slot writing, request passive refresh, request pet update only for a non-null player target (caster presence does not matter), and return 1. Actual attach/write/refresh effects remain STL-Q07; supplied request order is available independently.

**STL-R10b** Skill stat slots [unresolved]. Visit six slots in order. Invalid stat or zero expression result skips writing; zero leaves a stored value. Valid nonzero result requests a layer 0 list-stat write. Stat 68 also requests a separate operation for 69. The paired operation forwards to the same replacing list-stat writer after a null-list guard. Checked decoding of this caller and complete write/refresh effects remain STL-Q07; the shared writer is STL-R06c.

**STL-R11** Generic removal callback [verified in 1.13c code (static)]. Null unit is fatal. First request cancellation of unit timers kind 1/key state. Alive means aliveExclusionCondition false and either player mode outside{0,17} or monster mode outside{0,12}; all other types count as dead. If dead and StayDeath nonzero, return with state unchanged. Otherwise request state off, animation-rate refresh, passive refresh, and pet update for a player. These are ordered requests; refresh effects remain STL-Q07. A cold list carries CSA-R13's cold expiry callback instead of this one; that callback makes no timer cancellation, passive refresh or pet request, and its ordered requests and vector (CSA-V15) belong to CSA. A retained state does not retain the already-removed list.

**STL-R11a** Unit-timer cancellation [verified in 1.13c code (static)]. Missing unit/timer chain does nothing. Save successor before each entry. Matching kind/key: busy entry becomes pending removal, otherwise detach/release it. Nonmatching entries remain. These timers are separate from frame-event type 12 queues.

**STL-R12a** StayDeath consumer [verified in 1.13c code (static)]. Invalid signed state returns 0. A non-null monster selects monster mask; every other unit (player, object, missile, item or null) selects player mask. Return 2^(s mod 32) as an unsigned 32-bit value when the selected mask contains state s and 0 otherwise, not normalized 1; callers in this contract test only zero/nonzero. There is no boss branch. Supplied-mask acceptance does not establish STL-R12b's producer.

**STL-R12b** Stay-death mask producer [unresolved]. Candidate producer groups states by definition flags, including player/monster/boss stay-death fields. Loader bit-setting and byte decoding remain STL-Q07. The consumer can use supplied masks without asserting particular table cells.

**STL-R13a** Death requests/traversal [verified in 1.13c code (static)]. Non-staying-list removal returns without requests for null unit, missing base or a base without state storage. Otherwise walk the primary chain, saving each successor. Test in order: an item-owned list (owner type 4), a deathRemovalExempt list, then a list for which StayDeath(dying unit, list state) is nonzero is retained and the walk continues at the saved successor. Any other list is not retained: request FreeList unless its hasStateStorage condition is true, then restart at the current head in both cases (a hasStateStorage list can make this loop indefinitely, STL-Q05). There are two separate death operations, chosen by their callers rather than by a type test inside one operation. Player death requests a replacing unit-stat write (stat 6 `hitpoints` by GDF-R02, value 0, layer 0), non-staying-list removal, state partition with boss selector 0, then removal of the unit's pending type 8 and type 9 frame events. Monster death first, when an auxiliary effect is present, requests its removal (one logged request; its effects stay STL-Q07), then non-staying-list removal, then partition with boss selector `recordGate(monster's MonStats record, unit)` (CSA-R02: 64 when that record's GDA-R08a `boss` is set, else 0; vectors supply the record's `boss` value). Monster death completes only for a monster with loaded monster data: after the auxiliary step, any other unit (including a non-monster that had an auxiliary effect) reads its monster definition through a null reference and faults before any list request; a null unit skips the auxiliary step and faults the same way. Both death operations make further requests between these steps (for example player death's position/collision request after the partition) that are outside this contract; the ordered logs below list only lifecycle requests. Actual free/auxiliary effects and mask production remain STL-Q07; supplied-mask partition is STL-R13b.

**STL-R13b** Death partition [verified in 1.13c code (static)]. Return without requests for null unit, absent state storage or missing active/changed arrays (the two arrays are present or missing together). Test the boss selector first: a nonzero boss selector selects the boss mask for any unit type, including a player. Otherwise select the supplied player mask for a player and the monster mask for every other unit. Membership is processed as ceil(stateCount/32) words, state s being bit (s mod 32) of word floor(s/32). For each state word, first write `changed = changed OR (active AND NOT mask)`, then `active = active AND mask`. Request update marking afterward, including when state count is zero but the arrays exist. Actual mask production remains STL-R12b and update-queue integration RLI-Q02; supplied-mask state writes are independently verified.

**STL-R14** State-message selection [unresolved]. Candidate sender visits changed states with nosend false: inactive selects state-off 0xA9; active without list/exported stats selects state-on 0xA7; otherwise state-with-stats 0xA8 exports at most 16 stats. Real sync caller, eligibility, exporter and changed-state clearing remain STL-Q01/STL-Q07. Changed membership alone does not guarantee delivery.

**STL-R15** Client dispatch [unresolved]. Candidate set dispatch skips invalid unit/state and dead player/monster with notondead; valid non-null set selectors 0..30 invoke that callback. Remove selectors 0..29 invoke non-null selected callback. Client-event handling follows. Packet paths and event-operation arguments remain STL-Q02/STL-Q07; callback bodies/message layout are outside scope.

**STL-R16** Level/act transition [unresolved]. Retention/removal across transition is not established; STL-Q03 supplies no default.

**STL-R17** Auxiliary lifetime boundary [asm-only]. A separate event type 6 on an object compares auxiliary expiry to current frame unsigned. Future expiry requests type 6 rescheduling; due expiry requests removal, clears the auxiliary reference and requests update. No state identity is involved. Complete auxiliary removal/update effects remain STL-Q07.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| STL-R01 | Schedule | verified in 1.13c code (static) | STL-Q10: malformed/wrapped domain only | STL-V01, STL-V04 |
| STL-R02 | Exact frame | verified in 1.13c code (static) | none within supplied bucket | STL-V02, STL-V16 |
| STL-R02a | Queue order | unresolved | STL-Q07: group/immediate/insertion order | STL-V10 |
| STL-R03 | Expiry handler | verified in 1.13c code (static) | STL-Q10: object failure only | STL-V02, STL-V16 |
| STL-R04 | Traversal | verified in 1.13c code (static) | STL-Q07: actual free only | STL-V03, STL-V04, STL-V11, STL-V17 |
| STL-R05 | Free effects | unresolved | STL-Q07: complete free | STL-V03 |
| STL-R06a | Creation/expiry | verified in 1.13c code (static) | none at supplied-allocation boundary | STL-V05, STL-V17; RLI-V05 |
| STL-R06b | Attach | unresolved | STL-Q07: attach effects | none |
| STL-R06c | Replacing write | unresolved | STL-Q07: replacing effects | STL-V05 |
| STL-R07 | Membership | verified in 1.13c code (static) | STL-Q07: disguise only; RLI-Q02: changed-state/update-queue effects | STL-V06 |
| STL-R07a | Disguise | unresolved | STL-Q07: disguise condition | none |
| STL-R08 | Forced change | verified in 1.13c code (static) | RLI-Q02: changed-state/update-queue effects | STL-V12, STL-V17 |
| STL-R09 | Refresh policy | verified in 1.13c code (static) | Applier integration gates only | STL-V07 |
| STL-R10a | Skill requests | verified in 1.13c code (static) | STL-Q07: actual write/refresh | STL-V07, STL-V18 |
| STL-R10b | Skill slots | unresolved | STL-Q07: paired operation/refresh | none |
| STL-R11 | Removal callback | verified in 1.13c code (static) | STL-Q07: actual refresh | STL-V08 |
| STL-R11a | Timer cancellation | verified in 1.13c code (static) | none for supplied chain | STL-V13 |
| STL-R12a | StayDeath consumer | verified in 1.13c code (static) | STL-Q07: producer only | STL-V08, STL-V15 |
| STL-R12b | Mask producer | unresolved | STL-Q07: mask production | none |
| STL-R13a | Death traversal | verified in 1.13c code (static) | STL-Q07: free/partition/auxiliary effects | STL-V09, STL-V19 |
| STL-R13b | Death partition | verified in 1.13c code (static) | STL-Q07: data-derived mask producer only; RLI-Q02: changed-state/update-queue effects | STL-V14 |
| STL-R14 | Messages | unresolved | STL-Q01/STL-Q07: sender integration | none |
| STL-R15 | Client callbacks | unresolved | STL-Q02/STL-Q07: dispatch integration | none |
| STL-R16 | Transition | unresolved | STL-Q03: transition retention | none |
| STL-R17 | Auxiliary boundary | asm-only | STL-Q07: auxiliary effects | none |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| Scheduled event at clamped frame, linked first in its unit's event chain; fired event removed | STL-R01/STL-R02 |
| Empty list/owner/expiry fields; timed condition retained for nonpositive expiry | STL-R06a |
| Active/changed membership and update requests | STL-R07/STL-R08 |
| Due-list free requests and head reload | STL-R04 |
| Candidate contribution removal/callback/release, gated | STL-R05 |
| Candidate stored attach/replacing-write effects, gated | STL-R06b/STL-R06c |
| Ordered skill allocation/attach/stat/refresh requests; duration evaluation and its context; return 0/1 | STL-R10a |
| Timers detached/released or marked pending | STL-R11a |
| State-clear/animation/passive/pet requests or retained state | STL-R11 |
| Death cleanup requests and candidate partition | STL-R13a/STL-R13b |
| ExpireLists invocation by the type 12 handler | STL-R03 |
| Handler/override invocations (none at scheduling time; fired event still queued during its handler) | STL-R01/STL-R02 |
| Expiry getter value (0 for a null list) | STL-R06a |
| Traversal visit order (fixture observation of list examination) | STL-R04/STL-R11a/STL-R13a |
| StayDeath return value 0 or 2^(s mod 32) | STL-R12a |
| Fatal assertion (null removal-callback unit) or fault (monster death of a unit without monster data); no later request | STL-R11/STL-R13a |
| Candidate state-on/state-with-stats/state-off messages and callbacks | STL-R14/STL-R15 |
| Auxiliary removal/reschedule requests, separate lifetime | STL-R17 |

No local rule creates units or writes saves. Requests establish actual stored/delivered effects only under accepted dependency contracts.

## 6. Branches and edge cases

Zero/negative duration still schedules an ordinary expiry clamped to next frame. Nonpositive expiry writes never clear timed. Any type 12 event visits every due primary list; an extended list can survive an event at its scheduled expiry. Secondary lists are not traversed. An expiring stay-death state can retain membership after list contributions disappear; resurrection cleanup is STL-Q04. A retained due list with state storage can make head restart nonterminating (STL-Q05).

State lookup and membership use different guards (RLI-R03). Missing caster uses owner 6/-1. Outside explicit unit/dead-mode/boss branches, no local rule branches on difficulty, PvP, melee/missile origin or classic/expansion; evidence remains expansion 1.13c. Skill allocation failure returns 0. The nonplayer death cleanup faults for a non-monster unit before list removal (STL-R13a). Malformed event/frame behavior is STL-Q10.

## 7. Open questions

| Question | Kind / status | Evidence needed |
|---|---|---|
| **STL-Q01** | integration / unresolved | Real sync caller/predicate, exporter and changed-state clearing; actual messages only |
| **STL-Q02** | integration / unresolved | Packet handlers to set/remove/client-event dispatch; callback bodies remain outside |
| **STL-Q03** | integration / unresolved | Transition dispatch-to-list/state trace, retention/removal outcome |
| **STL-Q04** | outside-boundary / unresolved | Resurrection owner's retained-state cleanup; no local default |
| **STL-Q05** | outside-boundary / unresolved | Producer of primary hasStateStorage lists; supplied traversal remains available |
| **STL-Q06** | outside-boundary / unresolved | Wider meanings/producers of deathRemovalExempt and aliveExclusionCondition |
| **STL-Q07** | behavior / unresolved | Complete callee/decoding/output-write paths for each gated rule; separate boundaries in gap file |
| **STL-Q08** | observation / pending | Matching original 1.13c captures; none recorded |
| **STL-Q09** | outside-boundary / unresolved | Caller supplying 0 to countdown expiry; ordinary handlers supply game frame |
| **STL-Q10** | behavior / unresolved | Malformed signed type/wrapped frame and object type 12 failure behavior |

Attempts and exact blocked assertions are in the [gap file](../plans/gaps/state-timed-state-lifecycle.md). Supplied masks/operation returns do not verify their producers. Pending observation limits confirmation, not independently verified supplied-input assertions.

## 8. Test vectors

### 8.1 Executable vectors

Each vector uses supplied dependencies that log calls in order. FreeList removes the requested list; bit operations persist active/changed membership; allocation yields the expressly listed empty fields. Refresh operations log only. A supplied stat-write/removal model supports only explicitly gated assertions of STL-R05/STL-R06b/STL-R06c, not verified integrated effects. Each case starts independently unless a continuation is stated. StayDeath in STL-V08, STL-V09 and STL-V19 is STL-R12a evaluated over the supplied masks (a true bit for state 21 returns 0x00200000, not 1), not a stub. Unless a vector says otherwise, units have a base with state storage and lists have hasStateStorage false. In ordered request logs, the update-marking requests made inside SetState/ForceChanged are not separate entries; vectors that count update requests (STL-V06, STL-V12, STL-V14, STL-V17) say so. The six-slot stat operation is one supplied request logged once as `stat slots`. A visit is one examination of a list or unit-timer entry by a traversal, whether the list is then retained, freed or skipped. "Fatal" and "fault" both mean the operation does not return normally and makes no later request.

- **STL-V01** Schedule (STL-R01; boundary: supplied-input stage)
  - Inputs: now 1000; monster; type 12; no override; parameters 0,0; empty queues; event-handler invocation log initially empty; requested frames per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: Signed compare/clamp then remainder by 64.
  - Outputs: per independent case below.
  - (a) Inputs: requested frame 1000. Draws: none. Outputs: exactly one queued event, stored frame 1001, slot 41, monster group; immediately after scheduling at frame 1000 the handler invocation log is empty (count 0).
  - (b) Inputs: requested frame 999. Draws: none. Outputs: stored frame 1001, slot 41, monster group.
  - (c) Inputs: requested frame 1050. Draws: none. Outputs: stored frame 1050, slot 26, monster group.
  - (d) Inputs: requested frame -5. Draws: none. Outputs: stored frame 1001 in the monster group's ordinary slot 41; the immediate queue stays empty.
  - (e) Inputs: type 15, requested frame 1050. Draws: none. Outputs: no record.
  - (f) Inputs: requested frame -1. Draws: none. Outputs: exactly one record, in the monster group's immediate queue; no ordinary-slot record.
  - (g) Inputs: the monster's event chain already holds E0 (type 12, frame 1050); requested frame 1060. Draws: none. Outputs: new event stored frame 1060, slot 36; the monster's event chain is [new event, E0].
  - Discriminates: the frame-1000 scheduling call queues one future event and makes zero synchronous handler calls; storing 1001 while also invoking a handler at 1000 fails case (a); only frame -1 takes the immediate path (d, f); a tail append fails (g).
  - runtime_observed: no

- **STL-V02** Dispatch (STL-R02/STL-R03; boundary: supplied-input stage)
  - Inputs: now 1050; P player with type 12 events 986/1050/1114 in slot 26; M monster with type 12 event 1050; no overrides; valid bases/empty chains
  - Seeds: `no draw`
  - Draws: none
  - Trace: Match equality, not slot identity, then call unit expiry handler.
  - Outputs: Calls ExpireLists(P,1050) once and ExpireLists(M,1050) once; remove their 1050 events; retain P 986 (not fired at 1050) and P 1114 until frame 1114. Player-before-monster is gated STL-R02a.
  - Discriminates: 986 and 1114 share slot 26 but must not fire at 1050, so neither a slot-only nor a `frame <= now` test passes.
  - runtime_observed: no

- **STL-V03** Expiry restart (STL-R04/STL-R05; boundary: supplied-input stage)
  - Inputs: M monster; now 1050; primary A,B,C; A timed/state 11/expiry 1080/stats 67,68,69 each-50; B timed/state 2/expiry 1050/hpregen 74=-100; C untimed with expiry 1000; no list has state storage; secondary empty
  - Seeds: `no draw`
  - Draws: none
  - Trace: Visit A,B; supplied FreeList removes B; reload head and visit A,C. Continue independently at 1080 with resulting A,C.
  - Outputs: At 1050 visit[A,B,A,C], free[B], chain[A,C]; C is not freed although its expiry 1000 is below 1050; at 1080 visit[A,C], free[A], chain[C]. Gated STL-R05: B removal increases hpregen total 100 before callback(M,2,B); A removal increases totals 67/68/69 by 50 each.
  - Discriminates: Head restart revisits A and retains future expiry 1080; an untimed list is never due.
  - runtime_observed: no

- **STL-V04** Zero duration (STL-R01/STL-R04/STL-R06a; frame-1001 routing per STL-R02/STL-R03 as in STL-V02; boundary: supplied-input stage)
  - Inputs: now 1000; monster freeze list timed/expiry 1000 from supplied duration 0; requested event 1000; supplied FreeList removes it; separate timed L expiry 5 on a different unit's primary chain, then an STL-R06a expiry write of 0 to L
  - Seeds: `no draw`
  - Draws: none
  - Trace: Clamp event 1000 to 1001;1000<=1001 requests free. Set L expiry 0, retaining timed; positive next event sees 0<=now.
  - Outputs: Event 1001; immediately after scheduling at frame 1000 the free log is empty (no free request before frame 1001); at frame 1001 it calls ExpireLists(unit,1001), which requests free. L expiry 0/timed true; ExpireLists on L's unit at frame 1 (below the input expiry 5), requests free of L.
  - Discriminates: Zero duration does not remove synchronously or clear timed.
  - runtime_observed: no

- **STL-V05** Creation/expiry/write (STL-R06a/STL-R06c; boundary: supplied-input stage)
  - Inputs: successful allocation timed/expiry 1050/owner 0,7; independent untimed L; independent untimed L2 with expiry 7; independent empty stat 67/layer 0 and parent total 0 for gated write model; valid unit u with base whose only list is a primary list with state 0
  - Seeds: `no draw`
  - Draws: none
  - Trace: Initialize; write expiry 0 then -1 to new list,5 to independent untimed L,0 then -1 to independent untimed L2. Gated model writes-50 twice then 0. Lookup state 0.
  - Outputs: New list timed/1050/owner 0,7, other semantic fields 0/empty; after expiry 0 it is timed/0 and after -1 timed/-1; untimed L becomes timed/expiry 5; L2 stays untimed, expiry 0 then -1. Gated STL-R06c: two writes leave one-50 entry/total-50, then 0 removes entry/total 0 (delta+50). FindStateList(u,0)=null although u has a state-0 list.
  - Discriminates: Repeated replacement never yields-100; nonpositive expiry neither clears nor sets timed; state 0 lookup returns null without searching.
  - runtime_observed: no

- **STL-V06** Membership separation (STL-R07; boundary: supplied-input stage)
  - Inputs: stateCount 185; valid base; non-disguise state 21 active=false/changed=false; supplied update log
  - Seeds: `no draw`
  - Draws: none
  - Trace: Set 21 on, clear changed by a direct fixture write (no operation call, no update request), set 21 on again, attempt 185 on, set 21 off, attempt -1 on.
  - Outputs: First active 21=true/changed 21=true and update; second active 21=true/changed 21=false and update; 185 adds no changes/request; setting 21 off gives active 21=false/changed 21=true and the third update request; -1 adds no changes/request; no list allocated.
  - Discriminates: Equal membership does not force changed membership; wrapper still requests update.
  - runtime_observed: no

- **STL-V07** Skill replacement (STL-R09/STL-R10a; boundary: supplied-input stage)
  - Inputs: valid skill identity 1 and level 1 within a supplied skill count 2; aurastate 21/state count 185; current frame 1000; duration expression result 50; caster owner 0,7; monster target with state storage; supplied allocation succeeds; stat-slot/passive operations log only; no queued events
  - Seeds: `no draw`
  - Draws: none
  - Trace: 1000+50=1050 independently of stored expiry; execute the new-list prefix only when lookup returns no list.
  - Outputs: per independent case below; actual attachment/stat effects remain gated.
  - (a) Inputs: existing primary list for state 21, expiry 1200, active 21 true/changed 21 false. Draws: none. Outputs: no allocation/attachment; ordered[forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; active 21 true/changed 21 true; return 1; no pet update.
  - (b) Inputs: no list for state 21 (an unrelated primary list for state 11 exists), active 21 false/changed 21 false. Draws: none. Outputs: ordered[allocate(timed,1050,owner 0,7),set list state 21,set generic callback,full attach request,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; supplied bit operations leave active/changed 21 true; return 1; no pet update.
  - (c) Inputs: comparison-only policy fixture, supplied existing expiry 1200 and candidate expiry 1050; consume the CSA-R05 strict-extension comparison. Boundary: isolated comparison, with no cold-applier invocation. Seeds: `no draw`. Draws: none. Trace: 1200 < 1050 is false. Outputs: selected expiry 1200. Full cold-application requests and target-seed draws belong to CSA-V08/CSA-R06; this case asserts only the supplied-expiry comparison result.
  - (d) Inputs: existing primary list for state 21, expiry 900, active 21 true/changed 21 false. Draws: none. Outputs: no allocation/attachment; ordered[forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; stored expiry 1050; return 1; no pet update.
  - (e) Inputs: player target with state storage; existing primary list for state 21, expiry 1200, active 21 true/changed 21 false. Draws: none. Outputs: no allocation/attachment; ordered[forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh,pet update]; return 1.
  - Evaluation log for (a), (b), (d), (e): exactly one duration evaluation with context (caster 0,7, skill 1, level 1), made before the lookup.
  - Discriminates: skill refresh stores the new expiry whether shorter (a) or longer (d), so neither keep-maximum nor keep-minimum passes; cold does not shorten; the existing-list skill path skips allocation and does not request active state on; the pet request follows the target type on the existing-list path too (e).
  - runtime_observed: no

- **STL-V08** Removal gate (STL-R11/STL-R12a; boundary: supplied-input stage)
  - Inputs: state 21/count 185/active true; aliveExclusionCondition false; cancellation/refresh operations log; state writer persists; masks and unit type/mode supplied per independent case
  - Seeds: `no draw`
  - Draws: none
  - Trace: cancel(kind 1,key 21) first; choose mask and alive branch; return or clear/refresh.
  - Outputs: per case below; no list retention is asserted.
  - (a) Inputs: player mode 1, player mask bit true. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive,pet]; active false.
  - (b) Inputs: player mode 17, player mask bit true. Draws: none. Outputs: cancel(1,21) only; active true.
  - (c) Inputs: player mode 17, player mask bit false. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive,pet]; active false.
  - (d) Inputs: monster mode 12, monster mask bit true/player bit false. Draws: none. Outputs: cancel(1,21) only; active true.
  - (e) Inputs: object mode 1, player mask bit true/monster bit false. Draws: none. Outputs: cancel(1,21) only; active true.
  - (f) Inputs: player mode 1, aliveExclusionCondition true, player mask bit true. Draws: none. Outputs: cancel(1,21) only; active true.
  - (g) Inputs: null unit. Draws: none. Outputs: fatal assertion before any cancellation, state write or refresh; operation does not return normally.
  - (h) Inputs: player mode 0, player mask bit true. Draws: none. Outputs: cancel(1,21) only; active true.
  - (i) Inputs: player mode 12, player mask bit true. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive,pet]; active false.
  - (j) Inputs: monster mode 0, monster mask bit true/player bit false. Draws: none. Outputs: cancel(1,21) only; active true.
  - (k) Inputs: monster mode 12, monster mask bit false/player bit true. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive]; active false; no pet request.
  - (l) Inputs: monster mode 1, monster mask bit true/player bit true. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive]; active false; no pet request.
  - (m) Inputs: monster mode 17, monster mask bit true/player bit true. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive]; active false; no pet request.
  - (n) Inputs: monster mode 1, aliveExclusionCondition true, monster mask bit true/player bit false. Draws: none. Outputs: cancel(1,21) only; active true.
  - Discriminates: monster selects monster mask (d, k); each type's dead set is exactly its own pair, so player 12 and monster 17 are alive (i, m) and both mode-0 cases are dead (h, j); the additional death condition applies to monsters too (n); object never counts alive; the additional death condition changes a living-mode player to the retained-state branch; a living unit ignores a true mask bit (a, l); null asserts before callbacks; only players get the pet request.
  - runtime_observed: no

- **STL-V09** Death traversal (STL-R13a; boundary: supplied-input stage)
  - Inputs: player death operation; player unit with base and state storage; primary A,B,C; A owner 4/timed/state 11; B owner 0/timed/state 11; C owner 0/timed/state 21; supplied player mask contains only state 21, supplied monster mask contains only state 11; all deathRemovalExempt=false/hasStateStorage=false; secondary empty; supplied FreeList removes B
  - Seeds: `no draw`
  - Draws: none
  - Trace: Request the stat write, traverse A (item-owned, retained), B (player mask false, freed), restart at head A, then C (player mask true, retained), then partition/event removals.
  - Outputs: Visits[A,B,A,C]; ordered[write(stat 6,value 0,layer 0),FreeList(B),partition(0),remove type 8 events,remove type 9 events]; chain[A,C]. Actual callback/contributions gated STL-R05; actual mask production gated STL-R12b; the supplied-mask partition is STL-R13b.
  - Discriminates: Item owner keeps A although its state is outside the player mask; the player mask, not the monster mask, frees B and keeps C; the revisit of A shows restart at the head rather than continuing at B's successor. Swapping the stat write's value and layer is equivalent here because both are 0 in the contract.
  - runtime_observed: no

- **STL-V10** Queue order candidate (STL-R02a; boundary: supplied-input stage)
  - Inputs: now 1050; one immediate and one exact-frame event in each of five groups; logging handlers
  - Seeds: `no draw`
  - Draws: none
  - Trace: Candidate order is immediate then ordinary per group.
  - Outputs: Gated order[missile immediate,missile ordinary,player immediate,player ordinary,monster immediate,monster ordinary,object immediate,object ordinary,item immediate,item ordinary].
  - Discriminates: Rejects all-immediates-before-all-ordinary ordering.
  - runtime_observed: no

- **STL-V11** Countdown restart (STL-R04; boundary: supplied-input stage)
  - Inputs: now 0; primary A timed/expiry 2,B timed/expiry 1,C untimed/expiry 0; none hasStateStorage; supplied FreeList removes requested list
  - Seeds: `no draw`
  - Draws: none
  - Trace: A decrements 1; B decrements 0/free; restart A decrements 0/free; C untouched.
  - Outputs: Visits[A,B,A,C], free[B,A], chain[C], C expiry 0.
  - Discriminates: Restart can decrement a preceding list more than once.
  - runtime_observed: no

- **STL-V12** Forced change (STL-R08; boundary: supplied-input stage)
  - Inputs: non-null unit/state storage; count 185/state 21; active=false/changed=false; update log
  - Seeds: `no draw`
  - Draws: none
  - Trace: Force changed on twice then off.
  - Outputs: Changed values[true,true,false]; active=false throughout; three update requests.
  - Discriminates: Equal changed bit still requests update; active bit is independent.
  - runtime_observed: no

- **STL-V13** Timer cancellation (STL-R11a; boundary: supplied-input stage)
  - Inputs: kind 1/key 21; chain A(kind 1,key 21,nonbusy),B(kind 1,key 21,busy,pending=false),C(kind 1,key 11,nonbusy),D(kind 2,key 21,nonbusy)
  - Seeds: `no draw`
  - Draws: none
  - Trace: Save successor, release A, defer B, skip C,D.
  - Outputs: Visits[A,B,C,D]; chain[B,C,D]; B pending=true; only A released.
  - Discriminates: Busy match is retained; kind-only/key-only matching is insufficient.
  - runtime_observed: no

- **STL-V14** Supplied-mask death partition (STL-R13b; boundary: supplied-input stage)
  - Inputs: state count 185; non-null unit with state storage and six-word active/changed arrays; first active word 0x00200800 (states 11/21), first changed word 0x00000008 (state 3); other words zero; supplied player mask first word 0x00200000, monster mask 0x00000800, boss mask 0x00200800, remaining words zero; update operation logs only
  - Seeds: `no draw`
  - Draws: none
  - Trace: select mask, OR removed active bits into changed, AND active with mask, then request update.
  - Outputs: per independent case below; cases (a)-(c) leave all other words zero and request exactly one update.
  - (a) Inputs: player, boss selector 0. Draws: none. Outputs: active first word 0x00200000, changed 0x00000808.
  - (b) Inputs: monster, boss selector 0. Draws: none. Outputs: active first word 0x00000800, changed 0x00200008.
  - (c) Inputs: monster, boss selector 1. Draws: none. Outputs: active first word 0x00200800, changed 0x00000008.
  - (d) Inputs: player/boss selector 0, state storage absent. Draws: none. Outputs: no array writes and no update request.
  - (e) Inputs: player/boss selector 0, state storage present but active/changed arrays missing. Draws: none. Outputs: arrays remain missing and no update request.
  - (f) Inputs: player/boss selector 0, state count 0, both arrays present and empty. Draws: none. Outputs: empty arrays unchanged and exactly one update request.
  - (g) Inputs: player/boss selector 0; all array/mask words zero except active word 5=0x01000000 and changed word 5=0x00800000; player mask word 5=0. Draws: none. Outputs: active word 5=0, changed word 5=0x01800000; other words zero; exactly one update request.
  - (h) Inputs: player, boss selector 2. Draws: none. Outputs: active first word 0x00200800, changed 0x00000008; exactly one update request.
  - Discriminates: nonplayer uses monster mask unless boss selector is nonzero; any nonzero boss selector takes precedence for a player too (h); existing changed bits survive; missing arrays skip update but a zero-count present array requests it; processing reaches word 5.
  - runtime_observed: no

- **STL-V15** StayDeath mask return (STL-R12a; boundary: supplied-input stage)
  - Inputs: state count 185; player mask contains only state 21 unless a case supplies its own player mask; monster mask empty; independent unit/state cases below
  - Seeds: `no draw`
  - Draws: none
  - Trace: validate signed state, choose monster mask only for a non-null monster, test state bit without normalization.
  - Outputs: per case; no membership or list changes.
  - (a) Inputs: null unit, state 21. Draws: none. Outputs: 0x00200000.
  - (b) Inputs: player, state 21. Draws: none. Outputs: 0x00200000.
  - (c) Inputs: monster, state 21. Draws: none. Outputs: 0.
  - (d) Inputs: player, state 185. Draws: none. Outputs: 0.
  - (e) Inputs: player, state -1. Draws: none. Outputs: 0.
  - (f) Inputs: object, state 21. Draws: none. Outputs: 0x00200000.
  - (g) Inputs: player, state 53, with the player mask containing states 21, 53 and 54 (word 1, bits 21 and 22). Draws: none. Outputs: 0x00200000.
  - (h) Inputs: player, state 53, player mask containing only state 21. Draws: none. Outputs: 0.
  - (i) Inputs: player, state 22, player mask containing states 21 and 22. Draws: none. Outputs: 0x00400000.
  - (j) Inputs: player, state 19, player mask containing states 20 and 21. Draws: none. Outputs: 0.
  - Discriminates: true membership returns only the queried state's bit, not 1 and not the whole mask word (i, j); a null unit and an object still select the player mask; state 53 is tested in word 1, so a word-0-only reading fails (h), and a reading that returns 1 or the state index fails (g).
  - runtime_observed: no

- **STL-V16** Event handler routes (STL-R02/STL-R03; boundary: supplied-input stage)
  - Inputs: current frame 1050; independent case has one event at frame 1050; callback/type handlers log their semantic calls; ordinary handlers have valid targets and empty primary chains with state storage
  - Seeds: `no draw`
  - Draws: none
  - Trace: match the frame, choose override or group handler, remove the fired record.
  - Outputs: per independent case; event queue empty afterward.
  - (a) Inputs: player P/type 12 event with supplied override H; H records whether its own event is still queued when invoked. Draws: none. Outputs: log[H invoked once]; H observed its event still queued; expiry type handler not called; queue empty after dispatch.
  - (b) Inputs: missile M/type 6 event with no override. Draws: none. Outputs: log[single missile handler invoked once]; no type-6 group-table handler call.
  - (c) Inputs: item I/type 12 event with no override. Draws: none. Outputs: log[ExpireLists(I,1050)].
  - (d) Inputs: monster N/type 12 event with supplied override H. Draws: none. Outputs: log[H invoked once]; monster gate and expiry type handler not called.
  - (e) Inputs: missile M/type 6 event with supplied override H. Draws: none. Outputs: log[H invoked once]; single missile handler not called.
  - (f) Inputs: missile M/type 12 event with no override. Draws: none. Outputs: log[single missile handler invoked once]; no ExpireLists call.
  - (g) Inputs: monster N/type 12 event with no override; supplied monster gate logs any consultation and would drop the event; the type handler records whether its own event is still queued when invoked. Draws: none. Outputs: log[ExpireLists(N,1050)]; gate not consulted; the event was still queued during the handler; queue empty after dispatch.
  - Discriminates: override wins in the player, monster and missile groups; the same single missile handler runs for types 6 and 12 (b, f); without an override, missile dispatch does not use its event-type selector; items share expiry routing; type 12 skips the monster gate and the fired event is removed only after its group handler returns (g).
  - runtime_observed: no

- **STL-V17** Guarded local operations (STL-R04/STL-R06a/STL-R08; boundary: isolated rule)
  - Inputs: independent operations below; state count 185; effect/update request logs initially empty; any present active/changed membership initially false
  - Seeds: `no draw`
  - Draws: none
  - Trace: apply each operation's unit/list/storage/state guard before mutation.
  - Outputs: per case; no allocation, removal or update requests unless explicitly stated.
  - (a) Inputs: ExpireLists(null,1050). Draws: none. Outputs: no visits or effects.
  - (b) Inputs: ExpireLists on unit without base, now 1050. Draws: none. Outputs: no visits or effects.
  - (c) Inputs: ExpireLists on base without state storage and primary timed A/expiry 1000, now 1050. Draws: none. Outputs: A retained/expiry 1000; no visits or free requests.
  - (d) Inputs: null list; set state 21, set callback H, set expiry 1050, then read expiry. Draws: none. Outputs: no writes; expiry getter returns 0.
  - (e) Inputs: ForceChanged on null unit, state 21/on 1. Draws: none. Outputs: no writes or update request.
  - (f) Inputs: ForceChanged on unit without base, state 21/on 1. Draws: none. Outputs: no writes or update request.
  - (g) Inputs: ForceChanged on base without state storage, state 21/on 1. Draws: none. Outputs: no writes or update request.
  - (h) Inputs: ForceChanged on valid state-storage base, state 185/on 1. Draws: none. Outputs: active/changed membership remains false; no update request.
  - (i) Inputs: ForceChanged on valid state-storage base, state -1/on 1. Draws: none. Outputs: active/changed membership remains false; no update request.
  - (j) Inputs: ExpireLists on valid state-storage base, now 1050, single primary timed A with expiry -16 (0xFFFFFFF0), hasStateStorage false; supplied FreeList removes A. Draws: none. Outputs: free[A]; chain empty.
  - Discriminates: missing storage stops traversal even with a due primary list; the expiry comparison is signed, so an unsigned reading leaves A (j); invalid forced marks do not reach update marking.
  - runtime_observed: no

- **STL-V18** Skill refusal and owner/pet routes (STL-R10a; boundary: supplied-input stage)
  - Inputs: current frame 1000; skill count 2; skill identity 1/level 3; state count 185/aurastate 21; supplied duration result 50; no existing state list; caster owner 0,7; target monster with state storage and active/changed 21 false; no events; successful allocation except case (c); stat-slot and refresh operations log only; query operations supply values without adding effect-log entries; the duration evaluator records its invocations and context in a separate evaluation log; full attach returns success except case (i)
  - Seeds: `no draw`
  - Draws: none
  - Trace: reject selectors first; allocation failure stops before list setup; a missing caster uses sentinel owner, and a player success adds the final pet request.
  - Outputs: per independent case.
  - (a) Inputs: skill identity -1. Draws: none. Outputs: return 0; effect log empty; no list/event/membership changes.
  - (b) Inputs: aurastate 185. Draws: none. Outputs: return 0; effect log empty; no list/event/membership changes.
  - (c) Inputs: allocation returns null. Draws: none. Outputs: log[allocate(timed,1050,owner 0,7)] then return 0; no list setup/attachment/state/event/refresh requests.
  - (d) Inputs: caster absent; target player; allocation succeeds. Draws: none. Outputs: ordered[allocate(timed,1050,owner 6,-1),set list state 21,set generic callback,full attach,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh,pet update]; return 1; supplied bits active/changed 21 true. Actual attach/stat/refresh effects remain gated.
  - (e) Inputs: skill identity 2 (equal to the skill count). Draws: none. Outputs: return 0; effect log empty; no list/event/membership changes.
  - (f) Inputs: aurastate -1. Draws: none. Outputs: return 0; effect log empty; no list/event/membership changes.
  - (g) Inputs: caster absent; target monster; allocation succeeds. Draws: none. Outputs: ordered[allocate(timed,1050,owner 6,-1),set list state 21,set generic callback,full attach,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; no pet update; return 1.
  - (h) Inputs: caster owner 0,7 present; target player; allocation succeeds. Draws: none. Outputs: ordered[allocate(timed,1050,owner 0,7),set list state 21,set generic callback,full attach,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh,pet update]; return 1.
  - (i) Inputs: caster owner 0,7 present; target monster; allocation succeeds; full attach returns failure (0). Draws: none. Outputs: ordered[allocate(timed,1050,owner 0,7),set list state 21,set generic callback,full attach,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; no pet update; return 1.
  - Evaluation log: empty for (a), (b), (e), (f); exactly one duration evaluation for every other case, with context (caster 0,7 or null when absent, skill 1, level 3), so swapping skill and level is rejected.
  - Discriminates: both identity bounds are checked (a, e) and both aurastate bounds (b, f); no duration is evaluated before the checks pass; a failed attach does not stop the remaining requests or change the return (i); failure does not execute success-prefix writes; owner sentinel and final pet request are separate consequences of caster absence and target type (d, g, h).
  - runtime_observed: no

- **STL-V19** Nonplayer death requests (STL-R13a; boundary: supplied-input stage)
  - Inputs: monster death operation; unit with base and state storage; primary A,B,C, all owned by type 0, timed, hasStateStorage false; A state 11, deathRemovalExempt true; B state 21, deathRemovalExempt false; C state 3, deathRemovalExempt false; supplied monster mask contains only state 3, supplied player mask contains only state 21; secondary empty; supplied FreeList removes the requested list; auxiliary and partition operations log only; the partition log records its boss selector; MonStats `boss`, auxiliary presence and unit type below
  - Seeds: `no draw`
  - Draws: none
  - Trace: optional auxiliary removal precedes primary traversal; exemption keeps A; B is outside the monster mask and is freed, restarting at A; C is in the monster mask and retained; partition receives recordGate's 0 or 64.
  - Outputs: per independent case; no player stat-6 write and no type-8/type-9 event removal request in any case.
  - (a) Inputs: monster with loaded monster data, auxiliary effect present, MonStats `boss` false. Draws: none. Outputs: log[remove auxiliary effect,FreeList(B),partition(0)]; visits[A,B,A,C]; chain[A,C].
  - (b) Inputs: monster with loaded monster data, auxiliary effect absent, MonStats `boss` true. Draws: none. Outputs: log[FreeList(B),partition(64)]; visits[A,B,A,C]; chain[A,C].
  - (c) Inputs: object, auxiliary effect absent. Draws: none. Outputs: fault on the null monster-definition read; log empty; no FreeList or partition request; chain[A,B,C].
  - (d) Inputs: object, auxiliary effect present. Draws: none. Outputs: log[remove auxiliary effect], then fault; no FreeList or partition request; chain[A,B,C].
  - (e) Inputs: monster with loaded monster data, auxiliary effect absent, MonStats `boss` false; A instead owned by type 3 with deathRemovalExempt false (state 11, timed, hasStateStorage false). Draws: none. Outputs: log[FreeList(A),FreeList(B),partition(0)]; visits[A,B,C]; chain[C].
  - Discriminates: only owner type 4 is retained by ownership, so a missile-owned list is freed (e); death exemption preserves a non-item list without stay-death membership; monster death traverses and frees with the monster mask (a player-mask reading would free C and keep B); auxiliary removal precedes traversal and partition; the `boss` field's recordGate result 0/64, not auxiliary presence, reaches partition (a, b); a non-monster requests auxiliary removal and then faults before list removal or partition (c, d).
  - runtime_observed: no

### 8.2 Observation vectors

- **STL-O01** Expiry frame: measure list expiry, type 12 firing and removal; expect exact expiry, or next scheduling frame for a past/current request. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o01-expiry-frame).
- **STL-O02** Death: measure callback requests and active membership for supplied stay-death false/true, with partition integration separately gated. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o02-death).
- **STL-O03** Reapplication: measure forced changed state and selected/delivered message; state-with-stats requires exported stats and sync eligibility. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o03-reapplication).
- **STL-O04** Transition: record state/list across waypoint before expiry; no expected persistence default established. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o04-transition).

## 9. Acceptance and dependencies

Partial static acceptance is the verified assertions of STL-V01 through STL-V09 and STL-V11 through STL-V19 with expressly supplied operations. It covers schedule arithmetic, exact-frame routing, expiry/death traversal, creation/expiry fields, membership/forced marks, applier request order, callback gates and timer cancellation. It excludes every assertion attributed to a gated rule, actual free/attach/stat propagation, mask production, client delivery, refresh side effects and transition persistence. Full reproduction additionally needs STL-Q01/STL-Q02/STL-Q03/STL-Q07/STL-Q10 closed, executable discriminating vectors for resolved rules and matching observations of every verified behavioral rule (STL-Q08). Partial acceptance does not establish complete parity.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) frame units; [core-record-layouts-and-ids](core-record-layouts-and-ids.md) RLI-R03 state lookup/routing, RLI-R04b mode-bound lists (RLI-Q02/RLI-Q03 integration limits); [core-game-data-fields](core-game-data-fields.md) GDF-R02/GDF-R03 stat/state identities, GDF-R05a/GDF-R06a player/monster modes and the shared `aliveExclusionCondition` name; [core-game-data-table-records](core-game-data-table-records.md) GDA-R13 state-definition fields and GDA-R08a MonStats `boss`; [combat-status-application](combat-status-application.md) CSA-R04/CSA-R05/CSA-R08/CSA-R12 applier requests, CSA-R02 `recordGate` boss selector, CSA-R13 cold removal callback and CSA-R05 cold refresh; [skill-calc-expressions](skill-calc-expressions.md) SCX-R11 supplied expression results.

**Depended on by:** foundations state/mode-bound traversal, combat status application, [combat-leech-and-drain](combat-leech-and-drain.md) overlay lists, movement/freeze callbacks, [combat-open-wounds](combat-open-wounds.md) timed effects and damage-pipeline deferred expiry. Reverse consumers are not automatic prerequisites.

## 10. Related specs

- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): frames.
- [core-record-layouts-and-ids](core-record-layouts-and-ids.md): lookup, unit types and mode-bound lists.
- [combat-status-application](combat-status-application.md): status inputs and magnitudes.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): cold/freeze consumers.
- [skill-calc-expressions](skill-calc-expressions.md): duration/value expressions.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): damage hand-off to timed appliers.

## Appendix A. Provenance

Original Lord of Destruction Expansion 1.13c, Windows x86. D2Game/D2Common/D2Client identities are in the [source manifest](../research/source-manifest.md). Static records dated 2026-10-12 are reused within their actual boundaries; targeted readings dated 2026-10-14 and their limitations are in the [companion](../research/spec-evidence/state-timed-state-lifecycle.md). No gameplay observation or new archive-cell claim is recorded. The 2026-10-15 review correction to STL-R13a/STL-V19, the targeted original-code readings behind the review-pass vector additions, and the blind-reader passes are recorded in the companion; static integration and gameplay observations remain open.
