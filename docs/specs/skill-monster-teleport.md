# Monster teleport

**Description**

Monsters use teleport skills to change position immediately. Direct teleport chooses a destination from the monster's stored target point; imp teleport can instead move onto another unit or leave an attachment. The server also marks the monster for an update and changes attachment state, with collision, AI and complete synchronization limits stated below.

Status: draft
Evidence build: Lord of Destruction Expansion 1.13c
Foundations: [core-numeric-and-rng-foundations](core-numeric-and-rng-foundations.md), [core-record-layouts-and-ids](core-record-layouts-and-ids.md); names: [core-game-data-table-records](core-game-data-table-records.md)
Ids: MTP-R, MTP-Q, MTP-V, MTP-O; policy in [README conventions](README.md#conventions)
Evidence: [companion investigation records](../research/spec-evidence/skill-monster-teleport.md)
Open evidence gaps: [gap file](../plans/gaps/skill-monster-teleport.md)

## 1. Scope

This contract covers server action selectors 98 and 129: destination/target selection, local relocation requests and return handling, the reached same-room position-commit stage, update collection, and attachment/cleanup decisions and requests. It consumes skill dispatch rather than choosing when AI should teleport. Collision search, general room transfer, complete link-stat propagation, AI execution, death execution and client delivery retain their named integration boundaries. A supplied operation result proves its caller's decisions, not that operation's unverified effects.

## 2. Trigger and ownership

D2Game owns both callbacks. [SCD-R06](skill-cast-start-and-do-dispatch.md) selects the action callback from the prepared Skills `srvdofunc` selector. The callback runs synchronously within that action; no independent periodic cadence is established. Start/mode/AI eligibility remains the dispatch owner's boundary, including SCD-Q04 for complete monster triggering. Teleport synchronization depends on the relocation operation; this contract establishes no named client message.

## 3. Inputs and data

| Input | Semantic source | Width and format | Meaning |
|---|---|---|---|
| game, caster, skill identity, level | action invocation | game/unit references; i32 identity/level | Callback context |
| skill count and prepared record | loaded Skills collection; GDA-R11a | i32 count; record or absent | Target-aware route validates identity before further work |
| teleport state selector | prepared `Skills.aurastate` | signed 16-bit linked state identity | Target-aware route validates the field; attachment helpers separately use code-fixed state 143 |
| state count | loaded States collection | i32 | Exclusive upper bound for the teleport state selector |
| selected target and path target point | caster targeting/path operations | unit or absent; i32 coordinates | Potential destination |
| current room, destination room, attachment relation | room and attachment operations | references or absent | Routing inputs whose full production is MTP-Q02 |
| relocation result | relocation operation | signed i32; zero/nonzero | Outcome used by the reached local branch |
| registry and stored relationship key | supplied game's unit registry; unit relationship-present condition and stored counterpart kind/id | kinds i32 in 0..5, ids u32; unit reference or absent | Resolve by kind/id only when the relationship condition is set |
| target cache and key | caster path's cached target reference and stored target kind/id | reference or absent; i32/u32 | Clearing a stale cache preserves the saved key |
| positions and room candidates | dynamic current coordinates, unsigned 16-bit stored target X/Y; static unit positions; room rectangles and ordered adjacent rooms | current dynamic 16.16; other coordinates i32; rectangle sums wrap i32 | Position/room operations below |
| target life eligibility | unit kind, mode and alive-exclusion condition | kind/mode i32; boolean | Players in modes 0/17 and monsters in modes 0/12 fail |
| nearby-position operation | room, point, collision footprint, collision mask and search controls | footprint/coordinates i32; mask u32 | Supplies a possibly adjusted point and room; execution remains MTP-Q02 |
| path commit state | current room, room-update-request condition, room-synchronization condition, containing-room guard, movement condition, two progression counters and two auxiliary reset values | booleans, references, i32/u32 values | The two room conditions are distinct; MTP-R11's reached suffix and MTP-R10's synchronization guard use different conditions |
| update collection | unit queued condition/next reference; current room's suppression condition, changed marker and ordered update chain | booleans; unit references | MTP-R13's actual collection writes |
| monster base family | `MonStats.BaseId` through the base-family getter; GDA-R08a | signed 16-bit family; i32 loaded monster count | Helpers require caster family 492 and counterpart family 435 or 441; these are fixed code comparisons |
| attachment state and interaction conditions | active state 143, state storage/count; caster interaction conditions A/B/C | state identity i32; booleans | State calls consume RLI-R03/STL-R07; A/B/C name only the three locally cleared/restored conditions |
| event state | caster's ordered unit-event chain and monster persistent-event queue | logical events with kind, key, parameters, frame and override | Type-5/key matching and insertion below; processing cadence remains dependency-owned |
| health percentage inputs | supplied stat 6 `hitpoints` and stat 7 `maxhp` getter results | signed i32 raw 8.8 health | MTP-R14 discards fractional health before division |
| dependency results | link/unlink, event allocation/removal, AI change, death and room/client synchronization operations | supplied normal-return operations and logical results | Logs distinguish exact caller requests from dependency-produced effects |

| Mapping or constant | Origin | Binding |
|---|---|---|
| `MonTeleport.srvdofunc = 98`; `Imp Teleport.srvdofunc = 129` | effective `data/global/excel/skills.txt`, named rows and column; companion table record T01 | Data selects engine-defined handlers; the named selector fields have narrow compiled corroboration C01 |
| `Imp Teleport.aurastate = attached`; `States.attached.id = 143` | named cells in effective Skills and States files; T01/T02 and C02 | Linked numeric field 143 corroborated in compiled Skills/States; data resolves the field, while attachment helpers fix 143 independently in code |
| action selectors 98 and 129 | original dispatch-table code/data | Fixed handler association in the original executable |
| attachment families 492, 435, 441; type 5; AI selectors 16/0; health threshold 10; collision mask 7177; scan limit 50 | original engine code | Fixed by the identified consumers; these do not establish custom remapping |

## 4. Rules (execution order)

**MTP-R01** Named selectors [verified in 1.13c data]. The named text cells in section 3 associate MonTeleport with action selector 98 and Imp Teleport with selector 129. They do not prove either action executed.

**MTP-R02** Attachment selector identity [verified in 1.13c data]. Imp Teleport's `aurastate` text cell is `attached`; the effective States row `attached` has explicit id 143. The matching reported original profile's compiled skill 330 stores `aurastate = 143`, and compiled state 143 stores identity 143. This verifies those named scalar correspondences, not gameplay execution. The helpers' fixed use of 143 is a separate code binding in MTP-R12.

**MTP-R03** Target-aware validation [verified in 1.13c code (static)]. For selector 129, test signed skill identity first: negative or at least skill count returns 0 without accessing a skill record. Resolve the prepared record and return 0 if that resolution is absent. Otherwise read signed 16-bit `aurastate`. Negative or at least state count returns 0 before any target, relocation or attachment operation. A valid selector reaches the first relationship lookup for `(game,caster)`, then the routing rules below. The validation stage draws no random number and writes no gameplay state. Complete prepared-record storage under GDA-R11a supplies records for valid indices; absent resolution at an in-range index is a synthetic malformed-storage case. A valid `aurastate` is a gate here; it does not replace the helpers' fixed state 143.

**MTP-R04** Direct teleport request/result [verified in 1.13c code (static)]. For a present monster with a dynamic path, read stored target X then Y as unsigned 16-bit values and get the current room. Search that room first, then its adjacent rooms in stored order, skipping absent entries, for the first signed half-open rectangle containing the point. Rectangle upper bounds use i32-wrapped origin+extent. An absent initial room yields no match. If no match, a monster with present monster data, AI control and saved lookup context resolves the control's stored kind/id in that context's registry; a found unit supplies its current room directly, without another containment search in this caller. Otherwise destination room stays absent. The producer/meaning of that saved control context is outside this supplied-input consumer.

Request relocation `(game,caster,point,destinationRoom,skipNearbySearch=false,alternateMarker=false)`. Zero returns 0 without the caller's additional update or teleport-condition write. Nonzero requests update collection (MTP-R13), sets the caster's teleport-changed condition, then returns exactly 1. The relocation operation may have its own effects before either return; zero is not a rollback. This rule draws nothing directly. MTP-R10 owns its local relocation stage; collision/general movement/client integration remains MTP-Q02.

**MTP-R05** Linked-caster route [verified in 1.13c code (static)]. After MTP-R03 succeeds, resolve the caster relationship (MTP-R08), refresh its cached target (MTP-R09), read that cache and treat self as absent. If the saved relationship is present, request coordinate resolution (MTP-R09), then relocation with the produced point, no supplied room, `skipNearbySearch=false` and `alternateMarker=false`, then cleanup `(game,caster,saved counterpart,invocation skill identity)` under MTP-R12. Return exactly 1 after those calls return normally. Both the coordinate resolver's boolean result and the relocation result are ignored; the produced coordinates are still forwarded. Cleanup takes the skill identity, not the invocation level. No direct draw. Complete downstream effects remain MTP-Q02/MTP-Q03.

**MTP-R06** No-target fallback [verified in 1.13c code (static)]. After the relationship and target refresh above, with no caster relationship and no distinct cached target, delegate to selector 98 with the invocation game, caster, identity and level; return its result unchanged. No target-alive test or attachment request occurs on this branch. No direct draw. The direct stage's integration limits remain MTP-Q02.

**MTP-R07** Distinct-target route [verified in 1.13c code (static)]. With no caster relationship and a distinct refreshed target, first resolve that target's relationship; a present result returns 0. Otherwise reject when its alive-exclusion condition is set, its kind is neither player nor monster, or its mode is player 0/17 or monster 0/12. A target with a relationship condition set but an unresolved counterpart passes the relationship test. For a surviving target, obtain X, Y and current room in that order; dynamic X/Y use NRF-R10b, with the local wrappers returning 0 for an absent dynamic path. Request relocation using that point/room with `skipNearbySearch=true` and `alternateMarker=false`. Zero returns 0; nonzero requests attachment `(game,caster,target,invocation identity,invocation level)` under MTP-R12, then returns exactly 1. Attachment can decline its own family checks without changing this final result. No direct draw. Complete relocation/link/AI effects remain MTP-Q02/MTP-Q03.

**MTP-R08** Relationship resolution [verified in 1.13c code (static)]. An absent unit or clear relationship-present condition yields absent without registry lookup. Otherwise resolve its stored counterpart kind/id in the supplied game's registry and return that unit or absent. This consumer does not set a relationship or infer one from state membership. Valid registry kinds are 0..5; malformed stored kinds are outside this rule. No write or draw.

**MTP-R09** Target refresh and coordinates [verified in 1.13c code (static)]. With present game, caster and path, refresh does nothing when the cached target is absent. Otherwise resolve the path's saved target kind/id in the game registry. If the resolved reference differs from the cached reference, clear only the cached reference. Independently, a resolved item in mode 1 or 2 also clears that reference. Clearing preserves the saved kind/id; refresh never replaces a stale pointer with the resolved unit.

Coordinate resolution first performs that refresh, then reads the cache. For a distinct present target, get its X then Y: objects/items/tiles read their prepared static coordinates; other kinds use dynamic integer position parts, or 0 per coordinate when the target path is absent. For absent/self cache, read the caster's stored target X then Y. Write both output coordinates before returning 1 when both are nonzero, otherwise 0. No direct draw. Missing game/caster/path and invalid registry kinds are outside the valid-input stage.

**MTP-R10** Monster relocation local stage [verified in 1.13c code (static)]. With a present monster and dynamic path, retain its old room. Use a supplied destination room, or search the old room and its stored adjacent rooms using MTP-R04's rectangle test. No room returns 0. Unless `skipNearbySearch`, obtain the unit collision footprint and request the nearby-position operation with room, mutable point, footprint, mask 7177, scan limit 50 and fixed search controls 0/1; use its returned room and point. An absent returned room returns 0. The footprint/search production is MTP-Q02, not an asserted collision-success test.

Request the path relocation operation with path, caster, selected room and the low unsigned 16 bits of each point coordinate. Zero returns 0. Nonzero requests update collection, sets the ordinary teleport-changed condition when `alternateMarker=false` or the alternate relocation-changed condition otherwise, requests room/client synchronization, and returns 1. The synchronization helper immediately does nothing for a monster path whose room-synchronization condition is clear. Its other paths remain MTP-Q02. This stage draws nothing directly; dependency effects/draws are separate. The reached same-room path commit is MTP-R11.

**MTP-R11** Reached same-room position commit [verified in 1.13c code (static)]. This suffix begins after collision-footprint operations have returned normally, with a monster dynamic path, selected room equal to its current room, its room-update-request condition clear and its containing-room guard clear. The selected point is within that room. Let X and Y be the unsigned low 16 bits supplied to path relocation. Store centered 16.16 positions `FX=u32((X<<16)+32768)` and `FY=u32((Y<<16)+32768)`. Store projected path components `A=sar(i32((FX>>>11)-(FY>>>11)),1)` and `B=sar(i32((FX>>>11)+(FY>>>11)),2)`. The unsigned shifts precede the signed additions/subtractions. Snap again to the same centered coordinates, clear the moving condition, set both path progression counters to 0 and set the two path auxiliary reset values to 0; return 1. No room transfer, draw or event request occurs in this suffix. It verifies these actual stored fields at this reached boundary; reaching it and collision-map effects remain MTP-Q02. Wider meanings of the auxiliary reset values are not asserted.

**MTP-R12** Attachment and cleanup local decisions [verified in 1.13c code (static)]. Both helpers first require a present caster and counterpart, caster base family 492 within the loaded monster count, and counterpart base family 435 or 441 within that count. Other families return without state/link/event/AI requests. The base-family getter returns -1 for a non-monster; for a monster with a valid prepared MonStats record it reads signed `BaseId`, otherwise its class identity is returned and then checked by the helper's bounds.

Attachment requests, in order: state 143 on for the caster (RLI-R03/STL-R07); link caster to counterpart; link counterpart to caster; remove caster type-5 events whose key equals the invocation skill identity (key 0 removes all type-5 events); allocate/link a new caster event. With a supplied successful event allocation and valid queues, set event kind 5, frame -1, key/first parameter equal to skill identity, second parameter equal to level and no override; insert it at the front of the monster persistent-event queue as well as the unit chain. Clear caster interaction conditions A/B/C, then request its AI-control change with selector 16. Event processing/cadence, complete link propagation and AI execution remain MTP-Q03.

Cleanup requests state 143 off, unlink caster, then unlink counterpart, in that order. Evaluate caster health percentage under MTP-R14. If it is below 10, request caster death with absent attacker and control 1 and stop this helper. Otherwise set interaction conditions A/B/C, remove its type-5 events using the supplied skill identity as key (0 again means all), and request AI-control selector 0. Death execution is MTP-Q03. Neither helper performs a direct draw outside its dependencies.

The link operation's verified immediate prefix stores the counterpart kind/id before any optional state/stat integration; unlink stores zero in both saved fields before that integration. Complete post-return link conditions, state-98/list-stat propagation and allocation failure effects remain MTP-Q03. Supplied normal-return link operations therefore verify the caller's order and arguments, not those missing effects.

**MTP-R13** Update collection [verified in 1.13c code (static)]. With a present unit, get its current room. No room, a room's update-suppression condition, or an already-set unit queued condition makes no change. Otherwise set the unit queued condition, set the room changed marker, set the unit's next-update reference to the room's prior update head and make this unit the new head. Valid room queue storage is required. Repeated collection of an already-queued unit does not move it to the front again. This is collection only, not client delivery. No draw.

**MTP-R14** Cleanup health percentage [verified in 1.13c code (static)]. Obtain raw signed i32 stat 6 then stat 7 through the shared stat reader. For supplied returned values H and M, compute `h=sar(H,8)` and `m=sar(M,8)`. If m=0 return 0. Otherwise return signed `truncTowardZero(i32(h*100)/m)`. No clamp is applied; multiplication wraps before division and numeric-foundation division faults retain their contract. No draw. Actual stat aggregation is consumed from RLI-R02a/RLI-R02b; these supplied values do not verify a missing producer.

### Id map

| Id | Summary | Status label | Gates | Vectors |
|---|---|---|---|---|
| MTP-R01 | Named action selectors | verified in 1.13c data | none for text cells; C01 records narrow compiled-selector corroboration | T01 in companion |
| MTP-R02 | Attachment state identity | verified in 1.13c data | none for the named scalar correspondence | T01, T02, C02 in companion |
| MTP-R03 | Early validation | verified in 1.13c code (static) | none through the first dependency-entry boundary | MTP-V01, MTP-V02 |
| MTP-R04 | Direct request, update and result | verified in 1.13c code (static) | MTP-Q02: actual placement/general synchronization; supplied-result stage available | MTP-V03, MTP-V07 |
| MTP-R05 | Linked-caster call order | verified in 1.13c code (static) | MTP-Q02/MTP-Q03: dependency-produced effects | MTP-V04 |
| MTP-R06 | No-target delegation | verified in 1.13c code (static) | MTP-Q02: actual delegated placement | MTP-V05 |
| MTP-R07 | Distinct-target decisions | verified in 1.13c code (static) | MTP-Q02/MTP-Q03: actual placement/attachment integration | MTP-V05 |
| MTP-R08 | Relationship key lookup | verified in 1.13c code (static) | none for valid registry kinds and supplied registry | MTP-V06 |
| MTP-R09 | Target refresh/coordinates | verified in 1.13c code (static) | none for the stated valid-input stage | MTP-V06 |
| MTP-R10 | Monster relocation requests/result | verified in 1.13c code (static) | MTP-Q02: footprint/search/path/client producers | MTP-V08 |
| MTP-R11 | Reached same-room position writes | verified in 1.13c code (static) | MTP-Q02: reaching suffix/collision/general room transfer only | MTP-V09 |
| MTP-R12 | Attachment/cleanup local effects and requests | verified in 1.13c code (static) | MTP-Q03: link propagation, allocation failure, AI/death/event execution | MTP-V10, MTP-V11, MTP-V13 |
| MTP-R13 | Actual update collection | verified in 1.13c code (static) | none for supplied valid queue storage; delivery is MTP-Q02 | MTP-V03, MTP-V12 |
| MTP-R14 | Health-percentage arithmetic | verified in 1.13c code (static) | actual stat aggregation follows RLI; supplied raw results available | MTP-V11 |

## 5. Outputs and side effects

| Output | Rule | Established boundary |
|---|---|---|
| callback result 0 before dependency calls | MTP-R03 | Verified validation failures |
| first relationship-lookup request reached | MTP-R03 | Verified valid-input continuation only; dependency result and final callback result unasserted |
| relocation request, actual local teleport condition and update collection, result 0/1 | MTP-R04 | Local decisions with supplied relocation result |
| coordinate, relocation and cleanup requests; result 1 | MTP-R05 | Local order/arguments with normal-return dependencies |
| unchanged delegated result | MTP-R06 | Supplied delegated result |
| target relocation/attachment requests and result 0/1 | MTP-R07 | Verified eligibility and local requests |
| resolved relationship or absent; stale target cache cleared while key retained; coordinate outputs and boolean result | MTP-R08, MTP-R09 | Actual lookup/read/cache writes for valid inputs |
| nearby/path/update/synchronization requests, relocation marker, result | MTP-R10 | Local stage; dependency execution is separate |
| centered path position, projected coordinates, movement/counter/auxiliary resets and suffix result 1 | MTP-R11 | Actual reached same-room stores |
| state/link/event/AI/death requests; persistent event fields/insertion and interaction-condition writes | MTP-R12 | Stated successful allocation and normal-return dependency boundary |
| counterpart kind/id prefix stores | MTP-R12 | Immediate link/unlink prefix, before optional stat propagation |
| queued condition, room changed marker and linked update-chain writes | MTP-R13 | Actual collection; no client delivery claim |
| signed health percentage or numeric-foundation fault | MTP-R14 | Supplied stat-reader results |

No persistent-save or named-message claim is established here. Validation failures make no gameplay writes; that claim does not extend to later failures.

## 6. Branches and edge cases

Negative identities and state selectors use signed comparisons; equality with a count fails. A self target is absent after refresh. The linked-caster branch ignores relocation failure and still requests cleanup; the distinct-target branch does not. Refresh clears a stale cached reference without changing its stored key. Coordinates with one zero component still reach the linked-caster relocation request because that caller ignores the resolver result. Room bounds exclude the upper edge and first matching adjacent room wins. Family-ineligible attachment is a local no-op even after a successful target relocation. Type-5 event key 0 is a wildcard. Health percentage discards each raw value's fractional part before division, so the cleanup threshold is not a comparison of raw fractions. Update collection does not duplicate or reorder a unit already queued. No independent difficulty or player-versus-player branch exists in these local decisions. Absent caster/path, malformed registry kinds, collision failures, allocation failures and full AI/client behavior receive no invented safe default.

## 7. Open questions

**MTP-Q01** Resolved in this continuation: original Skills loader descriptors map the consumed signed word to `aurastate`; T02/C02 establish `attached` = 143 in the reported profile. This resolved question no longer gates a rule.

**MTP-Q02** [unresolved] Full collision-footprint search/mutation, general cross-room path transfer, room/client synchronization and any dependency draws remain unverified. Local room search, relationship/target consumers, same-room reached position stores and update collection are available. The saved AI-control lookup-context producer is outside the supplied-input boundary. Missing caster/path/invalid registry-kind behavior is not normalized. The companion identifies exact remaining operations and separately verified slices.

**MTP-Q03** [unresolved] Actual post-return link/stat-list propagation (including state 98), event-allocation failure and persistent-event execution, AI-control changes and death execution remain unverified. Family eligibility, fixed state-143 requests, exact skill-identity event keys, interaction-condition writes and event construction with successful supplied allocation are available. The helper's observable request order is not a claim that all these dependencies execute correctly.

**MTP-Q04** [unresolved] Gameplay teleport confirmation is pending. The 2026-10-29 Act I level-1 character supplied read-only data correspondence but did not encounter an eligible monster teleport. No handler invocation or gameplay effect was captured. Table correspondence does not confirm teleport execution; no gameplay vector is marked observed.

## 8. Test vectors

### 8.1 Executable vectors

These fixtures test only validation and arrival at the first relationship-lookup entry. A record-access observation log records which prepared record/state fields are inspected; a dependency-entry log records the first relationship lookup, without executing it. Both logs start empty in every case. Complete valid storage supplies a present record at identity 0 with state selector 0, which also defines the result a wrong clamp-to-zero implementation would try to use. The malformed-storage case explicitly overrides this default. No case uses another case's result or state.

- **MTP-V01** Signed boundaries reject before target work (MTP-R03; boundary: supplied-input stage)
  - Inputs: selector 129; game G and monster caster C present; level 1; skill count 400; state count 185; complete valid storage except where overridden below. Cases supply independent identity/record/state inputs.
  - Seeds: no draw.
  - Draws: none.
  - Trace: signed identity bounds precede record access; signed state bounds precede the first relationship lookup.
  - Outputs: each case returns exactly 0; dependency-call log empty; gameplay state unchanged.
  - Discriminates: clamping identity -1 to 0 would access the valid record 0 and reach the first relationship lookup; the empty record-access and dependency-entry logs reject that behavior. Equality with either exclusive count must not proceed. Reinterpreting an identity as unsigned and rejecting it with an equivalent unsigned bound is not distinguished by these fixtures.
  - runtime_observed: no.
  - (a) Inputs: identity -1. Outputs: 0; record-access log empty; dependency-entry log empty.
  - (b) Inputs: identity 400. Outputs: 0; record-access log empty; dependency-entry log empty.
  - (c) Inputs: identity 0; synthetic malformed storage resolves its prepared record as absent. Outputs: 0; no state-field read; dependency-entry log empty.
  - (d) Inputs: identity 330; record present; teleport state selector -1. Outputs: 0; state-field observation [(330,-1)]; dependency-entry log empty.
  - (e) Inputs: identity 330; record present; teleport state selector 185. Outputs: 0; state-field observation [(330,185)]; dependency-entry log empty.

- **MTP-V02** Valid selector reaches the first relationship lookup (MTP-R03; boundary: supplied-input stage ending at dependency entry)
  - Inputs: selector 129; game G and monster C present; level 1; skill count 400; complete storage; identity 330; present record; signed teleport state selector 0; state count 185.
  - Seeds: no draw before the dependency entry.
  - Draws: none.
  - Trace: 0<=330<400 and record present; 0<=0<185; reach the first relationship-lookup request for (G,C), then stop this fixture at that entry.
  - Outputs: state-field observation [(330,0)]; dependency-entry log [relationship lookup(G,C)]; gameplay state unchanged at this boundary; no final callback result asserted.
  - Discriminates: an implementation that rejects every input fails to reach the required entry. A valid state value 0 must not be rejected as absent.
  - runtime_observed: no.

### Additional executable fixtures

The following fixtures are independent of MTP-V01/MTP-V02 and of one another. Defaults: game G; monster caster C, kind 1/id 1, present dynamic path and room R; R rectangle (100,200,10,10), no adjacent rooms, valid update storage, update suppression false, changed false and update chain empty; C queued false, teleport/alternate markers false, target cache absent, relationship condition clear. State count is 185, monster count 600, invocation identity 330 and level 7. A registry maps each stated kind/id to the exact stated unit reference, or absent when no entry is listed. Other fixture state is unchanged unless the case supplies it.

The whole-relocation, underlying path-relocation, coordinate-resolution, link/unlink, AI and death operations are supplied normal-return operations wherever a fixture names their results. They log exact semantic arguments and have no unstated effects. Update collection executes MTP-R13 and persists its queue writes. State requests are logged without asserting their full post-return integration. Supplied event removal mutates exactly the matching logical entries; successful event allocation returns a zeroed logical event linked first in the caster's unit chain. These models prove the named caller stage and its actual local writes, not collision, link propagation, death, AI or client execution. Logs below list the selected rule's direct effect requests; internal getter calls are not implicitly added to those logs.

- **MTP-V03** Direct positive result collects update; zero does not (MTP-R04, MTP-R13; boundary: supplied-input stage)
  - Inputs: default C/G/R; stored target point (105,205); current path position (100,200); relocation operation supplies the result specified in each case and makes no position/room mutation.
  - Seeds: no draw in the caller or supplied relocation.
  - Draws: none.
  - Trace: stored point is inside R; request relocation(G,C,(105,205),R,false,false); test returned zero/nonzero before the caller's update and marker.
  - Outputs: independently specified below; no position change is asserted.
  - Discriminates: omitting successful collection/marker, marking on failure, or refusing a successful result violates a case.
  - runtime_observed: no.
  - (a) Inputs: supplied relocation result 1. Outputs: callback 1; direct effect log [relocation(G,C,(105,205),R,false,false), update collection(C)]; C queued true and teleport marker true; R changed true and update chain [C]; C next-update absent.
  - (b) Inputs: supplied relocation result 0. Outputs: callback 0; log [relocation(G,C,(105,205),R,false,false)]; C queued false, teleport marker false; R changed false and update chain empty.

- **MTP-V04** Linked route ignores coordinate/relocation failures and cleans by skill identity (MTP-R05; boundary: supplied-input stage)
  - Inputs: valid selector-129 record with `aurastate = 143`; C relationship condition true and saved key (1,9); registry (1,9) maps H; target cache absent; supplied coordinate resolver writes (0,0) and returns 0; supplied relocation returns 0; cleanup returns normally.
  - Seeds: no draw in the caller or supplied operations.
  - Draws: none.
  - Trace: relationship resolves H; refresh leaves target absent; despite resolver false request relocation at its produced zero point; despite relocation zero request cleanup with identity 330.
  - Outputs: callback 1; direct effect log [coordinate resolution(G,C), relocation(G,C,(0,0),absent,false,false), cleanup(G,C,H,330)]; no caller update collection or teleport-marker write; invocation level 7 is not the cleanup key. Dependency-produced state changes are unasserted.
  - Discriminates: aborting on either zero omits cleanup; passing level 7 instead of identity 330 changes its last argument.
  - runtime_observed: no.

- **MTP-V05** Distinct-target gates, bypass control and no-target delegation (MTP-R06, MTP-R07; boundary: supplied-input stage)
  - Inputs: valid selector-129 record; default C/G/R; target T kind 1/id 9, mode 1, alive-exclusion false, dynamic position (108,208), room R; cache T and saved target key (1,9); registry maps (1,9) to T. T relationship condition is clear. Relocation supplies 1 and attachment returns normally. Each case independently overrides defaults.
  - Seeds: no direct draw or supplied-operation draw.
  - Draws: none.
  - Trace: no caster relationship; refresh preserves valid T; a target without resolved relationship and passing the life test reaches the bypass-nearby-search relocation branch.
  - Outputs: specified per case; no dependency-produced attachment/position effect asserted.
  - Discriminates: using nearby-search false on a distinct target changes the relocation request; testing only a relationship flag would wrongly reject case (c); treating all branches as success fails (b), (d) and (e).
  - runtime_observed: no.
  - (a) Inputs: ordinary defaults. Outputs: callback 1; effect log [relocation(G,C,(108,208),R,true,false), attachment(G,C,T,330,7)].
  - (b) Inputs: cache absent and no caster relationship; supplied direct callback returns 0. Outputs: callback 0; log [direct teleport(G,C,330,7)]; no target-life, relocation or attachment request in this stage.
  - (c) Inputs: T relationship condition true, saved counterpart key (1,99), registry has no (1,99). Outputs: callback 1; log [relocation(G,C,(108,208),R,true,false), attachment(G,C,T,330,7)].
  - (d) Inputs: T mode 12, no relationship. Outputs: callback 0; effect log empty; no position/room/relocation/attachment operation after the failed life test.
  - (e) Inputs: T relationship condition true, key (1,99), registry additionally maps (1,99) to H. Outputs: callback 0; effect log empty; no life/position/room/relocation/attachment operation after the resolved relationship.

- **MTP-V06** Relationship flag and stale target cache are distinct (MTP-R08, MTP-R09; boundary: supplied-input stage)
  - Inputs: default C/G/R; saved counterpart key (1,9); saved target key (1,9); cached stale reference T; registry (1,9) maps different reference H; stored target point (10,0). Both T/H have kind 1, and the key is retained independently of the cached reference. Cases differ only in the relationship condition.
  - Seeds: no draw.
  - Draws: none.
  - Trace: relationship lookup depends on its own condition; refresh resolves the target key to H and clears cached T because object identities differ; coordinate resolution then uses stored (10,0), writes both components and returns false.
  - Outputs: target cache absent, saved target key still (1,9); coordinates exactly (10,0), coordinate result 0. Relationship outputs/query observations are specified below; no update collection.
  - Discriminates: replacing cache with H or clearing its saved key fails; suppressing coordinate writes on a zero component fails; inferring a relationship from the stored key fails case (a).
  - runtime_observed: no.
  - (a) Inputs: relationship condition false. Outputs: relationship result absent; no relationship-registry query; target refresh separately queries (1,9) and clears T; the common coordinate/key outputs above apply.
  - (b) Inputs: relationship condition true. Outputs: relationship result H; one relationship-registry query (1,9); target refresh separately queries (1,9) and clears T; the common coordinate/key outputs above apply.

- **MTP-V07** Upper room edge and alternate-context room handoff (MTP-R04; boundary: supplied-input room-selection stage)
  - Inputs: default C/G/R; stored target point (110,205); R2 rectangle (110,200,10,10); R3 has the same rectangle as R2; supplied relocation returns 0.
  - Seeds: no draw.
  - Draws: none.
  - Trace: R excludes x=110; choose first matching candidate, or use the alternate-context unit's room directly after a miss.
  - Outputs: callback 0, no caller marker/update; exact relocation requests below.
  - Discriminates: closed upper bounds choose R incorrectly; last-match search chooses R3; rechecking containment in this caller incorrectly drops the explicitly supplied alternate room RF.
  - runtime_observed: no.
  - (a) Inputs: R adjacent rooms [absent,R2,R3]; no alternate control context. Outputs: log [relocation(G,C,(110,205),R2,false,false)].
  - (b) Inputs: R adjacent rooms empty; monster AI control present with saved lookup context G2 and key (1,20); G2 registry maps (1,20) to H, whose current room is RF rectangle (500,500,10,10). Outputs: log [relocation(G,C,(110,205),RF,false,false)]; this caller performs no additional RF containment test.

- **MTP-V08** Shared monster relocation forwards adjusted point then marks update (MTP-R10, MTP-R13; boundary: supplied-input stage)
  - Inputs: default C/G/R; supplied destination R and point (104,204); skipNearbySearch false; alternateMarker false; footprint result 2; nearby operation returns R and adjusted point (105,205); underlying path relocation returns 1 without other changes; C path room-synchronization condition false.
  - Seeds: no direct or supplied-operation draw.
  - Draws: none.
  - Trace: nearby request uses mask 7177, scan limit 50 and controls 0/1; forward its changed point to path relocation, then collect, mark and request synchronization; clear synchronization condition makes its helper return immediately.
  - Outputs: result 1; log [nearby(R,(104,204),footprint 2,mask 7177,limit 50,controls 0/1), path relocation(C,R,(105,205)), update collection(C), room/client synchronization(C)]; C queued true and ordinary teleport marker true, alternate marker false; R changed true, update chain [C]. No position or client-message assertion follows from the supplied path result.
  - Discriminates: forwarding the original point fails; omitting update or choosing the alternate marker fails.
  - runtime_observed: no.

- **MTP-V09** Actual reached same-room position commit and movement reset (MTP-R11; boundary: isolated reached suffix)
  - Inputs: monster C and dynamic path at the post-collision suffix; selected/current room R rectangle (0,0,1000,1000); selected point (100,200); room-update-request condition false; containing-room guard false; movement condition true; progression counters 9/8; auxiliary reset values 7/6; initial position/projection fields 0. All preceding collision operations have returned normally; their effects are not simulated.
  - Seeds: no draw.
  - Draws: none.
  - Trace: FX=100*65536+32768=6586368; FY=200*65536+32768=13139968. Unsigned shifts by 11 give 3216 and 6416; A=(-3200)>>1=-1600; B=9632>>2=2408. Containment succeeds; centering a second time leaves the position unchanged; reset the named movement fields.
  - Outputs: suffix result 1; current path integer point (100,200), FX 6586368, FY 13139968, projected A -1600/B 2408; moving false; both progression counters 0; both auxiliary reset values 0; room still R; no room transfer or event request.
  - Discriminates: storing integer coordinates in the 16.16 fields, omitting the half-unit center, swapping projection sum/difference or retaining movement counters changes exact outputs.
  - runtime_observed: no.

- **MTP-V10** Attachment constructs the keyed persistent event (MTP-R12; boundary: supplied-input stage)
  - Inputs: present monster C base family 492 and target H base family 435; monster count 600; identity 330, level 7; state count 185; interaction A/B/C true. C's unit events initially [E1(kind 5,key 330),E2(kind 5,key 17),E3(kind 8,key 330)]; monster persistent queue [P]. Supplied allocation succeeds with zeroed event E; state/link/AI operations return normally with only logged effects; AI control K present.
  - Seeds: no direct or supplied-operation draw.
  - Draws: none.
  - Trace: family checks pass; request state 143 on and reciprocal links; remove only E1; link E first in the unit chain, fill kind/frame/key/parameters and prepend to persistent queue; clear interaction conditions and request AI 16.
  - Outputs: log [state(C,143,on), link(C,H), link(H,C), remove events(C,5,330), allocate/link event(C), AI change(G,C,K,16)]; C unit chain [E,E2,E3], persistent queue [E,P]; E target C/kind 5/frame -1/key 330/first parameter 330/second parameter 7/no override; C interaction A/B/C false. Full post-return link/state/AI integration is not asserted.
  - Discriminates: deleting all type-5 events wrongly removes E2; using level as key leaves E1 and writes 7; appending instead of prepending changes queue order.
  - runtime_observed: no.

- **MTP-V11** Fraction-discarding health percentage selects cleanup effect (MTP-R12, MTP-R14; boundary: supplied-input stage)
  - Inputs: C base family 492, counterpart H base family 441, monster count 600; identity 330; state 143 initially active; interaction A/B/C false; valid AI control K. Supplied stat 7 maximum health raw 23296; supplied stat 6 raw value per case. C events initially [E1(kind 5,key 330),E2(kind 5,key 7)]; state/unlink/death/AI operations return normally; event removal follows exact key matching.
  - Seeds: no draw.
  - Draws: none.
  - Trace: state off then unlink C then unlink H precede percentage calculation; arithmetic-shift raw values by 8 before multiplying/dividing; signed percentage below 10 selects death request.
  - Outputs: independently stated below; no final death/AI execution or cleanup return value asserted.
  - Discriminates: comparing raw fractions to 10% routes case (a) differently; using level 7 as removal key changes case (b)'s chain; restoring interactions before the threshold changes case (a).
  - runtime_observed: no.
  - (a) Inputs: raw stat 6=2559. Trace: h=9,m=91; trunc(900/91)=9. Outputs: percentage 9; log [state(C,143,off), unlink(C), unlink(H), death(G,C,attacker absent,control 1)]; interaction A/B/C remain false; event chain still [E1,E2]; no event-removal or AI request.
  - (b) Inputs: raw stat 6=2560. Trace: h=10,m=91; trunc(1000/91)=10. Outputs: log [state(C,143,off), unlink(C), unlink(H), remove events(C,5,330), AI change(G,C,K,0)]; interaction A/B/C true; event chain [E2]; no death request.

  - (c) Inputs: raw stat 6=128; override raw stat 7=255. Trace: h=0,m=0; return percentage 0 without division. Outputs: percentage 0; log [state(C,143,off), unlink(C), unlink(H), death(G,C,attacker absent,control 1)]; interactions remain false and event chain [E1,E2]; no event removal or AI request.

- **MTP-V12** Update collection never reorders an already-queued unit (MTP-R13; boundary: isolated rule)
  - Inputs: C present with room R; update suppression false; C queued true; R update chain [U,C], U next C and C next absent; R changed marker false.
  - Seeds: no draw.
  - Draws: none.
  - Trace: room present and not suppressed; C queued condition causes immediate return before writes.
  - Outputs: R chain remains [U,C], changed marker false; C remains queued and next absent; no writes or other requests.
  - Discriminates: removing and re-prepending C produces [C,U] and/or sets the changed marker, violating the outputs.
  - runtime_observed: no.


- **MTP-V13** Link and unlink immediately store their counterpart fields (MTP-R12; boundary: isolated helper prefix before stat integration)
  - Inputs: present unit C with stored counterpart kind/id (4,20), counterpart H kind 1/id 9; any optional stat storage is supplied but its processing is outside this fixture. Each case starts independently with its stated saved fields.
  - Seeds: no draw before the optional integration boundary.
  - Draws: none.
  - Trace: link copies H kind/id before testing optional stat storage; unlink stores two zeros before that test. Stop after those stores.
  - Outputs: per case below; relationship condition, state/list/stat propagation and final helper return remain unasserted.
  - Discriminates: waiting for a later allocation/stat result before storing the key fails the immediate prefix boundary; clearing only the relationship flag fails unlink's two key stores.
  - runtime_observed: no.
  - (a) Inputs: link C to H; initial stored fields (4,20). Outputs: stored counterpart exactly (1,9) before any optional stat operation.
  - (b) Inputs: unlink C; initial stored fields (1,9). Outputs: stored counterpart exactly (0,0) before any optional stat operation.

### 8.2 Observation vectors

**MTP-O01** On an imp teleport action, measure the selected handler, relationship route, relocation result, actual committed position/update collection, attachment event key and final callback result. Match only the captured input/stage assertions; full integration gates remain separate. The reproduction procedure and actual setup/data records are in the companion. No matching gameplay record is present.

## 9. Acceptance and dependencies

Partial acceptance comprises MTP-R01/MTP-R02's named data, MTP-V01 to MTP-V13 for the stated validation, local caller, reached position/queue and attachment-request stages. Supplied dependency results, successful event allocation and the reached same-room suffix are explicit fixture boundaries. Full teleport reproduction additionally requires the missing collision/general movement/client/link/AI/death/event integration in MTP-Q02/MTP-Q03 and matching gameplay observations MTP-Q04. Passing a local success-return fixture does not establish successful physical relocation; the position suffix checks actual stores only once reached. No zero result implies rollback.

**Depends on:** SCD-R06 for action selection/arguments; GDA-R11a and GDA-R08a for supplied Skills/MonStats records; [record layouts section 3.3](core-record-layouts-and-ids.md#33-unit-types) for kinds; NRF-R10b for dynamic integer coordinates; RLI-R02a/RLI-R02b for returned health-stat values; RLI-R03 and STL-R07 for state membership requests; STL-R01's successful allocation/unit-chain linkage convention for the supplied event constructor. Direct persistent-queue insertion belongs to MTP-R12 and does not consume an asserted processing cadence. Full collision/path/client/link/AI/death dependencies remain MTP-Q02/MTP-Q03 rather than unnamed completed contracts.

**Depended on by:** no known implementation or plan consumer. Other monster skills may refer to teleport as a supporting relationship without establishing a new prerequisite.

## 10. Related specs

[Skill dispatch](skill-cast-start-and-do-dispatch.md) owns callback selection and final dispatcher effects. [Record operations](core-record-layouts-and-ids.md) owns shared unit/state machinery. [Monster Baal](skill-monster-baal.md) describes separate Baal action handlers; its teleport association is a data relationship, not an additional teleport algorithm.

## Appendix A. Provenance

Original Windows x86 LoD Expansion 1.13c D2Game instructions and dispatch-table bytes were read on 2026-10-29 against `LOD113C-D2GAME` in the [source manifest](../research/source-manifest.md). T01/T02 record complete effective Skills/States text reads on owned vanilla-profile games; C01/C02 corroborate the selected compiled scalar fields. Original loader descriptors establish the consumed `aurastate` field association. Physical archive origin and loaded-module identities remain uncaptured. The continuation verifies local request, reached position-commit, attachment-event and update-queue stages with integration boundaries retained. No gameplay assertion is runtime-confirmed. The [companion](../research/spec-evidence/skill-monster-teleport.md) records anchors, limitations and observation procedure.
