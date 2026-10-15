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

This contract covers timed-state list creation, attachment, state membership, expiry scheduling and traversal, removal callbacks, refresh policies by reference, death cleanup, and client state-callback dispatch. Effect magnitudes belong to [combat-status-application](combat-status-application.md) and [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md); expression evaluation belongs to [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md). Curse overrides, aura ticks, network message layouts and client visual callback bodies are separate behaviors. STL-R16 records the unresolved transition boundary. STL-R17 bounds a separate auxiliary timed effect without state identity. No local rule draws randomness.

## 2. Trigger and ownership

D2Game owns server applications, timer dispatch, removal callbacks and death cleanup. D2Common owns effect lists and state membership. D2Client owns set/remove callback dispatch. Server callbacks identify original source folders `UNIT\SUnitDmg`, `SKILLS\Skills`, `SKILLS\SkillBar`, `SKILLS\SkillSor` and `SKILLS\SkilPal`; shared list operations identify `Stats\StatsEx`.

A damage or skill application creates or refreshes a timed list and schedules state-expiry event type 12. Per-unit frame-event dispatch reaches the expiry walker for players, monsters and items. An expiry event concerns the whole unit: it examines every due timed primary list. Player and monster death cleanup also reaches list removal. Binary dispatch/output anchors are in the [companion](../research/spec-evidence/state-timed-state-lifecycle.md#dispatch-and-output-anchors).

Dispatch matches exact stored frames once per server frame (STL-R02). Cross-group/immediate/equal-slot ordering is STL-R02a. Changed-state membership feeds state-on, state-with-stats and state-off message selection (STL-R14). Synchronization eligibility and changed-state clearing are STL-Q01; packet-to-client-callback integration is STL-Q02.

## 3. Inputs and data

Semantic operation names are illustrative and impose no API or storage layout.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| Current/requested frame, list expiry | Game frame; effect expiry | i32 frames | Signed comparisons, except unsigned auxiliary expiry |
| Target and owner identity | Unit type/id | u32 type; i32 id | Player 0, monster 1, object 2, missile 3, item 4; absent owner 6/-1 |
| Unit mode | GDF-R14a/GDF-R15a | u32 | Player death/dead 0/17; monster death/dead 0/12 |
| Additional death condition | `unit.expiryDeathCondition` | boolean | Treated as dead by removal callbacks; wider meaning STL-Q06 |
| Base list and chains | NRF-R12a | Optional base, ordered primary/secondary lists | State storage eligibility and effect-list traversal |
| Timed list | Owner, state, expiry, stats, callback | i32 state/expiry/stat values; callable | Effect metadata independent of active membership |
| List conditions | `timed`, `hasStateStorage`, `secondary`, `modeBound`, `filteredApply`, `deathRemovalExempt` | booleans | Code-tested conditions; unidentified category meanings STL-Q06 |
| State membership | Active and changed state sets | Bounded state identities | State count determines valid identities |
| Frame event | Type, target, frame, override, two parameters | Signed type; i32 frame/parameters | Timed-state callers use type 12/no override/parameters 0,0 |
| State definitions | GDF-R12/GDF-R13 | Named fields | `nosend`, `disguise`, `notondead`, player/monster/boss stay-death flags |
| Stay-death masks | Supplied player/monster/boss state sets | Membership bits | Consumer STL-R12a; producer STL-R12b |
| Skill fields | `Skills.aurastate`, `auralencalc`, `aurastat1..6`, `aurastatcalc1..6` | i16 selectors; expression references | State, duration and six ordered stat slots |
| Application inputs | Skill count, skill identity/level, caster, target | i32 identity/level; optional caster | Expression inputs and effect owner |
| Client selectors | `States.setfunc`, `remfunc`, `cltevent`, `clteventfunc` | i16 | Select dispatch behavior, not callback bodies |
| Boss condition | GDF-R05a `MonStats.boss` | boolean | Selects boss death partition |
| Unit timer entries | Kind, key, busy, pending removal | i32 kind/key; booleans | Kind 1/key state cancellation |
| Auxiliary effect | Optional effect and expiry | u32 expiry | Independent type 6 lifetime |

| Mapping / constant | Origin | Binding |
|---|---|---|
| Event types 12/6, type bound 15, 64 frame slots | Engine code | Fixed by original code |
| Unit/dead-mode tests above | Engine code; names through GDF | Fixed by original code |
| State/stat identities in policy table and vectors | GDF-R11/GDF-R12 data identity | Appliers use fixed identities; skill selectors resolve from data |
| Definition fields and stay-death membership | Table schema, GDF-R13 | Resolved from data; execution meaning established separately |
| Six stat slots; attackrate 68 and paired stat 69 | Engine code | Fixed by original code; paired operation STL-Q07 |

Vectors supply masks and selectors explicitly; they do not claim archive values for unnamed rows. Actual named cells require their own table-data evidence.

## 4. Rules (execution order)

**STL-R01** Schedule [verified in 1.13c code (static)]. Signed type >=15 produces no event. Frame -1 uses the immediate path. Otherwise signed `requestedFrame <= now` stores `now+1`; a future frame is unchanged. Addition wraps at 32 bits. For nonnegative stored frames, slot=`frame mod 64`. Unit-type groups are player, monster, missile, object and item. Store target, type, frame, override and parameters and link the event to its target. Timed-state callers supply type 12, no override and parameters 0,0. Negative event types and wrapped-negative frame indexing remain STL-Q10.

**STL-R02** Exact-frame dispatch [verified in 1.13c code (static)]. In the current slot, fire only events whose stored frame equals current frame. Retain other events. Invoke a supplied override when present, otherwise the group's event-type handler; remove the fired event afterward. Missiles call their single handler regardless of event type.

**STL-R02a** Queue ordering [unresolved]. Candidate group order is missiles, players, monsters, objects, items, immediate queue before ordinary slot in each group, oldest-first within a slot. Checked decoding and insertion/dispatch mutation behavior are required for acceptance (STL-Q07).

**STL-R03** State-expiry routing [verified in 1.13c code (static)]. Player, monster and item type 12 handlers call `ExpireLists(unit,now)`. Monster dispatch bypasses its ordinary validity test for types 3,4,5,8,12. The event names no particular list/state. The object dispatch table has no type 12 handler; its failure behavior remains STL-Q10.

**STL-R04** Expiry traversal [verified in 1.13c code (static)]. Return for null unit, missing base or absent state storage. Walk primary lists only, saving each successor before testing. When now!=0, timed lists with signed expiry<=now are due. When now==0, decrement each timed expiry with 32-bit wrap and treat signed result<1 as due. For a due list without state storage request FreeList, then restart at the current primary head. A due list with state storage is retained but still restarts; this can loop indefinitely. Actual free effects are STL-R05; supplied fixtures specify chain mutation. Countdown producers are STL-Q09.

**STL-R05** Free effects [unresolved]. Candidate observable behavior removes the list's applied contributions before invoking its callback with associated unit, state and still-readable list, then releases it. Secondary lists do not invoke that callback. Complete propagation/indirect effects have not been read through all output writes (STL-Q07); a callback-order reading alone does not establish stored totals.

**STL-R06a** Creation/expiry primitives [verified in 1.13c code (static)]. Successful allocation initializes supplied owner type/id, list conditions and expiry, with other semantic fields zero/empty. State/callback setters ignore null lists. Setting expiry always stores it; a positive value sets timed, and a nonpositive value never clears timed. Null-list expiry read returns 0. Lookup consumes NRF-R12a: state 0 returns null; otherwise the first primary match precedes secondary matches, with no state-range test.

**STL-R06b** Attach [unresolved]. Candidate observable behavior attaches secondary lists without applying stats, and primary lists newest-first with full or filtered stat application. Cycle refusal, reattachment callbacks, extended-list processing and stored totals require complete verification (STL-Q07). Timed states request full application.

**STL-R06c** Replacing list-stat write [unresolved]. Candidate behavior replaces the stat/layer value, does nothing for equality, removes a zero entry, and propagates new-minus-old. Direct replacement control flow is read; entry management, propagation and special-list effects remain STL-Q07.

**STL-R07** State membership [verified in 1.13c code (static)]. Signed s<0 or s>=stateCount makes no request. Otherwise request the active-bit operation carrying on unchanged, then update marking. With state storage present, nonzero on sets active membership and zero clears it; only a membership change sets changed membership. No list is allocated or found. A state can be active without a list and a list can exist with its state off. Disguise is STL-R07a; update-queue integration is NRF-Q17.

**STL-R07a** Disguise [unresolved]. A changed disguise state is intended to update the unit's disguise condition; clearing must account for another active disguise state. Byte decoding and the full state-test helper remain STL-Q07; no guessed default is established.

**STL-R08** Forced changed membership [verified in 1.13c code (static)]. For non-null unit with state storage and valid signed state, set/clear changed membership according to nonzero/zero on, leaving active membership unchanged; request update even when the changed bit already matched. Invalid state or missing unit/list/storage makes no request. Update-queue effects remain NRF-Q17.

**STL-R09** Per-applier refresh [verified in 1.13c code (static)]. There is no shared policy. Consume the contracts below. A changed expiry requests another type 12 event; a queued event for the scheduled expiry can fire with no due list.

| Applier | State | Existing expiry | Values | Contract |
|---|---|---|---|---|
| Cold |11 cold|Extend only|Unchanged|CSA-R05, MMF-R13d|
| Freeze |1 freeze|Extend only|No magnitude|CSA-R08|
| Stun |21 stunned|Replace, can shorten|No magnitude|CSA-R04|
| Poison/burn |2 poison/115 burning|Replace when new magnitude>=old|Replace hpregen|CSA-R12|
| Timed skill |Supplied aurastate|Replace, can shorten|Stat-slot operation|STL-R10a/STL-R10b|

**STL-R10a** Timed skill-state request order [verified in 1.13c code (static)]. Invalid signed skill identity or aurastate outside 0..stateCount-1 returns 0. Expiry is SCX-R11's supplied duration result+now with 32-bit wrap. Look up the target state list. If absent, request allocation with timed/expiry and caster type/id, or 6/-1 without caster; allocation failure returns 0. For a new list, set state, install generic removal callback, request full attachment, request active state on. Both paths then request forced changed on, store expiry, schedule type 12, invoke stat-slot writing, request passive refresh, request pet update only for a player, and return 1. Actual attach/write/refresh effects remain STL-Q07; supplied request order is available independently.

**STL-R10b** Skill stat slots [unresolved]. Visit six slots in order. Invalid stat or zero expression result skips writing; zero leaves a stored value. Valid nonzero result requests a layer 0 list-stat write. Stat 68 also requests a separate operation for 69. The paired operation forwards to the same replacing list-stat writer after a null-list guard. Checked decoding of this caller and complete write/refresh effects remain STL-Q07; the shared writer is STL-R06c.

**STL-R11** Generic removal callback [verified in 1.13c code (static)]. Null unit is fatal. First request cancellation of unit timers kind 1/key state. Alive means expiryDeathCondition false and either player mode outside{0,17} or monster mode outside{0,12}; all other types count as dead. If dead and StayDeath nonzero, return with state unchanged. Otherwise request state off, animation-rate refresh, passive refresh, and pet update for a player. These are ordered requests; refresh effects remain STL-Q07. Cold instead requests state off, then shatter 107 off when alive, then animation-rate refresh (MMF-R13c). A retained state does not retain the already-removed list.

**STL-R11a** Unit-timer cancellation [verified in 1.13c code (static)]. Missing unit/timer chain does nothing. Save successor before each entry. Matching kind/key: busy entry becomes pending removal, otherwise detach/release it. Nonmatching entries remain. These timers are separate from frame-event type 12 queues.

**STL-R12a** StayDeath consumer [verified in 1.13c code (static)]. Invalid signed state returns 0. A non-null monster selects monster mask; every other unit including null selects player mask. Return the tested mask bit, not normalized 1. There is no boss branch. Supplied-mask acceptance does not establish STL-R12b's producer.

**STL-R12b** Stay-death mask producer [unresolved]. Candidate producer groups states by definition flags, including player/monster/boss stay-death fields. Loader bit-setting and byte decoding remain STL-Q07. The consumer can use supplied masks without asserting particular table cells.

**STL-R13a** Death requests/traversal [verified in 1.13c code (static)]. Walk primary chain. Retain item-owned lists, deathRemovalExempt lists and nonzero StayDeath lists. Otherwise request FreeList unless hasStateStorage, then restart at current head. Player death requests stat 6/value 0/layer 0, non-staying-list removal, state partition(false), then frame-event removals types 8/9. Other-unit death first removes an auxiliary effect when present, then non-staying-list removal, then partition with supplied monster boss condition (false for other types). Actual free/auxiliary effects and mask production remain STL-Q07; supplied-mask partition is STL-R13b.

**STL-R13b** Death partition [verified in 1.13c code (static)]. Return without requests for null unit, absent state storage or missing active/changed arrays. Select the supplied player mask for player/nonboss, monster mask for nonplayer/nonboss, and boss mask for nonzero boss selector. For each state word, first write `changed = changed OR (active AND NOT mask)`, then `active = active AND mask`. Request update marking afterward, including when state count is zero but the arrays exist. Actual mask production remains STL-R12b and update-queue integration NRF-Q17; supplied-mask state writes are independently verified.

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
| STL-R06a | Creation/expiry | verified in 1.13c code (static) | none at supplied-allocation boundary | STL-V05, STL-V17; NRF-V33 |
| STL-R06b | Attach | unresolved | STL-Q07: attach effects | none |
| STL-R06c | Replacing write | unresolved | STL-Q07: replacing effects | STL-V05 |
| STL-R07 | Membership | verified in 1.13c code (static) | STL-Q07: disguise only; NRF-Q17: update queue | STL-V06 |
| STL-R07a | Disguise | unresolved | STL-Q07: disguise condition | none |
| STL-R08 | Forced change | verified in 1.13c code (static) | NRF-Q17: update queue | STL-V12, STL-V17 |
| STL-R09 | Refresh policy | verified in 1.13c code (static) | Applier integration gates only | STL-V07 |
| STL-R10a | Skill requests | verified in 1.13c code (static) | STL-Q07: actual write/refresh | STL-V07, STL-V18 |
| STL-R10b | Skill slots | unresolved | STL-Q07: paired operation/refresh | none |
| STL-R11 | Removal callback | verified in 1.13c code (static) | STL-Q07: actual refresh | STL-V08 |
| STL-R11a | Timer cancellation | verified in 1.13c code (static) | none for supplied chain | STL-V13 |
| STL-R12a | StayDeath consumer | verified in 1.13c code (static) | STL-Q07: producer only | STL-V08, STL-V15 |
| STL-R12b | Mask producer | unresolved | STL-Q07: mask production | none |
| STL-R13a | Death traversal | verified in 1.13c code (static) | STL-Q07: free/partition/auxiliary effects | STL-V09, STL-V19 |
| STL-R13b | Death partition | verified in 1.13c code (static) | STL-Q07: data-derived mask producer only; NRF-Q17: update queue | STL-V14 |
| STL-R14 | Messages | unresolved | STL-Q01/STL-Q07: sender integration | none |
| STL-R15 | Client callbacks | unresolved | STL-Q02/STL-Q07: dispatch integration | none |
| STL-R16 | Transition | unresolved | STL-Q03: transition retention | none |
| STL-R17 | Auxiliary boundary | asm-only | STL-Q07: auxiliary effects | none |

## 5. Outputs and side effects

| Output | Rule |
|---|---|
| Scheduled event at clamped frame; fired event removed | STL-R01/STL-R02 |
| Empty list/owner/expiry fields; timed condition retained for nonpositive expiry | STL-R06a |
| Active/changed membership and update requests | STL-R07/STL-R08 |
| Due-list free requests and head reload | STL-R04 |
| Candidate contribution removal/callback/release, gated | STL-R05 |
| Candidate stored attach/replacing-write effects, gated | STL-R06b/STL-R06c |
| Ordered skill allocation/attach/stat/refresh requests; return 0/1 | STL-R10a |
| Timers detached/released or marked pending | STL-R11a |
| State-clear/animation/passive/pet requests or retained state | STL-R11 |
| Death cleanup requests and candidate partition | STL-R13a/STL-R13b |
| Candidate state-on/state-with-stats/state-off messages and callbacks | STL-R14/STL-R15 |
| Auxiliary removal/reschedule requests, separate lifetime | STL-R17 |

No local rule creates units or writes saves. Requests establish actual stored/delivered effects only under accepted dependency contracts.

## 6. Branches and edge cases

Zero/negative duration still schedules an ordinary expiry clamped to next frame. Nonpositive expiry writes never clear timed. Any type 12 event visits every due primary list; an extended list can survive an event at its scheduled expiry. Secondary lists are not traversed. An expiring stay-death state can retain membership after list contributions disappear; resurrection cleanup is STL-Q04. A retained due list with state storage can make head restart nonterminating (STL-Q05).

State lookup and membership use different guards (NRF-R12a). Missing caster uses owner 6/-1. Outside explicit unit/dead-mode/boss branches, no local rule branches on difficulty, PvP, melee/missile origin or classic/expansion; evidence remains expansion 1.13c. Skill allocation failure returns 0. Malformed event/frame behavior is STL-Q10.

## 7. Open questions

| Question | Kind / status | Evidence needed |
|---|---|---|
| **STL-Q01** | integration / unresolved | Real sync caller/predicate, exporter and changed-state clearing; actual messages only |
| **STL-Q02** | integration / unresolved | Packet handlers to set/remove/client-event dispatch; callback bodies remain outside |
| **STL-Q03** | integration / unresolved | Transition dispatch-to-list/state trace, retention/removal outcome |
| **STL-Q04** | outside-boundary / unresolved | Resurrection owner's retained-state cleanup; no local default |
| **STL-Q05** | outside-boundary / unresolved | Producer of primary hasStateStorage lists; supplied traversal remains available |
| **STL-Q06** | outside-boundary / unresolved | Wider meanings/producers of deathRemovalExempt and expiryDeathCondition |
| **STL-Q07** | behavior / unresolved | Complete callee/decoding/output-write paths for each gated rule; separate boundaries in gap file |
| **STL-Q08** | observation / pending | Matching original 1.13c captures; none recorded |
| **STL-Q09** | outside-boundary / unresolved | Caller supplying 0 to countdown expiry; ordinary handlers supply game frame |
| **STL-Q10** | behavior / unresolved | Malformed signed type/wrapped frame and object type 12 failure behavior |

Attempts and exact blocked assertions are in the [gap file](../plans/gaps/state-timed-state-lifecycle.md). Supplied masks/operation returns do not verify their producers. Pending observation limits confirmation, not independently verified supplied-input assertions.

## 8. Test vectors

### 8.1 Executable vectors

Each vector uses supplied dependencies that log calls in order. FreeList removes the requested list; bit operations persist active/changed membership; allocation yields the expressly listed empty fields. Refresh operations log only. A supplied stat-write/removal model supports only explicitly gated assertions of STL-R05/STL-R06b/STL-R06c, not verified integrated effects. Each case starts independently unless a continuation is stated.

- **STL-V01** Schedule (STL-R01; boundary: supplied-input stage)
  - Inputs: now 1000; monster; type 12; no override; parameters 0,0; empty queues; event-handler invocation log initially empty; requested frames per case
  - Seeds: `no draw`
  - Draws: none
  - Trace: Signed compare/clamp then remainder by 64.
  - Outputs: per independent case below.
  - (a) Inputs: requested frame 1000. Draws: none. Outputs: exactly one queued event, stored frame 1001, slot 41, monster group; immediately after scheduling at frame 1000 the handler invocation log is empty (count 0).
  - (b) Inputs: requested frame 999. Draws: none. Outputs: stored frame 1001, slot 41, monster group.
  - (c) Inputs: requested frame 1050. Draws: none. Outputs: stored frame 1050, slot 26, monster group.
  - (d) Inputs: requested frame -5. Draws: none. Outputs: stored frame 1001, slot 41, monster group.
  - (e) Inputs: type 15, requested frame 1050. Draws: none. Outputs: no record.
  - Discriminates: the frame-1000 scheduling call queues one future event and makes zero synchronous handler calls; storing 1001 while also invoking a handler at 1000 fails case (a).
  - runtime_observed: no

- **STL-V02** Dispatch (STL-R02/STL-R03; boundary: supplied-input stage)
  - Inputs: now 1050; P player with type 12 events 1050/1114 in slot 26; M monster with type 12 event 1050; no overrides; valid bases/empty chains
  - Seeds: `no draw`
  - Draws: none
  - Trace: Match equality, not slot identity, then call unit expiry handler.
  - Outputs: Calls ExpireLists(P,1050) once and ExpireLists(M,1050) once; remove their 1050 events; retain P 1114 until frame 1114. Player-before-monster is gated STL-R02a.
  - Discriminates: 1114 shares slot 26 but must not fire at 1050.
  - runtime_observed: no

- **STL-V03** Expiry restart (STL-R04/STL-R05; boundary: supplied-input stage)
  - Inputs: M monster; now 1050; primary A,B,C; A timed/state 11/expiry 1080/stats 67,68,69 each-50; B timed/state 2/expiry 1050/hpregen 74=-100; C untimed; no list has state storage; secondary empty
  - Seeds: `no draw`
  - Draws: none
  - Trace: Visit A,B; supplied FreeList removes B; reload head and visit A,C. Continue independently at 1080 with resulting A,C.
  - Outputs: At 1050 visit[A,B,A,C], free[B], chain[A,C]; at 1080 free[A], chain[C]. Gated STL-R05: B removal increases hpregen total 100 before callback(M,2,B); A removal increases totals 67/68/69 by 50 each.
  - Discriminates: Head restart revisits A and retains future expiry 1080.
  - runtime_observed: no

- **STL-V04** Zero duration (STL-R01/STL-R04/STL-R06a; boundary: supplied-input stage)
  - Inputs: now 1000; monster freeze list timed/expiry 1000 from supplied duration 0; requested event 1000; supplied FreeList removes it; separate timed L expiry 5
  - Seeds: `no draw`
  - Draws: none
  - Trace: Clamp event 1000 to 1001;1000<=1001 requests free. Set L expiry 0, retaining timed; positive next event sees 0<=now.
  - Outputs: Event 1001; free at 1001. L expiry 0/timed true; next positive-frame expiry event requests free.
  - Discriminates: Zero duration does not remove synchronously or clear timed.
  - runtime_observed: no

- **STL-V05** Creation/expiry/write (STL-R06a/STL-R06c; boundary: supplied-input stage)
  - Inputs: successful allocation timed/expiry 1050/owner 0,7; independent untimed L; independent empty stat 67/layer 0 and parent total 0 for gated write model
  - Seeds: `no draw`
  - Draws: none
  - Trace: Initialize; write expiry 0 to new list,5 to independent untimed L. Gated model writes-50 twice then 0. Lookup state 0.
  - Outputs: New list timed/1050/owner 0,7, other semantic fields 0/empty; expiry 0 retains timed; untimed L becomes timed/expiry 5. Gated STL-R06c: two writes leave one-50 entry/total-50, then 0 removes entry/total 0 (delta+50). FindStateList(u,0)=null.
  - Discriminates: Repeated replacement never yields-100 and nonpositive expiry does not clear timed.
  - runtime_observed: no

- **STL-V06** Membership separation (STL-R07; boundary: supplied-input stage)
  - Inputs: stateCount 185; valid base; non-disguise state 21 active=false/changed=false; supplied update log
  - Seeds: `no draw`
  - Draws: none
  - Trace: Set 21 on, clear changed as fixture setup, set 21 on again, attempt 185 on, set 21 off, attempt -1 on.
  - Outputs: First active 21=true/changed 21=true and update; second active 21=true/changed 21=false and update; 185 adds no changes/request; setting 21 off gives active 21=false/changed 21=true and the third update request; -1 adds no changes/request; no list allocated.
  - Discriminates: Equal membership does not force changed membership; wrapper still requests update.
  - runtime_observed: no

- **STL-V07** Skill replacement (STL-R09/STL-R10a; boundary: supplied-input stage)
  - Inputs: valid skill identity 1 and level 1 within a supplied skill count 2; aurastate 21/state count 185; current frame 1000; duration expression result 50; caster owner 0,7; monster target; supplied allocation succeeds; stat-slot/passive operations log only; no queued events
  - Seeds: `no draw`
  - Draws: none
  - Trace: 1000+50=1050 independently of stored expiry; execute the new-list prefix only when lookup returns no list.
  - Outputs: per independent case below; actual attachment/stat effects remain gated.
  - (a) Inputs: existing primary list for state 21, expiry 1200, active 21 true/changed 21 false. Draws: none. Outputs: no allocation/attachment; ordered[forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; active 21 true/changed 21 true; return 1; no pet update.
  - (b) Inputs: no list, active 21 false/changed 21 false. Draws: none. Outputs: ordered[allocate(timed,1050,owner 0,7),set list state 21,set generic callback,full attach request,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh]; supplied bit operations leave active/changed 21 true; return 1; no pet update.
  - (c) Inputs: comparison-only policy fixture, supplied existing expiry 1200 and candidate expiry 1050; consume the CSA-R05 strict-extension comparison. Boundary: isolated comparison, with no cold-applier invocation. Seeds: `no draw`. Draws: none. Trace: 1200 < 1050 is false. Outputs: selected expiry 1200. Full cold-application requests and target-seed draws belong to CSA-V08/CSA-R06; this case asserts only the supplied-expiry comparison result.
  - Discriminates: skill refresh shortens, cold does not; the existing-list skill path skips allocation and does not request active state on.
  - runtime_observed: no

- **STL-V08** Removal gate (STL-R11/STL-R12a; boundary: supplied-input stage)
  - Inputs: state 21/count 185/active true; expiryDeathCondition false; cancellation/refresh operations log; state writer persists; masks and unit type/mode supplied per independent case
  - Seeds: `no draw`
  - Draws: none
  - Trace: cancel(kind 1,key 21) first; choose mask and alive branch; return or clear/refresh.
  - Outputs: per case below; no list retention is asserted.
  - (a) Inputs: player mode 1, player mask bit true. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive,pet]; active false.
  - (b) Inputs: player mode 17, player mask bit true. Draws: none. Outputs: cancel(1,21) only; active true.
  - (c) Inputs: player mode 17, player mask bit false. Draws: none. Outputs: ordered[cancel(1,21),state off,animation,passive,pet]; active false.
  - (d) Inputs: monster mode 12, monster mask bit true/player bit false. Draws: none. Outputs: cancel(1,21) only; active true.
  - (e) Inputs: object mode 1, player mask bit true/monster bit false. Draws: none. Outputs: cancel(1,21) only; active true.
  - (f) Inputs: player mode 1, expiryDeathCondition true, player mask bit true. Draws: none. Outputs: cancel(1,21) only; active true.
  - (g) Inputs: null unit. Draws: none. Outputs: fatal assertion before any cancellation, state write or refresh; operation does not return normally.
  - Discriminates: monster selects monster mask; object never counts alive; the additional death condition changes a living-mode player to the retained-state branch; null asserts before callbacks.
  - runtime_observed: no

- **STL-V09** Death traversal (STL-R13a; boundary: supplied-input stage)
  - Inputs: player primary A,B,C; A owner 0/timed/state 11/player mask=false; B owner 4/timed; C owner 0/timed/state 21/player mask=true; all deathRemovalExempt=false/hasStateStorage=false; secondary empty; supplied FreeList removes A
  - Seeds: `no draw`
  - Draws: none
  - Trace: Request stat 6=0/layer 0, traverse A then restart B,C, then partition/event removals.
  - Outputs: Ordered[write(6,0,0),FreeList(A),partition(false),remove event 8,remove event 9]; chain[B,C]. Actual callback/contributions gated STL-R05; actual mask production gated STL-R12b; the supplied-mask partition is STL-R13b.
  - Discriminates: Item owner keeps B regardless of mask.
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
  - Discriminates: nonplayer uses monster mask unless boss selector is nonzero; existing changed bits survive; missing arrays skip update but a zero-count present array requests it; processing reaches word 5.
  - runtime_observed: no

- **STL-V15** StayDeath mask return (STL-R12a; boundary: supplied-input stage)
  - Inputs: state count 185; player mask contains only state 21, monster mask empty; independent unit/state cases below
  - Seeds: `no draw`
  - Draws: none
  - Trace: validate signed state, choose monster mask only for a non-null monster, test state bit without normalization.
  - Outputs: per case; no membership or list changes.
  - (a) Inputs: null unit, state 21. Draws: none. Outputs: 0x00200000.
  - (b) Inputs: player, state 21. Draws: none. Outputs: 0x00200000.
  - (c) Inputs: monster, state 21. Draws: none. Outputs: 0.
  - (d) Inputs: player, state 185. Draws: none. Outputs: 0.
  - (e) Inputs: player, state -1. Draws: none. Outputs: 0.
  - Discriminates: true membership returns the mask value rather than 1; a null unit still selects the player mask.
  - runtime_observed: no

- **STL-V16** Event handler routes (STL-R02/STL-R03; boundary: supplied-input stage)
  - Inputs: current frame 1050; independent case has one event at frame 1050; callback/type handlers log their semantic calls; ordinary handlers have valid targets and empty primary chains with state storage
  - Seeds: `no draw`
  - Draws: none
  - Trace: match the frame, choose override or group handler, remove the fired record.
  - Outputs: per independent case; event queue empty afterward.
  - (a) Inputs: player P/type 12 event with supplied override H. Draws: none. Outputs: log[H invoked once]; expiry type handler not called.
  - (b) Inputs: missile M/type 6 event with no override. Draws: none. Outputs: log[single missile handler invoked once]; no type-6 group-table handler call.
  - (c) Inputs: item I/type 12 event with no override. Draws: none. Outputs: log[ExpireLists(I,1050)].
  - Discriminates: override wins for player events; missile dispatch does not use its event-type selector; items share expiry routing.
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
  - (d) Inputs: null list; set state 21, set callback H, then read expiry. Draws: none. Outputs: no writes; expiry getter returns 0.
  - (e) Inputs: ForceChanged on null unit, state 21/on 1. Draws: none. Outputs: no writes or update request.
  - (f) Inputs: ForceChanged on unit without base, state 21/on 1. Draws: none. Outputs: no writes or update request.
  - (g) Inputs: ForceChanged on base without state storage, state 21/on 1. Draws: none. Outputs: no writes or update request.
  - (h) Inputs: ForceChanged on valid state-storage base, state 185/on 1. Draws: none. Outputs: active/changed membership remains false; no update request.
  - (i) Inputs: ForceChanged on valid state-storage base, state -1/on 1. Draws: none. Outputs: active/changed membership remains false; no update request.
  - Discriminates: missing storage stops traversal even with a due primary list, and invalid forced marks do not reach update marking.
  - runtime_observed: no

- **STL-V18** Skill refusal and owner/pet routes (STL-R10a; boundary: supplied-input stage)
  - Inputs: current frame 1000; skill count 2; skill identity 1/level 1; state count 185/aurastate 21; supplied duration result 50; no existing state list; caster owner 0,7; target monster with state storage and active/changed 21 false; no events; successful allocation except case (c); stat-slot and refresh operations log only; query operations supply values without adding effect-log entries
  - Seeds: `no draw`
  - Draws: none
  - Trace: reject selectors first; allocation failure stops before list setup; a missing caster uses sentinel owner, and a player success adds the final pet request.
  - Outputs: per independent case.
  - (a) Inputs: skill identity -1. Draws: none. Outputs: return 0; effect log empty; no list/event/membership changes.
  - (b) Inputs: aurastate 185. Draws: none. Outputs: return 0; effect log empty; no list/event/membership changes.
  - (c) Inputs: allocation returns null. Draws: none. Outputs: log[allocate(timed,1050,owner 0,7)] then return 0; no list setup/attachment/state/event/refresh requests.
  - (d) Inputs: caster absent; target player; allocation succeeds. Draws: none. Outputs: ordered[allocate(timed,1050,owner 6,-1),set list state 21,set generic callback,full attach,state on,forced changed on,expiry 1050,schedule(type 12,1050),stat slots,passive refresh,pet update]; return 1; supplied bits active/changed 21 true. Actual attach/stat/refresh effects remain gated.
  - Discriminates: failure does not execute success-prefix writes; owner sentinel and final pet request are separate consequences of caster absence and target type.
  - runtime_observed: no

- **STL-V19** Nonplayer death requests (STL-R13a; boundary: supplied-input stage)
  - Inputs: primary A owned by type 0, timed, state 11 with supplied monster/player stay-death false, deathRemovalExempt true, hasStateStorage false; secondary empty; auxiliary and partition operations log only; supplied boss condition and unit type below
  - Seeds: `no draw`
  - Draws: none
  - Trace: optional auxiliary removal precedes primary traversal; exemption keeps A; partition receives the supplied boss selection.
  - Outputs: per independent case; A retained; no FreeList(A), player stat-6 write, or type-8/type-9 event removal request.
  - (a) Inputs: monster, boss condition true, auxiliary effect present. Draws: none. Outputs: log[remove auxiliary effect,partition(true)].
  - (b) Inputs: object, auxiliary effect absent; boss condition false. Draws: none. Outputs: log[partition(false)].
  - Discriminates: death exemption preserves a non-item list without stay-death membership; the nonplayer branch requests auxiliary removal before partition, and boss selection reaches partition.
  - runtime_observed: no

### 8.2 Observation vectors

- **STL-O01** Expiry frame: measure list expiry, type 12 firing and removal; expect exact expiry, or next scheduling frame for a past/current request. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o01-expiry-frame).
- **STL-O02** Death: measure callback requests and active membership for supplied stay-death false/true, with partition integration separately gated. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o02-death).
- **STL-O03** Reapplication: measure forced changed state and selected/delivered message; state-with-stats requires exported stats and sync eligibility. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o03-reapplication).
- **STL-O04** Transition: record state/list across waypoint before expiry; no expected persistence default established. [Procedure](../research/spec-evidence/state-timed-state-lifecycle.md#stl-o04-transition).

## 9. Acceptance and dependencies

Partial static acceptance is the verified assertions of STL-V01 through STL-V09 and STL-V11 through STL-V19 with expressly supplied operations. It covers schedule arithmetic, exact-frame routing, expiry/death traversal, creation/expiry fields, membership/forced marks, applier request order, callback gates and timer cancellation. It excludes every assertion attributed to a gated rule, actual free/attach/stat propagation, mask production, client delivery, refresh side effects and transition persistence. Full reproduction additionally needs STL-Q01/STL-Q02/STL-Q03/STL-Q07/STL-Q10 closed, executable discriminating vectors for resolved rules and matching observations of every verified behavioral rule (STL-Q08). Partial acceptance does not establish complete parity.

**Depends on:** [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md) NRF-R09a frames, NRF-R12a state lookup/routing, NRF-R12b stat reads, NRF-R14b mode-bound lists (NRF-Q17/NRF-Q20 integration limits); [core-game-data-fields](core-game-data-fields.md) GDF-R05a/GDF-R11/GDF-R12/GDF-R13/GDF-R14a/GDF-R15a names/schema/modes/boss; [combat-status-application](combat-status-application.md) CSA-R04/CSA-R05/CSA-R08/CSA-R12 applier requests; [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md) SCX-R11 supplied expression results; [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md) MMF-R13c/MMF-R13d cold removal/refresh.

**Depended on by:** foundations state/mode-bound traversal, combat status application, [combat-leech-and-status-effects](combat-leech-and-status-effects.md) overlay lists, movement/freeze callbacks, [combat-open-wounds](combat-open-wounds.md) timed effects and damage-pipeline deferred expiry. Reverse consumers are not automatic prerequisites.

## 10. Related specs

- [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md): lookup, stat reads and mode-bound lists.
- [combat-status-application](combat-status-application.md): status inputs and magnitudes.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): cold/freeze consumers.
- [skill-calc-expressions-and-synergies](skill-calc-expressions-and-synergies.md): duration/value expressions.
- [combat-damage-pipeline-order](combat-damage-pipeline-order.md): damage hand-off to timed appliers.

## Appendix A. Provenance

Original Lord of Destruction Expansion 1.13c, Windows x86. D2Game/D2Common/D2Client identities are in the [source manifest](../research/source-manifest.md). Static records dated 2026-10-12 are reused within their actual boundaries; targeted readings dated 2026-10-14 and their limitations are in the [companion](../research/spec-evidence/state-timed-state-lifecycle.md). No gameplay observation or new archive-cell claim is recorded. The completed reader review applies to the supplied-input contract; static integration and gameplay observations remain open.
