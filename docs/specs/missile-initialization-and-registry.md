# Missile Initialization and Registry

**Description**

Right after a projectile is created and numbered, the server finishes preparing it within the same request. It registers the projectile so the game can find it by number, places it on the map, sets how long it lives and when it becomes active, and sets its speed, target and what it collides with. It then records the skill and level, prepares the damage, records who fired it, rolls how many targets it can pierce and returns it. If the projectile cannot be given a path, setup stops and the request fails, though the projectile has already been numbered and registered.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MIR-R, MIR-Q, MIR-V, MIR-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/missile-initialization-and-registry.md)
Open evidence gaps: [gap file](../plans/gaps/missile-initialization-and-registry.md)

## 1. Scope

This contract owns the second part of one server missile-creation request, from the missile's metadata through its final return: metadata and the fresh creation event, placement and missile-registry insertion, range/lifetime/activation, the missing-path return, path setup, destroy condition, moving-speed and planner requests, last-collision source, acceleration/maximum speed, point-distance lifetime, base stat-list setup, callback/skill/level, damage-setup requests, owner attribution, pierce-count calculation, stat-write requests and the monster-mod event request. Request validation, start/aim/speed computation, allocation, seeding, identity, the client-notification boundary and the request's draw accounting belong to [missile-creation-and-seeding](missile-creation-and-seeding.md). Per-frame movement, collision, countdown and explosions belong to [missile-movement-and-collision](missile-movement-and-collision.md); computed damage belongs to [missile-damage-parameters](missile-damage-parameters.md); hit resolution belongs to [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md). Supplied dependency results permit the verified local stages below; they do not establish complete path, stat-lifecycle or monster-mod integration.

## 2. Trigger and ownership

D2Game's server `MISSILES` subsystem runs these stages synchronously inside the same creator invocation, immediately after [MCI-R11](missile-creation-and-seeding.md) has numbered the missile. They run once per successful creation and have no independent trigger. The invocation returns absent at MIR-R04 or the missile at MIR-R16. Later missile updates are owned by the movement contract.

D2Common supplies missile metadata, room/path operations, state/stat primitives and owner attribution. The stages request update marking during placement and attribution; they send no client message directly (the later notification boundary is MCI-R12). Binary entry points, call chains and original request/layout mappings are in the companion.

## 3. Inputs and data

Runtime integers are i32 unless specified. A bit test uses the u32 request flags. All arithmetic wraps to 32 bits before the next operation unless a rule explicitly widens it. Signed division truncates toward zero. `low16`/`low32` retain the indicated low bits. `sat16(v)` interprets v as i32 and clamps to -32768..32767. Positions are path units; speed is signed 8.8; lifetime/activation are frame counters.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| created missile | MCI-R09..R11 | unit with type, class, id, act/environment, mode and seed | The allocated, seeded and numbered missile |
| owner, source | request unit references accepted by MCI-R01/R02 | unit; type/id i32 | Owner for attribution, damage and pierce; source as damage parent |
| start room/point, retained point, surviving targetUnit, final speed | MCI-R02..R08 | room; i32 path units; unit or absent; signed 8.8 | Placement, target storage, path speed and point-distance lifetime inputs |
| flags | request conditions below | u32 | Select range, elapsed, activation, duration and stat operations |
| skill, level, loops | request | i32 | Skill metadata, range scaling and loop count |
| elapsed, activateOverride, toHit, rangeOverride | request | i32 | Already elapsed frames, activation subtraction, attack-rating request and total range override |
| callback, callbackArg | request | optional operation; opaque supplied argument | Called with the new missile and argument; callback effects/draws are supplied by its owner |
| allocation and placement results | supplied event, missile-data, path, base-list and source-list operations | success, absent or fatal outcome as specified by their owner | A missing path result is distinguishable from a dependency fault |
| existing missile metadata/path/base list | outputs of earlier stages or supplied stage fixtures | logical fields with widths in rules | Subsequent initialization consumes their current values |
| current path position | path state at MIR-R10 | i32 integer path components | Point-distance lifetime input |
| owner stat data | ordered existing stat arrays and definitions | RLI-R02a/RLI-R02b | Pierce aggregate and own-base values |
| owner source-list data | state count, extended-base condition, existing state-98 list; allocation result | i32 count; booleans; list or absent | Attribution requests and its early allocation-failure branch |
| registry bucket | current ordered chain for missile identity | u32 identities, unique within type | Insertion fixture; bucket selector is id modulo 128 |
| monster-mod context | owner monster modifiers | supplied operations/results | Their complete production/execution remains MIR-Q08 |

| Request flag | Established local behavior |
|---|---|
| 0x8 | Add animation sub-loop duration to range |
| 0x200 / 0x400 | Subtract elapsed frames / replace lifetime with point-distance duration |
| 0x800 / 0x1000 | Activation override / request attack-rating write |
| 0x8000 / 0x10000 | Range override / set missile metadata condition 2 |

Start, aim and speed flags are consumed by [missile-creation-and-seeding](missile-creation-and-seeding.md) section 3.

| Data field | Source / logical width | Consumer |
|---|---|---|
| `Range`, `LevRange` | `data/global/excel/missiles.txt`, i16 | MIR-R03 |
| `MaxVel`, `Activate` | Same file, u8 | MIR-R09/R03 |
| `Accel` | Same file, i16 | MIR-R09 |
| `SubLoop`, `SubStart`, `SubStop` | Same file, u8 | MIR-R03 |
| `CollideType`, `Collision` | Same file, u8 | MIR-R05; accepted mask domain 0..8 |
| `DamageRate` | Same file, i32 | MIR-R16 |
| `LastCollide`, `Pierce`, `CanDestroy` | Same file, booleans; GDA-R10 | MIR-R08/R15/R06 |

| Named identity | Read / request | Origin | Binding |
|---|---|---|---|
| stats 156 `item_pierce`, 166 `skill_pierce`, 328 `pierce_idx` | Aggregate chance components; own-base seed input; count write request | Named ItemStatCost cells in [table record](../research/spec-evidence/missile-initialization-and-registry.md#named-cells-2026-10-15) | Mapping from data; ids fixed by code |
| stats 19 `tohit`, 327 `damage_framerate` | Replacing layer-0 unit-stat requests | Same named-cell record | Mapping from data; ids fixed by code |
| state 98 `sourceunit`; stats 353 `source_unit_type`, 354 `source_unit_id` | Source membership and list-stat requests | Same named-cell record | Mapping from data; ids fixed by code |
| pierce cap 4, high seed word 666, collision masks, clamp endpoints | Rules below; NRF-R03a for seed initializer | Engine code/constants | Fixed by original code |
| missile fields and flags | Missiles schema identified by original loader; GDA-R03/GDA-R10 | Table data; flag consumers in code | Data supplies values; code fixes their interpretation |

Named table records establish text-cell identities only within their recorded profile. Synthetic row values below do not assert a shipped missile row or complete text/compiled correspondence. Actual stat mutation/propagation is separate from a logged write request (RLI-Q03, STL-R06b/c).

## 4. Rules (execution order)

**MIR-R01** Missile metadata/registration.
- **MIR-R01a** [verified in 1.13c code (static)]. If missile metadata is absent request zeroed metadata allocation and initialize last-collision source id to -1. Existing metadata is retained. Store the unit id and clear creation conditions C and D. Metadata-allocation failure has no checked absent-result return in this wrapper; dependency faults are not converted to null creation.
- **MIR-R01b** [verified in 1.13c code (static)]. For a fresh unit with empty frame-event chain and a supplied successful event allocation, associate a creation event with this missile, type 0, frame -1, persistent-creation condition true, no override, parameters 0,0 and auxiliary value 0. Insert it first in the missile creation-event group and in this unit's event chain. Subsequent execution/removal and replacement of a nonempty old chain remain MIR-Q01. These are frame-event entries, distinct from STL-R11a unit timers.

**MIR-R02** Placement.
- **MIR-R02a** [verified in 1.13c code (static)]. Reassociate act/environment as MCI-R09. Request path creation with game pool, start room, start coordinates, missile and extra condition 0; then request registry insertion. If the resulting unit has a room, request update marking. The supplied path result and room test determine this local request order. Complete path allocation/room-collision publication and update-queue effects remain MIR-Q02/RLI-Q02; faults are not specified as a successful absent-path result.
- **MIR-R02b** [verified in 1.13c code (static)]. In the supplied missile registry bucket selected by id modulo 128, compare ids unsigned. Insert before the first id smaller than the new id, yielding descending order; append if none. Equal id takes a fatal assertion path, not a null return. No draw occurs. Wider registry lifecycle/lookup policy is outside this insertion boundary.

**MIR-R03** Range/lifetime/activation [verified in 1.13c code (static)]. If 0x8000 use rangeOverride; otherwise compute `r=i32(i32(LevRange*level)+Range)` from sign-extended i16 fields. Only on the nonoverride branch, if 0x8 and SubLoop != 0, add `i32((SubStop-SubStart)*loops)` with wrap. Bytes are unsigned before subtraction. If 0x200, subtract elapsed with wrap and store animation accumulator `low32(elapsed<<8)`. Store totalRange=`sat16`(r), lifetime=`sat16`(r). Choose a=activateOverride if 0x800, else unsigned Activate; store activationCounter=`sat16`(i32(r-a)), using unclamped r. Later countdown/activation is movement-owned.

**MIR-R04** Missing path [verified in 1.13c code (static)]. After R03, absent path returns absent without freeing the allocated unit or undoing prior calls/writes. Its seed step and id (MCI-R10/R11) remain consumed. Registration/placement stored effects are only those established or supplied in R01/R02.

**MIR-R05** Path setup [verified in 1.13c code (static)]. With a present path:
1. Set speed and maximum speed to 0; set path reset selector 15 only if the entering speed was nonzero (NRF-R08b).
2. With surviving targetUnit, store tracking unit/type/id. Otherwise store low16 of each retained target coordinate and clear tracking unit. Speed 0 still preserves the request's targetUnit because MCI-R08 was skipped.
3. Request collision-footprint value 64 if Collision != 0, otherwise 0; room collision-map effects are MIR-Q02. Store the collision mask selected by CollideType: 0→0, 1→0x84, 2→0x104, 3→0x184, 4→0, 5→0x104, 6→4, 7→0x40, 8→0x185. This code table is not a data-remappable policy; selectors above 8 remain MIR-Q03.

**MIR-R06** Destroy condition [verified in 1.13c code (static)]. When CanDestroy is set, set the missile's destroy condition true. No clear is requested when it is false. Its wider reader behavior is MCI-Q02.

**MIR-R07** Moving missile initialization.
- **MIR-R07a** [verified in 1.13c code (static)]. Only when final speed != 0, set path speed and maximum speed to that speed, including the speed setter's conditional reset-selector behavior (NRF-R08b).
- **MIR-R07b** [verified in 1.13c code (static)]. Immediately after that write request path building with this path, missile and condition 0; ignore its numeric return. Speed 0 skips both stages. The planner's path/direction/room mutations and draw count remain MIR-Q04; this local request is available with a supplied planner operation.

**MIR-R08** Last-collision source [verified in 1.13c code (static)]. With LastCollide set, store owner type/id in missile metadata. Otherwise retain existing values, including -1 source id after fresh allocation.

**MIR-R09** Acceleration/maximum speed [verified in 1.13c code (static)]. Store sign-extended Accel directly, without a fixed-point shift; then overwrite path maximum speed with `MaxVel<<8`. This runs even for speed 0 and does not receive MCI-R07's 75% scale.

**MIR-R10** Point-distance lifetime [verified in 1.13c code (static)]. Only with 0x400, read the missile's current integer path position at this stage and the retained MCI-R04/MCI-R08 point, even when tracking a targetUnit. Per axis compute wrapped delta point-position, negate if negative, then replace a still-negative result with 0. Thus an INT_MIN delta contributes 0. Let `s=i32(2*max(dx,dy)+min(dx,dy))`, wrapping before signed truncating division: `d=trunc(s/2)`. Replace exactly d=0 with 1. Set denominator=`low32(speed<<4)`. If denominator=0 set q=0; otherwise divide unsigned `low32(d<<16)` by unsigned denominator, floor, giving u32 q. Store lifetime=`sat16`(i32(q)). TotalRange/activationCounter are unchanged. A nonzero denominator is at least 16, so q<=268431360 and this caller cannot produce a negative signed q. Full actual position preparation remains MIR-Q04; supplied-current-position arithmetic is independently available.

**MIR-R11** Base stat-list setup.
- **MIR-R11a** [verified in 1.13c code (static)]. With no existing base list and successful allocations, create an empty extended base list associated with missile type/id and pool, empty own/resolved entries, empty attached lists, zero active/changed state membership, zero callback/context and base condition 0. State storage has enough words for the supplied state count. Allocation failure is unchecked here and has no defined null fallback.
- **MIR-R11b** [unresolved]. With an existing base list, setup can request cleanup before replacing it. Complete cleanup/propagation/callback effects are not established (MIR-Q05). This does not block R11a's fresh-empty branch.

**MIR-R12** Callback/skill/level [verified in 1.13c code (static)]. After base-list setup, invoke a present callback once with missile and callbackArg; ignore its return. Then store skill clamped to 0..32767, and store level's low16 unchanged. Reading that level sign-extends i16, so 40000 reads -25536 and 70000 reads 4464. Callback-owned effects/draws remain with its caller; later writes can overwrite callback changes to these fields.

**MIR-R13** Damage-setup requests [verified in 1.13c code (static)]. Request owner weapon-stat activation with selection condition 0, then the MDP damage builder with MDP source=owner, MDP parent=request source, new missile and original request level; then the MDP store with owner, missile, its result and the original request level. Consume [missile-damage-parameters](missile-damage-parameters.md) at its supplied-input/request boundaries. Complete equipment preparation, builder/storer integration and stored stat effects retain MIR-Q06 and the exact MDP gates; this rule guarantees caller arguments/order, not a complete damage or draw result.

**MIR-R14** Owner attribution.
- **MIR-R14a** [verified in 1.13c code (static)]. Store direct owner type/id. Without a base list, set owner-link-valid and finish attribution. With a base list and stateCount > 98, request state-98 active/changed bookkeeping and update marking; with extended state storage the membership bit is set and changed membership set only if newly active. Look up state-98 list (primary before secondary); if absent request allocation associated with missile type/id, state 98 and attachment with full application. Allocation failure returns from attribution before owner-link-valid is set. A present/successfully allocated list receives ordered layer-0 list-stat requests 353=owner.type, 354=owner.id, then owner-link-valid is set. Even when stateCount <= 98, list lookup/allocation/write requests still run when base list exists. The creator then clears creation condition D regardless of that early attribution return. Disguise and update bookkeeping consume STL-R07a/RLI-Q02.
- **MIR-R14b** [unresolved]. Complete list attachment and resulting stored source stats are STL-R06b/c; source disguise/queue integration also remains gated. Direct link fields and local request order are separately available under R14a (MIR-Q05).

**MIR-R15** Pierce count.
- **MIR-R15a** [verified in 1.13c code (static)]. Require player/monster owner, missile recipient and Pierce true. Set c to wrapped sum of owner aggregate stats 156 and 166. If c=0, draw nothing and make no count-write request. Otherwise initialize a local seed from owner's own-base stat 328 with high word 666. Start n=0; while n<4, step the local seed (NRF-R01), take unsigned lo modulo 100, and stop if the signed remainder >= c; otherwise increment n. Request missile stat 328=n, layer 0. There are min(n+1,4) draws: negative c consumes one failed draw; c>=100 consumes four passing draws. The owner's stat/seed, game seed and missile seed are unchanged by this helper. Equal supplied seed value/chance repeat the same count (MIR-Q07 does not gate this consumer).
- **MIR-R15b** [verified in 1.13c code (static)]. The chance components use aggregate/current unit reads and seed input uses own-base read, via RLI-R02a/RLI-R02b. Only a replacing unit-stat request is established for the resulting count; actual stored effects remain RLI-Q03/STL-R06c (MIR-Q05).

**MIR-R16** Request stats/condition/return [verified in 1.13c code (static)]. After pierce, flag 0x1000 requests stat 19=toHit, layer 0. Flag 0x10000 sets missile metadata condition 2 while retaining other metadata conditions. Then invoke the monster-owner wrapper (R17). Finally request stat 327=DamageRate, layer 0, and return the missile. Complete stat-write propagation remains RLI-Q03; logged requests and the direct metadata write are independent.

**MIR-R17** Monster-mod event.
- **MIR-R17a** [verified in 1.13c code (static)]. Only a monster owner and present missile request the monster-mod operation with game, owner, missile and event 5. Players skip it. It runs between the optional metadata condition write and the final damage-rate request.
- **MIR-R17b** [unresolved]. The event-5 dispatcher selects callbacks from up to nine owner mod bytes. Which selected callbacks act, their effects and seed owners/counts require complete handler verification (MIR-Q08). Dispatcher byte decoding alone does not establish those effects.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MIR-R01a | Metadata/id | verified in 1.13c code (static) | none | MIR-V01, MCI-V07 |
| MIR-R01b | Fresh creation event | verified in 1.13c code (static) | Q01: subsequent lifecycle/nonfresh chain only | MIR-V02 |
| MIR-R02a | Placement requests | verified in 1.13c code (static) | Q02/RLI-Q02: actual path/room/update effects | MIR-V02 |
| MIR-R02b | Descending registry insert | verified in 1.13c code (static) | none | MIR-V03 |
| MIR-R03 | Range/lifetime/activation | verified in 1.13c code (static) | none | MIR-V04, MIR-V05 |
| MIR-R04 | Missing path | verified in 1.13c code (static) | none | MIR-V06 |
| MIR-R05 | Path setup | verified in 1.13c code (static) | Q02: collision-map integration; Q03: selectors >8 | MIR-V07, MCI-V06 |
| MIR-R06 | Destroy condition | verified in 1.13c code (static) | none | MIR-V07, MIR-V01 |
| MIR-R07a | Moving speed | verified in 1.13c code (static) | none | MIR-V08 |
| MIR-R07b | Planner request | verified in 1.13c code (static) | Q04: planner-produced path/draws | MIR-V02 |
| MIR-R08 | Source metadata | verified in 1.13c code (static) | none | MIR-V07 |
| MIR-R09 | Acceleration/max speed | verified in 1.13c code (static) | none | MIR-V08 |
| MIR-R10 | Point-distance lifetime | verified in 1.13c code (static) | Q04: actual current-position production only | MIR-V09 |
| MIR-R11a | Fresh base list | verified in 1.13c code (static) | none | MIR-V10 |
| MIR-R11b | Existing-list cleanup | unresolved | Q05: cleanup effects | none |
| MIR-R12 | Callback/metadata | verified in 1.13c code (static) | none | MIR-V11 |
| MIR-R13 | Damage requests | verified in 1.13c code (static) | Q06: damage/equipment integration and full draws | MIR-V12 |
| MIR-R14a | Attribution fields/requests | verified in 1.13c code (static) | Q05/STL-R07a/RLI-Q02: attachment/write/disguise/queue effects only | MIR-V12 |
| MIR-R14b | Attribution stored effects | unresolved | Q05: stored source-list effects | none |
| MIR-R15a | Local pierce draws | verified in 1.13c code (static) | none | MIR-V13 |
| MIR-R15b | Pierce reads/write request | verified in 1.13c code (static) | Q05/RLI-Q03: actual stored count | MIR-V13 |
| MIR-R16 | Stat requests/condition/return | verified in 1.13c code (static) | Q05/RLI-Q03: actual stored stats | MIR-V12 |
| MIR-R17a | Monster-event request | verified in 1.13c code (static) | Q08: handler execution only | MIR-V12 |
| MIR-R17b | Mod handler effects | unresolved | Q08: complete handlers | none |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| Return absent for a later missing path; return missile after final stat request | MIR-R04/R16 |
| Fatal duplicate-identity assertion; unchecked dependency allocation/fault outcomes have no null fallback | MIR-R01a/R02b/R11a |
| Creation conditions C/D, destroy condition, direct owner type/id and owner-valid condition | MIR-R01a/R06/R14a |
| Metadata allocation; source id sentinel; range/lifetime/activation, skill/level, source type/id and condition 2 | MIR-R01a/R03/R08/R10/R12/R16 |
| Animation accumulator when elapsed flag set | MIR-R03 |
| Fresh creation-event association/group insertion; registry insertion | MIR-R01b/R02b |
| Path allocation/planner/update/collision-footprint requests; direct speed/reset/target/mask/acceleration/max-speed writes | MIR-R02a/R05/R07/R09 |
| Fresh base-list state and storage; existing-base cleanup request | MIR-R11 |
| Callback invocation, original request level to damage setup, owner weapon activation and MDP build/store requests | MIR-R12/R13 |
| State membership/update/list allocation/attach/write requests | MIR-R14 |
| Local pierce seed/count; ordered stat 328/19/327 requests; metadata condition write | MIR-R15/R16 |
| Monster event-5 request and possible handler effects | MIR-R17 |

No save/persistence operation is directly requested. Actual stat/attachment, path/room/update effects retain the gates named above.

## 6. Branches and edge cases

Zero final speed skips the moving-speed write and planner call, while targetUnit storage, acceleration/max-speed and optional point-distance lifetime still run. Range override bypasses loop addition; elapsed subtraction still applies. Negative arithmetic wraps and truncates before clamps. Skill is clamped, level is narrowed without clamp. Players skip the monster-mod request.

No path after placement returns absent after seed/id/range effects. A dependency assertion or failed unchecked allocation is not that branch. Count and stat-write requests do not guarantee persisted values. Larger collision selectors and nonfresh event/base cleanup remain gated. Tracking a unit does not redirect the retained point used for optional lifetime calculation. A failed source-list allocation ends attribution before owner-link-valid is set, but the request still continues to pierce, stats and return.

## 7. Open questions

The [gap file](../plans/gaps/missile-initialization-and-registry.md) records attempted evidence and precise blocked assertions.

- **MIR-Q01** | unresolved; integration | Complete persistent creation-event consumption/removal and replacement of a nonempty old chain. Fresh supplied-event association/insertion is available. Close by reading the applicable frame-event consumer/removal path through its effects.
- **MIR-Q02** | unresolved; integration | Complete path allocation, collision-room publication and update-queue effects. Supplied path/room outcomes support the local consumer requests. Close with these owner contracts; successful supplied absence must not stand in for an unchecked allocation fault.
- **MIR-Q03** | unresolved; behavior | Collision mask selection for CollideType >8 and reachable loader/input domain. Values 0..8 have an exact contract; no default is supplied above it. Close with producer bounds or checked original-image behavior for the wider domain.
- **MIR-Q04** | unresolved; integration | Actual planner target/path/position mutations and all planner draw sites. Local call order and supplied-current-position lifetime arithmetic remain available. Close with a path/movement owner contract covering the reached missile route.
- **MIR-Q05** | unresolved; integration | Existing-base cleanup, source-list attach/write propagation and actual replacing unit-stat effects (RLI-Q03/STL-R06b/c), including indirect notifications. Fresh empty-base setup and caller requests remain available. Close through the lifecycle owners.
- **MIR-Q06** | unresolved; integration | Complete owner equipment preparation, MDP builder/store execution and total creation damage draws. Consume MDP's verified arithmetic/request subsets; its producer/stat-lifecycle gates still apply. Close through those owners, without inferring effects from caller arguments.
- **MIR-Q07** | unresolved; outside-boundary | Upstream production/change of the owner's own-base stat 328 seed value. It does not block deterministic supplied-seed pierce calculation. Close with an identified owner-writer contract; a missile-count decrement is a different write.
- **MIR-Q08** | unresolved; integration | Selected event-5 monster-mod callbacks, their writes and seed owners/draw counts. Wrapper guard/order is available. Close by reading selected handlers with their inputs and table selection.
- **MIR-Q09** | unresolved; observation | No executable vector has a matching gameplay capture. Metadata/path/pierce effects are observable; confirmation remains pending, not impossible. No gameplay campaign is part of these records.

## 8. Test vectors

### 8.1 Executable vectors

Each vector is isolated at its named stage unless explicitly stated. All table values are supplied synthetic records. Unlisted request flags are 0; units and records are valid for the named stage. `no draw` means the local stage, excluding an invoked supplied operation's own budget. Supplied operations append ordered arguments to an initially empty call log; allocation supplies the expressly named success/absence, and supplied direct metadata/path writes persist for the vector. Stat/attachment/update/planner/damage operations log requests only and imply no original stored effect. Fixtures never silently substitute an absent result for a fatal dependency failure. Each case starts anew and inherits only the vector's shared inputs. All vectors have runtime_observed: no.

- **MIR-V01** Existing metadata and false destroy condition (MIR-R01a/R06; boundary: isolated supplied-metadata stages)
  - Inputs: existing metadata N with source type 0/id 55, skill 9, lifetime 17 and all other metadata fields 0; supplied missile id 42, creation conditions C/D true; CanDestroy false. This fixture enters R01a and then R06 with those supplied fields; intervening creator stages are outside its boundary.
  - Seeds: no draw.
  - Draws: none.
  - Trace: metadata already exists so do not allocate/initialize it; write id and clear C/D; CanDestroy false makes no destroy-condition write.
  - (a)
    - Inputs: destroy condition already true.
    - Draws: none.
    - Outputs: metadata remains N with source (0,55), skill 9, lifetime 17, other metadata fields 0; unit id 42, C/D false, destroy true; metadata-allocation log [], destroy-write log [].
  - (b)
    - Inputs: destroy condition false.
    - Draws: none.
    - Outputs: metadata remains N with source (0,55), skill 9, lifetime 17, other metadata fields 0; unit id 42, C/D false, destroy false; metadata-allocation log [], destroy-write log [].
  - Discriminates: resetting existing metadata loses source id 55; clearing destroy in (a) or setting it in (b) violates the false-CanDestroy no-write branch.
  - runtime_observed: no

- **MIR-V02** Registration and movement request order (MIR-R01b/R02a/R07b; boundary: supplied-input stage)
  - Inputs: fresh missile M id 42 with empty event chain, supplied event E, existing missile-group entries[F], pool P/room A/start(10,20), supplied path H and room A, act 2/environment E2, final speed 2304, path speed/max 0/reset selector 2. Placement insertion is supplied/logged; planner returns 0 and makes no supplied mutations.
  - Seeds: no draw locally; supplied operations no draw.
  - Draws: none.
  - Trace: initialize creation entry, place/reassociate, allocate path then register/update, write speed then call planner; ignore planner 0.
  - (a)
    - Inputs: final speed 2304 and shared placement/event inputs.
    - Draws: none.
    - Outputs: E targets M/type 0/frame -1/persistent true/no override/parameters 0,0/auxiliary 0; unit events [E], group [E,F], act 2/environment E2. Exact request log [path(P,A,10,20,M,0), registry(M), update(M), planner(H,M,0)]; path speed/max 2304/reset 15. Control continues to R08 despite planner result 0.
  - (b)
    - Inputs: final speed 0 and shared placement/event inputs.
    - Draws: none.
    - Outputs: E targets M/type 0/frame -1/persistent true/no override/parameters 0,0/auxiliary 0; unit events [E], group [E,F], act 2/environment E2. Exact request log [path(P,A,10,20,M,0), registry(M), update(M)]; path speed/max 0/reset 2. Control continues to R08, without a moving setter or planner call.
  - (c)
    - Inputs: final speed 0; supplied path H is present but the post-placement unit-room test returns absent; other shared placement/event inputs apply.
    - Draws: none locally; supplied operations no draw.
    - Outputs: E targets M/type 0/frame -1/persistent true/no override/parameters 0,0/auxiliary 0; unit events [E], group [E,F], act 2/environment E2. Exact request log [path(P,A,10,20,M,0), registry(M)]; no update or planner request; path speed/max 0/reset 2. Control continues to R08. This asserts the supplied room-test branch, not an original allocator/room-publication outcome.
  - Discriminates: treating planner 0 as creation failure or calling planner at speed 0 fails log/continuation assertion; original planner effects remain gated. Case (c) rejects update marking without a post-placement room.
  - runtime_observed: no

- **MIR-V03** Descending registry insertion (MIR-R02b; boundary: supplied-input stage)
  - Inputs: bucket 1 initially contains [513,129]; supplied type missile.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: 513>257, then 129<257; splice between. Equality in (b) invokes fatal duplicate assertion. Case (c) compares unsigned 2147483905 > 513 at the head; case (d) reaches the empty tail.
  - (a)
    - Inputs: new id 257.
    - Draws: none.
    - Outputs: bucket chain [513,257,129], all other buckets unchanged.
  - (b)
    - Inputs: new id 129.
    - Draws: none.
    - Outputs: fatal duplicate-id assertion; no null-success result.
  - (c)
    - Inputs: new id 0x80000101; bucket 1 initially [513,129].
    - Draws: none.
    - Outputs: bucket chain [0x80000101,513,129], all other buckets unchanged.
  - (d)
    - Inputs: new id 1; bucket 1 initially [513,129].
    - Draws: none.
    - Outputs: bucket chain [513,129,1], all other buckets unchanged.
  - Discriminates: ascending insertion or silently replacing equal id fails. A signed-id comparison appends the high-bit id in (c), violating its exact head insertion.
  - runtime_observed: no

- **MIR-V04** Range/loops/elapsed/activation (MIR-R03; boundary: isolated rule)
  - Inputs: Range 20, LevRange 2, level 5, Activate 3, SubLoop 1, SubStart 4, SubStop 10, loops 2, elapsed 5, activateOverride 0.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: base r=30; loop adds 12; elapsed subtracts 5 and stores 1280; activation subtracts selected a from unclamped r; override skips loops. Case (f): 2*(-11)+20=-2, activation -5. Case (g): low32(2*2147483647)=-2, then +20=18 and activation 15.
  - (a)
    - Inputs: flags 0.
    - Draws: none.
    - Outputs: totalRange/lifetime/activationCounter (30,30,27); no accumulator write.
  - (b)
    - Inputs: flags 0x8.
    - Draws: none.
    - Outputs: (42,42,39); no accumulator write.
  - (c)
    - Inputs: flags 0x208.
    - Draws: none.
    - Outputs: (37,37,34); accumulator 1280.
  - (d)
    - Inputs: flags 0xA08.
    - Draws: none.
    - Outputs: (37,37,37); accumulator 1280.
  - (e)
    - Inputs: flags 0x8008; rangeOverride 17.
    - Draws: none.
    - Outputs: (17,17,14); no accumulator write.
  - (f)
    - Inputs: flags 0, level -11; shared Range 20, LevRange 2 and Activate 3.
    - Draws: none.
    - Outputs: totalRange/lifetime/activationCounter (-2,-2,-5); no accumulator write.
  - (g)
    - Inputs: flags 0, level 2147483647; shared Range 20, LevRange 2 and Activate 3.
    - Draws: none.
    - Outputs: totalRange/lifetime/activationCounter (18,18,15); no accumulator write.
  - Discriminates: adding loops to rangeOverride fails (e). Widening the product and clamping only at the end makes (g) 32767 instead of 18; clamping negative r at 0 fails (f).
  - runtime_observed: no

- **MIR-V05** Independent clamps (MIR-R03; boundary: isolated rule)
  - Inputs: flag 0x8000; Activate 3.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: clamp r independently from r-3, rather than subtract from clamped totalRange.
  - (a)
    - Inputs: rangeOverride 40000.
    - Draws: none.
    - Outputs: totalRange/lifetime/activationCounter (32767,32767,32767).
  - (b)
    - Inputs: rangeOverride -40000.
    - Draws: none.
    - Outputs: (-32768,-32768,-32768).
  - (c)
    - Inputs: rangeOverride 32768.
    - Draws: none.
    - Outputs: (32767,32767,32765).
  - Discriminates: subtracting after clamping would make (c) activation 32764.
  - runtime_observed: no

- **MIR-V06** Late missing path (MIR-R04; boundary: supplied-input stage)
  - Inputs: post-R03 allocated missile id 42, counter 42, supplied registration completed, path absent, totalRange/lifetime 30, activation 27; unit not released.
  - Seeds: game already `{0x6AC6935F,0}` from one reached seeder; child `{0x6AC6935F,666}`.
  - Draws: none in this stage; earlier game step is input state.
  - Trace: path test returns absent without rollback.
  - Outputs: absent return; unit/counter/seeds/range fields unchanged; no release, callback, planner or damage request.
  - Discriminates: rollback or continuing to callback changes state/log.
  - runtime_observed: no

- **MIR-V07** Path targets/collision/source (MIR-R05/R06/R08; boundary: supplied-input stage)
  - Inputs: present path speed 7/reset selector 2; absent targetUnit, point (101,101), Collision 1, CollideType 1, CanDestroy/LastCollide true, owner player id 7; fresh source type 0/id -1.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: clear speed/max, choose point low16, request footprint, select code mask, set destroy and optional source.
  - (a)
    - Inputs: shared defaults.
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0x84, destroy true, source (0,7).
  - (b)
    - Inputs: Collision 0; CollideType 8.
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 0, mask 0x185, destroy true, source (0,7).
  - (c)
    - Inputs: LastCollide false.
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0x84, destroy true, source (0,-1), no source write.
  - (d)
    - Inputs: point (-1,65537).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (65535,1), no tracking, footprint request 64, mask 0x84, destroy true, source (0,7).
  - (e)
    - Inputs: CollideType 2; targetUnit T is monster id 99, entering point target fields (7,8); all other shared inputs apply.
    - Draws: none.
    - Outputs: speed/max 0, reset 15, tracking T/type 1/id 99, no point-target write, footprint request 64, mask 0x104, destroy true, source (0,7).
  - (f)
    - Inputs: CollideType 3; shared absent targetUnit and point (101,101).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0x184, destroy true, source (0,7).
  - (g)
    - Inputs: CollideType 6; shared absent targetUnit and point (101,101).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 4, destroy true, source (0,7).
  - (h)
    - Inputs: CollideType 7; shared absent targetUnit and point (101,101).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0x40, destroy true, source (0,7).
  - (i)
    - Inputs: CollideType 0; shared absent targetUnit and point (101,101).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0, destroy true, source (0,7).
  - (j)
    - Inputs: CollideType 4; shared absent targetUnit and point (101,101).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0, destroy true, source (0,7).
  - (k)
    - Inputs: CollideType 5; shared absent targetUnit and point (101,101).
    - Draws: none.
    - Outputs: speed/max 0, reset 15, point (101,101), no tracking, footprint request 64, mask 0x104, destroy true, source (0,7).
  - Discriminates: signed rather than low16 coordinate storage fails (d). Case (e) requires all tracking identity fields and distinguishes tracking from a point-target write; cases (e)-(i) exercise every additional distinct mask result. Selectors 4/5 are exercised separately in cases (j)/(k). Cases (j)/(k) exercise the exact selector-4/5 mappings despite sharing outputs with other selectors; an incorrect value at either code-table position now fails explicitly.
  - runtime_observed: no

- **MIR-V08** Speed and maximum overwrite (MIR-R07a/R09, with supplied speed from MCI-R05/MCI-R07; boundary: supplied-input stage)
  - Inputs: Vel 10, VelLev 4, level 5, MaxVel 30, Accel -2, CanSlow false; supplied path speed 0/reset selector 2.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: (a) trunc(20/8)=2; 12<<8=3072; 230400/100=2304; moving setter then max overwrite. (b) 7<<8=1792, scaled 1344. (c) trunc(525/100)=5.
  - (a)
    - Inputs: flags 0.
    - Draws: none.
    - Outputs: speed/max/acceleration (2304,7680,-2), reset selector 15.
  - (b)
    - Inputs: flag 0x4; override 7.
    - Draws: none.
    - Outputs: speed/max/acceleration (1344,7680,-2), reset selector 15.
  - (c)
    - Inputs: flags 0x14; override 7.
    - Draws: none.
    - Outputs: speed/max/acceleration (5,7680,-2), reset selector 15.
  - Discriminates: scaling maximum or treating Accel as 8.8 changes 7680/-2. The speed values 2304/1344/5 are MCI-R05/MCI-R07 assertions.
  - runtime_observed: no

- **MIR-V09** Retained-point duration and width (MIR-R10; boundary: supplied-current-position stage)
  - Inputs: flag 0x400, current path position (10,20), retained point (16,24), final speed 2304, totalRange 30/activationCounter 27, tracking unit at (99,99). No path/seed writes occur.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: (a) dx 6/dy 4, s=16/d=8, numerator 524288/denominator 36864 =>14. (b) d=0→1,65536/36864=>1. (c)/(d)denominator 0=>0. (e) dx 65535,d=65535,numerator 4294901760/16=>268431360, clamp 32767. (f) wrapped delta INT_MIN negates to itself then clamps 0; d=0→1 =>1. (g) dx 1073741824,s INT_MIN,d-1073741824, low32(d<<16)=0 =>0. Case (h): dx=4, dy=6, s=16, d=8; unsigned 524288/36864=14.
  - (a)
    - Inputs: shared defaults.
    - Draws: none.
    - Outputs: lifetime 14, totalRange 30/activationCounter 27.
  - (b)
    - Inputs: retained point (10,20).
    - Draws: none.
    - Outputs: lifetime 1, totalRange 30/activationCounter 27.
  - (c)
    - Inputs: speed 0.
    - Draws: none.
    - Outputs: lifetime 0, totalRange 30/activationCounter 27.
  - (d)
    - Inputs: speed 0x10000000.
    - Draws: none.
    - Outputs: lifetime 0, totalRange 30/activationCounter 27.
  - (e)
    - Inputs: point (65545,20), speed 1.
    - Draws: none.
    - Outputs: lifetime 32767, totalRange 30/activationCounter 27.
  - (f)
    - Inputs: point (-2147483638,20), speed 2304.
    - Draws: none.
    - Outputs: lifetime 1, totalRange 30/activationCounter 27.
  - (g)
    - Inputs: point (1073741834,20), speed 1.
    - Draws: none.
    - Outputs: lifetime 0, totalRange 30/activationCounter 27.
  - (h)
    - Inputs: retained point (14,26), current path position (10,20), final speed 2304; shared range/activation and tracking inputs apply.
    - Draws: none.
    - Outputs: lifetime 14, totalRange 30/activationCounter 27; no path/seed writes.
  - Discriminates: using tracking position, retaining full numerator, signed division, ignoring shift wrap or treating INT_MIN absolute as positive changes outputs. Case (h) rejects always treating dx as the larger component: that computes d=7 and lifetime 12.
  - runtime_observed: no

- **MIR-V10** Fresh extended base (MIR-R11a; boundary: supplied-allocation stage)
  - Inputs: no base list, successful list/storage allocations, missile type 3/id 42/pool P, stateCount 99, setup conditions/callback/context 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: zero fresh base and state storage; associate missile; required state words ceil(99/32)=4 per active/changed collection.
  - Outputs: base associated(type 3,id 42,pool P), extended true/base condition 0, own/resolved/primary/secondary empty, active/changed state words[0,0,0,0], callback/context 0; no cleanup request.
  - Discriminates: retaining stale membership or allocating only 3 words fails exact initialized state.
  - runtime_observed: no

- **MIR-V11** Callback and narrow metadata (MIR-R12; boundary: supplied-input stage)
  - Inputs: base setup completed; callback records (missile,99), writes skill 77/level 88 then returns 123; request level 70000.
  - Seeds: no draw locally; supplied callback makes no draw.
  - Draws: per independent case below.
  - Trace: callback runs first; skill clamp and level narrowing overwrite its metadata changes; return 123 is ignored.
  - (a)
    - Inputs: request skill 0.
    - Draws: none.
    - Outputs: callback log [(missile,99)]; skill 0; level bits/read 4464/4464.
  - (b)
    - Inputs: request skill 40000.
    - Draws: none.
    - Outputs: callback log [(missile,99)]; skill 32767; level bits/read 4464/4464.
  - (c)
    - Inputs: request skill 54.
    - Draws: none.
    - Outputs: callback log [(missile,99)]; skill 54; level bits/read 4464/4464.
  - (d)
    - Inputs: request skill 54/level 40000.
    - Draws: none.
    - Outputs: callback log [(missile,99)]; skill 54; level bits/read 40000/-25536.
  - (e)
    - Inputs: callback absent, entering skill 77/level 88, request skill -1/level 70000.
    - Draws: none.
    - Outputs: callback log []; skill 0; level bits/read 4464/4464.
  - Discriminates: callback after metadata or clamping level changes outputs. Case (e) rejects an unconditional callback and negative-skill narrowing without the required clamp.
  - runtime_observed: no

- **MIR-V12** Late ordered requests (MIR-R13/R14a/R16/R17a; boundary: supplied-input stage)
  - Inputs: owner player id 7, request source S a distinct missile id 99, new missile M id 42 with stored skill level 4464, original request level 70000/toHit 123, flags 0x11000, metadata conditions 8, DamageRate 9, owner-valid false, creation condition D true, Pierce false; base present/extended, stateCount 99, state 98 inactive, matching source list L already present; supplied damage record D with every value/condition 0; stat/lifecycle operations log only.
  - Seeds: no draw locally; supplied weapon/damage/monster operations no draw.
  - Draws: none.
  - Trace: activation→builder→store→attribution→optional attack stat→metadata condition→monster wrapper→rate stat→return. Attribution fails before valid marker when its new list allocation fails, but creator continues.
  - (a)
    - Inputs: shared player owner and existing source list L.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), update(M), listWrite(L,353,0,0), listWrite(L,354,7,0), unitWrite(M,19,123,0), unitWrite(M,327,9,0)]; direct owner (0,7), owner-valid true, state 98 active/changed, creation condition D false, metadata conditions 10; final creator return M.
  - (b)
    - Inputs: monster owner id 8; existing source list L.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), update(M), listWrite(L,353,1,0), listWrite(L,354,8,0), unitWrite(M,19,123,0), monsterEvent(game,owner,M,5), unitWrite(M,327,9,0)]; direct owner (1,8), owner-valid true, state 98 active/changed, creation condition D false, metadata conditions 10 already at monsterEvent; final creator return M.
  - (c)
    - Inputs: player owner id 7; source list absent; its allocation returns absent.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), update(M), allocate(type 3,id 42,state 98), unitWrite(M,19,123,0), unitWrite(M,327,9,0)]; no attach/source write; direct owner (0,7), owner-valid false, state 98 active/changed, creation condition D false, metadata conditions 10; final creator return M.
  - (d)
    - Inputs: player owner id 7; base list absent.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), unitWrite(M,19,123,0), unitWrite(M,327,9,0)]; direct owner (0,7), owner-valid true, creation condition D false, metadata conditions 10, no state/list/update request; final creator return M. Actual damage/list/stat integration is gated in every case.
  - (e)
    - Inputs: player owner id 7, present extended base, stateCount 98, inactive/unchanged state 98, existing source list L; all other shared inputs apply.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), listWrite(L,353,0,0), listWrite(L,354,7,0), unitWrite(M,19,123,0), unitWrite(M,327,9,0)]; direct owner (0,7), owner-valid true, state 98 remains inactive/unchanged, no update request, creation condition D false, metadata conditions 10; final creator return M. Stored source-stat effects remain gated.
  - (f)
    - Inputs: player owner id 7, present extended base, stateCount 99, state 98 already active with changed membership false, existing source list L; all other shared inputs apply.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), update(M), listWrite(L,353,0,0), listWrite(L,354,7,0), unitWrite(M,19,123,0), unitWrite(M,327,9,0)]; direct owner (0,7), owner-valid true, state 98 active with changed membership still false, creation condition D false, metadata conditions 10; final creator return M.
  - (g)
    - Inputs: player owner id 7, present extended base, stateCount 99, state 98 inactive/unchanged, source list absent, source-list allocation succeeds with list Lnew; attachment operation logs only; all other shared inputs apply.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), update(M), allocate(type 3,id 42,state 98), attach(M,Lnew,full), listWrite(Lnew,353,0,0), listWrite(Lnew,354,7,0), unitWrite(M,19,123,0), unitWrite(M,327,9,0)]; Lnew associated with missile type 3/id 42 and state 98; direct owner (0,7), owner-valid true, state 98 active/changed, creation condition D false, metadata conditions 10; final creator return M. Actual attachment and stored source-stat propagation remain gated.
  - (h)
    - Inputs: flags 0, shared player owner id 7 and existing source list L; entering metadata conditions 8; all other shared inputs apply.
    - Draws: none locally; supplied dependencies no draw.
    - Outputs: ordered log [weapons(owner,0), build(owner,S,M,70000), store(owner,M,D,70000), update(M), listWrite(L,353,0,0), listWrite(L,354,7,0), unitWrite(M,327,9,0)]; no stat-19 request or metadata-condition write; direct owner (0,7), owner-valid true, state 98 active/changed, creation condition D false, metadata conditions remain 8; final creator return M.
  - Discriminates: passing stored narrowed level 4464 to builder, setting valid before failed allocation, or placing rate before monster event changes exact outputs/order. Case (e) separates the state-count membership/update gate from the subsequent source-list work; skipping all attribution below count 99 fails its request log. Case (f) rejects forced changed membership for an already active state; (g) requires allocation before full-attach request before list writes; (h) rejects unconditional optional writes.
  - runtime_observed: no

- **MIR-V13** Local pierce seed/count requests (MIR-R15a/R15b, with MCI-R13's draw accounting; boundary: supplied-input stage)
  - Inputs: player owner, missile, Pierce true; owner own-base stat 328=1, aggregate stat 156=c/stat 166=0, unrelated resolved stat 328=12345. Game seed {9,10}, owner seed {11,12}, missile seed {13,14}, game counter 41 are untouched markers; only count requests are asserted.
  - Seeds: local initialization `{1,666}` for (a)-(c)/(f)/(j)-(l); `{12345,666}` for (e); (d)/(g)-(i) no local initialization/step.
  - Draws: per independent case below.
  - Trace: compare signed remainder with c and stop before fifth draw; c=0 skips entire helper body. Case (k) obtains chance from stat 166 alone. Case (l) wraps 2147483647+1 before the signed remainder comparison, so the first roll fails.
  - (a)
    - Inputs: c=60.
    - Draws: local mod 100: 51,31,12,93.
    - Outputs: layer-0 missile stat 328 request 3; local final seed {0xB6AAB839,0x10B1BFE6}; game/owner/missile seeds, counter and owner stat unchanged.
  - (b)
    - Inputs: c=20.
    - Draws: local mod 100: 51.
    - Outputs: layer-0 missile stat 328 request 0; local final seed {0x6AC6935F,0}; game/owner/missile seeds, counter and owner stat unchanged.
  - (c)
    - Inputs: c=100.
    - Draws: local mod 100: 51,31,12,93.
    - Outputs: layer-0 missile stat 328 request 4; local final seed {0xB6AAB839,0x10B1BFE6}; game/owner/missile seeds, counter and owner stat unchanged.
  - (d)
    - Inputs: c=0.
    - Draws: none.
    - Outputs: no count request or local initialization; game/owner/missile seeds, counter and owner stat unchanged; supplied empty list still reads 0.
  - (e)
    - Inputs: own-base stat 328=12345; c=70; local initial seed {12345,666}.
    - Draws: local mod 100: 87.
    - Outputs: layer-0 missile stat 328 request 0; local final seed {0x15B2E77,0x141D}; game/owner/missile seeds, counter and owner stat unchanged.
  - (f)
    - Inputs: c=-1.
    - Draws: local mod 100: 51.
    - Outputs: layer-0 missile stat 328 request 0; local final seed {0x6AC6935F,0}; game/owner/missile seeds, counter and owner stat unchanged.
  - (g)
    - Inputs: c=60, owner object; all other shared inputs apply.
    - Draws: none; no local initialization.
    - Outputs: no count request; supplied game/owner/missile seeds, counter and owner stat unchanged.
  - (h)
    - Inputs: c=60, recipient player instead of missile; all other shared inputs apply.
    - Draws: none; no local initialization.
    - Outputs: no count request; supplied game/owner/recipient seeds, counter and owner stat unchanged.
  - (i)
    - Inputs: c=60, Pierce false; all other shared inputs apply.
    - Draws: none; no local initialization.
    - Outputs: no count request; supplied game/owner/missile seeds, counter and owner stat unchanged.
  - (j)
    - Inputs: c=60, owner monster; all other shared inputs apply; local initial seed {1,666}.
    - Draws: local mod 100 results 51,31,12,93.
    - Outputs: layer-0 missile stat 328 request 3; local final seed {0xB6AAB839,0x10B1BFE6}; game/owner/missile seeds, counter and owner stat unchanged.
  - (k)
    - Inputs: aggregate stat 156=0, stat 166=60, so c=60; owner own-base stat 328=1; local initial seed {1,666}.
    - Draws: local mod 100 results 51,31,12,93.
    - Outputs: layer-0 missile stat 328 request 3; local final seed {0xB6AAB839,0x10B1BFE6}; game/owner/missile seeds, counter and owner stat unchanged.
  - (l)
    - Inputs: aggregate stat 156=2147483647, stat 166=1, so wrapped c=-2147483648; owner own-base stat 328=1; local initial seed {1,666}.
    - Draws: local mod 100 result 51.
    - Outputs: layer-0 missile stat 328 request 0; local final seed {0x6AC6935F,0}; game/owner/missile seeds, counter and owner stat unchanged.
  - Discriminates: using resolved 328, stepping an owner seed, masking with 99 or taking a fifth draw fails exact results/logs. Cases (g)-(i) each isolate one helper eligibility guard; (j) distinguishes accepted monster owners from a player-only shortcut. The helper fixtures do not claim those refused requests reach this stage through creator validation. Ignoring stat 166 fails (k); widening the chance sum fails (l).
  - runtime_observed: no

### 8.2 Observation vectors

- **MIR-O01** Range/loops/elapsed (MIR-R03): capture request and row values and metadata at creation; expect the exact independent clamp arithmetic in MIR-V04/V05. [Procedure](../research/spec-evidence/missile-initialization-and-registry.md#mir-o01).
- **MIR-O02** Pierce producer/consumer (MIR-R15/MIR-Q07): read owner own-base 328 and chance before successive shots, local draw/count requests and resulting missile stat; expect deterministic helper results for each captured input. [Procedure](../research/spec-evidence/missile-initialization-and-registry.md#mir-o02).

## 9. Acceptance and dependencies

Partial acceptance consists of MIR-V01..V13's exact local arithmetic, guards, fields, returns and ordered requests at their supplied-input boundaries, including the explicit fatal/unchecked-failure distinctions. Verified-rule rows name their vectors; gated source-list/stat, path, event lifecycle, damage and mod effects are not inferred from these passing requests. Full behavior reproduction additionally requires MIR-Q01/Q02/Q03/Q04/Q05/Q06/Q08's in-scope integration/assertions, [missile-creation-and-seeding](missile-creation-and-seeding.md)'s acceptance for the stages before metadata, and applicable original-game observations, including MIR-O01. MIR-Q07 and MCI-Q02 are upstream/wider meaning questions and impose no local supplied-input gate; MIR-Q09 blocks runtime-confirmation claims. A partial pass never establishes complete creation parity.

**Depends on:** [missile creation and seeding](missile-creation-and-seeding.md) MCI-R01..R11 for the allocated, seeded and numbered missile, start room/point, retained point, surviving targetUnit and final speed; [foundations](core-numeric-and-rng-foundations.md) NRF-R01/R03a for the local pierce seed, NRF-R08b for speed-setter behavior, RLI-R03/RLI-R02a/RLI-R02b for existing state/stat reads and RLI-Q02/RLI-Q03 for update/stat integration; [game data](core-game-data-fields.md) GDA-R03/GDA-R10/GDA-R12/GDF-R02/GDF-R03 for schemas and identity conventions; [state lifecycle](state-timed-state-lifecycle.md) STL-R01 for supplied successful frame-event allocation, STL-R06a for list allocation and STL-R06b/c/R07/R07a for the exact attachment/write/membership limits; [missile damage parameters](missile-damage-parameters.md) MDP-R01..R12 at its supplied-input/request boundaries, especially MDP-R07b/R08b/R11b/R12b gates. Wider planner/room producers require their named open integration contracts. Owners of the cited GDA, RLI ids: [core-game-data-table-records](core-game-data-table-records.md) (GDA); [core-record-layouts-and-ids](core-record-layouts-and-ids.md) (RLI).

**Depended on by:** [missile creation and seeding](missile-creation-and-seeding.md) for the request's final return and draw accounting; [missile movement](missile-movement-and-collision.md) for initialized speed/lifetime/activation/collision; [missile contact](missile-hit-and-damage-calculation.md) for count and attribution; [missile damage parameters](missile-damage-parameters.md) for stored skill level/context; [PvP attribution](combat-pvp-damage-and-kill-credit.md) PVP-R07 for direct owner link; creating skill contracts including [Blade skills](skill-assassin-blade-sentinel-and-fury.md). These consumers do not establish the upstream operations here.

## 10. Related specs

- [missile-creation-and-seeding](missile-creation-and-seeding.md): the same request's validation, aim, speed, allocation, seed and identity stages.
- [missile-movement-and-collision](missile-movement-and-collision.md): later planner/movement/countdown/collision ownership.
- [missile-damage-parameters](missile-damage-parameters.md): damage initialization, element selectors and damage-owned draws.
- [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md): pierce count use and hit processing.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): direct owner-link resolution at impact.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): concrete missile caller and caller-owned damage/RNG boundaries.

## Appendix A. Provenance

Original LoD Expansion 1.13c Windows x86, identified by [source manifest](../research/source-manifest.md) artifacts LOD113C-D2GAME and LOD113C-D2COMMON. Compatible 2026-10-12 static readings are retained at their recorded boundaries. Targeted 2026-10-15 source readings and original-image decoding establish registry order, point-distance arithmetic, base-list setup and local request/failure distinctions; named text-cell records reuse complete 2026-10-15 loader payloads with profile limits. No gameplay vector is confirmed. Detailed original artifact/function locators, corrections, evidence assessment, observation procedures and review status are in the [companion](../research/spec-evidence/missile-initialization-and-registry.md); remaining evidence is in the [gap file](../plans/gaps/missile-initialization-and-registry.md).
