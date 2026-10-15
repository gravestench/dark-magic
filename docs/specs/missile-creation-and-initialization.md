# Missile Creation and Initialization

**Description**

When an attack, skill or another game action asks for a projectile, the server checks its owner and starting place before creating it. It chooses the projectile's aim, speed and lifetime, prepares its movement and damage, and records who created it. Some requests fail before a projectile exists; a later failure can leave an already created projectile registered with the game.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md); names: [core-game-data-fields](core-game-data-fields.md)
Ids: MCI-R, MCI-Q, MCI-V, MCI-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/missile-creation-and-initialization.md)
Open evidence gaps: [gap file](../plans/gaps/missile-creation-and-initialization.md)

## 1. Scope

This contract owns one server missile-creation request: refusals, start room and point, aim, initial speed, unit identity and seed, initial lifetime/activation, path-setup requests, skill metadata, damage-setup requests, owner attribution, pierce-count calculation and stat-write requests. It also identifies the subsequent client-notification boundary. Per-frame movement, collision, countdown and explosions belong to [missile-movement-and-collision](missile-movement-and-collision.md); computed damage belongs to [missile-damage-parameters](missile-damage-parameters.md); hit resolution belongs to [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md). Skill callers and message contents have their own contracts. Supplied dependency results permit the verified local stages below; they do not establish complete path, stat-lifecycle, monster-mod or synchronization integration.

## 2. Trigger and ownership

D2Game's server `MISSILES` subsystem runs synchronously when a skill, missile, object or AI action supplies a creation request. One identified route is the skill server do-function table, basic-attack entry, positioned-missile request, then this creator. Each invocation returns at most one missile. It is not a periodic operation and has no independently fixed position within a frame beyond its caller's execution. Later missile updates are owned by the movement contract.

D2Common supplies missile metadata, room/path operations, state/stat primitives and owner attribution. The creator requests update marking during placement; it sends no client message directly. A later synchronization caller can select the missile-creation message; recipient selection and complete delivery remain MCI-Q08. Binary entry points, call chains and original request/layout mappings are in the companion.

## 3. Inputs and data

Runtime integers are i32 unless specified. A bit test uses the u32 request flags. All arithmetic wraps to 32 bits before the next operation unless a rule explicitly widens it. Signed division truncates toward zero. `low16`/`low32` retain the indicated low bits. `sat16(v)` interprets v as i32 and clamps to -32768..32767. Positions are path units; speed is signed 8.8; lifetime/activation are frame counters.

| Input | Semantic source | Width / format | Meaning |
|---|---|---|---|
| owner, source, targetUnit | request unit references | unit or absent; type/id i32 | Owner must be player or monster; source supplies position and damage context; targetUnit supplies tracking aim |
| missileId, record count | request; loaded Missiles collection | i32 | Zero-based missile record index and signed bounds |
| flags | request conditions below | u32 | Select supplied start, aim, speed and metadata operations |
| srcX/srcY, tgtX/tgtY | request | i32 path units | Explicit start and absolute target or target offset |
| velOverride, skill, level, loops | request | i32; speed override whole units or 8.8 | Speed, skill metadata, speed/range scaling and loop count |
| elapsed, activateOverride, toHit, rangeOverride | request | i32 | Already elapsed frames, activation subtraction, attack-rating request and total range override |
| callback, callbackArg | request | optional operation; opaque supplied argument | Called with the new missile and argument; callback effects/draws are supplied by its owner |
| owner/source/target position | unit position operation; NRF-R10b for dynamic paths | unsigned 16-bit integer path components for player/monster/missile; i32 for static positions | Source start and target/owner coincidence tests |
| owner room and candidates | current room and its ordered adjacent-room collection | room or absent; candidate count u32 | Supplied rectangle/candidate data for start-room selection |
| room rectangle / level | room x0,y0,width,height; level identity | i32 | Half-open signed rectangle tests and act association |
| game context | allocation pool, expansion condition, act environments, missile counter, game seed | logical pool/environments; boolean; u32 counter; two u32 seed words | Allocation, base fields, identity and seeding |
| allocation and placement results | supplied unit, event, missile-data, path and base-list operations | success, absent or fatal outcome as specified by their owner | A missing path result is distinguishable from a dependency fault |
| existing missile metadata/path/base list | outputs of earlier stages or supplied stage fixtures | logical fields with widths in rules | Subsequent initialization consumes their current values |
| owner state/list data | state 87; ordered existing stat arrays and definitions | NRF-R12a/b/c | State predicate and own-list slow percentage; pierce aggregate and own-base values |
| owner source-list data | state count, extended-base condition, existing state-98 list; allocation result | i32 count; booleans; list or absent | Attribution requests and its early allocation-failure branch |
| registry bucket | current ordered chain for missile identity | u32 identities, unique within type | Insertion fixture; bucket selector is id modulo 128 |
| monster-mod/client context | owner monster modifiers; supplied synchronization eligibility/owner lookup | supplied operations/results | Their complete production/execution remains MCI-Q07/MCI-Q08 |

| Request flag | Established local behavior |
|---|---|
| 0x1 | Explicit start point |
| 0x2 / 0x20 | Start-relative / absolute target point; 0x2 wins |
| 0x4 / 0x10 | Speed override / override already in 8.8 |
| 0x8 | Add animation sub-loop duration to range |
| 0x200 / 0x400 | Subtract elapsed frames / replace lifetime with point-distance duration |
| 0x800 / 0x1000 | Activation override / request attack-rating write |
| 0x8000 / 0x10000 | Range override / set missile metadata condition 2 |

| Data field | Source / logical width | Consumer |
|---|---|---|
| `Range`, `LevRange` | `data/global/excel/missiles.txt`, i16 | MCI-R14 |
| `Vel`, `VelLev`, `MaxVel`, `Activate` | Same file, u8 | MCI-R05/R14/R20 |
| `Accel` | Same file, i16 | MCI-R20 |
| `SubLoop`, `SubStart`, `SubStop` | Same file, u8 | MCI-R14 |
| `CollideType`, `Collision` | Same file, u8 | MCI-R11/R16; accepted mask domain 0..8 |
| `DamageRate` | Same file, i32 | MCI-R27 |
| `LastCollide`, `Pierce`, `CanSlow`, `CanDestroy`, `ClientSend` | Same file, booleans; GDF-R07 | MCI-R19/R26/R06/R17/R29 |
| `Light`, `RandStart`, `HitShift` | Same file; not consumed directly here | MCI-R30; damage-owned `HitShift` use is MDP |

| Named identity | Read / request | Origin | Binding |
|---|---|---|---|
| state 87 `slowmissiles`; stat 161 `skill_handofathena` | Active-state test; layer-0 own state-list percentage | Named States/ItemStatCost cells in [table record](../research/spec-evidence/missile-creation-and-initialization.md#named-cells-2026-10-15) | Name/id mapping from data; consumer identities fixed in code |
| stats 156 `item_pierce`, 166 `skill_pierce`, 328 `pierce_idx` | Aggregate chance components; own-base seed input; count write request | Same named-cell record | Mapping from data; ids fixed by code |
| stats 19 `tohit`, 327 `damage_framerate` | Replacing layer-0 unit-stat requests | Same named-cell record | Mapping from data; ids fixed by code |
| state 98 `sourceunit`; stats 353 `source_unit_type`, 354 `source_unit_id` | Source membership and list-stat requests | Same named-cell record | Mapping from data; ids fixed by code |
| 75%, 100-unit rejection, pierce cap 4, high seed word 666, collision masks, act thresholds, clamp endpoints | Rules below; NRF-R03a for seed initializer | Engine code/constants | Fixed by original code |
| missile fields and flags | Missiles schema identified by original loader; GDF-R01/GDF-R07 | Table data; flag consumers in code | Data supplies values; code fixes their interpretation |

Named table records establish text-cell identities only within their recorded profile. Synthetic row values below do not assert a shipped missile row or complete text/compiled correspondence. Actual stat mutation/propagation is separate from a logged write request (NRF-Q20, STL-R06b/c).

## 4. Rules (execution order)

**MCI-R01** Request validation [verified in 1.13c code (static)]. Return absent with no allocation, counter write or draw for absent owner, signed missileId < 0 or >= count, owner type outside player/monster, or absent record after lookup.

**MCI-R02** Start point [verified in 1.13c code (static)]. Flag 0x1 selects `(srcX,srcY)`. Otherwise an absent source returns absent without draws; a present source supplies both coordinates. The owner need not be the source.

**MCI-R03** Start room [verified in 1.13c code (static)]. An absent owner room yields absent creation. Test the owner room first with signed half-open bounds `x0 <= X < low32(x0+width)` and `y0 <= Y < low32(y0+height)`. Otherwise visit its supplied adjacent rooms in stored order, skipping absent entries, and choose the first containing point. No match returns absent before allocation/seeding. The candidate getter returns the stored collection/count; it does not sort or construct rooms. Their producer is outside this consumer stage (MCI-Q11).

**MCI-R04** Retained target point [verified in 1.13c code (static)]. If flag 0x2, set point to `low32(start+request offset)` per axis; else if 0x20, use request target; otherwise use start. This retained point is distinct from the target-unit position used for R08's distance test.

**MCI-R05** Initial speed [verified in 1.13c code (static)]. With flag 0x4, use velOverride directly if 0x10, otherwise `low32(velOverride<<8)`. Without 0x4, compute `low32((trunc(i32(VelLev*level)/8)+Vel)<<8)`, wrapping the product and addition separately. Both table velocity fields are unsigned bytes.

**MCI-R06** Slow-missiles scale [verified in 1.13c code (static)]. When CanSlow is set, owner state 87 is active and its state-87 list exists, read layer-0 stat 161 from that list's own array, then set speed to `trunc(i32(value*speed)/100)`. The unit aggregate is not read. Missing state or list leaves speed unchanged. This consumes NRF-R12a/b; it does not produce the state/list.

**MCI-R07** Global speed scale [verified in 1.13c code (static)]. If speed != 0, set it to `trunc(i32(speed*75)/100)`. R08/R18/R21 use this result. Maximum speed is scaled separately by R20 and does not receive this percentage.

**MCI-R08** Aim and distance refusal [verified in 1.13c code (static)]. Skip this entire rule when final speed is 0. Otherwise, with no targetUnit and retained point equal to start, increment both retained coordinates with 32-bit wrap. With a targetUnit different from owner whose position equals the owner's position, increment both retained coordinates and drop targetUnit; this test uses owner position, even when source/start differs. Keep targetUnit when it is the owner. For distance only, use the remaining targetUnit's current position when present, otherwise the retained point. Per axis form the wrapped signed delta from start, negate with wrap if negative, then reject if its signed value >= 100. Thus ordinary deltas 99 pass and 100 fail; an INT_MIN delta remains INT_MIN and does not fail this signed comparison. A surviving targetUnit is stored by R16; its snapshot does not replace the retained point used by R21.

**MCI-R09** Unit allocation/base fields [verified in 1.13c code (static)]. Request zeroed missile allocation from the game pool. An absent allocator result, if supplied to the creator, returns absent before seeding. The original allocation helper zeroes its allocation without a null check; low-level allocation failure therefore has no established safe null-return fallback. On success set type missile, class missileId, game/pool references, expansion condition equal to the game's expansion condition, creation condition A true, auxiliary reference empty, and act/environment from start-room level. Signed level thresholds select act 0 below 40, 1 below 75, 2 below 103, 3 below 109, 4 below 1024, else 0. These associations do not give condition A a wider meaning (MCI-Q09).

**MCI-R10** Unit seed [verified in 1.13c code (static)]. Consume NRF-R03a/b and NRF-R17a at the reached missile seeder: initialize the child seed to `{1,666}`, step the game seed once, store its new lo as the missile's initial seed copy, then set child seed to `{new game.lo,666}`. The child is not stepped. R01-R09's early refusals never reach this operation; this is not a whole-creation draw guarantee.

**MCI-R11** Identity/mode/conditions [verified in 1.13c code (static)]. Set creation condition B true and creation selector 11. Set id to `low32(game.missileCounter+1)`, replacing exactly 0xFFFFFFFF with 1, then store counter=id. Zero is not separately skipped. Set missile mode to the unsigned CollideType value. Meanings of B and selector 11 remain MCI-Q09.

**MCI-R12** Missile metadata/registration.
- **MCI-R12a** [verified in 1.13c code (static)]. If missile metadata is absent request zeroed metadata allocation and initialize last-collision source id to -1. Existing metadata is retained. Store the unit id and clear creation conditions C and D. Metadata-allocation failure has no checked absent-result return in this wrapper; dependency faults are not converted to null creation.
- **MCI-R12b** [verified in 1.13c code (static)]. For a fresh unit with empty frame-event chain and a supplied successful event allocation, associate a creation event with this missile, type 0, frame -1, persistent-creation condition true, no override, parameters 0,0 and auxiliary value 0. Insert it first in the missile creation-event group and in this unit's event chain. Subsequent execution/removal and replacement of a nonempty old chain remain MCI-Q03. These are frame-event entries, distinct from STL-R11a unit timers.

**MCI-R13** Placement.
- **MCI-R13a** [verified in 1.13c code (static)]. Reassociate act/environment as R09. Request path creation with game pool, start room, start coordinates, missile and extra condition 0; then request registry insertion. If the resulting unit has a room, request update marking. The supplied path result and room test determine this local request order. Complete path allocation/room-collision publication and update-queue effects remain MCI-Q11/NRF-Q17; faults are not specified as a successful absent-path result.
- **MCI-R13b** [verified in 1.13c code (static)]. In the supplied missile registry bucket selected by id modulo 128, compare ids unsigned. Insert before the first id smaller than the new id, yielding descending order; append if none. Equal id takes a fatal assertion path, not a null return. No draw occurs. Wider registry lifecycle/lookup policy is outside this insertion boundary.

**MCI-R14** Range/lifetime/activation [verified in 1.13c code (static)]. If 0x8000 use rangeOverride; otherwise compute `r=i32(i32(LevRange*level)+Range)` from sign-extended i16 fields. Only on the nonoverride branch, if 0x8 and SubLoop != 0, add `i32((SubStop-SubStart)*loops)` with wrap. Bytes are unsigned before subtraction. If 0x200, subtract elapsed with wrap and store animation accumulator `low32(elapsed<<8)`. Store totalRange=`sat16`(r), lifetime=`sat16`(r). Choose a=activateOverride if 0x800, else unsigned Activate; store activationCounter=`sat16`(i32(r-a)), using unclamped r. Later countdown/activation is movement-owned.

**MCI-R15** Missing path [verified in 1.13c code (static)]. After R14, absent path returns absent without freeing the allocated unit or undoing prior calls/writes. Its seed step and id remain consumed. Registration/placement stored effects are only those established or supplied in R12/R13.

**MCI-R16** Path setup [verified in 1.13c code (static)]. With a present path:
1. Set speed and maximum speed to 0; set path reset selector 15 only if the entering speed was nonzero (NRF-R08b).
2. With surviving targetUnit, store tracking unit/type/id. Otherwise store low16 of each retained target coordinate and clear tracking unit. Speed 0 still preserves the request's targetUnit because R08 was skipped.
3. Request collision-footprint value 64 if Collision != 0, otherwise 0; room collision-map effects are MCI-Q11. Store the collision mask selected by CollideType: 0→0, 1→0x84, 2→0x104, 3→0x184, 4→0, 5→0x104, 6→4, 7→0x40, 8→0x185. This code table is not a data-remappable policy; selectors above 8 remain MCI-Q12.

**MCI-R17** Destroy condition [verified in 1.13c code (static)]. When CanDestroy is set, set the missile's destroy condition true. No clear is requested when it is false. Its wider reader behavior is MCI-Q09.

**MCI-R18** Moving missile initialization.
- **MCI-R18a** [verified in 1.13c code (static)]. Only when final speed != 0, set path speed and maximum speed to that speed, including the speed setter's conditional reset-selector behavior (NRF-R08b).
- **MCI-R18b** [verified in 1.13c code (static)]. Immediately after that write request path building with this path, missile and condition 0; ignore its numeric return. Speed 0 skips both stages. The planner's path/direction/room mutations and draw count remain MCI-Q04; this local request is available with a supplied planner operation.

**MCI-R19** Last-collision source [verified in 1.13c code (static)]. With LastCollide set, store owner type/id in missile metadata. Otherwise retain existing values, including -1 source id after fresh allocation.

**MCI-R20** Acceleration/maximum speed [verified in 1.13c code (static)]. Store sign-extended Accel directly, without a fixed-point shift; then overwrite path maximum speed with `MaxVel<<8`. This runs even for speed 0 and does not receive R07's 75% scale.

**MCI-R21** Point-distance lifetime [verified in 1.13c code (static)]. Only with 0x400, read the missile's current integer path position at this stage and the retained R04/R08 point, even when tracking a targetUnit. Per axis compute wrapped delta point-position, negate if negative, then replace a still-negative result with 0. Thus an INT_MIN delta contributes 0. Let `s=i32(2*max(dx,dy)+min(dx,dy))`, wrapping before signed truncating division: `d=trunc(s/2)`. Replace exactly d=0 with 1. Set denominator=`low32(speed<<4)`. If denominator=0 set q=0; otherwise divide unsigned `low32(d<<16)` by unsigned denominator, floor, giving u32 q. Store lifetime=`sat16`(i32(q)). TotalRange/activationCounter are unchanged. A nonzero denominator is at least 16, so q<=268431360 and this caller cannot produce a negative signed q. Full actual position preparation remains MCI-Q04; supplied-current-position arithmetic is independently available.

**MCI-R22** Base stat-list setup.
- **MCI-R22a** [verified in 1.13c code (static)]. With no existing base list and successful allocations, create an empty extended base list associated with missile type/id and pool, empty own/resolved entries, empty attached lists, zero active/changed state membership, zero callback/context and base condition 0. State storage has enough words for the supplied state count. Allocation failure is unchecked here and has no defined null fallback.
- **MCI-R22b** [unresolved]. With an existing base list, setup can request cleanup before replacing it. Complete cleanup/propagation/callback effects are not established (MCI-Q13). This does not block R22a's fresh-empty branch.

**MCI-R23** Callback/skill/level [verified in 1.13c code (static)]. After base-list setup, invoke a present callback once with missile and callbackArg; ignore its return. Then store skill clamped to 0..32767, and store level's low16 unchanged. Reading that level sign-extends i16, so 40000 reads -25536 and 70000 reads 4464. Callback-owned effects/draws remain with its caller; later writes can overwrite callback changes to these fields.

**MCI-R24** Damage-setup requests [verified in 1.13c code (static)]. Request owner weapon-stat activation with selection condition 0, then the MDP damage builder with MDP source=owner, MDP parent=request source, new missile and original request level; then the MDP store with owner, missile, its result and the original request level. Consume [missile-damage-parameters](missile-damage-parameters.md) at its supplied-input/request boundaries. Complete equipment preparation, builder/storer integration and stored stat effects retain MCI-Q06 and the exact MDP gates; this rule guarantees caller arguments/order, not a complete damage or draw result.

**MCI-R25** Owner attribution.
- **MCI-R25a** [verified in 1.13c code (static)]. Store direct owner type/id. Without a base list, set owner-link-valid and finish attribution. With a base list and stateCount > 98, request state-98 active/changed bookkeeping and update marking; with extended state storage the membership bit is set and changed membership set only if newly active. Look up state-98 list (primary before secondary); if absent request allocation associated with missile type/id, state 98 and attachment with full application. Allocation failure returns from attribution before owner-link-valid is set. A present/successfully allocated list receives ordered layer-0 list-stat requests 353=owner.type, 354=owner.id, then owner-link-valid is set. Even when stateCount <= 98, list lookup/allocation/write requests still run when base list exists. The creator then clears creation condition D regardless of that early attribution return. Disguise and update bookkeeping consume STL-R07a/NRF-Q17.
- **MCI-R25b** [unresolved]. Complete list attachment and resulting stored source stats are STL-R06b/c; source disguise/queue integration also remains gated. Direct link fields and local request order are separately available under R25a (MCI-Q13).

**MCI-R26** Pierce count.
- **MCI-R26a** [verified in 1.13c code (static)]. Require player/monster owner, missile recipient and Pierce true. Set c to wrapped sum of owner aggregate stats 156 and 166. If c=0, draw nothing and make no count-write request. Otherwise initialize a local seed from owner's own-base stat 328 with high word 666. Start n=0; while n<4, step the local seed (NRF-R01), take unsigned lo modulo 100, and stop if the signed remainder >= c; otherwise increment n. Request missile stat 328=n, layer 0. There are min(n+1,4) draws: negative c consumes one failed draw; c>=100 consumes four passing draws. The owner's stat/seed, game seed and missile seed are unchanged by this helper. Equal supplied seed value/chance repeat the same count (MCI-Q02 does not gate this consumer).
- **MCI-R26b** [verified in 1.13c code (static)]. The chance components use aggregate/current unit reads and seed input uses own-base read, via NRF-R12b/c. Only a replacing unit-stat request is established for the resulting count; actual stored effects remain NRF-Q20/STL-R06c (MCI-Q13).

**MCI-R27** Request stats/condition/return [verified in 1.13c code (static)]. After pierce, flag 0x1000 requests stat 19=toHit, layer 0. Flag 0x10000 sets missile metadata condition 2 while retaining other metadata conditions. Then invoke the monster-owner wrapper (R28). Finally request stat 327=DamageRate, layer 0, and return the missile. Complete stat-write propagation remains NRF-Q20; logged requests and the direct metadata write are independent.

**MCI-R28** Monster-mod event.
- **MCI-R28a** [verified in 1.13c code (static)]. Only a monster owner and present missile request the monster-mod operation with game, owner, missile and event 5. Players skip it. It runs between the optional metadata condition write and the final damage-rate request.
- **MCI-R28b** [unresolved]. The event-5 dispatcher selects callbacks from up to nine owner mod bytes. Which selected callbacks act, their effects and seed owners/counts require complete handler verification (MCI-Q07). Dispatcher byte decoding alone does not establish those effects.

**MCI-R29** Client notification [unresolved]. Later synchronization reaches a missile sender that requires a valid direct owner link, resolvable owner, valid missile record and ClientSend true before requesting missile-creation message. The creator only requests update marking. Actual recipient/cadence/eligibility and delivery remain MCI-Q08; no unconditional creation-message assertion follows from the mark.

**MCI-R30** RNG accounting [verified in 1.13c code (static)]. Before the unit seeder, the creator's own validations, room selection, coordinate/speed arithmetic and allocation request make no draws. At the seeder, R10 takes one game-seed step. The creator's other directly verified stages make no draws except R26a's 0..4 local steps. Callback, complete damage preparation, actual path building and monster-mod handlers retain their own draw budgets/gates; adding these known steps never proves a whole-creation total. Light/RandStart/HitShift are not directly read by the creator or the specifically checked non-damage setup helpers. Damage-owned HitShift use remains MDP, and unread helper interiors are not covered by that absence claim.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MCI-R01 | Validation | verified in 1.13c code (static) | none | MCI-V01 |
| MCI-R02 | Start | verified in 1.13c code (static) | none | MCI-V02 |
| MCI-R03 | Room selection | verified in 1.13c code (static) | Q11: candidate production only; supplied order available | MCI-V13 |
| MCI-R04 | Retained point | verified in 1.13c code (static) | none | MCI-V02 |
| MCI-R05 | Initial speed | verified in 1.13c code (static) | none | MCI-V03, MCI-V19 |
| MCI-R06 | Own-list slow scale | verified in 1.13c code (static) | none | MCI-V04 |
| MCI-R07 | Global scale | verified in 1.13c code (static) | none | MCI-V03, MCI-V04, MCI-V19 |
| MCI-R08 | Aim/distance | verified in 1.13c code (static) | none | MCI-V05, MCI-V19 |
| MCI-R09 | Allocation/base/act | verified in 1.13c code (static) | none | MCI-V06, MCI-V20, MCI-V21 |
| MCI-R10 | Game step/child seed | verified in 1.13c code (static) | none | MCI-V06 |
| MCI-R11 | Identity/mode | verified in 1.13c code (static) | none | MCI-V06 |
| MCI-R12a | Metadata/id | verified in 1.13c code (static) | none | MCI-V06, MCI-V22 |
| MCI-R12b | Fresh creation event | verified in 1.13c code (static) | Q03: subsequent lifecycle/nonfresh chain only | MCI-V14 |
| MCI-R13a | Placement requests | verified in 1.13c code (static) | Q11/NRF-Q17: actual path/room/update effects | MCI-V14 |
| MCI-R13b | Descending registry insert | verified in 1.13c code (static) | none | MCI-V15 |
| MCI-R14 | Range/lifetime/activation | verified in 1.13c code (static) | none | MCI-V07, MCI-V08 |
| MCI-R15 | Missing path | verified in 1.13c code (static) | none | MCI-V09 |
| MCI-R16 | Path setup | verified in 1.13c code (static) | Q11: collision-map integration; Q12: selectors >8 | MCI-V10 |
| MCI-R17 | Destroy condition | verified in 1.13c code (static) | none | MCI-V10, MCI-V22 |
| MCI-R18a | Moving speed | verified in 1.13c code (static) | none | MCI-V03 |
| MCI-R18b | Planner request | verified in 1.13c code (static) | Q04: planner-produced path/draws | MCI-V14 |
| MCI-R19 | Source metadata | verified in 1.13c code (static) | none | MCI-V10 |
| MCI-R20 | Acceleration/max speed | verified in 1.13c code (static) | none | MCI-V03 |
| MCI-R21 | Point-distance lifetime | verified in 1.13c code (static) | Q04: actual current-position production only | MCI-V16 |
| MCI-R22a | Fresh base list | verified in 1.13c code (static) | none | MCI-V17 |
| MCI-R22b | Existing-list cleanup | unresolved | Q13: cleanup effects | none |
| MCI-R23 | Callback/metadata | verified in 1.13c code (static) | none | MCI-V11 |
| MCI-R24 | Damage requests | verified in 1.13c code (static) | Q06: damage/equipment integration and full draws | MCI-V18 |
| MCI-R25a | Attribution fields/requests | verified in 1.13c code (static) | Q13/STL-R07a/NRF-Q17: attachment/write/disguise/queue effects only | MCI-V18 |
| MCI-R25b | Attribution stored effects | unresolved | Q13: stored source-list effects | none |
| MCI-R26a | Local pierce draws | verified in 1.13c code (static) | none | MCI-V12 |
| MCI-R26b | Pierce reads/write request | verified in 1.13c code (static) | Q13/NRF-Q20: actual stored count | MCI-V12 |
| MCI-R27 | Stat requests/condition/return | verified in 1.13c code (static) | Q13/NRF-Q20: actual stored stats | MCI-V18 |
| MCI-R28a | Monster-event request | verified in 1.13c code (static) | Q07: handler execution only | MCI-V18 |
| MCI-R28b | Mod handler effects | unresolved | Q07: complete handlers | none |
| MCI-R29 | Later client notification | unresolved | Q08: synchronization/delivery | none |
| MCI-R30 | Bounded draw accounting | verified in 1.13c code (static) | Q04/Q06/Q07: aggregate draw total; supplied callback budget external | MCI-V01, MCI-V06, MCI-V12 |

## 5. Outputs and side effects

| Output / effect | Rule |
|---|---|
| Return absent for early refusal, allocation absence or later missing path; return missile after final stat request | MCI-R01/R02/R03/R08/R09/R15/R27 |
| Fatal duplicate-identity assertion; unchecked dependency allocation/fault outcomes have no null fallback | MCI-R12a/R13b/R22a |
| Allocated missile, type/class/game/pool, act/environment, expansion/creation conditions, auxiliary empty, mode, id/counter | MCI-R09/R11/R12a/R17/R25a |
| Game seed step, missile initialLo and seed writes | MCI-R10 |
| Metadata allocation; source id sentinel; range/lifetime/activation, skill/level, source type/id and condition 2 | MCI-R12a/R14/R19/R21/R23/R27 |
| Animation accumulator when elapsed flag set | MCI-R14 |
| Fresh creation-event association/group insertion; registry insertion | MCI-R12b/R13b |
| Path allocation/planner/update/collision-footprint requests; direct speed/reset/target/mask/acceleration/max-speed writes | MCI-R13a/R16/R18/R20 |
| Fresh base-list state and storage; existing-base cleanup request | MCI-R22 |
| Callback invocation, original request level to damage setup, owner weapon activation and MDP build/store requests | MCI-R23/R24 |
| Direct owner type/id, owner-valid condition; state membership/update/list allocation/attach/write requests | MCI-R25 |
| Local pierce seed/count; ordered stat 328/19/327 requests; metadata condition write | MCI-R26/R27 |
| Monster event-5 request and possible handler effects; later possible missile-creation message | MCI-R28/R29 |

No save/persistence operation is directly requested. Actual stat/attachment, path/room/update and synchronization effects retain the gates named above.

## 6. Branches and edge cases

Only players/monsters own requests here. Difficulty is not read; expansion changes only the named base condition in this creator. Source and owner may differ. Zero final speed skips nudge, distance refusal, moving-speed write and planner call, while targetUnit storage, acceleration/max-speed and optional point-distance lifetime still run. A raw override of 1 becomes 0 after scaling.

Range override bypasses loop addition; elapsed subtraction still applies. Negative arithmetic wraps and truncates before clamps. Skill is clamped, level is narrowed without clamp. A targetUnit equal to owner is retained. Owner-position coincidence, rather than start-position coincidence, drops another targetUnit. Distance 99/100 and wrapped INT_MIN differ as R08 specifies.

No path after placement returns absent after seed/id/range effects. A dependency assertion or failed unchecked allocation is not that branch. Count and stat-write requests do not guarantee persisted values. Larger collision selectors and nonfresh event/base cleanup remain gated. Tracking a unit does not redirect the retained point used for optional lifetime calculation.

## 7. Open questions

The [gap file](../plans/gaps/missile-creation-and-initialization.md) records attempted evidence and precise blocked assertions.

- **MCI-Q02** | unresolved; outside-boundary | Upstream production/change of the owner's own-base stat 328 seed value. It does not block deterministic supplied-seed pierce calculation. Close with an identified owner-writer contract; a missile-count decrement is a different write.
- **MCI-Q03** | unresolved; integration | Complete persistent creation-event consumption/removal and replacement of a nonempty old chain. Fresh supplied-event association/insertion is available. Close by reading the applicable frame-event consumer/removal path through its effects.
- **MCI-Q04** | unresolved; integration | Actual planner target/path/position mutations and all planner draw sites. Local call order and supplied-current-position lifetime arithmetic remain available. Close with a path/movement owner contract covering the reached missile route.
- **MCI-Q06** | unresolved; integration | Complete owner equipment preparation, MDP builder/store execution and total creation damage draws. Consume MDP's verified arithmetic/request subsets; its producer/stat-lifecycle gates still apply. Close through those owners, without inferring effects from caller arguments.
- **MCI-Q07** | unresolved; integration | Selected event-5 monster-mod callbacks, their writes and seed owners/draw counts. Wrapper guard/order is available. Close by reading selected handlers with their inputs and table selection.
- **MCI-Q08** | unresolved; integration | Actual synchronization scheduling, recipient selection, sender eligibility and delivery. Checked sender call-site bytes do not verify the full caller chain. Close through the synchronization/message owner.
- **MCI-Q09** | unresolved; outside-boundary | Wider readers/meaning of the neutral creation conditions, auxiliary field and selector 11. Local writes are established; no local assertion depends on assigning wider meanings. Close by tracing their relevant consumers.
- **MCI-Q10** | unresolved; observation | No executable vector has a matching gameplay capture. Source/seed/metadata/path effects are observable; confirmation remains pending, not impossible. No gameplay campaign is part of these records.
- **MCI-Q11** | unresolved; integration/outside-boundary | Complete path allocation, collision-room publication and update-queue effects; adjacent-room collection production. Supplied room order/path/room outcomes support the local consumer requests. Close with these owner contracts; successful supplied absence must not stand in for an unchecked allocation fault.
- **MCI-Q12** | unresolved; behavior | Collision mask selection for CollideType >8 and reachable loader/input domain. Values 0..8 have an exact contract; no default is supplied above it. Close with producer bounds or checked original-image behavior for the wider domain.
- **MCI-Q13** | unresolved; integration | Existing-base cleanup, source-list attach/write propagation and actual replacing unit-stat effects (NRF-Q20/STL-R06b/c), including indirect notifications. Fresh empty-base setup and caller requests remain available. Close through the lifecycle owners.

## 8. Test vectors

### 8.1 Executable vectors

Each vector is isolated at its named stage unless explicitly stated. All table values are supplied synthetic records. Unlisted request flags are 0; units and records are valid for the named stage. `no draw` means the local stage, excluding an invoked supplied operation's own budget. Supplied operations append ordered arguments to an initially empty call log; allocation supplies the expressly named success/absence, and supplied direct metadata/path writes persist for the vector. Stat/attachment/update/planner/damage operations log requests only and imply no original stored effect. Fixtures never silently substitute an absent result for a fatal dependency failure. Each case starts anew and inherits only the vector's shared inputs. All vectors have runtime_observed: no.

- **MCI-V01** Early refusals (MCI-R01/R30; boundary: isolated rule)
  - Inputs: count 10, counter 41; otherwise valid player owner/id 1 and present record.
  - Seeds: game `{1,666}`.
  - Draws: per independent case below.
  - Trace: (a)-(e) reach their first failing validation before allocation/seeder; (f) passes all validation conditions.
  - (a)
    - Inputs: owner absent.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (b)
    - Inputs: missileId -1.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (c)
    - Inputs: missileId 10.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (d)
    - Inputs: owner object.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (e)
    - Inputs: record lookup absent.
    - Draws: none.
    - Outputs: absent return; counter 41, game seed {1,666}; no allocation/callback.
  - (f)
    - Inputs: owner player, missileId 1, count 10, record present, counter 41.
    - Draws: none.
    - Outputs: validation succeeds and control continues to R02; counter 41, game seed {1,666}; allocation/callback logs empty at this validation-stage boundary.
  - Discriminates: accepting equality with count or allocating before owner validation changes return/log. Case (f) rejects unconditional refusal of otherwise valid requests.
  - runtime_observed: no

- **MCI-V02** Start and point precedence (MCI-R02/R04; boundary: isolated rule)
  - Inputs: source position (100,200), request target (5,-3). Stage ends before R08.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: select start, then relative before absolute before start-default.
  - (a)
    - Inputs: flags 0.
    - Draws: none.
    - Outputs: start/point (100,200); nonzero-speed R08 would nudge to (101,201).
  - (b)
    - Inputs: flags 0x3; start (10,20).
    - Draws: none.
    - Outputs: start (10,20), point (15,17).
  - (c)
    - Inputs: flags 0x22.
    - Draws: none.
    - Outputs: start (100,200), point (105,197).
  - (d)
    - Inputs: flag 0x20; target (50,60).
    - Draws: none.
    - Outputs: start (100,200), point (50,60); nonzero-speed R08 would refuse its 140-unit Y distance.
  - (e)
    - Inputs: flags 0; source absent.
    - Draws: none.
    - Outputs: absent return.
  - Discriminates: choosing absolute first makes (c) point (5,-3).
  - runtime_observed: no

- **MCI-V03** Speed and maximum overwrite (MCI-R05/R07/R18a/R20; boundary: supplied-input stage)
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
  - Discriminates: scaling maximum or treating Accel as 8.8 changes 7680/-2.
  - runtime_observed: no

- **MCI-V04** Own slow list and signed truncation (MCI-R06/R07; boundary: supplied-input stage)
  - Inputs: initial speed 3072, CanSlow true; unit aggregate stat 161=90.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: (a) 3072*50/100=1536, *75/100=1152. (b) -30720/100=-307, -23025/100=-230. (c)/(d) skip first scale, 3072*75/100=2304. (e) 0 remains 0.
  - (a)
    - Inputs: active state 87, matching own list stat 161=50.
    - Draws: none.
    - Outputs: speed 1152.
  - (b)
    - Inputs: active state 87, matching own list stat 161=-10.
    - Draws: none.
    - Outputs: speed -230.
  - (c)
    - Inputs: state absent.
    - Draws: none.
    - Outputs: speed 2304.
  - (d)
    - Inputs: state active but matching list absent.
    - Draws: none.
    - Outputs: speed 2304.
  - (e)
    - Inputs: state active, matching own list value 0.
    - Draws: none.
    - Outputs: speed 0.
  - Discriminates: aggregate value 90 or floor(-307.2) fails (a)/(b).
  - runtime_observed: no

- **MCI-V05** Aim/gate boundaries (MCI-R08/R16; boundary: supplied-input stage)
  - Inputs: final speed 2304, start (100,100), owner position (100,100), retained point (100,100), absent targetUnit; stage includes target selection of R16 only, not its collision operations.
  - Seeds: game `{1,666}`; this stage never reaches seeder.
  - Draws: per independent case below.
  - Trace: nudge absent/owner-coincident target, then compare the selected gate coordinates per axis.
  - (a)
    - Inputs: shared defaults.
    - Draws: none.
    - Outputs: point target (101,101), no tracking, pass.
  - (b)
    - Inputs: target T different from owner at (150,100).
    - Draws: none.
    - Outputs: tracking T, retained point (100,100), gate delta (50,0), pass.
  - (c)
    - Inputs: target T at (200,100).
    - Draws: none.
    - Outputs: absent return, game seed {1,666} unchanged.
  - (d)
    - Inputs: point (120,130), T different from owner at (100,100).
    - Draws: none.
    - Outputs: drop T; point target (121,131), pass.
  - (e)
    - Inputs: speed 0, T at (400,400).
    - Draws: none.
    - Outputs: tracking T, retained point (100,100), no nudge/gate, pass.
  - (f)
    - Inputs: T at (199,100).
    - Draws: none.
    - Outputs: tracking T, retained point (100,100), delta (99,0), pass.
  - (g)
    - Inputs: start (90,90), point (120,130), T different from owner at (100,100).
    - Draws: none.
    - Outputs: drop T; point target (121,131), pass.
  - (h)
    - Inputs: targetUnit is owner at (100,100).
    - Draws: none.
    - Outputs: keep tracking owner, retained point (100,100), pass.
  - (i)
    - Inputs: speed 2304, start (100,100), retained point (200,100), targetUnit absent; owner at (100,100).
    - Draws: none.
    - Outputs: no nudge; gate delta (100,0); absent return, game seed {1,666} unchanged.
  - Discriminates: 100-inclusive distance, start-position coincidence or dropping the owner changes (c)/(g)/(h). Case (i) rejects a distance gate incorrectly restricted to tracking-unit targets.
  - runtime_observed: no

- **MCI-V06** Allocation, seeding and identity (MCI-R09/R10/R11/R12a/R30; boundary: supplied-input stage)
  - Inputs: successful zeroed unit/metadata allocation, expansion true, CollideType 3, room level 80, no later stages.
  - Seeds: game `{1,666}`, child initially zeroed.
  - Draws: per independent case below.
  - Trace: child first `{1,666}`; game product plus 666 gives 1791398751; final child `{1791398751,666}`; room 80 selects act 2; counter increment skips exactly all-ones.
  - (a)
    - Inputs: counter 41.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 42, mode 3, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - (b)
    - Inputs: counter 0xFFFFFFFE.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 1, mode 3, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - (c)
    - Inputs: counter 0xFFFFFFFF.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 0, mode 3, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - (d)
    - Inputs: counter 41, CollideType 8; shared successful allocations and room level 80.
    - Draws: one game step to {0x6AC6935F,0}; no child step.
    - Outputs: game {0x6AC6935F,0}; child {0x6AC6935F,666}, initialLo 0x6AC6935F; type missile, id/counter 42, mode 8, act 2, expansion/A/B true, selector 11, auxiliary empty, C/D false, source id -1.
  - Discriminates: stepping child, seeding before successful allocation or skipping zero changes exact outputs. Case (d) rejects a fixed mode 3 while retaining the existing counter boundary cases.
  - runtime_observed: no

- **MCI-V07** Range/loops/elapsed/activation (MCI-R14; boundary: isolated rule)
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

- **MCI-V08** Independent clamps (MCI-R14; boundary: isolated rule)
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

- **MCI-V09** Late missing path (MCI-R15; boundary: supplied-input stage)
  - Inputs: post-R14 allocated missile id 42, counter 42, supplied registration completed, path absent, totalRange/lifetime 30, activation 27; unit not released.
  - Seeds: game already `{0x6AC6935F,0}` from one reached seeder; child `{0x6AC6935F,666}`.
  - Draws: none in this stage; earlier game step is input state.
  - Trace: path test returns absent without rollback.
  - Outputs: absent return; unit/counter/seeds/range fields unchanged; no release, callback, planner or damage request.
  - Discriminates: rollback or continuing to callback changes state/log.
  - runtime_observed: no

- **MCI-V10** Path targets/collision/source (MCI-R16/R17/R19; boundary: supplied-input stage)
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

- **MCI-V11** Callback and narrow metadata (MCI-R23; boundary: supplied-input stage)
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

- **MCI-V12** Local pierce seed/count requests (MCI-R26a/R26b/R30; boundary: supplied-input stage)
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

- **MCI-V13** Owner room first, ordered neighbors (MCI-R03; boundary: supplied-input stage)
  - Inputs: owner room A rectangle (0,0,10,10); ordered candidates [absent,B,C], B/C both rectangle (10,0,10,10).
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: half-open upper bound excludes 10 from A; visit candidates in given order.
  - (a)
    - Inputs: point (9,9).
    - Draws: none.
    - Outputs: select A.
  - (b)
    - Inputs: point (10,5).
    - Draws: none.
    - Outputs: select B.
  - (c)
    - Inputs: point (20,5).
    - Draws: none.
    - Outputs: absent creation; no allocation/seeder request.
  - (d)
    - Inputs: owner room absent; point (9,9).
    - Draws: none.
    - Outputs: absent creation; no allocation/seeder request.
  - Discriminates: inclusive bound chooses A in (b); sorting/reversing matching candidates may choose C.
  - runtime_observed: no

- **MCI-V14** Registration and movement request order (MCI-R12b/R13a/R18b; boundary: supplied-input stage)
  - Inputs: fresh missile M id 42 with empty event chain, supplied event E, existing missile-group entries[F], pool P/room A/start(10,20), supplied path H and room A, act 2/environment E2, final speed 2304, path speed/max 0/reset selector 2. Placement insertion is supplied/logged; planner returns 0 and makes no supplied mutations.
  - Seeds: no draw locally; supplied operations no draw.
  - Draws: none.
  - Trace: initialize creation entry, place/reassociate, allocate path then register/update, write speed then call planner; ignore planner 0.
  - (a)
    - Inputs: final speed 2304 and shared placement/event inputs.
    - Draws: none.
    - Outputs: E targets M/type 0/frame -1/persistent true/no override/parameters 0,0/auxiliary 0; unit events [E], group [E,F], act 2/environment E2. Exact request log [path(P,A,10,20,M,0), registry(M), update(M), planner(H,M,0)]; path speed/max 2304/reset 15. Control continues to R19 despite planner result 0.
  - (b)
    - Inputs: final speed 0 and shared placement/event inputs.
    - Draws: none.
    - Outputs: E targets M/type 0/frame -1/persistent true/no override/parameters 0,0/auxiliary 0; unit events [E], group [E,F], act 2/environment E2. Exact request log [path(P,A,10,20,M,0), registry(M), update(M)]; path speed/max 0/reset 2. Control continues to R19, without a moving setter or planner call.
  - (c)
    - Inputs: final speed 0; supplied path H is present but the post-placement unit-room test returns absent; other shared placement/event inputs apply.
    - Draws: none locally; supplied operations no draw.
    - Outputs: E targets M/type 0/frame -1/persistent true/no override/parameters 0,0/auxiliary 0; unit events [E], group [E,F], act 2/environment E2. Exact request log [path(P,A,10,20,M,0), registry(M)]; no update or planner request; path speed/max 0/reset 2. Control continues to R19. This asserts the supplied room-test branch, not an original allocator/room-publication outcome.
  - Discriminates: treating planner 0 as creation failure or calling planner at speed 0 fails log/continuation assertion; original planner effects remain gated. Case (c) rejects update marking without a post-placement room.
  - runtime_observed: no

- **MCI-V15** Descending registry insertion (MCI-R13b; boundary: supplied-input stage)
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

- **MCI-V16** Retained-point duration and width (MCI-R21; boundary: supplied-current-position stage)
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

- **MCI-V17** Fresh extended base (MCI-R22a; boundary: supplied-allocation stage)
  - Inputs: no base list, successful list/storage allocations, missile type 3/id 42/pool P, stateCount 99, setup conditions/callback/context 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: zero fresh base and state storage; associate missile; required state words ceil(99/32)=4 per active/changed collection.
  - Outputs: base associated(type 3,id 42,pool P), extended true/base condition 0, own/resolved/primary/secondary empty, active/changed state words[0,0,0,0], callback/context 0; no cleanup request.
  - Discriminates: retaining stale membership or allocating only 3 words fails exact initialized state.
  - runtime_observed: no

- **MCI-V18** Late ordered requests (MCI-R24/R25a/R27/R28a; boundary: supplied-input stage)
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

- **MCI-V19** Signed/wrapped speed and zero gate (MCI-R05/R07/R08; boundary: isolated stages)
  - Inputs: arithmetic stages only; valid records/units as needed.
  - Seeds: no draw.
  - Draws: per independent case below.
  - Trace: (a) trunc(-20/8)=-2; 8<<8=2048;153600/100=1536. (b) 75/100=0. (c) low32(2147483647*75)=2147483573; /100=21474835. (d) negating INT_MIN retains INT_MIN, signed comparison>=100 false; Y0 passes.
  - (a)
    - Inputs: Vel 10/VelLev 4/level -5/no override.
    - Draws: none.
    - Outputs: speed 1536.
  - (b)
    - Inputs: flags 0x14/override 1; start (0,0), point (400,400), targetUnit absent.
    - Draws: none.
    - Outputs: speed 0, R08 skipped; retained point (400,400), no refusal.
  - (c)
    - Inputs: flags 0x14/override 2147483647.
    - Draws: none.
    - Outputs: speed 21474835.
  - (d)
    - Inputs: final speed 1/start (0,0)/point (-2147483648,0)/targetUnit absent.
    - Draws: none.
    - Outputs: distance gate passes; retained point (-2147483648,0).
  - Discriminates: floor negative division, widened product or mathematical absolute value changes output.
  - runtime_observed: no

- **MCI-V20** Allocation failure seed boundary (MCI-R09; boundary: supplied-input stage)
  - Inputs: all validations/room/aim pass, unit allocation returns absent, game counter 41.
  - Seeds: game `{1,666}`.
  - Draws: none.
  - Trace: a supplied absent allocator return reaches the creator guard before base writes/seeder; it does not verify safe null-return behavior from the unchecked original allocator.
  - Outputs: absent return, counter 41/seed `{1,666}`, no metadata/event/path/callback requests.
  - Discriminates: seeding or incrementing before allocation changes seed/counter.
  - runtime_observed: no

- **MCI-V21** Act-selection boundaries (MCI-R09; boundary: isolated act-selection stage)
  - Inputs: supplied room level and corresponding game act environments E0..E4. This stage asserts only the act/environment association, with successful base allocation supplied by its caller.
  - Seeds: no draw.
  - Draws: none.
  - Trace: test signed level against 40,75,103,109,1024 in order; select the matching act and that same-index environment.
  - (a)
    - Inputs: room level 39.
    - Draws: none.
    - Outputs: act 0, environment E0.
  - (b)
    - Inputs: room level 40.
    - Draws: none.
    - Outputs: act 1, environment E1.
  - (c)
    - Inputs: room level 74.
    - Draws: none.
    - Outputs: act 1, environment E1.
  - (d)
    - Inputs: room level 75.
    - Draws: none.
    - Outputs: act 2, environment E2.
  - (e)
    - Inputs: room level 102.
    - Draws: none.
    - Outputs: act 2, environment E2.
  - (f)
    - Inputs: room level 103.
    - Draws: none.
    - Outputs: act 3, environment E3.
  - (g)
    - Inputs: room level 108.
    - Draws: none.
    - Outputs: act 3, environment E3.
  - (h)
    - Inputs: room level 109.
    - Draws: none.
    - Outputs: act 4, environment E4.
  - (i)
    - Inputs: room level 1023.
    - Draws: none.
    - Outputs: act 4, environment E4.
  - (j)
    - Inputs: room level 1024.
    - Draws: none.
    - Outputs: act 0, environment E0.
  - (k)
    - Inputs: room level -1.
    - Draws: none.
    - Outputs: act 0, environment E0.
  - Discriminates: every threshold's strict comparison and the high-level return to act 0; clamping all high levels to act 4 fails case (j).
  - runtime_observed: no

- **MCI-V22** Existing metadata and false destroy condition (MCI-R12a/R17; boundary: isolated supplied-metadata stages)
  - Inputs: existing metadata N with source type 0/id 55, skill 9, lifetime 17 and all other metadata fields 0; supplied missile id 42, creation conditions C/D true; CanDestroy false. This fixture enters R12a and then R17 with those supplied fields; intervening creator stages are outside its boundary.
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

### 8.2 Observation vectors

- **MCI-O01** Initial speed/max/acceleration (MCI-R05/R07/R20): capture one returned missile's supplied row and level plus final path values; expected arithmetic is MCI-V03/R06 when applicable. [Procedure](../research/spec-evidence/missile-creation-and-initialization.md#mci-o01).
- **MCI-O02** Seeder boundary (MCI-R10/R30): record game seed immediately around the seeder and initial child writes, then account separately for downstream draws. Expect NRF-R03a's one game step and unstepped child seed. [Procedure](../research/spec-evidence/missile-creation-and-initialization.md#mci-o02).
- **MCI-O03** Range/loops/elapsed (MCI-R14): capture request and row values and metadata at creation; expect the exact independent clamp arithmetic in MCI-V07/V08. [Procedure](../research/spec-evidence/missile-creation-and-initialization.md#mci-o03).
- **MCI-O04** Pierce producer/consumer (MCI-R26/MCI-Q02): read owner own-base 328 and chance before successive shots, local draw/count requests and resulting missile stat; expect deterministic helper results for each captured input. [Procedure](../research/spec-evidence/missile-creation-and-initialization.md#mci-o04).
- **MCI-O05** Distance rejection (MCI-R08): find a naturally reachable request with a 100-axis delta and compare 99 under equal relevant inputs. Expect early refusal at 100 and acceptance through this gate at 99. [Procedure](../research/spec-evidence/missile-creation-and-initialization.md#mci-o05).

## 9. Acceptance and dependencies

Partial acceptance consists of MCI-V01..V22's exact local arithmetic, guards, fields, returns and ordered requests at their supplied-input boundaries, including the explicit fatal/unchecked-failure distinctions. Verified-rule rows name their vectors; gated source-list/stat, path, event lifecycle, damage, mod and synchronization effects are not inferred from these passing requests. Full behavior reproduction additionally requires MCI-Q03/Q04/Q06/Q07/Q08/Q11/Q12/Q13's in-scope integration/assertions and applicable original-game observations, including MCI-O01/O02. MCI-Q02/Q09 are upstream/wider meaning questions and impose no local supplied-input gate; MCI-Q10 blocks runtime-confirmation claims. A partial pass never establishes complete creation parity.

**Depends on:** [foundations](core-numeric-and-rng-foundations.md) NRF-R01/R03a/b/R17a for exact seeds, NRF-R06a for percent format, NRF-R08b for speed-setter behavior, NRF-R10b for dynamic coordinates, NRF-R12a/b/c for existing state/stat reads and NRF-Q17/Q20 for update/stat integration; [game data](core-game-data-fields.md) GDF-R01/GDF-R07/GDF-R09/GDF-R11a/GDF-R12 for schemas and identity conventions; [state lifecycle](state-timed-state-lifecycle.md) STL-R01 for supplied successful frame-event allocation, STL-R06a for list allocation and STL-R06b/c/R07/R07a for the exact attachment/write/membership limits; [missile damage parameters](missile-damage-parameters.md) MDP-R01..R12 at its supplied-input/request boundaries, especially MDP-R07b/R08b/R11b/R12b gates. Wider planner/room producers require their named open integration contracts.

**Depended on by:** [missile movement](missile-movement-and-collision.md) for initialized speed/lifetime/activation/collision; [missile contact](missile-hit-and-damage-calculation.md) for count and attribution; [missile damage parameters](missile-damage-parameters.md) for stored skill level/context; [PvP attribution](combat-pvp-damage-and-kill-credit.md) PVP-R07 for direct owner link; creating skill contracts including [Blade skills](skill-assassin-blade-sentinel-and-fury.md). These consumers do not establish the upstream operations here.

## 10. Related specs

- [missile-movement-and-collision](missile-movement-and-collision.md): later planner/movement/countdown/collision ownership.
- [missile-damage-parameters](missile-damage-parameters.md): damage initialization, element selectors and damage-owned draws.
- [missile-hit-and-damage-calculation](missile-hit-and-damage-calculation.md): pierce count use and hit processing.
- [combat-pvp-damage-and-kill-credit](combat-pvp-damage-and-kill-credit.md): direct owner-link resolution at impact.
- [state-movement-modifiers-and-freeze](state-movement-modifiers-and-freeze.md): separate movement-state operations; it does not define this slow-list producer.
- [skill-assassin-blade-sentinel-and-fury](skill-assassin-blade-sentinel-and-fury.md): concrete missile caller and caller-owned damage/RNG boundaries.

## Appendix A. Provenance

Original LoD Expansion 1.13c Windows x86, identified by [source manifest](../research/source-manifest.md) artifacts LOD113C-D2GAME and LOD113C-D2COMMON. Compatible 2026-10-12 static readings are retained at their recorded boundaries. Targeted 2026-10-15 source readings and original-image decoding establish room selection, registry order, distance arithmetic and local request/failure distinctions; named text-cell records reuse complete 2026-10-15 loader payloads with profile limits. No gameplay vector is confirmed. Detailed original artifact/function locators, corrections, evidence assessment, observation procedures and review status are in the [companion](../research/spec-evidence/missile-creation-and-initialization.md); remaining evidence is in the [gap file](../plans/gaps/missile-creation-and-initialization.md).
